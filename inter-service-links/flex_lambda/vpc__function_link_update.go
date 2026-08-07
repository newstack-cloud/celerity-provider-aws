package flexlambda

import (
	"context"
	"fmt"
	"sort"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// UpdateResourceA is a no-op: placing a function does not modify the VPC. The
// security-group rules a placed function needs to reach its targets are opened by the
// access links between the function and those targets.
func (l *vpcFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

// UpdateResourceB sets (or, on destroy, clears) the Lambda function's VpcConfig so it
// runs attached to the flex VPC. Subnets are selected from the VPC's computed outputs by
// the aws.flexvpc.lambda.subnetType tier, and the VPC's Bluelink-managed security group is
// attached.
func (l *vpcFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, err := l.getLambdaService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	functionARN, hasFunctionARN := utils.ExtractARNFromResourceInfo(input.ResourceInfo)
	if !hasFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the %q function resource",
			pluginutils.GetResourceName(input.ResourceInfo),
		)
	}
	functionName := pluginutils.GetResourceName(input.ResourceInfo)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// The flex VPC is the link's other resource; its computed subnets and topology
	// are read from its resource state (populated for create and reference-mode).
	flexVPCState := input.OtherResourceInfo.CurrentResourceState.SpecData
	identity, err := workloadIdentity(flexVPCState, functionName, input.ResourceInfo.InstanceID)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot place function %q in flex VPC %q: %w",
			functionName,
			pluginutils.GetResourceName(input.OtherResourceInfo),
			err,
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		// Detaching the function is all this link does on destroy. Its security group
		// is removed by the flex VPC's own teardown: the group is referenced by rules
		// on groups the access links own, and this link cannot make a peer link revoke
		// them first, so deleting it here fails whenever an access link has not run yet.
		//
		// The detach comes first, and only then the permissions are revoked: Lambda
		// deletes the function's network interfaces using the execution role, so a role
		// stripped of those permissions while still attached leaves interfaces behind
		// that hold the VPC's security groups, and with them the VPC, undeletable.
		if err := linkutils.UpdateLambdaVPCConfig(
			ctx, lambdaService, functionARN,
			/* subnetIDs */ []string{},
			/* securityGroupIDs */ []string{},
			/* ipv6AllowedForDualStack */ false,
		); err != nil {
			return nil, err
		}

		if _, err := l.revokeENIPermissions(ctx, input, providerCtx); err != nil {
			return nil, err
		}

		return &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields(functionName, core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		}, nil
	}

	subnetType := subnetTypeAnnotation(input.ResourceInfo)

	subnetIDs, err := subnetIDsByTier(flexVPCState, subnetType)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot place function %q in flex VPC %q: %w",
			functionName,
			pluginutils.GetResourceName(input.OtherResourceInfo),
			err,
		)
	}

	// The function gets its own group rather than the VPC's shared one, so reach
	// within the VPC is granted per link instead of to everything placed in it.
	securityGroupID, err := resolveWorkloadSecurityGroup(ctx, ec2Service, identity)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot create a security group for function %q in flex VPC %q: %w",
			functionName,
			pluginutils.GetResourceName(input.OtherResourceInfo),
			err,
		)
	}

	// Reach outside the VPC is declared, not derived from links, since the public
	// internet is not a resource the link graph can name.
	egress, err := resolveEgressPlan(input.ResourceInfo, flexVPCState, subnetType)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot resolve outbound access for function %q: %w",
			functionName,
			err,
		)
	}
	if err := authorizeWorkloadEgress(ctx, ec2Service, securityGroupID, egress); err != nil {
		return nil, fmt.Errorf(
			"cannot authorise outbound access for function %q: %w",
			functionName,
			err,
		)
	}

	// Lambda validates the execution role's network interface permissions at the moment
	// the attachment is set, so the grant has to land first. UpdateLambdaVPCConfig
	// retries the rejection IAM's eventual consistency can still produce here.
	grant, err := l.grantENIPermissions(ctx, input, providerCtx)
	if err != nil {
		return nil, err
	}

	securityGroupIDs := []string{securityGroupID}
	ipv6Allowed := subnetsAreDualStack(flexVPCState, subnetIDs)
	if err := linkutils.UpdateLambdaVPCConfig(
		ctx,
		lambdaService,
		functionARN,
		subnetIDs,
		securityGroupIDs,
		ipv6Allowed,
	); err != nil {
		return nil, err
	}

	output := vpcConfigLinkOutput(functionName, subnetIDs, securityGroupIDs, ipv6Allowed)

	return addENIGrantToOutput(output, input, grant), nil
}

// UpdateIntermediaryResources is a no-op: placement creates no intermediary resources.
func (l *vpcFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func subnetTypeAnnotation(resourceInfo *provider.ResourceInfo) string {
	subnetType, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     "aws.flexvpc.lambda.subnetType",
			Default: "private",
		},
	)
	return subnetType
}

// Collects the IDs of the flex VPC's subnets in the requested tier,
// sorted for deterministic output.
func subnetIDsByTier(flexVPCSpecData *core.MappingNode, tier string) ([]string, error) {
	subnetsNode, ok := pluginutils.GetValueByPath("$.subnets", flexVPCSpecData)
	if !ok || subnetsNode == nil || subnetsNode.Fields == nil {
		return nil, fmt.Errorf("the linked flex VPC exposes no subnets")
	}

	ids := []string{}
	for _, subnet := range subnetsNode.Fields {
		typeNode, _ := pluginutils.GetValueByPath("$.subnetType", subnet)
		if core.StringValue(typeNode) != tier {
			continue
		}
		idNode, _ := pluginutils.GetValueByPath("$.id", subnet)
		if id := core.StringValue(idNode); id != "" {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("the linked flex VPC has no %q subnets", tier)
	}

	sort.Strings(ids)
	return ids, nil
}

// Whether every subnet the function is placed in carries an IPv6 CIDR. A function is
// placed in all subnets of its tier, so IPv6 is only enabled when all of them can
// provide an address; a partially dual-stack VPC would otherwise give the function an
// egress path that depends on which subnet an invocation happened to land in.
func subnetsAreDualStack(flexVPCSpecData *core.MappingNode, subnetIDs []string) bool {
	subnetsNode, ok := pluginutils.GetValueByPath("$.subnets", flexVPCSpecData)
	if !ok || subnetsNode == nil || subnetsNode.Fields == nil {
		return false
	}

	dualStack := map[string]bool{}
	for _, subnet := range subnetsNode.Fields {
		idNode, _ := pluginutils.GetValueByPath("$.id", subnet)
		ipv6Node, _ := pluginutils.GetValueByPath("$.ipv6CidrBlock", subnet)
		if id := core.StringValue(idNode); id != "" {
			dualStack[id] = core.StringValue(ipv6Node) != ""
		}
	}

	for _, subnetID := range subnetIDs {
		if !dualStack[subnetID] {
			return false
		}
	}

	return len(subnetIDs) > 0
}

func vpcConfigLinkOutput(
	functionName string,
	subnetIDs, securityGroupIDs []string,
	ipv6AllowedForDualStack bool,
) *provider.LinkUpdateResourceOutput {
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			functionName,
			core.MappingNodeFields(
				"vpcConfig",
				core.MappingNodeFields(
					"subnetIds", stringItems(subnetIDs),
					"securityGroupIds", stringItems(securityGroupIDs),
					"ipv6AllowedForDualStack", core.MappingNodeFromBool(ipv6AllowedForDualStack),
				),
			),
		),
		ResourceDataMappings: map[string]string{
			fmt.Sprintf("%s::spec.vpcConfig.subnetIds", functionName):        fmt.Sprintf("%s.vpcConfig.subnetIds", functionName),
			fmt.Sprintf("%s::spec.vpcConfig.securityGroupIds", functionName): fmt.Sprintf("%s.vpcConfig.securityGroupIds", functionName),
			fmt.Sprintf("%s::spec.vpcConfig.ipv6AllowedForDualStack", functionName): fmt.Sprintf(
				"%s.vpcConfig.ipv6AllowedForDualStack",
				functionName,
			),
		},
	}
}

func stringItems(values []string) *core.MappingNode {
	items := make([]*core.MappingNode, len(values))
	for i, value := range values {
		items[i] = core.MappingNodeFromString(value)
	}
	return &core.MappingNode{Items: items}
}
