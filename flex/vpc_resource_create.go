package flex

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	resgrouptagtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// The timeout for waiting for a VPC to be available.
const vpcWaitTimeout = 10 * time.Minute

// The timeout for waiting for a NAT gateway to be available.
const natGatewayWaitTimeout = 10 * time.Minute

// The timeout for waiting for an Amazon-provided IPv6 CIDR block to be
// associated with a VPC, and how often the association is re-checked.
const (
	vpcIPv6CIDRWaitTimeout  = 2 * time.Minute
	vpcIPv6CIDRPollInterval = 2 * time.Second
)

func (l *vpcResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	// Make sure the AWS config is created with the correct region
	// so requests to the EC2 service are made to the user-defined region.
	meta := awsConfigMetaWithRegion(input)
	service, region, err := l.getEC2ServiceWithRegion(ctx, input.ProviderContext, meta)
	if err != nil {
		return nil, err
	}

	resourceSpecData := pluginutils.GetResolvedResourceSpecData(input.Changes)

	mode, hasMode := pluginutils.GetValueByPath("$.mode", resourceSpecData)
	if !hasMode {
		return nil, errors.New("mode is required")
	}

	modeStr := core.StringValue(mode)

	if modeStr == "reference" {
		// When in reference mode, use get external state to synchronise the resource
		// in the current blueprint with the resource state in the target AWS account.
		externalStateOutput, err := l.GetExternalState(ctx, &provider.ResourceGetExternalStateInput{
			ProviderContext:     input.ProviderContext,
			CurrentResourceSpec: resourceSpecData,
			InstanceID:          input.InstanceID,
			InstanceName:        input.InstanceName,
		})
		if err != nil {
			return nil, err
		}

		return l.referenceModeDeployOutput(ctx, service, input, resourceSpecData, externalStateOutput)
	}

	// When in create mode, if the VPC already exists, then the deployment should fail
	// as it indicates that the VPC has been created as a part of a different deployment
	// and should be defined using the reference mode in the current blueprint.
	err = l.validateVPCDoesNotExist(ctx, service, resourceSpecData)
	if err != nil {
		return nil, err
	}

	resourceGroupTaggingService, err := l.getResourceGroupTaggingService(
		ctx,
		input.ProviderContext,
		meta,
	)
	if err != nil {
		return nil, err
	}

	err = l.validateVPCResourcesDoNotExist(ctx, resourceGroupTaggingService, resourceSpecData)
	if err != nil {
		return nil, err
	}

	return l.createFlexVPCResources(ctx, service, resourceSpecData, region, input)
}

func (l *vpcResourceActions) validateVPCDoesNotExist(
	ctx context.Context,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) error {
	vpcs, name, err := l.describeVPCsByName(ctx, service, resourceSpecData)
	if err != nil {
		return err
	}

	if len(vpcs) > 0 {
		return fmt.Errorf(
			"VPC can not be created in this blueprint as one already exists with the flex VPC name: %s",
			name,
		)
	}

	return nil
}

func (l *vpcResourceActions) describeVPCsByName(
	ctx context.Context,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) ([]types.Vpc, string, error) {
	name, hasName := pluginutils.GetValueByPath("$.name", resourceSpecData)
	if !hasName {
		return nil, "", errors.New("name is required for flex VPC")
	}

	vpcsOutput, err := service.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []types.Filter{
			TagFilterFlexVPCName(core.StringValue(name)),
		},
	})
	if err != nil {
		return nil, "", err
	}

	return vpcsOutput.Vpcs, core.StringValue(name), nil
}

func (l *vpcResourceActions) validateVPCResourcesDoNotExist(
	ctx context.Context,
	resourceGroupTaggingService resgrouptagservice.Service,
	resourceSpecData *core.MappingNode,
) error {
	vpcName, hasVPCName := pluginutils.GetValueByPath("$.name", resourceSpecData)
	if !hasVPCName {
		return errors.New("name is required for flex VPC")
	}

	resources, err := resourceGroupTaggingService.GetResources(
		ctx,
		&resourcegroupstaggingapi.GetResourcesInput{
			ResourceTypeFilters: []string{
				"ec2:subnet",
				"ec2:route-table",
				"ec2:internet-gateway",
				"ec2:vpc",
				"ec2:security-group",
				"ec2:network-acl",
				"ec2:nat-gateway",
				"ec2:elastic-ip",
			},
			TagFilters: []resgrouptagtypes.TagFilter{
				{
					Key: aws.String(TagFlexVPCName),
					Values: []string{
						core.StringValue(vpcName),
					},
				},
			},
		},
	)
	if err != nil {
		return err
	}

	if len(resources.ResourceTagMappingList) > 0 {
		return fmt.Errorf(
			"VPC can not be created in this blueprint as "+
				"VPC resources exist with the flex VPC name: %s",
			core.StringValue(vpcName),
		)
	}

	return nil
}

func (l *vpcResourceActions) createFlexVPCResources(
	ctx context.Context,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
	region string,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	// Get Bluelink system tags for provenance tracking
	bluelinkTags := getBluelinkTagsAsEC2Tags(input)

	createVPCInput, vpcName, err := createVPCInputFromSpec(resourceSpecData, bluelinkTags)
	if err != nil {
		return nil, err
	}
	createVPCOutput, err := service.CreateVpc(ctx, createVPCInput)
	if err != nil {
		return nil, err
	}

	err = waitForVPCAvailable(ctx, service, createVPCOutput.Vpc.VpcId)
	if err != nil {
		return nil, err
	}

	err = l.modifyVPCAttributes(ctx, service, createVPCOutput.Vpc.VpcId, resourceSpecData)
	if err != nil {
		return nil, err
	}

	vpcIPv6CIDRBlock, err := waitForVPCIPv6CIDRBlock(ctx, service, createVPCOutput.Vpc.VpcId)
	if err != nil {
		return nil, err
	}

	vpcCtx := &vpcContext{
		resourceSpecData: resourceSpecData,
		vpcName:          vpcName,
		tags:             createVPCInput.TagSpecifications[0].Tags,
		instanceID:       input.InstanceID,
		vpcID:            createVPCOutput.Vpc.VpcId,
		vpcCIDRBlock:     aws.ToString(createVPCInput.CidrBlock),
		vpcIPv6CIDRBlock: vpcIPv6CIDRBlock,
		region:           region,
	}
	subnets, err := l.createAndConfigureSubnets(
		ctx,
		service,
		vpcCtx,
	)
	if err != nil {
		return nil, err
	}

	routingInfo, err := l.setupRoutingForSubnets(ctx, service, subnets, vpcCtx)
	if err != nil {
		return nil, err
	}

	securityGroupOutput, err := l.createInitialSecurityGroup(ctx, service, vpcCtx)
	if err != nil {
		return nil, err
	}

	namedGroupIDs, err := l.createNamedSecurityGroups(ctx, service, vpcCtx)
	if err != nil {
		return nil, err
	}

	networkACL, err := l.createInitialNACL(ctx, service, subnets, vpcCtx)
	if err != nil {
		return nil, err
	}

	specSubnets := toSpecComputedSubnets(subnets)
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.vpcId": core.MappingNodeFromString(
				aws.ToString(createVPCOutput.Vpc.VpcId),
			),
			"spec.subnets":          specSubnets,
			"spec.privateSubnetIds": specSubnetIdsByTier(specSubnets, "private"),
			"spec.publicSubnetIds":  specSubnetIdsByTier(specSubnets, "public"),
			"spec.routeTables":      toSpecComputedRouteTables(routingInfo.routeTables),
			"spec.securityGroupIds": toSpecComputedSecurityGroupIds(
				[]*ec2.CreateSecurityGroupOutput{securityGroupOutput},
			),
			"spec.securityGroupIdsByName": toSpecComputedSecurityGroupIdsByName(namedGroupIDs),
			"spec.networkAcls": toSpecComputedNetworkACLs(
				[]*networkACLInfo{networkACL},
			),
			"spec.gateways": toSpecComputedGateways(
				routingInfo.internetGateway,
				routingInfo.egressOnlyInternetGateway,
				routingInfo.natGateways,
			),
		},
	}, nil
}

func (l *vpcResourceActions) setupRoutingForSubnets(
	ctx context.Context,
	service ec2service.Service,
	subnetsConfig *subnets,
	vpcCtx *vpcContext,
) (*routingInfo, error) {
	routeTables := []*routeTableInfo{}

	igw, err := l.createInternetGateway(ctx, service, subnetsConfig, vpcCtx)
	if err != nil {
		return nil, err
	}

	eigw, err := l.createEgressOnlyInternetGateway(ctx, service, subnetsConfig, vpcCtx)
	if err != nil {
		return nil, err
	}

	publicSubnetsSorted := sortSubnetMap(subnetsConfig.public)
	for _, subnetInfo := range publicSubnetsSorted {
		routeTableInfo, err := l.setupPublicSubnetRouting(
			ctx,
			service,
			subnetInfo.subnet,
			subnetInfo.name,
			subnetsConfig,
			igw,
			vpcCtx,
		)
		if err != nil {
			return nil, err
		}
		routeTables = append(routeTables, routeTableInfo)
	}

	var natGateways []*natGatewayInfo
	privateSubnetsSorted := sortSubnetMap(subnetsConfig.private)
	for _, subnetInfo := range privateSubnetsSorted {
		routeTableInfo, err := l.setupPrivateSubnetRouting(
			ctx,
			service,
			subnetInfo.subnet,
			subnetInfo.name,
			subnetsConfig,
			eigw,
			vpcCtx,
		)
		if err != nil {
			return nil, err
		}
		routeTables = append(routeTables, routeTableInfo)
		natGateways = append(natGateways, routeTableInfo.natGateways...)
	}

	return &routingInfo{
		routeTables:               routeTables,
		internetGateway:           igw,
		egressOnlyInternetGateway: eigw,
		natGateways:               natGateways,
	}, nil
}

type routingInfo struct {
	routeTables               []*routeTableInfo
	natGateways               []*natGatewayInfo
	internetGateway           *types.InternetGateway
	egressOnlyInternetGateway *types.EgressOnlyInternetGateway
}

type routeTableInfo struct {
	id          string
	subnetIds   []string
	natGateways []*natGatewayInfo
}

func (l *vpcResourceActions) setupPublicSubnetRouting(
	ctx context.Context,
	service ec2service.Service,
	subnet *types.Subnet,
	subnetName string,
	subnetsConfig *subnets,
	igw *types.InternetGateway,
	vpcCtx *vpcContext,
) (*routeTableInfo, error) {
	subnetRoute := subnetsConfig.publicRoutes[subnetName]
	routeTableOutput, err := service.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{
		VpcId: subnet.VpcId,
		TagSpecifications: createRouteTableTagSpecifications(
			tagsWithRouteTable(vpcCtx.tags, vpcCtx.vpcName, subnetName),
		),
	})
	if err != nil {
		return nil, err
	}

	_, err = service.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{
		RouteTableId: routeTableOutput.RouteTable.RouteTableId,
		SubnetId:     subnet.SubnetId,
	})
	if err != nil {
		return nil, err
	}

	if subnetRoute.connectsToInternet {
		_, err = service.CreateRoute(ctx, &ec2.CreateRouteInput{
			RouteTableId:         routeTableOutput.RouteTable.RouteTableId,
			GatewayId:            igw.InternetGatewayId,
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
		})
		if err != nil {
			return nil, err
		}

		_, err = service.CreateRoute(ctx, &ec2.CreateRouteInput{
			RouteTableId:             routeTableOutput.RouteTable.RouteTableId,
			GatewayId:                igw.InternetGatewayId,
			DestinationIpv6CidrBlock: aws.String("::/0"),
		})
		if err != nil {
			return nil, err
		}
	}

	return &routeTableInfo{
		id:        aws.ToString(routeTableOutput.RouteTable.RouteTableId),
		subnetIds: []string{aws.ToString(subnet.SubnetId)},
	}, nil
}

type natGatewayInfo struct {
	id                 string
	elasticIPID        string
	inPublicSubnetID   string
	forPrivateSubnetID string
}

func (l *vpcResourceActions) setupPrivateSubnetRouting(
	ctx context.Context,
	service ec2service.Service,
	subnet *types.Subnet,
	subnetName string,
	subnetsConfig *subnets,
	eigw *types.EgressOnlyInternetGateway,
	vpcCtx *vpcContext,
) (*routeTableInfo, error) {
	natGateways := []*natGatewayInfo{}

	subnetRoute := subnetsConfig.privateRoutes[subnetName]
	publicSubnet := subnetsConfig.public[subnetRoute.connectViaPublicSubnet]
	routeTableOutput, err := service.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{
		VpcId: subnet.VpcId,
		TagSpecifications: createRouteTableTagSpecifications(
			tagsWithRouteTable(vpcCtx.tags, vpcCtx.vpcName, subnetName),
		),
	})
	if err != nil {
		return nil, err
	}

	_, err = service.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{
		RouteTableId: routeTableOutput.RouteTable.RouteTableId,
		SubnetId:     subnet.SubnetId,
	})
	if err != nil {
		return nil, err
	}

	if subnetRoute.connectsToInternet {
		elasticIPOutput, err := service.AllocateAddress(ctx, &ec2.AllocateAddressInput{
			Domain: types.DomainTypeVpc,
			TagSpecifications: createElasticIPTagSpecifications(
				tagsWithElasticIP(vpcCtx.tags, vpcCtx.vpcName, subnetName),
			),
		})
		if err != nil {
			return nil, err
		}

		natGatewayOutput, err := service.CreateNatGateway(ctx, &ec2.CreateNatGatewayInput{
			SubnetId:     publicSubnet.SubnetId,
			AllocationId: elasticIPOutput.AllocationId,
			TagSpecifications: createNatGatewayTagSpecifications(
				tagsWithNatGateway(vpcCtx.tags, vpcCtx.vpcName, subnetName),
			),
		})
		if err != nil {
			return nil, err
		}

		natGateways = append(natGateways, &natGatewayInfo{
			id:                 aws.ToString(natGatewayOutput.NatGateway.NatGatewayId),
			elasticIPID:        aws.ToString(elasticIPOutput.AllocationId),
			inPublicSubnetID:   aws.ToString(publicSubnet.SubnetId),
			forPrivateSubnetID: aws.ToString(subnet.SubnetId),
		})

		err = waitForNATGatewayAvailable(ctx, service, natGatewayOutput.NatGateway.NatGatewayId)
		if err != nil {
			return nil, err
		}

		_, err = service.CreateRoute(ctx, &ec2.CreateRouteInput{
			RouteTableId:         routeTableOutput.RouteTable.RouteTableId,
			NatGatewayId:         natGatewayOutput.NatGateway.NatGatewayId,
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
		})
		if err != nil {
			return nil, err
		}

		if eigw == nil {
			return nil, fmt.Errorf(
				"no egress-only internet gateway was created for subnet %q, which routes to the internet",
				subnetName,
			)
		}

		// Outbound IPv6 goes through the egress-only gateway rather than the internet
		// gateway, so the subnet stays private in both address families.
		_, err = service.CreateRoute(ctx, &ec2.CreateRouteInput{
			RouteTableId:                routeTableOutput.RouteTable.RouteTableId,
			EgressOnlyInternetGatewayId: eigw.EgressOnlyInternetGatewayId,
			DestinationIpv6CidrBlock:    aws.String("::/0"),
		})
		if err != nil {
			return nil, err
		}
	}

	return &routeTableInfo{
		id:          aws.ToString(routeTableOutput.RouteTable.RouteTableId),
		subnetIds:   []string{aws.ToString(subnet.SubnetId)},
		natGateways: natGateways,
	}, nil
}

// IPv6 has no NAT: the equivalent of a NAT gateway for outbound-only IPv6 is an
// egress-only internet gateway. Routing ::/0 from a private subnet to the ordinary
// internet gateway would work for outbound traffic but an internet gateway is
// bidirectional for IPv6, leaving anything in the subnet that holds an IPv6 address
// reachable from the internet.
//
// One gateway serves every private subnet in the VPC, so it is created once here
// rather than per subnet as NAT gateways are.
func (l *vpcResourceActions) createEgressOnlyInternetGateway(
	ctx context.Context,
	service ec2service.Service,
	subnetsConfig *subnets,
	vpcCtx *vpcContext,
) (*types.EgressOnlyInternetGateway, error) {
	if !anyPrivateSubnetConnectsToInternet(subnetsConfig) {
		return nil, nil
	}

	output, err := service.CreateEgressOnlyInternetGateway(ctx, &ec2.CreateEgressOnlyInternetGatewayInput{
		VpcId: vpcCtx.vpcID,
		TagSpecifications: createEgressOnlyInternetGatewayTagSpecifications(
			tagsWithEgressOnlyInternetGateway(vpcCtx.tags, vpcCtx.vpcName),
		),
	})
	if err != nil {
		return nil, err
	}

	if output == nil || output.EgressOnlyInternetGateway == nil {
		return nil, errors.New(
			"no egress-only internet gateway was returned when creating one for the VPC",
		)
	}

	return output.EgressOnlyInternetGateway, nil
}

func anyPrivateSubnetConnectsToInternet(subnetsConfig *subnets) bool {
	for _, route := range subnetsConfig.privateRoutes {
		if route.connectsToInternet {
			return true
		}
	}

	return false
}

func (l *vpcResourceActions) createInternetGateway(
	ctx context.Context,
	service ec2service.Service,
	subnetsConfig *subnets,
	vpcCtx *vpcContext,
) (*types.InternetGateway, error) {
	if len(subnetsConfig.public) > 0 {
		igwOutput, err := service.CreateInternetGateway(ctx, &ec2.CreateInternetGatewayInput{
			TagSpecifications: createInternetGatewayTagSpecifications(
				tagsWithInternetGateway(vpcCtx.tags, vpcCtx.vpcName),
			),
		})
		if err != nil {
			return nil, err
		}

		_, err = service.AttachInternetGateway(ctx, &ec2.AttachInternetGatewayInput{
			InternetGatewayId: igwOutput.InternetGateway.InternetGatewayId,
			VpcId:             vpcCtx.vpcID,
		})
		if err != nil {
			return nil, err
		}

		return igwOutput.InternetGateway, nil
	}

	return nil, nil
}

func (l *vpcResourceActions) modifyVPCAttributes(
	ctx context.Context,
	service ec2service.Service,
	vpcID *string,
	resourceSpecData *core.MappingNode,
) error {
	enableDNSSupport, hasEnableDNSSupport := pluginutils.GetValueByPath(
		"$.enableDNSSupport",
		resourceSpecData,
	)
	if hasEnableDNSSupport && !core.BoolValue(enableDNSSupport) {
		// DNS support is enabled by default, we only need to modify the attribute if it is set to false.
		_, err := service.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
			VpcId: vpcID,
			EnableDnsSupport: &types.AttributeBooleanValue{
				Value: aws.Bool(false),
			},
		})
		if err != nil {
			return err
		}
	}

	enableDNSHostnames, hasEnableDNSHostnames := pluginutils.GetValueByPath(
		"$.enableDNSHostnames",
		resourceSpecData,
	)
	if hasEnableDNSHostnames && core.BoolValue(enableDNSHostnames) {
		// DNS hostnames are disabled by default, we only need to modify the attribute if it is set to true.
		_, err := service.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
			VpcId: vpcID,
			EnableDnsHostnames: &types.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *vpcResourceActions) createAndConfigureSubnets(
	ctx context.Context,
	service ec2service.Service,
	vpcCtx *vpcContext,
) (*subnets, error) {
	subnets, err := l.createSubnets(ctx, service, vpcCtx)
	if err != nil {
		return nil, err
	}

	for _, subnet := range subnets.public {
		err := l.modifyPublicSubnetAttributes(ctx, service, subnet.SubnetId)
		if err != nil {
			return nil, err
		}
	}

	for _, subnet := range subnets.private {
		err := l.setAssignIpv6AddressOnCreation(ctx, service, subnet.SubnetId)
		if err != nil {
			return nil, err
		}
	}

	return subnets, nil
}

func (l *vpcResourceActions) modifyPublicSubnetAttributes(
	ctx context.Context,
	service ec2service.Service,
	subnetID *string,
) error {
	_, err := service.ModifySubnetAttribute(ctx, &ec2.ModifySubnetAttributeInput{
		SubnetId: subnetID,
		MapPublicIpOnLaunch: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	})
	if err != nil {
		return err
	}

	return l.setAssignIpv6AddressOnCreation(ctx, service, subnetID)
}

func (l *vpcResourceActions) setAssignIpv6AddressOnCreation(
	ctx context.Context,
	service ec2service.Service,
	subnetID *string,
) error {
	_, err := service.ModifySubnetAttribute(ctx, &ec2.ModifySubnetAttributeInput{
		SubnetId: subnetID,
		AssignIpv6AddressOnCreation: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	})
	return err
}

type subnets struct {
	private       map[string]*types.Subnet
	public        map[string]*types.Subnet
	privateRoutes map[string]*subnetRoute
	publicRoutes  map[string]*subnetRoute
}

func (l *vpcResourceActions) createSubnets(
	ctx context.Context,
	service ec2service.Service,
	vpcCtx *vpcContext,
) (*subnets, error) {
	regionAZs, err := l.getRegionAZs(ctx, service, vpcCtx.region)
	if err != nil {
		return nil, err
	}

	presetConfig, err := getPresetConfig(vpcContextWithRegionAZs(vpcCtx, regionAZs))
	if err != nil {
		return nil, err
	}

	privateSubnets := map[string]*types.Subnet{}
	for _, presetSubnet := range presetConfig.privateSubnets {
		output, err := service.CreateSubnet(ctx, presetSubnet.input)
		if err != nil {
			return nil, err
		}
		privateSubnets[presetSubnet.name] = output.Subnet
	}

	publicSubnets := map[string]*types.Subnet{}
	for _, presetSubnet := range presetConfig.publicSubnets {
		output, err := service.CreateSubnet(ctx, presetSubnet.input)
		if err != nil {
			return nil, err
		}
		publicSubnets[presetSubnet.name] = output.Subnet
	}

	return &subnets{
		private:       privateSubnets,
		public:        publicSubnets,
		privateRoutes: presetConfig.privateRoutes,
		publicRoutes:  presetConfig.publicRoutes,
	}, nil
}

func (l *vpcResourceActions) getRegionAZs(
	ctx context.Context,
	service ec2service.Service,
	region string,
) ([]string, error) {
	azsOutput, err := service.DescribeAvailabilityZones(ctx, &ec2.DescribeAvailabilityZonesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("region-name"),
				Values: []string{region},
			},
			{
				Name:   aws.String("zone-type"),
				Values: []string{"availability-zone"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	azs := []string{}
	for _, az := range azsOutput.AvailabilityZones {
		azs = append(azs, aws.ToString(az.ZoneName))
	}

	return azs, nil
}

func (l *vpcResourceActions) createInitialSecurityGroup(
	ctx context.Context,
	service ec2service.Service,
	vpcCtx *vpcContext,
) (*ec2.CreateSecurityGroupOutput, error) {
	securityGroupOutput, err := service.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(vpcCtx.vpcName),
		Description: aws.String("Security group for " + vpcCtx.vpcName),
		VpcId:       vpcCtx.vpcID,
		TagSpecifications: createSecurityGroupTagSpecifications(
			tagsWithSecurityGroup(vpcCtx.tags, vpcCtx.vpcName),
		),
	})
	if err != nil {
		return nil, err
	}

	// Links must activate egress rules to allow traffic to the internet
	// to fine-grained targets where possible.
	_, err = service.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
		GroupId: securityGroupOutput.GroupId,
		IpPermissions: []types.IpPermission{
			{
				// All protocols.
				IpProtocol: aws.String("-1"),
				IpRanges: []types.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
				Ipv6Ranges: []types.Ipv6Range{
					{
						CidrIpv6: aws.String("::/0"),
					},
				},
			},
		},
	})
	return securityGroupOutput, err
}

func (l *vpcResourceActions) createInitialNACL(
	ctx context.Context,
	service ec2service.Service,
	subnets *subnets,
	vpcCtx *vpcContext,
) (*networkACLInfo, error) {
	networkACL := &networkACLInfo{}
	if len(subnets.public) > 0 && len(subnets.private) == 0 {
		// Only create a deny-all NACL for public-only VPCs.
		networkAclOutput, err := service.CreateNetworkAcl(ctx, &ec2.CreateNetworkAclInput{
			VpcId: vpcCtx.vpcID,
			TagSpecifications: createNetworkAclTagSpecifications(
				tagsWithNetworkAcl(vpcCtx.tags, vpcCtx.vpcName),
			),
		})
		if err != nil {
			return nil, err
		}

		networkACL.id = aws.ToString(networkAclOutput.NetworkAcl.NetworkAclId)

		networkACLSubnets, err := l.associateNetworkACLWithSubnets(
			ctx,
			service,
			networkAclOutput.NetworkAcl,
			subnets,
		)
		if err != nil {
			return nil, err
		}

		networkACL.subnetIDs = networkACLSubnets

		// deny-all IPv4 ingress rule.
		_, err = service.CreateNetworkAclEntry(ctx, &ec2.CreateNetworkAclEntryInput{
			NetworkAclId: networkAclOutput.NetworkAcl.NetworkAclId,
			Egress:       aws.Bool(false),
			RuleNumber:   aws.Int32(100),
			Protocol:     aws.String("-1"),
			RuleAction:   types.RuleActionDeny,
			CidrBlock:    aws.String("0.0.0.0/0"),
		})
		if err != nil {
			return nil, err
		}

		// deny-all IPv6 ingress rule.
		_, err = service.CreateNetworkAclEntry(ctx, &ec2.CreateNetworkAclEntryInput{
			NetworkAclId:  networkAclOutput.NetworkAcl.NetworkAclId,
			Egress:        aws.Bool(false),
			RuleNumber:    aws.Int32(101),
			Protocol:      aws.String("-1"),
			RuleAction:    types.RuleActionDeny,
			Ipv6CidrBlock: aws.String("::/0"),
		})
		if err != nil {
			return nil, err
		}

		// deny-all IPv4 egress rule.
		_, err = service.CreateNetworkAclEntry(ctx, &ec2.CreateNetworkAclEntryInput{
			NetworkAclId: networkAclOutput.NetworkAcl.NetworkAclId,
			Egress:       aws.Bool(true),
			RuleNumber:   aws.Int32(100),
			Protocol:     aws.String("-1"),
			RuleAction:   types.RuleActionDeny,
			CidrBlock:    aws.String("0.0.0.0/0"),
		})
		if err != nil {
			return nil, err
		}

		// deny-all IPv6 egress rule.
		_, err = service.CreateNetworkAclEntry(ctx, &ec2.CreateNetworkAclEntryInput{
			NetworkAclId:  networkAclOutput.NetworkAcl.NetworkAclId,
			Egress:        aws.Bool(true),
			RuleNumber:    aws.Int32(101),
			Protocol:      aws.String("-1"),
			RuleAction:    types.RuleActionDeny,
			Ipv6CidrBlock: aws.String("::/0"),
		})
		if err != nil {
			return nil, err
		}
	}

	return networkACL, nil
}

func (l *vpcResourceActions) associateNetworkACLWithSubnets(
	ctx context.Context,
	service ec2service.Service,
	networkAcl *types.NetworkAcl,
	subnets *subnets,
) ([]string, error) {
	subnetIDs := []string{}

	for _, subnet := range subnets.public {
		subnetIDs = append(subnetIDs, aws.ToString(subnet.SubnetId))
		// A subnet is always associated with a default network ACL,
		// we need to find the association so we can replace it with the
		// new network ACL.
		networkAclsOutput, err := service.DescribeNetworkAcls(
			ctx,
			&ec2.DescribeNetworkAclsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("association.subnet-id"),
						Values: []string{aws.ToString(subnet.SubnetId)},
					},
				},
			},
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to describe NACLs for subnet %s: %w",
				*subnet.SubnetId,
				err,
			)
		}

		if len(networkAclsOutput.NetworkAcls) == 0 {
			return nil, fmt.Errorf(
				"no NACL found for subnet %s",
				*subnet.SubnetId,
			)
		}

		if len(networkAclsOutput.NetworkAcls[0].Associations) == 0 {
			return nil, fmt.Errorf(
				"no NACL association found for subnet %s",
				*subnet.SubnetId,
			)
		}

		_, err = service.ReplaceNetworkAclAssociation(
			ctx,
			&ec2.ReplaceNetworkAclAssociationInput{
				NetworkAclId:  networkAcl.NetworkAclId,
				AssociationId: networkAclsOutput.NetworkAcls[0].Associations[0].NetworkAclAssociationId,
			},
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to replace NACL association for subnet %s: %w",
				*subnet.SubnetId,
				err,
			)
		}
	}

	return subnetIDs, nil
}

func waitForVPCAvailable(
	ctx context.Context,
	service ec2service.Service,
	vpcID *string,
) error {
	waiter := ec2.NewVpcAvailableWaiter(service)
	return waiter.Wait(
		ctx,
		&ec2.DescribeVpcsInput{
			VpcIds: []string{aws.ToString(vpcID)},
		},
		vpcWaitTimeout,
	)
}

// An Amazon-provided IPv6 block is assigned asynchronously. The CreateVpc
// response carries the association in an "associating" state with an empty CIDR,
// and the VpcAvailable waiter reports on the VPC's own state rather than the
// association's. Subnet CIDRs are derived from this block, so the VPC is re-read
// until the association completes instead of trusting the create response.
func waitForVPCIPv6CIDRBlock(
	ctx context.Context,
	service ec2service.Service,
	vpcID *string,
) (string, error) {
	deadline := time.Now().Add(vpcIPv6CIDRWaitTimeout)
	for {
		vpcsOutput, err := service.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
			VpcIds: []string{aws.ToString(vpcID)},
		})
		if err != nil {
			return "", err
		}

		if vpcsOutput != nil && len(vpcsOutput.Vpcs) > 0 {
			if cidrBlock := associatedIPv6CIDRBlock(vpcsOutput.Vpcs[0]); cidrBlock != "" {
				return cidrBlock, nil
			}
		}

		if time.Now().After(deadline) {
			return "", fmt.Errorf(
				"timed out waiting for an IPv6 CIDR block to be associated with VPC %q",
				aws.ToString(vpcID),
			)
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(vpcIPv6CIDRPollInterval):
		}
	}
}

func associatedIPv6CIDRBlock(vpc types.Vpc) string {
	for _, association := range vpc.Ipv6CidrBlockAssociationSet {
		if association.Ipv6CidrBlockState == nil ||
			association.Ipv6CidrBlockState.State != types.VpcCidrBlockStateCodeAssociated {
			continue
		}
		if cidrBlock := aws.ToString(association.Ipv6CidrBlock); cidrBlock != "" {
			return cidrBlock
		}
	}

	return ""
}

func waitForNATGatewayAvailable(
	ctx context.Context,
	service ec2service.Service,
	natGatewayID *string,
) error {
	waiter := ec2.NewNatGatewayAvailableWaiter(service)
	return waiter.Wait(
		ctx,
		&ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []string{aws.ToString(natGatewayID)},
		},
		natGatewayWaitTimeout,
	)
}

func createVPCInputFromSpec(
	resourceSpecData *core.MappingNode,
	bluelinkTags []types.Tag,
) (*ec2.CreateVpcInput, string, error) {
	cidrBlock, hasCidrBlock := pluginutils.GetValueByPath("$.cidrBlock", resourceSpecData)
	if !hasCidrBlock {
		return nil, "", errors.New("cidrBlock is required to create a flex VPC")
	}

	name, hasName := pluginutils.GetValueByPath("$.name", resourceSpecData)
	if !hasName {
		return nil, "", errors.New("name is required to create a flex VPC")
	}

	tags, _ := pluginutils.GetValueByPath("$.tags", resourceSpecData)
	if tags == nil {
		tags = core.MappingNodeFields()
	}

	return &ec2.CreateVpcInput{
		CidrBlock:                   aws.String(core.StringValue(cidrBlock)),
		AmazonProvidedIpv6CidrBlock: aws.Bool(true),
		TagSpecifications:           vpcElementTagSpecifications(core.StringValue(name), tags, bluelinkTags),
	}, core.StringValue(name), nil
}

// getBluelinkTagsAsEC2Tags extracts Bluelink system tags from the deploy input
// and converts them to EC2 tag format for resource provenance tracking.
func getBluelinkTagsAsEC2Tags(input *provider.ResourceDeployInput) []types.Tag {
	bluelinkTags := providerv1.GetBluelinkTags(input)
	if bluelinkTags == nil {
		return nil
	}

	tagsMap := providerv1.ToMap(bluelinkTags)
	ec2Tags := make([]types.Tag, 0, len(tagsMap))
	for key, value := range tagsMap {
		ec2Tags = append(ec2Tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	return ec2Tags
}

func vpcElementTagSpecifications(
	name string,
	tags *core.MappingNode,
	bluelinkTags []types.Tag,
) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeVpc,
			Tags:         vpcElementTags(name, tags, bluelinkTags),
		},
	}
}

func vpcElementTags(name string, tags *core.MappingNode, bluelinkTags []types.Tag) []types.Tag {
	// Start with Bluelink system tags for provenance tracking
	tagList := append([]types.Tag{}, bluelinkTags...)

	// Add flex VPC identification tags
	tagList = append(tagList,
		types.Tag{
			Key:   aws.String(TagFlexVPCName),
			Value: aws.String(name),
		},
		types.Tag{
			Key:   aws.String(TagFlexVPCResource),
			Value: aws.String("true"),
		},
	)

	// Add user-provided tags (these take precedence on key conflicts)
	if core.IsObjectMappingNode(tags) {
		for key, value := range tags.Fields {
			tagList = append(tagList, types.Tag{
				Key:   aws.String(key),
				Value: aws.String(core.StringValue(value)),
			})
		}
	}

	return tagList
}

type presetConfig struct {
	publicSubnets  []*presetSubnet
	privateSubnets []*presetSubnet
	// A mapping of public subnet to routing config.
	publicRoutes map[string]*subnetRoute
	// A mapping of private subnet to routing config.
	privateRoutes map[string]*subnetRoute
}

type presetSubnet struct {
	name  string
	input *ec2.CreateSubnetInput
}

type subnetRoute struct {
	subnetType         string
	connectsToInternet bool
	// The name of the public subnet in the same AZ that
	// will contain the NAT gateway for a private subnet.
	// This is only used for private subnets.
	connectViaPublicSubnet string
}

func getPresetConfig(vpcCtx *vpcContext) (*presetConfig, error) {
	preset, hasPreset := pluginutils.GetValueByPath("$.preset", vpcCtx.resourceSpecData)
	if !hasPreset {
		return nil, nil
	}

	presetStr := core.StringValue(preset)
	switch presetStr {
	case "standard":
		return standardPresetConfig(vpcCtx)
	case "public":
		return publicPresetConfig(vpcCtx)
	case "isolated":
		return isolatedPresetConfig(vpcCtx)
	case "light":
		return lightPresetConfig(vpcCtx)
	case "light-public":
		return lightPublicPresetConfig(vpcCtx)
	default:
		return nil, fmt.Errorf("invalid preset: %s", presetStr)
	}
}

func standardPresetConfig(
	vpcCtx *vpcContext,
) (*presetConfig, error) {
	if len(vpcCtx.regionAZs) < 3 {
		return nil, fmt.Errorf(
			"region %s must have at least 3 availability zones for the standard preset",
			vpcCtx.region,
		)
	}

	ipv4SubnetCIDRBlocks, err := utils.CalculateIPv4SubnetCIDRBlocks(vpcCtx.vpcCIDRBlock, 6)
	if err != nil {
		return nil, err
	}

	ipv6SubnetCIDRBlocks, err := utils.CalculateIPv6SubnetCIDRBlocks(vpcCtx.vpcIPv6CIDRBlock, 6, 64)
	if err != nil {
		return nil, err
	}

	// 6 subnets across 3 availability zones.
	return &presetConfig{
		privateSubnets: []*presetSubnet{
			{
				name: "private-az-1",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[0]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[0]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPrivateSubnet(vpcCtx.tags, vpcCtx.vpcName, "private-az-1"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[0]),
				},
			},
			{
				name: "private-az-2",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[1]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[1]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPrivateSubnet(vpcCtx.tags, vpcCtx.vpcName, "private-az-2"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[1]),
				},
			},
			{
				name: "private-az-3",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[2]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[2]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPrivateSubnet(vpcCtx.tags, vpcCtx.vpcName, "private-az-3"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[2]),
				},
			},
		},
		privateRoutes: map[string]*subnetRoute{
			"private-az-1": {
				subnetType:         "private",
				connectsToInternet: true,
				// Route traffic through the public subnet in the same AZ
				// that will contain the nat gateway.
				connectViaPublicSubnet: "public-az-1",
			},
			"private-az-2": {
				subnetType:             "private",
				connectsToInternet:     true,
				connectViaPublicSubnet: "public-az-2",
			},
			"private-az-3": {
				subnetType:             "private",
				connectsToInternet:     true,
				connectViaPublicSubnet: "public-az-3",
			},
		},
		publicSubnets: []*presetSubnet{
			{
				name: "public-az-1",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[3]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[3]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPublicSubnet(vpcCtx.tags, vpcCtx.vpcName, "public-az-1"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[0]),
				},
			},
			{
				name: "public-az-2",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[4]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[4]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPublicSubnet(vpcCtx.tags, vpcCtx.vpcName, "public-az-2"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[1]),
				},
			},
			{
				name: "public-az-3",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[5]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[5]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPublicSubnet(vpcCtx.tags, vpcCtx.vpcName, "public-az-3"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[2]),
				},
			},
		},
		publicRoutes: map[string]*subnetRoute{
			"public-az-1": {
				subnetType:         "public",
				connectsToInternet: true,
			},
			"public-az-2": {
				subnetType:         "public",
				connectsToInternet: true,
			},
			"public-az-3": {
				subnetType:         "public",
				connectsToInternet: true,
			},
		},
	}, nil
}

func publicPresetConfig(
	vpcCtx *vpcContext,
) (*presetConfig, error) {
	if len(vpcCtx.regionAZs) < 3 {
		return nil, fmt.Errorf(
			"region %s must have at least 3 availability zones for the public preset",
			vpcCtx.region,
		)
	}

	ipv4SubnetCIDRBlocks, err := utils.CalculateIPv4SubnetCIDRBlocks(vpcCtx.vpcCIDRBlock, 3)
	if err != nil {
		return nil, err
	}

	ipv6SubnetCIDRBlocks, err := utils.CalculateIPv6SubnetCIDRBlocks(vpcCtx.vpcIPv6CIDRBlock, 3, 64)
	if err != nil {
		return nil, err
	}

	// 3 public subnets are created across 3 availability zones.
	return &presetConfig{
		publicSubnets: []*presetSubnet{
			{
				name: "public-az-1",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[0]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[0]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPublicSubnet(vpcCtx.tags, vpcCtx.vpcName, "public-az-1"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[0]),
				},
			},
			{
				name: "public-az-2",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[1]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[1]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPublicSubnet(vpcCtx.tags, vpcCtx.vpcName, "public-az-2"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[1]),
				},
			},
			{
				name: "public-az-3",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[2]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[2]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPublicSubnet(vpcCtx.tags, vpcCtx.vpcName, "public-az-3"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[2]),
				},
			},
		},
		publicRoutes: map[string]*subnetRoute{
			"public-az-1": {
				subnetType:         "public",
				connectsToInternet: true,
			},
			"public-az-2": {
				subnetType:         "public",
				connectsToInternet: true,
			},
			"public-az-3": {
				subnetType:         "public",
				connectsToInternet: true,
			},
		},
	}, nil
}

func isolatedPresetConfig(
	vpcCtx *vpcContext,
) (*presetConfig, error) {
	if len(vpcCtx.regionAZs) < 3 {
		return nil, fmt.Errorf(
			"region %s must have at least 3 availability zones for the isolated preset",
			vpcCtx.region,
		)
	}

	ipv4SubnetCIDRBlocks, err := utils.CalculateIPv4SubnetCIDRBlocks(vpcCtx.vpcCIDRBlock, 3)
	if err != nil {
		return nil, err
	}

	ipv6SubnetCIDRBlocks, err := utils.CalculateIPv6SubnetCIDRBlocks(vpcCtx.vpcIPv6CIDRBlock, 3, 64)
	if err != nil {
		return nil, err
	}

	// 3 isolated subnets are created across 3 availability zones.
	return &presetConfig{
		privateSubnets: []*presetSubnet{
			{
				name: "private-az-1",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[0]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[0]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPrivateSubnet(vpcCtx.tags, vpcCtx.vpcName, "private-az-1"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[0]),
				},
			},
			{
				name: "private-az-2",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[1]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[1]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPrivateSubnet(vpcCtx.tags, vpcCtx.vpcName, "private-az-2"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[1]),
				},
			},
			{
				name: "private-az-3",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[2]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[2]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPrivateSubnet(vpcCtx.tags, vpcCtx.vpcName, "private-az-3"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[2]),
				},
			},
		},
		privateRoutes: map[string]*subnetRoute{
			"private-az-1": {
				subnetType:         "private",
				connectsToInternet: false,
			},
			"private-az-2": {
				subnetType:         "private",
				connectsToInternet: false,
			},
			"private-az-3": {
				subnetType:         "private",
				connectsToInternet: false,
			},
		},
	}, nil
}

func lightPresetConfig(
	vpcCtx *vpcContext,
) (*presetConfig, error) {
	if len(vpcCtx.regionAZs) < 1 {
		return nil, fmt.Errorf(
			"region %s must have at least 1 availability zone for the light preset",
			vpcCtx.region,
		)
	}

	ipv4SubnetCIDRBlocks, err := utils.CalculateIPv4SubnetCIDRBlocks(vpcCtx.vpcCIDRBlock, 2)
	if err != nil {
		return nil, err
	}

	ipv6SubnetCIDRBlocks, err := utils.CalculateIPv6SubnetCIDRBlocks(vpcCtx.vpcIPv6CIDRBlock, 2, 64)
	if err != nil {
		return nil, err
	}

	// 1 private subnet and 1 public subnet in a single availability zone.
	return &presetConfig{
		privateSubnets: []*presetSubnet{
			{
				name: "private-az-1",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[0]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[0]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPrivateSubnet(vpcCtx.tags, vpcCtx.vpcName, "private-az-1"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[0]),
				},
			},
		},
		privateRoutes: map[string]*subnetRoute{
			"private-az-1": {
				subnetType:             "private",
				connectsToInternet:     true,
				connectViaPublicSubnet: "public-az-1",
			},
		},
		publicSubnets: []*presetSubnet{
			{
				name: "public-az-1",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[1]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[1]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPublicSubnet(vpcCtx.tags, vpcCtx.vpcName, "public-az-1"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[0]),
				},
			},
		},
		publicRoutes: map[string]*subnetRoute{
			"public-az-1": {
				subnetType:         "public",
				connectsToInternet: true,
			},
		},
	}, nil
}

func lightPublicPresetConfig(
	vpcCtx *vpcContext,
) (*presetConfig, error) {
	if len(vpcCtx.regionAZs) < 1 {
		return nil, fmt.Errorf(
			"region %s must have at least 1 availability zone for the light public preset",
			vpcCtx.region,
		)
	}

	ipv4SubnetCIDRBlocks, err := utils.CalculateIPv4SubnetCIDRBlocks(vpcCtx.vpcCIDRBlock, 1)
	if err != nil {
		return nil, err
	}

	ipv6SubnetCIDRBlocks, err := utils.CalculateIPv6SubnetCIDRBlocks(vpcCtx.vpcIPv6CIDRBlock, 1, 64)
	if err != nil {
		return nil, err
	}

	// 1 public subnet in a single availability zone.
	return &presetConfig{
		publicSubnets: []*presetSubnet{
			{
				name: "public-az-1",
				input: &ec2.CreateSubnetInput{
					VpcId:         vpcCtx.vpcID,
					CidrBlock:     aws.String(ipv4SubnetCIDRBlocks[0]),
					Ipv6CidrBlock: aws.String(ipv6SubnetCIDRBlocks[0]),
					TagSpecifications: createSubnetTagSpecifications(
						tagsWithPublicSubnet(vpcCtx.tags, vpcCtx.vpcName, "public-az-1"),
					),
					AvailabilityZone: aws.String(vpcCtx.regionAZs[0]),
				},
			},
		},
		publicRoutes: map[string]*subnetRoute{
			"public-az-1": {
				subnetType:         "public",
				connectsToInternet: true,
			},
		},
	}, nil
}

// The single constructor for the tags on any resource created as part of a flex
// VPC: the VPC's base tag set (Bluelink provenance and the user's own tags) with
// the flex VPC name, the flex VPC resource marker, and the tags identifying this
// particular resource upserted onto it.
//
// Every flex VPC resource must be tagged through this. The name and resource
// marker are applied here rather than by each caller because every lookup in
// GetExternalState filters on the name, so a resource missing it is invisible to
// the provider even though it exists in the account.
func flexVPCResourceTags(base []types.Tag, vpcName string, identifying ...types.Tag) []types.Tag {
	tags := append([]types.Tag{}, base...)
	tags = upsertTag(tags, TagFlexVPCName, vpcName)
	tags = upsertTag(tags, TagFlexVPCResource, "true")
	for _, tag := range identifying {
		tags = upsertTag(tags, aws.ToString(tag.Key), aws.ToString(tag.Value))
	}

	return tags
}

func flexVPCTag(key string, value string) types.Tag {
	return types.Tag{Key: aws.String(key), Value: aws.String(value)}
}

func upsertTag(tags []types.Tag, key string, value string) []types.Tag {
	for i := range tags {
		if aws.ToString(tags[i].Key) == key {
			tags[i].Value = aws.String(value)
			return tags
		}
	}

	return append(tags, types.Tag{Key: aws.String(key), Value: aws.String(value)})
}

func tagsWithPrivateSubnet(tags []types.Tag, vpcName string, subnetName string) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCSubnetType, "private"),
		flexVPCTag(TagFlexVPCSubnetName, subnetName),
	)
}

func tagsWithPublicSubnet(tags []types.Tag, vpcName string, subnetName string) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCSubnetType, "public"),
		flexVPCTag(TagFlexVPCSubnetName, subnetName),
	)
}

func tagsWithRouteTable(tags []types.Tag, vpcName string, subnetName string) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCRouteTable, "true"),
		// Each route table created for a flex VPC is associated with a specific subnet.
		flexVPCTag(TagFlexVPCSubnetName, subnetName),
	)
}

func tagsWithNatGateway(tags []types.Tag, vpcName string, subnetName string) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCNATGateway, "true"),
		flexVPCTag(TagFlexVPCSubnetName, subnetName),
	)
}

func tagsWithInternetGateway(tags []types.Tag, vpcName string) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCInternetGateway, "true"),
	)
}

func tagsWithEgressOnlyInternetGateway(tags []types.Tag, vpcName string) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCEgressOnlyInternetGateway, "true"),
	)
}

func tagsWithElasticIP(tags []types.Tag, vpcName string, subnetName string) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCElasticIP, "true"),
		flexVPCTag(TagFlexVPCSubnetName, subnetName),
	)
}

func tagsWithSecurityGroup(tags []types.Tag, vpcName string) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCSecurityGroup, "true"),
	)
}

func tagsWithNetworkAcl(tags []types.Tag, vpcName string) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCNetworkACL, "true"),
	)
}

func createSubnetTagSpecifications(tags []types.Tag) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeSubnet,
			Tags:         tags,
		},
	}
}

func createRouteTableTagSpecifications(tags []types.Tag) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeRouteTable,
			Tags:         tags,
		},
	}
}

func createInternetGatewayTagSpecifications(tags []types.Tag) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeInternetGateway,
			Tags:         tags,
		},
	}
}

func createEgressOnlyInternetGatewayTagSpecifications(tags []types.Tag) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeEgressOnlyInternetGateway,
			Tags:         tags,
		},
	}
}

func createNatGatewayTagSpecifications(tags []types.Tag) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeNatgateway,
			Tags:         tags,
		},
	}
}

func createElasticIPTagSpecifications(tags []types.Tag) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeElasticIp,
			Tags:         tags,
		},
	}
}

func createSecurityGroupTagSpecifications(tags []types.Tag) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeSecurityGroup,
			Tags:         tags,
		},
	}
}

func createNetworkAclTagSpecifications(tags []types.Tag) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeNetworkAcl,
			Tags:         tags,
		},
	}
}

type vpcContext struct {
	resourceSpecData *core.MappingNode
	vpcName          string
	tags             []types.Tag
	// The blueprint instance that owns any policy security groups created from
	// this context; a referenced VPC is shared, so groups must be attributable.
	instanceID       string
	vpcID            *string
	vpcCIDRBlock     string
	vpcIPv6CIDRBlock string
	region           string
	regionAZs        []string
}

func vpcContextWithRegionAZs(input *vpcContext, regionAZs []string) *vpcContext {
	return &vpcContext{
		resourceSpecData: input.resourceSpecData,
		vpcName:          input.vpcName,
		tags:             input.tags,
		instanceID:       input.instanceID,
		vpcID:            input.vpcID,
		vpcCIDRBlock:     input.vpcCIDRBlock,
		vpcIPv6CIDRBlock: input.vpcIPv6CIDRBlock,
		region:           input.region,
		regionAZs:        regionAZs,
	}
}

func awsConfigMetaWithRegion(input *provider.ResourceDeployInput) map[string]*core.MappingNode {
	resourceSpecData := pluginutils.GetResolvedResourceSpecData(input.Changes)
	region, hasRegion := pluginutils.GetValueByPath("$.region", resourceSpecData)
	if !hasRegion {
		return map[string]*core.MappingNode{}
	}

	return map[string]*core.MappingNode{
		"region": region,
	}
}

func awsConfigMetaWithRegionGetExternalState(
	input *provider.ResourceGetExternalStateInput,
) map[string]*core.MappingNode {
	region, hasRegion := pluginutils.GetValueByPath("$.region", input.CurrentResourceSpec)
	if !hasRegion {
		return map[string]*core.MappingNode{}
	}

	return map[string]*core.MappingNode{
		"region": region,
	}
}

func toSpecComputedSubnets(subnets *subnets) *core.MappingNode {
	outputSubnets := core.MappingNodeFields()
	for subnetName, subnet := range subnets.public {
		outputSubnets.Fields[subnetName] = toSpecComputedSubnet(subnet, "public")
	}
	for subnetName, subnet := range subnets.private {
		outputSubnets.Fields[subnetName] = toSpecComputedSubnet(subnet, "private")
	}
	return outputSubnets
}

func toSpecComputedSubnet(subnet *types.Subnet, subnetType string) *core.MappingNode {
	computed := core.MappingNodeFields(
		"id",
		core.MappingNodeFromString(aws.ToString(subnet.SubnetId)),
		"availabilityZone",
		core.MappingNodeFromString(aws.ToString(subnet.AvailabilityZone)),
		"subnetType",
		core.MappingNodeFromString(subnetType),
	)

	if ipv6CIDRBlock := subnetIPv6CIDRBlock(subnet); ipv6CIDRBlock != "" {
		computed.Fields["ipv6CidrBlock"] = core.MappingNodeFromString(ipv6CIDRBlock)
	}

	return computed
}

// Whether a subnet is dual-stack decides whether a workload placed in it can be
// given outbound IPv6, so the association is surfaced rather than left for a caller
// to rediscover.
func subnetIPv6CIDRBlock(subnet *types.Subnet) string {
	for _, association := range subnet.Ipv6CidrBlockAssociationSet {
		if association.Ipv6CidrBlock != nil {
			return aws.ToString(association.Ipv6CidrBlock)
		}
	}

	return ""
}

// Derives a directly-referenceable list of subnet IDs in the
// given tier from the computed `subnets` map node, sorted by availability zone
// (then by ID) so the output is stable between plans.
func specSubnetIdsByTier(subnetsNode *core.MappingNode, tier string) *core.MappingNode {
	type subnetEntry struct {
		id               string
		availabilityZone string
	}

	entries := []subnetEntry{}
	if subnetsNode != nil {
		for _, subnet := range subnetsNode.Fields {
			typeNode, _ := pluginutils.GetValueByPath("$.subnetType", subnet)
			if core.StringValue(typeNode) != tier {
				continue
			}
			idNode, _ := pluginutils.GetValueByPath("$.id", subnet)
			azNode, _ := pluginutils.GetValueByPath("$.availabilityZone", subnet)
			if id := core.StringValue(idNode); id != "" {
				entries = append(entries, subnetEntry{
					id:               id,
					availabilityZone: core.StringValue(azNode),
				})
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].availabilityZone != entries[j].availabilityZone {
			return entries[i].availabilityZone < entries[j].availabilityZone
		}
		return entries[i].id < entries[j].id
	})

	items := []*core.MappingNode{}
	for _, entry := range entries {
		items = append(items, core.MappingNodeFromString(entry.id))
	}
	return core.MappingNodeItems(items...)
}

func toSpecComputedRouteTables(routeTables []*routeTableInfo) *core.MappingNode {
	specRouteTables := core.MappingNodeItems()
	for _, routeTable := range routeTables {
		specRouteTables.Items = append(specRouteTables.Items, core.MappingNodeFields(
			"id",
			core.MappingNodeFromString(routeTable.id),
			"subnetIds",
			core.MappingNodeFromStringSlice(routeTable.subnetIds),
		))
	}

	return specRouteTables
}

func toSpecComputedSecurityGroupIds(securityGroups []*ec2.CreateSecurityGroupOutput) *core.MappingNode {
	specSecurityGroups := core.MappingNodeItems()
	for _, securityGroup := range securityGroups {
		specSecurityGroups.Items = append(
			specSecurityGroups.Items,
			core.MappingNodeFromString(aws.ToString(securityGroup.GroupId)),
		)
	}
	return specSecurityGroups
}

type networkACLInfo struct {
	id        string
	subnetIDs []string
}

func toSpecComputedNetworkACLs(networkACLs []*networkACLInfo) *core.MappingNode {
	specNetworkACLs := core.MappingNodeItems()
	for _, networkACL := range networkACLs {
		if networkACL.id != "" {
			specNetworkACLs.Items = append(specNetworkACLs.Items, core.MappingNodeFields(
				"id",
				core.MappingNodeFromString(networkACL.id),
				"subnetIds",
				core.MappingNodeFromStringSlice(networkACL.subnetIDs),
			))
		}
	}
	return specNetworkACLs
}

func toSpecComputedGateways(
	internetGateway *types.InternetGateway,
	egressOnlyInternetGateway *types.EgressOnlyInternetGateway,
	natGateways []*natGatewayInfo,
) *core.MappingNode {
	gateways := core.MappingNodeFields(
		"natGateways",
		toSpecComputedNATGateways(natGateways),
	)

	// The isolated preset provisions no public subnets, so it has no internet gateway
	// and this is nil. Dereferencing it unconditionally made that preset impossible to
	// create at all.
	if internetGateway != nil {
		gateways.Fields["internetGatewayId"] = core.MappingNodeFromString(
			aws.ToString(internetGateway.InternetGatewayId),
		)
	}

	// Presets with no internet-connected private subnet have no egress-only gateway,
	// so the field is absent rather than empty.
	if egressOnlyInternetGateway != nil {
		gateways.Fields["egressOnlyInternetGatewayId"] = core.MappingNodeFromString(
			aws.ToString(egressOnlyInternetGateway.EgressOnlyInternetGatewayId),
		)
	}

	return gateways
}

func toSpecComputedNATGateways(natGateways []*natGatewayInfo) *core.MappingNode {
	specNATGateways := &core.MappingNode{
		Items: []*core.MappingNode{},
	}
	for _, natGateway := range natGateways {
		specNATGateways.Items = append(specNATGateways.Items, core.MappingNodeFields(
			"id",
			core.MappingNodeFromString(natGateway.id),
			"elasticIpId",
			core.MappingNodeFromString(natGateway.elasticIPID),
			"inPublicSubnetId",
			core.MappingNodeFromString(natGateway.inPublicSubnetID),
			"forPrivateSubnetId",
			core.MappingNodeFromString(natGateway.forPrivateSubnetID),
		))
	}
	return specNATGateways
}

func extractComputedFieldsFromExternalState(
	externalState *core.MappingNode,
) map[string]*core.MappingNode {
	vpcID, _ := pluginutils.GetValueByPath("$.vpcId", externalState)
	subnets, _ := pluginutils.GetValueByPath("$.subnets", externalState)
	privateSubnetIds, _ := pluginutils.GetValueByPath("$.privateSubnetIds", externalState)
	publicSubnetIds, _ := pluginutils.GetValueByPath("$.publicSubnetIds", externalState)
	routeTables, _ := pluginutils.GetValueByPath("$.routeTables", externalState)
	securityGroups, _ := pluginutils.GetValueByPath("$.securityGroupIds", externalState)
	securityGroupIdsByName, _ := pluginutils.GetValueByPath(
		"$.securityGroupIdsByName",
		externalState,
	)
	networkACLs, _ := pluginutils.GetValueByPath("$.networkAcls", externalState)
	gateways, _ := pluginutils.GetValueByPath("$.gateways", externalState)

	return map[string]*core.MappingNode{
		"spec.vpcId":                  vpcID,
		"spec.subnets":                subnets,
		"spec.privateSubnetIds":       privateSubnetIds,
		"spec.publicSubnetIds":        publicSubnetIds,
		"spec.routeTables":            routeTables,
		"spec.securityGroupIds":       securityGroups,
		"spec.securityGroupIdsByName": securityGroupIdsByName,
		"spec.networkAcls":            networkACLs,
		"spec.gateways":               gateways,
	}
}

type subnetWithName struct {
	name   string
	subnet *types.Subnet
}

func sortSubnetMap(subnetMap map[string]*types.Subnet) []*subnetWithName {
	subnetInfos := make([]*subnetWithName, 0, len(subnetMap))
	for name, subnet := range subnetMap {
		subnetInfos = append(subnetInfos, &subnetWithName{name: name, subnet: subnet})
	}
	sort.Slice(subnetInfos, func(i, j int) bool {
		return subnetInfos[i].name < subnetInfos[j].name
	})
	return subnetInfos
}
