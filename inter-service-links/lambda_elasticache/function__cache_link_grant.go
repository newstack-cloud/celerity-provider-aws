package lambdaelasticache

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionCacheLinkActions) reconcileConnectGrant(
	ctx context.Context,
	providerCtx provider.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	annotations *cacheLinkAnnotations,
	region string,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	if err := input.ResourceService.AcquireResourceLock(ctx, &provider.AcquireResourceLockInput{
		InstanceID:      pluginutils.GetInstanceID(input.ResourceAInfo),
		ResourceName:    setupCtx.RoleResourceName,
		ProviderContext: providerCtx,
		AcquiredBy:      input.LinkID,
	}); err != nil {
		return nil, err
	}

	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	sid := createConnectSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if _, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setupCtx.RoleName,
			SID:      sid,
		}); err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}, nil
	}

	connectResources, err := connectResourceARNs(input.ResourceBInfo, setupCtx.FunctionARN, region, annotations.userId)
	if err != nil {
		return nil, err
	}

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: connectStatement(sid, connectResources),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q elasticache:Connect on cache %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	return connectLinkOutput(input, setupCtx.RoleResourceName, sid, connectResources, result), nil
}

func connectStatement(sid string, connectResources []string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   []string{"elasticache:Connect"},
		"Resource": connectResources,
	}
}

func connectLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid string,
	connectResources []string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	resourceItems := make([]*core.MappingNode, 0, len(connectResources))
	for _, resource := range connectResources {
		resourceItems = append(resourceItems, core.MappingNodeFromString(resource))
	}
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		core.MappingNodeFields(
			"sid", core.MappingNodeFromString(sid),
			"effect", core.MappingNodeFromString("Allow"),
			"action", &core.MappingNode{Items: []*core.MappingNode{
				core.MappingNodeFromString("elasticache:Connect"),
			}},
			"resource", &core.MappingNode{Items: resourceItems},
		),
	)

	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

// Builds the elasticache:Connect resource ARNs. The action must be scoped to both the replication
// group and the user in a single statement:
//
//	arn:aws:elasticache:<region>:<account>:replicationgroup:<replicationGroupId>
//	arn:aws:elasticache:<region>:<account>:user:<userId>
//
// The replication group exposes no ARN attribute, so both are constructed from the region, the
// account (parsed from the function ARN) and the replication group / user ids.
func connectResourceARNs(
	cacheInfo *provider.ResourceInfo,
	functionARN, region, userId string,
) ([]string, error) {
	replicationGroupId, hasID := extractReplicationGroupID(cacheInfo)
	if !hasID {
		return nil, fmt.Errorf(
			"replication group id could not be retrieved from the linked to %q ElastiCache replication group resource",
			pluginutils.GetResourceName(cacheInfo),
		)
	}
	account, err := accountFromARN(functionARN)
	if err != nil {
		return nil, err
	}
	return []string{
		fmt.Sprintf("arn:aws:elasticache:%s:%s:replicationgroup:%s", region, account, replicationGroupId),
		fmt.Sprintf("arn:aws:elasticache:%s:%s:user:%s", region, account, userId),
	}, nil
}

func extractReplicationGroupID(cacheInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(cacheInfo)
	node, has := pluginutils.GetValueByPath("$.replicationGroupId", spec)
	if !has || node == nil || core.StringValue(node) == "" {
		return "", false
	}
	return core.StringValue(node), true
}

// Extracts the account id (the 5th colon-separated segment) from an ARN such as
// arn:aws:lambda:<region>:<account>:function:<name>.
func accountFromARN(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 5 || parts[4] == "" {
		return "", fmt.Errorf("account id could not be determined from ARN: %q", arn)
	}
	return parts[4], nil
}

func createConnectSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("ElastiCacheConnect%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}
