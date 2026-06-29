package snssqs

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

func (l *subscriptionQueueLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The subscription is not modified by this link; it owns the topic -> queue wiring
	// and all of its configuration.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *subscriptionQueueLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The queue is not modified by this link; the send-message permission is a separate,
	// link-owned aws/sqs/queueInlinePolicy intermediary handled in
	// UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) SNS permission to send messages to the
// queue by deploying a single link-owned aws/sqs/queueInlinePolicy resource. The
// statement is scoped to the subscription's topic via aws:SourceArn.
func (l *subscriptionQueueLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")

	identity := subscriptionQueueIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)
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

	topicARN, hasTopicARN := subscriptionTopicARN(input.ResourceAInfo)
	if !hasTopicARN {
		return nil, fmt.Errorf(
			"topic ARN could not be retrieved from the %q subscription resource",
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}

	sid := snsStatementID(input.ResourceAInfo)
	policyDocument := buildQueuePolicyDocument(sid, queueARN, topicARN)

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
				"sourceArn": core.MappingNodeFromString(topicARN),
				"queueArn":  core.MappingNodeFromString(queueARN),
			},
		}),
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{intermediaryState},
	}, nil
}

func subscriptionTopicARN(subscriptionInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(subscriptionInfo)
	topicARNNode, has := pluginutils.GetValueByPath("$.topicArn", spec)
	if !has {
		return "", false
	}
	return core.StringValue(topicARNNode), true
}

// Builds the resource-based queue policy granting
// sns.amazonaws.com sqs:SendMessage on the queue, scoped by aws:SourceArn to the topic.
// The QueueInlinePolicy "policyDocument" property is re-typed by the policy-document
// overlay as a structured object, so the document is built with camelCase fields (the
// engine translates them to the PascalCase shape Cloud Control expects). The condition
// block is a free-form map, so its operator and key names are kept verbatim.
func buildQueuePolicyDocument(sid, queueARN, topicARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"version", core.MappingNodeFromString("2012-10-17"),
		"statement", core.MappingNodeItems(
			core.MappingNodeFields(
				"sid", core.MappingNodeFromString(sid),
				"effect", core.MappingNodeFromString("Allow"),
				"principal", core.MappingNodeFields(
					"service", core.MappingNodeFromString("sns.amazonaws.com"),
				),
				"action", core.MappingNodeFromString("sqs:SendMessage"),
				"resource", core.MappingNodeFromString(queueARN),
				"condition", core.MappingNodeFields(
					"ArnEquals", core.MappingNodeFields(
						"aws:SourceArn", core.MappingNodeFromString(topicARN),
					),
				),
			),
		),
	)
}

func snsStatementID(subscriptionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"SNS%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(subscriptionInfo)),
	)
}

func subscriptionQueueIntermediaryIdentity(subscriptionInfo, queueInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/sqs/queueInlinePolicy",
		ResourceID:   subscriptionQueuePolicyResourceID(subscriptionInfo, queueInfo),
		ResourceName: subscriptionQueuePolicyResourceName(subscriptionInfo, queueInfo),
	}
}

func subscriptionQueuePolicyResourceID(subscriptionInfo, queueInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%s__%s__sns-send-policy",
		pluginutils.GetResourceName(subscriptionInfo),
		pluginutils.GetResourceName(queueInfo),
	)
}

func subscriptionQueuePolicyResourceName(subscriptionInfo, queueInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sSend%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(subscriptionInfo)),
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(queueInfo)),
	)
}
