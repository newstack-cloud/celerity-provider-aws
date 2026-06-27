package eventssqs

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *ruleQueueLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The rule is not modified by this link; the rule -> queue wiring is modelled
	// inline in the rule's targets[] and owned by the aws/events/rule resource.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *ruleQueueLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The queue is not modified by this link; the send-message permission is a
	// separate, link-owned aws/sqs/queueInlinePolicy intermediary handled in
	// UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) EventBridge permission to send
// messages to the queue by deploying a single link-owned aws/sqs/queueInlinePolicy
// resource. The statement is scoped to the rule via aws:SourceArn = the rule's ARN.
func (l *ruleQueueLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")

	identity := ruleQueueIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)
	priorState := linkutils.FindIntermediaryState(input.CurrentLinkState, identity.ResourceID)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if err := linkutils.DestroyManagedIntermediary(
			ctx, input.ResourceService, pluginutils.GetInstanceID(input.ResourceAInfo), providerCtx, priorState,
		); err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	queueARN, hasQueueARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasQueueARN {
		return nil, fmt.Errorf(
			"queue ARN could not be retrieved from the linked to %q queue resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	queueSpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(input.ResourceBInfo)
	queueURLNode, hasQueueURL := pluginutils.GetValueByPath("$.queueUrl", queueSpec)
	if !hasQueueURL {
		return nil, fmt.Errorf(
			"queue URL could not be retrieved from the linked to %q queue resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}
	queueURL := core.StringValue(queueURLNode)

	ruleARN, hasRuleARN := utils.ExtractARNFromResourceInfo(input.ResourceAInfo)
	if !hasRuleARN {
		return nil, fmt.Errorf(
			"rule ARN could not be retrieved from the %q rule resource",
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}

	sid := eventBridgeStatementID(input.ResourceAInfo)
	policyDocument := buildQueuePolicyDocument(sid, queueARN, ruleARN)

	intermediaryState, err := linkutils.DeployManagedIntermediary(
		ctx,
		input.ResourceService,
		pluginutils.GetInstanceID(input.ResourceAInfo),
		input.InstanceName,
		providerCtx,
		priorState,
		linkutils.ManagedIntermediary{
			ResourceType: identity.ResourceType,
			ResourceID:   identity.ResourceID,
			ResourceName: identity.ResourceName,
			Spec: core.MappingNodeFields(
				"queue", core.MappingNodeFromString(queueURL),
				"policyDocument", policyDocument,
			),
		},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData: linkutils.IntermediaryLinkData(linkutils.DeployedIntermediary{
			Identity: identity,
			Leaves: map[string]*core.MappingNode{
				"sourceArn": core.MappingNodeFromString(ruleARN),
				"queueArn":  core.MappingNodeFromString(queueARN),
			},
		}),
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{intermediaryState},
	}, nil
}

// Builds the resource-based queue policy granting
// events.amazonaws.com sqs:SendMessage on the queue, scoped by aws:SourceArn to the
// rule. The QueueInlinePolicy "policyDocument" property is re-typed by the
// policy-document overlay as a structured object, so the document is built with
// camelCase fields (the engine translates them to the PascalCase shape Cloud
// Control expects). The condition block is a free-form map, so its operator and key
// names ("ArnEquals", "aws:SourceArn") are kept verbatim.
func buildQueuePolicyDocument(sid, queueARN, ruleARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"version", core.MappingNodeFromString("2012-10-17"),
		"statement", core.MappingNodeItems(
			core.MappingNodeFields(
				"sid", core.MappingNodeFromString(sid),
				"effect", core.MappingNodeFromString("Allow"),
				"principal", core.MappingNodeFields(
					"service", core.MappingNodeFromString("events.amazonaws.com"),
				),
				"action", core.MappingNodeFromString("sqs:SendMessage"),
				"resource", core.MappingNodeFromString(queueARN),
				"condition", core.MappingNodeFields(
					"ArnEquals", core.MappingNodeFields(
						"aws:SourceArn", core.MappingNodeFromString(ruleARN),
					),
				),
			),
		),
	)
}

func eventBridgeStatementID(ruleResourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"EventBridge%s",
		pluginutils.StripNonAlphaNumericChars(
			pluginutils.GetResourceName(ruleResourceInfo),
		),
	)
}

func ruleQueueIntermediaryIdentity(ruleInfo, queueInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/sqs/queueInlinePolicy",
		ResourceID:   ruleQueuePolicyResourceID(ruleInfo, queueInfo),
		ResourceName: ruleQueuePolicyResourceName(ruleInfo, queueInfo),
	}
}

func ruleQueuePolicyResourceID(ruleInfo, queueInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%s__%s__eventbridge-send-policy",
		pluginutils.GetResourceName(ruleInfo),
		pluginutils.GetResourceName(queueInfo),
	)
}

func ruleQueuePolicyResourceName(ruleInfo, queueInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sSend%s",
		pluginutils.StripNonAlphaNumericChars(
			pluginutils.GetResourceName(ruleInfo),
		),
		pluginutils.StripNonAlphaNumericChars(
			pluginutils.GetResourceName(queueInfo),
		),
	)
}
