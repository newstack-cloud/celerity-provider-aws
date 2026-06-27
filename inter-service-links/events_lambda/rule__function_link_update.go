package eventslambda

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

func (l *ruleFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The rule is not modified by this link; the rule -> function wiring is modelled
	// inline in the rule's targets array and owned by the aws/events/rule resource.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *ruleFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The function is not modified by this link; the invoke permission is a separate,
	// link-owned aws/lambda/permission intermediary handled in
	// UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) EventBridge permission to invoke
// the function by deploying a single link-owned aws/lambda/permission resource. The
// permission is scoped to the rule via SourceArn = the rule's ARN.
func (l *ruleFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")

	identity := ruleFunctionIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)
	priorState := linkutils.FindIntermediaryState(input.CurrentLinkState, identity.ResourceID)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		err := linkutils.DestroyManagedIntermediary(
			ctx,
			input.ResourceService,
			pluginutils.GetInstanceID(input.ResourceAInfo),
			providerCtx,
			priorState,
		)
		if err != nil {
			return nil, err
		}

		return &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	functionARN, hasFunctionARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked to %q function resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	ruleARN, hasRuleARN := utils.ExtractARNFromResourceInfo(input.ResourceAInfo)
	if !hasRuleARN {
		return nil, fmt.Errorf(
			"rule ARN could not be retrieved from the %q rule resource",
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}

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
				"functionName", core.MappingNodeFromString(functionARN),
				"action", core.MappingNodeFromString("lambda:InvokeFunction"),
				"principal", core.MappingNodeFromString("events.amazonaws.com"),
				"sourceArn", core.MappingNodeFromString(ruleARN),
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
				"sourceArn":    core.MappingNodeFromString(ruleARN),
				"functionName": core.MappingNodeFromString(functionARN),
			},
		}),
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{intermediaryState},
	}, nil
}

// The single source of truth for the link-owned
// Lambda permission's identity, used by both UpdateIntermediaryResources (to deploy and
// persist it) and StageChanges (to project it into link changes).
func ruleFunctionIntermediaryIdentity(ruleInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/lambda/permission",
		ResourceID:   rulePermissionResourceID(ruleInfo, functionInfo),
		ResourceName: rulePermissionResourceName(ruleInfo, functionInfo),
	}
}

func rulePermissionResourceID(ruleInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%s__%s__eventbridge-invoke-permission",
		pluginutils.GetResourceName(ruleInfo),
		pluginutils.GetResourceName(functionInfo),
	)
}

func rulePermissionResourceName(ruleInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sInvoke%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(ruleInfo)),
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(functionInfo)),
	)
}
