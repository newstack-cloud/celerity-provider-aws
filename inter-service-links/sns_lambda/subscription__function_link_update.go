package snslambda

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

func (l *subscriptionFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The subscription is not modified by this link; it owns the topic -> function wiring
	// and all of its configuration.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
	}, nil
}

func (l *subscriptionFunctionLinkActions) UpdateResourceB(
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

// UpdateIntermediaryResources grants (or revokes) SNS permission to invoke the function
// by deploying a single link-owned aws/lambda/permission resource. The permission is
// scoped to the subscription's topic via SourceArn.
func (l *subscriptionFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")

	identity := subscriptionFunctionIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)
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

	topicARN, hasTopicARN := subscriptionTopicARN(input.ResourceAInfo)
	if !hasTopicARN {
		return nil, fmt.Errorf(
			"topic ARN could not be retrieved from the %q subscription resource",
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
				"principal", core.MappingNodeFromString("sns.amazonaws.com"),
				"sourceArn", core.MappingNodeFromString(topicARN),
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
				"sourceArn":    core.MappingNodeFromString(topicARN),
				"functionName": core.MappingNodeFromString(functionARN),
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

func subscriptionFunctionIntermediaryIdentity(subscriptionInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/lambda/permission",
		ResourceID:   subscriptionPermissionResourceID(subscriptionInfo, functionInfo),
		ResourceName: subscriptionPermissionResourceName(subscriptionInfo, functionInfo),
	}
}

func subscriptionPermissionResourceID(subscriptionInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%s__%s__sns-invoke-permission",
		pluginutils.GetResourceName(subscriptionInfo),
		pluginutils.GetResourceName(functionInfo),
	)
}

func subscriptionPermissionResourceName(subscriptionInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sInvoke%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(subscriptionInfo)),
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(functionInfo)),
	)
}
