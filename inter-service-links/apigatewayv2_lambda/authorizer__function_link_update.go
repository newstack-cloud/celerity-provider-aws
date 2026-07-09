package apigatewayv2lambda

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

func (l *authorizerFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The authorizer is not modified by this link; the invoke permission is a link-owned
	// aws/lambda/permission intermediary handled in UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{LinkData: core.MappingNodeFields()}, nil
}

func (l *authorizerFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The function is not modified by this link.
	return &provider.LinkUpdateResourceOutput{LinkData: core.MappingNodeFields()}, nil
}

// UpdateIntermediaryResources grants API Gateway permission to invoke the authorizer's backing
// function, scoped to the authorizer's execute-api ARN.
func (l *authorizerFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	instanceID := pluginutils.GetInstanceID(input.ResourceAInfo)

	identity := authorizerPermissionIdentity(input.ResourceAInfo, input.ResourceBInfo)
	prior := linkutils.FindIntermediaryState(input.CurrentLinkState, identity.ResourceID)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if err := linkutils.DestroyManagedIntermediary(
			ctx, input.ResourceService, instanceID, providerCtx, prior,
		); err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}, nil
	}

	apiID, hasAPIID := authorizerAPIID(input.ResourceAInfo)
	authorizerID, hasAuthorizerID := authorizerID(input.ResourceAInfo)
	if !hasAPIID || !hasAuthorizerID {
		return nil, fmt.Errorf(
			"API id and authorizer id could not be retrieved from the API Gateway v2 authorizer %q",
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}
	functionARN, hasFunctionARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the Lambda function %q",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	region, err := l.getRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}
	account, err := accountFromARN(functionARN)
	if err != nil {
		return nil, err
	}
	sourceARN := authorizerSourceARN(region, account, apiID, authorizerID)

	permissionState, err := linkutils.DeployManagedIntermediary(
		ctx, input.ResourceService, instanceID, input.InstanceName, providerCtx, prior,
		linkutils.ManagedIntermediary{
			ResourceType: identity.ResourceType,
			ResourceID:   identity.ResourceID,
			ResourceName: identity.ResourceName,
			Spec: core.MappingNodeFields(
				"functionName", core.MappingNodeFromString(functionARN),
				"action", core.MappingNodeFromString("lambda:InvokeFunction"),
				"principal", core.MappingNodeFromString("apigateway.amazonaws.com"),
				"sourceArn", core.MappingNodeFromString(sourceARN),
			),
		},
	)
	if err != nil {
		return nil, err
	}

	linkData := linkutils.IntermediaryLinkData(linkutils.DeployedIntermediary{
		Identity: identity,
		Leaves:   map[string]*core.MappingNode{"functionArn": core.MappingNodeFromString(functionARN)},
	})

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:                   linkData,
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{permissionState},
	}, nil
}
