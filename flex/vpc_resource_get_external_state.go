package flex

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (v *vpcResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	meta := awsConfigMetaWithRegionGetExternalState(input)
	service, region, err := v.getEC2ServiceWithRegion(ctx, input.ProviderContext, meta)
	if err != nil {
		return nil, err
	}

	vpcNameValue, hasVPCName := pluginutils.GetValueByPath("$.name", input.CurrentResourceSpec)
	if !hasVPCName {
		return nil, errors.New(
			"name is required to get external state for flex VPC resource",
		)
	}
	vpcName := core.StringValue(vpcNameValue)

	// Fields such as mode and preset are not reflected in the external state
	// so must be sourced from the input resource spec.
	mode, _ := pluginutils.GetValueByPath("$.mode", input.CurrentResourceSpec)
	preset, _ := pluginutils.GetValueByPath("$.preset", input.CurrentResourceSpec)
	// Diffing tags across multiple resources against the flex VPC resource tags
	// doesn't provide much benefit, as tags are used to query flex VPC resources,
	// failure to find a resource with a given tag will surface an error indicating
	// unexpected changes to the resources that power the "flex VPC".
	tags, _ := pluginutils.GetValueByPath("$.tags", input.CurrentResourceSpec)

	vpc, err := v.getVPCByNameTag(ctx, service, vpcName)
	if err != nil {
		// If no VPC found by flex VPC name tag, attempt fallback lookup by Bluelink tags
		// This handles interrupted resource creation where the VPC may exist but without
		// the expected flex VPC tags.
		fallbackVPC, fallbackErr := v.lookupVPCByBluelinkTags(ctx, input, meta)
		if fallbackErr != nil {
			return nil, fmt.Errorf("failed to lookup VPC by tags: %w", fallbackErr)
		}
		if fallbackVPC == nil {
			// Resource doesn't exist yet
			return &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				},
			}, nil
		}
		vpc = fallbackVPC
	}

	vpcAttributes, err := v.getVPCAttributes(ctx, service, vpc.VpcId)
	if err != nil {
		return nil, err
	}

	subnets, err := v.getSubnets(ctx, service, vpcName)
	if err != nil {
		return nil, err
	}

	routeTables, natGatewaySubnetMappings, err := v.getRouteTablesAndSubnetNATGatewayMappings(ctx, service, vpcName)
	if err != nil {
		return nil, err
	}

	securityGroups, err := v.getSecurityGroups(ctx, service, vpcName)
	if err != nil {
		return nil, err
	}

	securityGroupIdsByName, err := namedSecurityGroupIDs(ctx, service, vpcName, input.InstanceID)
	if err != nil {
		return nil, err
	}

	networkACLs, err := v.getNetworkACLs(ctx, service, vpcName)
	if err != nil {
		return nil, err
	}

	gateways, err := v.getGateways(ctx, service, vpcName, natGatewaySubnetMappings)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: core.MappingNodeFields(
			"name",
			vpcNameValue,
			"mode",
			mode,
			"preset",
			preset,
			"cidrBlock",
			core.MappingNodeFromString(aws.ToString(vpc.CidrBlock)),
			"enableDNSSupport",
			core.MappingNodeFromBool(vpcAttributes.dnsSupport),
			"enableDNSHostnames",
			core.MappingNodeFromBool(vpcAttributes.dnsHostnames),
			"region",
			core.MappingNodeFromString(region),
			"tags",
			tags,
			"vpcId",
			core.MappingNodeFromString(aws.ToString(vpc.VpcId)),
			"subnets",
			subnets,
			"privateSubnetIds",
			specSubnetIdsByTier(subnets, "private"),
			"publicSubnetIds",
			specSubnetIdsByTier(subnets, "public"),
			"routeTables",
			routeTables,
			"securityGroupIds",
			securityGroups,
			"securityGroupIdsByName",
			toSpecComputedSecurityGroupIdsByName(securityGroupIdsByName),
			"networkAcls",
			networkACLs,
			"gateways",
			gateways,
		),
	}, nil
}

func (v *vpcResourceActions) getVPCByNameTag(
	ctx context.Context,
	service ec2service.Service,
	vpcName string,
) (*types.Vpc, error) {
	vpcsOutput, err := service.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
				Values: []string{vpcName},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(vpcsOutput.Vpcs) == 0 {
		return nil, errors.New("no VPC found with name")
	}

	return &vpcsOutput.Vpcs[0], nil
}

type vpcAttributes struct {
	dnsSupport   bool
	dnsHostnames bool
}

func (v *vpcResourceActions) getVPCAttributes(
	ctx context.Context,
	service ec2service.Service,
	vpcID *string,
) (*vpcAttributes, error) {
	dnsSupportOutput, err := service.DescribeVpcAttribute(ctx, &ec2.DescribeVpcAttributeInput{
		Attribute: types.VpcAttributeNameEnableDnsSupport,
		VpcId:     vpcID,
	})
	if err != nil {
		return nil, err
	}

	dnsHostnamesOutput, err := service.DescribeVpcAttribute(ctx, &ec2.DescribeVpcAttributeInput{
		Attribute: types.VpcAttributeNameEnableDnsHostnames,
		VpcId:     vpcID,
	})
	if err != nil {
		return nil, err
	}

	var dnsSupport bool
	if dnsSupportOutput.EnableDnsSupport != nil && dnsSupportOutput.EnableDnsSupport.Value != nil {
		dnsSupport = aws.ToBool(dnsSupportOutput.EnableDnsSupport.Value)
	}

	var dnsHostnames bool
	if dnsHostnamesOutput.EnableDnsHostnames != nil && dnsHostnamesOutput.EnableDnsHostnames.Value != nil {
		dnsHostnames = aws.ToBool(dnsHostnamesOutput.EnableDnsHostnames.Value)
	}

	return &vpcAttributes{
		dnsSupport:   dnsSupport,
		dnsHostnames: dnsHostnames,
	}, nil
}

func (v *vpcResourceActions) getSubnets(
	ctx context.Context,
	service ec2service.Service,
	vpcName string,
) (*core.MappingNode, error) {
	subnetsOutput, err := service.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
				Values: []string{vpcName},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	subnets := core.MappingNodeFields()
	for _, subnet := range subnetsOutput.Subnets {
		subnetName := ec2TagValue(subnet.Tags, TagFlexVPCSubnetName)
		if subnetName == "" {
			subnetName = aws.ToString(subnet.SubnetId)
		}
		// Shared with the create path so a referenced VPC and a managed one expose
		// subnets in the same shape.
		subnets.Fields[subnetName] = toSpecComputedSubnet(
			&subnet,
			ec2TagValue(subnet.Tags, TagFlexVPCSubnetType),
		)
	}

	return subnets, nil
}

func ec2TagValue(tags []types.Tag, key string) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

func (v *vpcResourceActions) getRouteTablesAndSubnetNATGatewayMappings(
	ctx context.Context,
	service ec2service.Service,
	vpcName string,
) (*core.MappingNode, *core.MappingNode, error) {
	routeTablesOutput, err := service.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
				Values: []string{vpcName},
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}

	if len(routeTablesOutput.RouteTables) == 0 {
		return &core.MappingNode{}, &core.MappingNode{}, nil
	}

	items := []*core.MappingNode{}
	natGatewaySubnetMappings := map[string]*core.MappingNode{}
	for _, routeTable := range routeTablesOutput.RouteTables {
		subnetIds := []*core.MappingNode{}
		for _, association := range routeTable.Associations {
			subnetIds = append(
				subnetIds,
				core.MappingNodeFromString(aws.ToString(association.SubnetId)),
			)
			natGatewayID := getSubnetNATGateway(
				routeTable.Routes,
			)
			natGatewaySubnetMappings[core.StringValue(natGatewayID)] = core.MappingNodeFromString(aws.ToString(association.SubnetId))
		}

		items = append(items, core.MappingNodeFields(
			"id",
			core.MappingNodeFromString(aws.ToString(routeTable.RouteTableId)),
			"subnetIds",
			core.MappingNodeItems(subnetIds...),
		))
	}

	return core.MappingNodeItems(items...), &core.MappingNode{
		Fields: natGatewaySubnetMappings,
	}, nil
}

func (v *vpcResourceActions) getSecurityGroups(
	ctx context.Context,
	service ec2service.Service,
	vpcName string,
) (*core.MappingNode, error) {
	// Filtered on the base-group tag as well as the VPC name so the per-policy
	// groups, which carry the VPC name but not this tag, are not swept in here.
	securityGroupsOutput, err := service.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
				Values: []string{vpcName},
			},
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCSecurityGroup)),
				Values: []string{"true"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(securityGroupsOutput.SecurityGroups) == 0 {
		return &core.MappingNode{}, nil
	}

	items := []*core.MappingNode{}
	for _, securityGroup := range securityGroupsOutput.SecurityGroups {
		items = append(
			items,
			core.MappingNodeFromString(aws.ToString(securityGroup.GroupId)),
		)
	}

	return core.MappingNodeItems(items...), nil
}

func (v *vpcResourceActions) getNetworkACLs(
	ctx context.Context,
	service ec2service.Service,
	vpcName string,
) (*core.MappingNode, error) {
	networkACLsOutput, err := service.DescribeNetworkAcls(ctx, &ec2.DescribeNetworkAclsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
				Values: []string{vpcName},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(networkACLsOutput.NetworkAcls) == 0 {
		return &core.MappingNode{}, nil
	}

	items := []*core.MappingNode{}
	for _, networkACL := range networkACLsOutput.NetworkAcls {
		subnetIds := []*core.MappingNode{}
		for _, association := range networkACL.Associations {
			subnetIds = append(subnetIds, core.MappingNodeFromString(aws.ToString(association.SubnetId)))
		}

		items = append(items, core.MappingNodeFields(
			"id",
			core.MappingNodeFromString(aws.ToString(networkACL.NetworkAclId)),
			"subnetIds",
			core.MappingNodeItems(subnetIds...),
		))
	}

	return core.MappingNodeItems(items...), nil
}

func (v *vpcResourceActions) getGateways(
	ctx context.Context,
	service ec2service.Service,
	vpcName string,
	natGatewaySubnetMappings *core.MappingNode,
) (*core.MappingNode, error) {
	igwOutput, err := service.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
				Values: []string{vpcName},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	natGateways, err := v.getNatGateways(ctx, service, vpcName, natGatewaySubnetMappings)
	if err != nil {
		return nil, err
	}

	gateways := core.MappingNodeFields(
		"natGateways",
		natGateways,
	)

	// An isolated VPC has no internet gateway, so its absence is expected rather than
	// a failure to discover one.
	if len(igwOutput.InternetGateways) > 0 {
		gateways.Fields["internetGatewayId"] = core.MappingNodeFromString(
			aws.ToString(igwOutput.InternetGateways[0].InternetGatewayId),
		)
	}

	eigwOutput, err := service.DescribeEgressOnlyInternetGateways(ctx, &ec2.DescribeEgressOnlyInternetGatewaysInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
				Values: []string{vpcName},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Presets without an internet-connected private subnet have none, so its absence
	// is not an error the way a missing internet gateway is.
	if eigwOutput != nil && len(eigwOutput.EgressOnlyInternetGateways) > 0 {
		gateways.Fields["egressOnlyInternetGatewayId"] = core.MappingNodeFromString(
			aws.ToString(eigwOutput.EgressOnlyInternetGateways[0].EgressOnlyInternetGatewayId),
		)
	}

	return gateways, nil
}

func (v *vpcResourceActions) getNatGateways(
	ctx context.Context,
	service ec2service.Service,
	vpcName string,
	natGatewaySubnetMappings *core.MappingNode,
) (*core.MappingNode, error) {
	natGatewaysOutput, err := service.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{
		Filter: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
				Values: []string{vpcName},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(natGatewaysOutput.NatGateways) == 0 {
		return &core.MappingNode{}, nil
	}

	items := []*core.MappingNode{}
	for _, natGateway := range natGatewaysOutput.NatGateways {
		items = append(items, core.MappingNodeFields(
			"id",
			core.MappingNodeFromString(aws.ToString(natGateway.NatGatewayId)),
			"elasticIpId",
			core.MappingNodeFromString(aws.ToString(natGateway.NatGatewayAddresses[0].AllocationId)),
			"inPublicSubnetId",
			core.MappingNodeFromString(aws.ToString(natGateway.SubnetId)),
			"forPrivateSubnetId",
			natGatewaySubnetMappings.Fields[aws.ToString(natGateway.NatGatewayId)],
		))
	}

	return core.MappingNodeItems(items...), nil
}

func getSubnetNATGateway(
	routes []types.Route,
) *core.MappingNode {
	for _, route := range routes {
		// Any route with a NAT gateway ID in the current route table
		// will apply to any subnet associated with the route table.
		if route.NatGatewayId != nil {
			return core.MappingNodeFromString(aws.ToString(route.NatGatewayId))
		}
	}
	return nil
}

// lookupVPCByBluelinkTags attempts to find a VPC by its Bluelink provenance tags.
// This is used as a fallback when the VPC cannot be found by its flex VPC name tag
// (e.g., interrupted resource creation).
// Returns nil if no matching VPC is found.
func (v *vpcResourceActions) lookupVPCByBluelinkTags(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
	meta map[string]*core.MappingNode,
) (*types.Vpc, error) {
	tagFilters := utils.BuildBluelinkTagFiltersForLookup(input)
	if tagFilters == nil {
		// Tagging is not enabled, cannot perform fallback lookup
		return nil, nil
	}

	taggingService, err := v.getResourceGroupTaggingService(ctx, input.ProviderContext, meta)
	if err != nil {
		return nil, err
	}

	result, err := taggingService.GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters:          tagFilters,
		ResourceTypeFilters: []string{"ec2:vpc"},
	})
	if err != nil {
		return nil, err
	}

	if len(result.ResourceTagMappingList) == 0 {
		return nil, nil
	}

	// Extract VPC ID from ARN
	// ARN format: arn:aws:ec2:region:account-id:vpc/vpc-id
	vpcARN := aws.ToString(result.ResourceTagMappingList[0].ResourceARN)
	vpcID := extractVPCIDFromARN(vpcARN)
	if vpcID == "" {
		return nil, fmt.Errorf("failed to extract VPC ID from ARN: %s", vpcARN)
	}

	// Fetch the full VPC details using the VPC ID
	service, _, err := v.getEC2ServiceWithRegion(ctx, input.ProviderContext, meta)
	if err != nil {
		return nil, err
	}

	vpcsOutput, err := service.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcID},
	})
	if err != nil {
		return nil, err
	}

	if len(vpcsOutput.Vpcs) == 0 {
		return nil, nil
	}

	return &vpcsOutput.Vpcs[0], nil
}

// extractVPCIDFromARN extracts the VPC ID from an ARN.
// ARN format: arn:aws:ec2:region:account-id:vpc/vpc-id.
func extractVPCIDFromARN(arn string) string {
	// Find the last "/" and extract everything after it
	for i := len(arn) - 1; i >= 0; i-- {
		if arn[i] == '/' {
			return arn[i+1:]
		}
	}
	return ""
}
