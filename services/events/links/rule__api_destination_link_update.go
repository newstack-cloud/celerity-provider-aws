package eventslinks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *ruleAPIDestinationLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The rule itself is not modified by this link; the rule -> destination wiring is
	// modelled inline in the rule's targets[] and owned by the aws/events/rule
	// resource.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
	}, nil
}

func (l *ruleAPIDestinationLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The API destination is not modified by this link; only the IAM role referenced
	// by the rule's matching target entry is updated.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
	}, nil
}

type apiDestinationRoleSetup struct {
	roleName          string
	roleResourceState *state.ResourceState
}

// UpdateIntermediaryResources grants (or revokes) the EventBridge rule's target
// role permission to invoke the linked API destination by packing a single IAM
// statement into the role's allocator-managed policies. The role is an existing
// intermediary in the blueprint (referenced by the rule's target entry roleArn),
// so the role lock is held while its shared policy set is read-modify-written.
func (l *ruleAPIDestinationLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")

	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	destinationARN, hasARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasARN {
		return nil, fmt.Errorf(
			"API destination ARN could not be retrieved from the linked to %q API destination resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	setup, err := l.setupRole(ctx, input, providerCtx, destinationARN)
	if err != nil {
		return nil, err
	}

	sid := createInvokeAPIDestinationSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		_, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setup.roleName,
			SID:      sid,
		})
		if err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setup.roleName,
		SID:       sid,
		Statement: invokeAPIDestinationStatement(sid, destinationARN),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant EventBridge rule %q permission to invoke API destination %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	return invokeLinkOutput(input, setup.roleResourceState.Name, sid, destinationARN, result), nil
}

func (l *ruleAPIDestinationLinkActions) setupRole(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	providerCtx provider.Context,
	destinationARN string,
) (*apiDestinationRoleSetup, error) {
	roleARN, hasRoleARN := targetEntryRoleARN(input.ResourceAInfo, destinationARN)
	if !hasRoleARN {
		return nil, fmt.Errorf(
			"roleArn could not be retrieved from the matching target entry for API "+
				"destination %q in the %q rule resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}

	// The rule's target stores the role as an ARN, but an aws/iam/role's external ID
	// in state is its role name (the Cloud Control primary identifier), so the lookup
	// must match on the name derived from the ARN.
	roleName := roleNameFromARN(roleARN)

	roleResourceState, err := input.ResourceService.LookupResourceInState(
		ctx,
		&provider.ResourceLookupInput{
			InstanceID:      pluginutils.GetInstanceID(input.ResourceAInfo),
			ResourceType:    "aws/iam/role",
			ExternalID:      roleName,
			ProviderContext: providerCtx,
		},
	)
	if err != nil {
		return nil, err
	}
	if roleResourceState == nil {
		return nil, fmt.Errorf(
			"intermediary resource of type 'aws/iam/role' is not present in the same " +
				"blueprint as this link (aws/events/rule::aws/events/apiDestination). " +
				"Links can only update intermediary resources that are defined in the " +
				"same blueprint",
		)
	}

	// The target's role is shared by every link that grants it access, so lock it
	// for the read-modify-write of its policy set.
	err = input.ResourceService.AcquireResourceLock(
		ctx,
		&provider.AcquireResourceLockInput{
			InstanceID:      pluginutils.GetInstanceID(input.ResourceAInfo),
			ResourceName:    roleResourceState.Name,
			ProviderContext: providerCtx,
		},
	)
	if err != nil {
		return nil, err
	}

	return &apiDestinationRoleSetup{
		roleName:          roleName,
		roleResourceState: roleResourceState,
	}, nil
}

func targetEntryRoleARN(ruleResourceInfo *provider.ResourceInfo, destinationARN string) (string, bool) {
	ruleSpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(ruleResourceInfo)
	targetsNode, hasTargets := pluginutils.GetValueByPath("$.targets", ruleSpec)
	if !hasTargets || targetsNode == nil {
		return "", false
	}

	for _, target := range targetsNode.Items {
		if target == nil {
			continue
		}
		arn := core.StringValue(target.Fields["arn"])
		if arn != destinationARN {
			continue
		}
		roleARN := core.StringValue(target.Fields["roleArn"])
		if roleARN == "" {
			return "", false
		}
		return roleARN, true
	}

	return "", false
}

// Builds the IAM policy statement (canonical
// PascalCase, as the IAM API expects) granting permission to invoke the API
// destination.
func invokeAPIDestinationStatement(sid, destinationARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   []any{"events:InvokeApiDestination"},
		"Resource": destinationARN,
	}
}

// Records the granted statement in link data and, for inline
// placements, maps it onto the role's spec so the framework attributes the
// statement to this link and does not treat it as drift / strip it on redeploy.
func invokeLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, destinationARN string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specInvokeStatementNode(sid, destinationARN),
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it:
	// inline placements map the statement by Sid; managed (overflow) placements map
	// the attached managed policy ARN.
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

// Builds the statement in the camelCase spec form the
// role's external state uses (after Cloud Control name translation), so the drift
// comparison against link data matches.
func specInvokeStatementNode(sid, destinationARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{
			Items: []*core.MappingNode{
				core.MappingNodeFromString("events:InvokeApiDestination"),
			},
		},
		"resource", core.MappingNodeFromString(destinationARN),
	)
}

func createInvokeAPIDestinationSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"InvokeApiDestination%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(resourceInfo)),
	)
}

func createLinkDataRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sRole", pluginutils.GetResourceName(resourceInfo))
}

func roleNameFromARN(roleARN string) string {
	idx := strings.LastIndex(roleARN, "/")
	if idx == -1 {
		return roleARN
	}
	return roleARN[idx+1:]
}
