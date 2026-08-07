package flexlambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// A VPC-attached Lambda function runs behind an elastic network interface that Lambda
// creates, and deletes, using the function's own execution role. Lambda checks for those
// permissions at the moment the attachment is set and rejects the update outright if they
// are missing, so this grant has to be applied before the vpcConfig, not after it.
//
// Without this, placing a function in a VPC only works if the user has separately
// attached AWSLambdaVPCAccessExecutionRole to the role, which is exactly the kind of
// hidden prerequisite the link is supposed to remove.
//
// This mirrors the AWS-managed AWSLambdaVPCAccessExecutionRole policy: the same actions,
// on all resources. Narrowing it would take three statements rather than one, because the
// describe calls do not accept resource-level permissions while the interface calls do,
// and the role's statement budget is shared with every access link granting permissions
// on the same role.
var eniManagementActions = []string{
	"ec2:CreateNetworkInterface",
	"ec2:DescribeNetworkInterfaces",
	"ec2:DescribeSubnets",
	"ec2:DeleteNetworkInterface",
	"ec2:AssignPrivateIpAddresses",
	"ec2:UnassignPrivateIpAddresses",
}

func eniManagementStatement(sid string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   eniManagementActions,
		"Resource": "*",
	}
}

// The SID is derived from the function rather than the VPC, because the permission is a
// property of the function being VPC-attached at all. A function can only be placed in
// one VPC, so there is never more than one of these per role per function.
func createENIManagementSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"VPCNetworkInterfaces%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

// Where the grant landed, so the caller can attribute the statement to this link and
// suppress drift on the role.
type eniGrantResult struct {
	roleResourceName string
	sid              string
	placement        linkutils.RoleAccessResult
	// granted is false when the function has no execution role to grant on, in which
	// case there is nothing to attribute and nothing to revoke.
	granted bool
}

func (l *vpcFunctionLinkActions) grantENIPermissions(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	providerCtx provider.Context,
) (eniGrantResult, error) {
	return l.reconcileENIGrant(ctx, input, providerCtx, eniManagementStatement)
}

func (l *vpcFunctionLinkActions) revokeENIPermissions(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	providerCtx provider.Context,
) (eniGrantResult, error) {
	// A nil statement removes the grant.
	return l.reconcileENIGrant(ctx, input, providerCtx, nil)
}

// Applies the grant, or removes it when buildStatement is nil, holding the per-role lock
// across the read-modify-write so it does not race the access links reconciling their own
// statements on the same role.
func (l *vpcFunctionLinkActions) reconcileENIGrant(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	providerCtx provider.Context,
	buildStatement func(sid string) map[string]any,
) (eniGrantResult, error) {
	lambdaService, err := l.getLambdaService(ctx, providerCtx)
	if err != nil {
		return eniGrantResult{}, err
	}

	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceInfo,
			LoadRoleInfo:           true,
		},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	// Unlike the access links, placement is not here to manage permissions: it is here to
	// put the function in a VPC. A function whose execution role the user manages
	// themselves is still placed, with the network interface permissions left as their
	// responsibility, rather than being refused outright over a role the link has no
	// standing to modify.
	if errors.Is(err, linkutils.ErrExecutionRoleNotInBlueprint) {
		return eniGrantResult{}, nil
	}
	if err != nil {
		return eniGrantResult{}, err
	}

	err = input.ResourceService.AcquireResourceLock(ctx, &provider.AcquireResourceLockInput{
		InstanceID:      pluginutils.GetInstanceID(input.ResourceInfo),
		ResourceName:    setupCtx.RoleResourceName,
		ProviderContext: providerCtx,
		AcquiredBy:      input.LinkID,
	})
	if err != nil {
		return eniGrantResult{}, err
	}

	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return eniGrantResult{}, err
	}

	sid := createENIManagementSID(input.ResourceInfo)
	grant := linkutils.RoleAccessGrant{
		RoleName: setupCtx.RoleName,
		SID:      sid,
	}
	if buildStatement != nil {
		grant.Statement = buildStatement(sid)
	}

	placement, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, grant)
	if err != nil {
		return eniGrantResult{}, fmt.Errorf(
			"cannot grant function %q the network interface permissions a VPC-attached "+
				"function requires: %w",
			pluginutils.GetResourceName(input.ResourceInfo),
			err,
		)
	}

	return eniGrantResult{
		roleResourceName: setupCtx.RoleResourceName,
		sid:              sid,
		placement:        placement,
		granted:          true,
	}, nil
}

// Attributes the ENI statement to this link in the link's output, so the role does not
// report drift for a statement the link put there.
func addENIGrantToOutput(
	output *provider.LinkUpdateResourceOutput,
	input *provider.LinkUpdateResourceInput,
	grant eniGrantResult,
) *provider.LinkUpdateResourceOutput {
	if !grant.granted {
		return output
	}

	linkDataKey := fmt.Sprintf("%sExecutionRole", pluginutils.GetResourceName(input.ResourceInfo))
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specENIStatementNode(grant.sid),
	)

	if output.ResourceDataMappings == nil {
		output.ResourceDataMappings = map[string]string{}
	}
	linkutils.AppendRoleAccessMapping(
		output.ResourceDataMappings,
		roleLinkData,
		grant.roleResourceName,
		linkDataKey,
		grant.sid,
		grant.placement,
	)

	if output.LinkData == nil {
		output.LinkData = core.MappingNodeFields()
	}
	if output.LinkData.Fields == nil {
		output.LinkData.Fields = map[string]*core.MappingNode{}
	}
	output.LinkData.Fields[linkDataKey] = roleLinkData

	return output
}

// The same statement as eniManagementStatement, in the mapping node form the link data
// and drift suppression use.
func specENIStatementNode(sid string) *core.MappingNode {
	return core.MappingNodeFields(
		"Sid", core.MappingNodeFromString(sid),
		"Effect", core.MappingNodeFromString("Allow"),
		"Action", stringItems(eniManagementActions),
		"Resource", core.MappingNodeFromString("*"),
	)
}
