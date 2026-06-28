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

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if err := linkutils.UpdateLambdaVPCConfig(ctx, lambdaService, functionARN, []string{}, []string{}); err != nil {
			return nil, err
		}
		return &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields(functionName, core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		}, nil
	}

	subnetType := subnetTypeAnnotation(input.ResourceInfo)

	// The flex VPC is the link's other resource; its computed subnets and security
	// group are read from its resource state (populated for create and reference-mode).
	flexVPCState := input.OtherResourceInfo.CurrentResourceState.SpecData
	subnetIDs, err := subnetIDsByTier(flexVPCState, subnetType)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot place function %q in flex VPC %q: %w",
			functionName,
			pluginutils.GetResourceName(input.OtherResourceInfo),
			err,
		)
	}
	securityGroupID, hasSecurityGroup := firstSecurityGroup(flexVPCState)
	if !hasSecurityGroup {
		return nil, fmt.Errorf(
			"no security group found on the linked flex VPC %q",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	securityGroupIDs := []string{securityGroupID}
	if err := linkutils.UpdateLambdaVPCConfig(ctx, lambdaService, functionARN, subnetIDs, securityGroupIDs); err != nil {
		return nil, err
	}

	return vpcConfigLinkOutput(functionName, subnetIDs, securityGroupIDs), nil
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

func firstSecurityGroup(flexVPCSpecData *core.MappingNode) (string, bool) {
	sgNode, ok := pluginutils.GetValueByPath("$.securityGroups", flexVPCSpecData)
	if !ok || sgNode == nil || len(sgNode.Items) == 0 {
		return "", false
	}
	return core.StringValue(sgNode.Items[0]), true
}

func vpcConfigLinkOutput(
	functionName string,
	subnetIDs, securityGroupIDs []string,
) *provider.LinkUpdateResourceOutput {
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			functionName,
			core.MappingNodeFields(
				"vpcConfig",
				core.MappingNodeFields(
					"subnetIds", stringItems(subnetIDs),
					"securityGroupIds", stringItems(securityGroupIDs),
				),
			),
		),
		ResourceDataMappings: map[string]string{
			fmt.Sprintf("%s::spec.vpcConfig.subnetIds", functionName):        fmt.Sprintf("%s.vpcConfig.subnetIds", functionName),
			fmt.Sprintf("%s::spec.vpcConfig.securityGroupIds", functionName): fmt.Sprintf("%s.vpcConfig.securityGroupIds", functionName),
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
