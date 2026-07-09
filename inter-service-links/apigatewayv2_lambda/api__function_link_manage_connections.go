package apigatewayv2lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Grants (or revokes) execute-api:ManageConnections on the WebSocket
// route handler's execution role so it can push messages to clients via the @connections API. It is
// a no-op for HTTP APIs and for WebSocket handlers that opt out, unless a prior grant is recorded in
// the link state (in which case the prior grant is revoked). Returns the grant's linkData and
// resource-data mappings to merge into the link output.
func (l *apiFunctionLinkActions) reconcileManageConnectionsGrant(
	ctx context.Context,
	providerCtx provider.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	apiID, region, account string,
	annotations *apiLinkAnnotations,
) (*core.MappingNode, map[string]string, error) {
	isDestroy := input.LinkUpdateType == provider.LinkUpdateTypeDestroy
	applicable := annotations.manageConnections && !isHTTPProtocol(input.ResourceAInfo)

	// Nothing to grant and nothing recorded to revoke: skip the role lookup entirely (the common
	// case for HTTP APIs and opted-out handlers).
	if !applicable && !hasRecordedManageConnectionsGrant(input) {
		return core.MappingNodeFields(), map[string]string{}, nil
	}

	lambdaService, err := l.getLambdaService(ctx, providerCtx)
	if err != nil {
		return nil, nil, err
	}
	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{LambdaFuncResourceInfo: input.ResourceBInfo, LoadRoleInfo: true},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	if err != nil {
		return nil, nil, err
	}

	if err := input.ResourceService.AcquireResourceLock(ctx, &provider.AcquireResourceLockInput{
		InstanceID:      pluginutils.GetInstanceID(input.ResourceBInfo),
		ResourceName:    setupCtx.RoleResourceName,
		ProviderContext: providerCtx,
		AcquiredBy:      input.LinkID,
	}); err != nil {
		return nil, nil, err
	}

	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, nil, err
	}

	sid := manageConnectionsSID(input.ResourceBInfo)

	if isDestroy || !applicable {
		if _, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setupCtx.RoleName,
			SID:      sid,
		}); err != nil {
			return nil, nil, err
		}
		return core.MappingNodeFields(), map[string]string{}, nil
	}

	sourceARN := manageConnectionsSourceARN(region, account, apiID, annotations.manageConnectionsStage)
	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: manageConnectionsStatement(sid, sourceARN),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, nil, fmt.Errorf(
				"cannot grant Lambda %q execute-api:ManageConnections on API %q: %w",
				pluginutils.GetResourceName(input.ResourceBInfo),
				pluginutils.GetResourceName(input.ResourceAInfo),
				err,
			)
		}
		return nil, nil, err
	}

	linkDataKey := executionRoleLinkDataName(input.ResourceBInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		core.MappingNodeFields(
			"sid", core.MappingNodeFromString(sid),
			"effect", core.MappingNodeFromString("Allow"),
			"action", core.MappingNodeFromString("execute-api:ManageConnections"),
			"resource", core.MappingNodeFromString(sourceARN),
		),
	)
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, setupCtx.RoleResourceName, linkDataKey, sid, result)

	return core.MappingNodeFields(linkDataKey, roleLinkData), mappings, nil
}

func manageConnectionsStatement(sid, sourceARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   "execute-api:ManageConnections",
		"Resource": sourceARN,
	}
}

func hasRecordedManageConnectionsGrant(input *provider.LinkUpdateIntermediaryResourcesInput) bool {
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)
	if currentLinkData == nil || len(currentLinkData.Fields) == 0 {
		return false
	}
	key := executionRoleLinkDataName(input.ResourceBInfo)
	node, has := pluginutils.GetValueByPath(
		fmt.Sprintf("$[%q][%q]", key, linkutils.PermissionFieldName),
		currentLinkData,
	)
	return has && node != nil
}
