package flex

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *vpcResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	service, _, err := l.getEC2ServiceWithRegion(ctx, input.ProviderContext, nil)
	if err != nil {
		return nil, err
	}

	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)
	updatedResourceSpecData := pluginutils.GetResolvedResourceSpecData(input.Changes)

	mode, hasMode := pluginutils.GetValueByPath("$.mode", updatedResourceSpecData)
	if !hasMode {
		return nil, errors.New("mode is required")
	}

	modeStr := core.StringValue(mode)

	if modeStr == "reference" {
		return l.updateReferencedVPC(
			ctx,
			input,
			service,
			currentStateSpecData,
			updatedResourceSpecData,
		)
	}

	vpcIDValue, hasVPCID := pluginutils.GetValueByPath("$.vpcId", currentStateSpecData)
	if !hasVPCID {
		return nil, errors.New("vpcId is not set in resource spec data")
	}
	vpcID := core.StringValue(vpcIDValue)

	err = l.updateConcreteVPC(
		ctx,
		input,
		vpcID,
		service,
		updatedResourceSpecData,
	)
	if err != nil {
		return nil, err
	}

	err = l.updateResourceTags(
		ctx,
		input,
		vpcID,
		service,
		currentStateSpecData,
	)
	if err != nil {
		return nil, err
	}

	userTags, _ := pluginutils.GetValueByPath("$.tags", updatedResourceSpecData)
	if userTags == nil {
		userTags = core.MappingNodeFields()
	}

	namedGroupIDs, err := l.updateNamedSecurityGroups(
		ctx,
		input,
		service,
		vpcID,
		currentStateSpecData,
		updatedResourceSpecData,
		userTags,
	)
	if err != nil {
		return nil, err
	}

	// Apart from the security groups, only user-defined fields are updated, so the
	// remaining computed field values are sourced from the current state.
	computedFields := computedFieldsFromCurrentState(currentStateSpecData)
	computedFields["spec.securityGroupIdsByName"] = toSpecComputedSecurityGroupIdsByName(
		namedGroupIDs,
	)

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}

// The shared topology of a referenced VPC belongs to another blueprint, so an update
// reconciles this application's own named security groups and nothing else.
func (l *vpcResourceActions) updateReferencedVPC(
	ctx context.Context,
	input *provider.ResourceDeployInput,
	service ec2service.Service,
	currentStateSpecData *core.MappingNode,
	updatedResourceSpecData *core.MappingNode,
) (*provider.ResourceDeployOutput, error) {
	computedFields := computedFieldsFromCurrentState(currentStateSpecData)

	vpcIDValue, hasVPCID := pluginutils.GetValueByPath("$.vpcId", currentStateSpecData)
	if !hasVPCID {
		return &provider.ResourceDeployOutput{ComputedFieldValues: computedFields}, nil
	}

	namedGroupIDs, err := l.updateNamedSecurityGroups(
		ctx,
		input,
		service,
		core.StringValue(vpcIDValue),
		currentStateSpecData,
		updatedResourceSpecData,
		// A referenced VPC's user tags describe the owner's VPC and are ignored in
		// this mode, so they are not applied to the groups created here.
		core.MappingNodeFields(),
	)
	if err != nil {
		return nil, err
	}

	computedFields["spec.securityGroupIdsByName"] = toSpecComputedSecurityGroupIdsByName(
		namedGroupIDs,
	)

	return &provider.ResourceDeployOutput{ComputedFieldValues: computedFields}, nil
}

// Reconciles the VPC's named security groups against what the updated spec declares.
//
// The reconcile reads the groups back from AWS by tag rather than trusting a list carried
// in state, so for a VPC that declares any it runs on every update, which also repairs a
// group deleted outside Bluelink. A VPC that has never declared one skips it entirely
// rather than paying for a describe on every update.
//
// The guard has to consider the current state as well as the updated spec, or removing
// the last name would skip the reconcile that deletes its group.
func (l *vpcResourceActions) updateNamedSecurityGroups(
	ctx context.Context,
	input *provider.ResourceDeployInput,
	service ec2service.Service,
	vpcID string,
	currentStateSpecData *core.MappingNode,
	updatedResourceSpecData *core.MappingNode,
	userTags *core.MappingNode,
) (map[string]string, error) {
	declared := parseSecurityGroupNames(updatedResourceSpecData)
	current, _ := pluginutils.GetValueByPath("$.securityGroupIdsByName", currentStateSpecData)
	if len(declared) == 0 && (current == nil || len(current.Fields) == 0) {
		return map[string]string{}, nil
	}

	name, hasName := pluginutils.GetValueByPath("$.name", updatedResourceSpecData)
	if !hasName {
		return nil, errors.New("name is not set in resource spec data")
	}

	vpcName := core.StringValue(name)

	vpcCtx := &vpcContext{
		resourceSpecData: updatedResourceSpecData,
		vpcName:          vpcName,
		tags:             vpcElementTags(vpcName, userTags, getBluelinkTagsAsEC2Tags(input)),
		instanceID:       input.InstanceID,
		vpcID:            aws.String(vpcID),
	}

	return l.reconcileNamedSecurityGroups(ctx, service, vpcCtx)
}

func (l *vpcResourceActions) updateConcreteVPC(
	ctx context.Context,
	input *provider.ResourceDeployInput,
	vpcID string,
	service ec2service.Service,
	updatedResourceSpecData *core.MappingNode,
) error {
	if pluginutils.HasModifiedField(input.Changes, "spec.enableDNSSupport") {
		enableDNSSupport, hasEnableDNSSupport := pluginutils.GetValueByPath(
			"$.enableDNSSupport",
			updatedResourceSpecData,
		)
		if !hasEnableDNSSupport {
			return errors.New("enableDNSSupport is not set in resource spec data")
		}

		_, err := service.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
			VpcId: aws.String(vpcID),
			EnableDnsSupport: &types.AttributeBooleanValue{
				Value: aws.Bool(core.BoolValue(enableDNSSupport)),
			},
		})
		if err != nil {
			return err
		}
	}

	if pluginutils.HasModifiedField(input.Changes, "spec.enableDNSHostnames") {
		enableDNSHostnames, hasEnableDNSHostnames := pluginutils.GetValueByPath(
			"$.enableDNSHostnames",
			updatedResourceSpecData,
		)
		if !hasEnableDNSHostnames {
			return errors.New("enableDNSHostnames is not set in resource spec data")
		}

		_, err := service.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
			VpcId: aws.String(vpcID),
			EnableDnsHostnames: &types.AttributeBooleanValue{
				Value: aws.Bool(core.BoolValue(enableDNSHostnames)),
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *vpcResourceActions) updateResourceTags(
	ctx context.Context,
	input *provider.ResourceDeployInput,
	vpcID string,
	service ec2service.Service,
	currentStateSpecData *core.MappingNode,
) error {
	diffResult := utils.DiffTags(
		input.Changes,
		"$.tags",
		func(tag *utils.Tag) types.Tag {
			return types.Tag{
				Key:   aws.String(tag.Key),
				Value: aws.String(tag.Value),
			}
		},
	)

	if len(diffResult.ToSet) == 0 && len(diffResult.ToRemove) == 0 {
		return nil
	}

	resourceIDs := collectVPCResourceIDs(currentStateSpecData, vpcID)

	if len(diffResult.ToSet) > 0 {
		_, err := service.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: resourceIDs,
			Tags:      diffResult.ToSet,
		})
		if err != nil {
			return err
		}
	}

	if len(diffResult.ToRemove) > 0 {
		_, err := service.DeleteTags(ctx, &ec2.DeleteTagsInput{
			Resources: resourceIDs,
			Tags:      toTagKeys(diffResult.ToRemove),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func toTagKeys(tagKeys []string) []types.Tag {
	tags := []types.Tag{}
	for _, tagKey := range tagKeys {
		tags = append(
			tags,
			types.Tag{
				Key: aws.String(tagKey),
			},
		)
	}
	return tags
}

func collectVPCResourceIDs(
	currentStateSpecData *core.MappingNode,
	vpcID string,
) []string {
	resourceIDs := []string{vpcID}

	subnetResourceIDs := collectVPCSubnetResourceIDs(currentStateSpecData)
	resourceIDs = append(resourceIDs, subnetResourceIDs...)

	routeTableResourceIDs := collectVPCRouteTableResourceIDs(currentStateSpecData)
	resourceIDs = append(resourceIDs, routeTableResourceIDs...)

	securityGroupResourceIDs := collectVPCSecurityGroupResourceIDs(currentStateSpecData)
	resourceIDs = append(resourceIDs, securityGroupResourceIDs...)

	networkACLResourceIDs := collectVPCNetworkACLResourceIDs(currentStateSpecData)
	resourceIDs = append(resourceIDs, networkACLResourceIDs...)

	gatewayResourceIDs := collectVPCGatewayResourceIDs(currentStateSpecData)
	resourceIDs = append(resourceIDs, gatewayResourceIDs...)

	return resourceIDs
}

func collectVPCSubnetResourceIDs(
	currentStateSpecData *core.MappingNode,
) []string {
	subnetResourceIDs := []string{}

	subnets, hasSubnets := pluginutils.GetValueByPath("$.subnets", currentStateSpecData)
	if !hasSubnets {
		return subnetResourceIDs
	}

	// Subnets are held in state as a map keyed by the preset's subnet name.
	for _, subnetName := range sortedFieldNames(subnets) {
		subnetID, hasSubnetID := pluginutils.GetValueByPath("$.id", subnets.Fields[subnetName])
		if hasSubnetID {
			subnetResourceIDs = append(subnetResourceIDs, core.StringValue(subnetID))
		}
	}

	return subnetResourceIDs
}

func collectVPCRouteTableResourceIDs(
	currentStateSpecData *core.MappingNode,
) []string {
	routeTableResourceIDs := []string{}

	routeTables, hasRouteTables := pluginutils.GetValueByPath("$.routeTables", currentStateSpecData)
	if !hasRouteTables {
		return routeTableResourceIDs
	}

	for _, routeTable := range routeTables.Items {
		routeTableID, hasRouteTableID := pluginutils.GetValueByPath("$.id", routeTable)
		if hasRouteTableID {
			routeTableResourceIDs = append(
				routeTableResourceIDs,
				core.StringValue(routeTableID),
			)
		}
	}

	return routeTableResourceIDs
}

func collectVPCSecurityGroupResourceIDs(
	currentStateSpecData *core.MappingNode,
) []string {
	securityGroupResourceIDs := []string{}

	securityGroups, hasSecurityGroups := pluginutils.GetValueByPath("$.securityGroupIds", currentStateSpecData)
	if !hasSecurityGroups {
		return securityGroupResourceIDs
	}

	for _, securityGroup := range securityGroups.Items {
		securityGroupResourceIDs = append(
			securityGroupResourceIDs,
			core.StringValue(securityGroup),
		)
	}

	return securityGroupResourceIDs
}

func collectVPCNetworkACLResourceIDs(
	currentStateSpecData *core.MappingNode,
) []string {
	networkACLResourceIDs := []string{}

	networkACLs, hasNetworkACLs := pluginutils.GetValueByPath("$.networkAcls", currentStateSpecData)
	if !hasNetworkACLs {
		return networkACLResourceIDs
	}

	for _, networkACL := range networkACLs.Items {
		networkACLID, hasNetworkACLID := pluginutils.GetValueByPath("$.id", networkACL)
		if hasNetworkACLID {
			networkACLResourceIDs = append(networkACLResourceIDs, core.StringValue(networkACLID))
		}
	}

	return networkACLResourceIDs
}

func collectVPCGatewayResourceIDs(
	currentStateSpecData *core.MappingNode,
) []string {
	gatewayResourceIDs := []string{}

	gateways, hasGateways := pluginutils.GetValueByPath("$.gateways", currentStateSpecData)
	if !hasGateways {
		return gatewayResourceIDs
	}

	internetGatewayID, hasInternetGatewayID := pluginutils.GetValueByPath("$.internetGatewayId", gateways)
	if hasInternetGatewayID {
		gatewayResourceIDs = append(gatewayResourceIDs, core.StringValue(internetGatewayID))
	}

	natGateways, hasNatGateways := pluginutils.GetValueByPath("$.natGateways", gateways)
	if !hasNatGateways {
		return gatewayResourceIDs
	}

	for _, natGateway := range natGateways.Items {
		natGatewayID, hasNatGatewayID := pluginutils.GetValueByPath("$.id", natGateway)
		if hasNatGatewayID {
			gatewayResourceIDs = append(gatewayResourceIDs, core.StringValue(natGatewayID))
		}

		elasticIPID, hasElasticIPID := pluginutils.GetValueByPath("$.elasticIpId", natGateway)
		if hasElasticIPID {
			gatewayResourceIDs = append(gatewayResourceIDs, core.StringValue(elasticIPID))
		}
	}

	return gatewayResourceIDs
}

func computedFieldsFromCurrentState(currentStateSpecData *core.MappingNode) map[string]*core.MappingNode {
	vpcID, _ := pluginutils.GetValueByPath("$.vpcId", currentStateSpecData)
	subnets, _ := pluginutils.GetValueByPath("$.subnets", currentStateSpecData)
	routeTables, _ := pluginutils.GetValueByPath("$.routeTables", currentStateSpecData)
	securityGroups, _ := pluginutils.GetValueByPath("$.securityGroupIds", currentStateSpecData)
	securityGroupIdsByName, _ := pluginutils.GetValueByPath(
		"$.securityGroupIdsByName",
		currentStateSpecData,
	)
	networkACLs, _ := pluginutils.GetValueByPath("$.networkAcls", currentStateSpecData)
	gateways, _ := pluginutils.GetValueByPath("$.gateways", currentStateSpecData)

	return map[string]*core.MappingNode{
		"spec.vpcId":   core.MappingNodeFromString(core.StringValue(vpcID)),
		"spec.subnets": subnets,
		// Derived from the subnets map rather than carried forward so that state
		// saved before these fields existed is populated on the next update.
		"spec.privateSubnetIds":       specSubnetIdsByTier(subnets, "private"),
		"spec.publicSubnetIds":        specSubnetIdsByTier(subnets, "public"),
		"spec.routeTables":            routeTables,
		"spec.securityGroupIds":       securityGroups,
		"spec.securityGroupIdsByName": securityGroupIdsByName,
		"spec.networkAcls":            networkACLs,
		"spec.gateways":               gateways,
	}
}
