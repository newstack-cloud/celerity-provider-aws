package flex

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *vpcResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	service, _, err := l.getEC2ServiceWithRegion(ctx, input.ProviderContext, nil)
	if err != nil {
		return err
	}

	resourceSpecData := input.ResourceState.SpecData
	vpcIDValue, hasVPCID := pluginutils.GetValueByPath("$.vpcId", resourceSpecData)
	if !hasVPCID {
		return errors.New("vpcId missing in flex VPC state")
	}
	vpcID := core.StringValue(vpcIDValue)

	mode, hasMode := pluginutils.GetValueByPath("$.mode", resourceSpecData)
	if !hasMode {
		return errors.New("mode missing in flex VPC state")
	}

	modeStr := core.StringValue(mode)
	if modeStr == "reference" {
		// The flex VPC resources shouldn't be deleted when in reference mode.
		return nil
	}

	networkACLIDs, err := l.disassociateNetworkACLsFromSubnets(ctx, vpcID, service, resourceSpecData)
	if err != nil {
		return err
	}

	err = l.deleteNetworkACLs(ctx, service, networkACLIDs)
	if err != nil {
		return err
	}

	err = l.deleteSecurityGroups(ctx, service, resourceSpecData)
	if err != nil {
		return err
	}

	err = l.deleteRoutes(ctx, vpcID, service, resourceSpecData)
	if err != nil {
		return err
	}

	err = l.deleteNATGatewaysAndReleaseElasticIPs(ctx, service, resourceSpecData)
	if err != nil {
		return err
	}

	err = l.detachAndDeleteInternetGateway(ctx, vpcID, service, resourceSpecData)
	if err != nil {
		return err
	}

	err = l.deleteSubnets(ctx, service, resourceSpecData)
	if err != nil {
		return err
	}

	return l.deleteVPC(ctx, vpcID, service)
}

func (l *vpcResourceActions) disassociateNetworkACLsFromSubnets(
	ctx context.Context,
	vpcID string,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) (
	[]string,
	error,
) {
	networkACLs, hasNetworkACLs := pluginutils.GetValueByPath("$.networkAcls", resourceSpecData)
	if !hasNetworkACLs || len(networkACLs.Items) == 0 {
		return []string{}, nil
	}

	defaultNACLOutput, err := service.DescribeNetworkAcls(ctx, &ec2.DescribeNetworkAclsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("default"),
				Values: []string{"true"},
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return []string{}, err
	}

	if len(defaultNACLOutput.NetworkAcls) == 0 {
		return []string{}, errors.New("no default network ACL found for VPC")
	}

	networkACLIDs := []string{}
	for _, networkACL := range networkACLs.Items {
		networkACLID, err := l.replaceNACLAssociationWithVPCDefault(
			ctx,
			service,
			networkACL,
			defaultNACLOutput.NetworkAcls[0].NetworkAclId,
		)
		if err != nil {
			return []string{}, err
		}
		if networkACLID != "" {
			networkACLIDs = append(networkACLIDs, networkACLID)
		}
	}

	return networkACLIDs, nil
}

func (l *vpcResourceActions) replaceNACLAssociationWithVPCDefault(
	ctx context.Context,
	service ec2service.Service,
	networkACL *core.MappingNode,
	defaultNetworkACLID *string,
) (string, error) {
	networkACLID, hasNetworkACLID := pluginutils.GetValueByPath("$.id", networkACL)
	if hasNetworkACLID {
		networkACLsOutput, err := service.DescribeNetworkAcls(ctx, &ec2.DescribeNetworkAclsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("network-acl-id"),
					Values: []string{core.StringValue(networkACLID)},
				},
			},
		})
		if err != nil {
			return "", err
		}

		if len(networkACLsOutput.NetworkAcls) == 0 {
			return "", errors.New("no network ACL found for ID in state")
		}

		networkACL := networkACLsOutput.NetworkAcls[0]
		if len(networkACL.Associations) > 0 {
			associationID := networkACL.Associations[0].NetworkAclAssociationId
			_, err := service.ReplaceNetworkAclAssociation(ctx, &ec2.ReplaceNetworkAclAssociationInput{
				AssociationId: associationID,
				NetworkAclId:  defaultNetworkACLID,
			})
			if err != nil {
				return "", err
			}
		}

		return core.StringValue(networkACLID), nil
	}

	return "", nil
}

func (l *vpcResourceActions) deleteNetworkACLs(
	ctx context.Context,
	service ec2service.Service,
	networkACLIDs []string,
) error {
	if len(networkACLIDs) == 0 {
		return nil
	}

	for _, networkACLID := range networkACLIDs {
		_, err := service.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: aws.String(networkACLID),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *vpcResourceActions) deleteSecurityGroups(
	ctx context.Context,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) error {
	securityGroups, hasSecurityGroups := pluginutils.GetValueByPath(
		"$.securityGroups",
		resourceSpecData,
	)
	if !hasSecurityGroups || len(securityGroups.Items) == 0 {
		return nil
	}

	for _, securityGroupID := range securityGroups.Items {
		_, err := service.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(core.StringValue(securityGroupID)),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *vpcResourceActions) deleteRoutes(
	ctx context.Context,
	vpcID string,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) error {
	err := l.disassociateRouteTablesFromSubnets(ctx, vpcID, service, resourceSpecData)
	if err != nil {
		return err
	}

	return l.deleteRouteTables(ctx, service, resourceSpecData)
}

func (l *vpcResourceActions) disassociateRouteTablesFromSubnets(
	ctx context.Context,
	vpcID string,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) error {
	routeTables, hasRouteTables := pluginutils.GetValueByPath(
		"$.routeTables",
		resourceSpecData,
	)
	if !hasRouteTables || len(routeTables.Items) == 0 {
		return nil
	}

	routeTablesOutput, err := service.DescribeRouteTables(
		ctx,
		&ec2.DescribeRouteTablesInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{vpcID},
				},
			},
		},
	)
	if err != nil {
		return err
	}

	for _, routeTable := range routeTablesOutput.RouteTables {
		for _, association := range routeTable.Associations {
			_, err := service.DisassociateRouteTable(
				ctx,
				&ec2.DisassociateRouteTableInput{
					AssociationId: association.RouteTableAssociationId,
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *vpcResourceActions) deleteRouteTables(
	ctx context.Context,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) error {
	routeTables, hasRouteTables := pluginutils.GetValueByPath(
		"$.routeTables",
		resourceSpecData,
	)
	if !hasRouteTables || len(routeTables.Items) == 0 {
		return nil
	}

	for _, routeTable := range routeTables.Items {
		routeTableID, hasRouteTableID := pluginutils.GetValueByPath("$.id", routeTable)
		if hasRouteTableID {
			_, err := service.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
				RouteTableId: aws.String(core.StringValue(routeTableID)),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *vpcResourceActions) deleteNATGatewaysAndReleaseElasticIPs(
	ctx context.Context,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) error {
	natGateways, hasNatGateways := pluginutils.GetValueByPath(
		"$.gateways.natGateways",
		resourceSpecData,
	)
	if !hasNatGateways || len(natGateways.Items) == 0 {
		return nil
	}

	for _, natGateway := range natGateways.Items {
		natGatewayID, hasNatGatewayID := pluginutils.GetValueByPath("$.id", natGateway)
		if hasNatGatewayID {
			_, err := service.DeleteNatGateway(ctx, &ec2.DeleteNatGatewayInput{
				NatGatewayId: aws.String(core.StringValue(natGatewayID)),
			})
			if err != nil {
				return err
			}
		}

		elasticIPID, hasElasticIPID := pluginutils.GetValueByPath("$.elasticIpId", natGateway)
		if hasElasticIPID {
			_, err := service.ReleaseAddress(ctx, &ec2.ReleaseAddressInput{
				AllocationId: aws.String(core.StringValue(elasticIPID)),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *vpcResourceActions) detachAndDeleteInternetGateway(
	ctx context.Context,
	vpcID string,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) error {
	internetGatewayID, hasInternetGatewayID := pluginutils.GetValueByPath(
		"$.gateways.internetGatewayId",
		resourceSpecData,
	)
	if !hasInternetGatewayID {
		return nil
	}

	_, err := service.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(core.StringValue(internetGatewayID)),
		VpcId:             aws.String(vpcID),
	})
	if err != nil {
		return err
	}

	_, err = service.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(core.StringValue(internetGatewayID)),
	})
	return err
}

func (l *vpcResourceActions) deleteSubnets(
	ctx context.Context,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) error {
	subnets, hasSubnets := pluginutils.GetValueByPath(
		"$.subnets",
		resourceSpecData,
	)
	if !hasSubnets || len(subnets.Items) == 0 {
		return nil
	}

	for _, subnet := range subnets.Items {
		subnetID, hasSubnetID := pluginutils.GetValueByPath("$.id", subnet)
		if hasSubnetID {
			_, err := service.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
				SubnetId: aws.String(core.StringValue(subnetID)),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *vpcResourceActions) deleteVPC(
	ctx context.Context,
	vpcID string,
	service ec2service.Service,
) error {
	_, err := service.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(vpcID),
	})
	return err
}
