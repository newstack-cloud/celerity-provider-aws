package flex

import (
	"context"
	"errors"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Destroy tears down the VPC and everything created for it.
//
// Errors are wrapped so a failure names the step that produced it. The framework treats
// an error that is not one of its own types as fatal and reports it without a failure
// state, which collapses every distinct failure in the sequence below into the same
// opaque "failed to destroy" message. That made several real defects impossible to
// diagnose from a deployment's output alone.
func (l *vpcResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	if err := l.destroySteps(ctx, input); err != nil {
		// A retryable error must pass through unchanged: wrapping it would turn a
		// "come back later" into a terminal failure.
		var retryable *provider.RetryableError
		if provider.AsRetryableError(err, &retryable) {
			return err
		}

		var destroyErr *provider.ResourceDestroyError
		if provider.AsResourceDestroyError(err, &destroyErr) {
			return err
		}

		return &provider.ResourceDestroyError{
			FailureReasons: []string{err.Error()},
			ChildError:     err,
		}
	}

	return nil
}

func (l *vpcResourceActions) destroySteps(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	service, _, err := l.getEC2ServiceWithRegion(ctx, input.ProviderContext, nil)
	if err != nil {
		return err
	}

	resourceSpecData := input.ResourceState.SpecData
	resourceSpecData, err = l.resolveSpecDataForDestroy(ctx, input, resourceSpecData)
	if err != nil {
		return err
	}
	if resourceSpecData == nil {
		// Nothing was created in the target environment, so there is nothing to
		// tear down.
		return nil
	}

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
		return l.destroyReferencedVPC(ctx, service, resourceSpecData, input.InstanceID, input.InstanceName)
	}

	networkACLIDs, err := l.disassociateNetworkACLsFromSubnets(ctx, vpcID, service, resourceSpecData)
	if err != nil {
		return err
	}

	err = l.deleteNetworkACLs(ctx, service, networkACLIDs)
	if err != nil {
		return err
	}

	// Links create groups they cannot always delete: a workload's group is referenced
	// by rules the access links own, and an endpoint's group can outlive the link when
	// its cleanup yields and the link is then skipped as already destroyed. Both are
	// swept here, after every link in the instance has run.
	err = l.deleteLinkOwnedSecurityGroups(ctx, service, vpcID, input.InstanceID, input.InstanceName)
	if err != nil {
		return err
	}

	// After the link-owned sweep, never before it. A workload's group holds egress to
	// these groups, and EC2 refuses to delete a group that another group's rule still
	// references.
	err = l.deleteNamedSecurityGroups(
		ctx,
		service,
		vpcNameFromSpec(resourceSpecData),
		input.InstanceID,
	)
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

	err = l.deleteEgressOnlyInternetGateway(ctx, service, resourceSpecData)
	if err != nil {
		return err
	}

	err = l.deleteSubnets(ctx, service, resourceSpecData)
	if err != nil {
		return err
	}

	return l.deleteVPC(ctx, vpcID, service)
}

// The shared topology belongs to the blueprint that owns the VPC and must survive, but
// the security groups created for this application's own workloads are this
// blueprint's to remove.
func (l *vpcResourceActions) destroyReferencedVPC(
	ctx context.Context,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
	instanceID string,
	instanceName string,
) error {
	vpcIDValue, _ := pluginutils.GetValueByPath("$.vpcId", resourceSpecData)
	vpcID := core.StringValue(vpcIDValue)

	if err := l.deleteLinkOwnedSecurityGroups(
		ctx, service, vpcID, instanceID, instanceName,
	); err != nil {
		return err
	}

	return l.deleteNamedSecurityGroups(
		ctx,
		service,
		vpcNameFromSpec(resourceSpecData),
		instanceID,
	)
}

func vpcNameFromSpec(resourceSpecData *core.MappingNode) string {
	nameValue, _ := pluginutils.GetValueByPath("$.name", resourceSpecData)
	return core.StringValue(nameValue)
}

// A create that fails partway leaves state without the computed fields the
// teardown steps read, so nothing tracks the resources that were created and they
// stay in the account. Discovery by tag rebuilds the equivalent of that state,
// which is exactly what GetExternalState already does, so the normal teardown can
// proceed against a partial deployment.
//
// Returns nil when the VPC cannot be found, meaning nothing reached the target
// environment.
func (l *vpcResourceActions) resolveSpecDataForDestroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
	resourceSpecData *core.MappingNode,
) (*core.MappingNode, error) {
	if vpcID, hasVPCID := pluginutils.GetValueByPath("$.vpcId", resourceSpecData); hasVPCID &&
		core.StringValue(vpcID) != "" {
		return resourceSpecData, nil
	}

	if name, hasName := pluginutils.GetValueByPath("$.name", resourceSpecData); !hasName ||
		core.StringValue(name) == "" {
		return nil, errors.New("vpcId and name are both missing in flex VPC state")
	}

	externalState, err := l.GetExternalState(ctx, &provider.ResourceGetExternalStateInput{
		ProviderContext:     input.ProviderContext,
		CurrentResourceSpec: resourceSpecData,
		InstanceID:          input.InstanceID,
		InstanceName:        input.InstanceName,
	})
	if err != nil {
		return nil, err
	}

	discovered := externalState.ResourceSpecState
	if vpcID, hasVPCID := pluginutils.GetValueByPath("$.vpcId", discovered); !hasVPCID ||
		core.StringValue(vpcID) == "" {
		return nil, nil
	}

	// The mode is not reflected in external state, so it is carried over from the
	// spec: a referenced VPC must not have its topology torn down here.
	if mode, hasMode := pluginutils.GetValueByPath("$.mode", resourceSpecData); hasMode {
		discovered.Fields["mode"] = mode
	}

	return discovered, nil
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
		"$.securityGroupIds",
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
			// Every VPC has a main route table association that EC2 refuses to
			// disassociate; it disappears with the VPC itself. The query above
			// returns every route table in the VPC, not just the ones created for
			// the flex VPC, so the main association is always among them.
			if aws.ToBool(association.Main) {
				continue
			}

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

	natGatewayIDs := []string{}
	elasticIPIDs := []string{}
	for _, natGateway := range natGateways.Items {
		if natGatewayID, ok := pluginutils.GetValueByPath("$.id", natGateway); ok {
			natGatewayIDs = append(natGatewayIDs, core.StringValue(natGatewayID))
		}
		if elasticIPID, ok := pluginutils.GetValueByPath("$.elasticIpId", natGateway); ok {
			elasticIPIDs = append(elasticIPIDs, core.StringValue(elasticIPID))
		}
	}

	for _, natGatewayID := range natGatewayIDs {
		_, err := service.DeleteNatGateway(ctx, &ec2.DeleteNatGatewayInput{
			NatGatewayId: aws.String(natGatewayID),
		})
		if err != nil {
			return err
		}
	}

	// An elastic IP stays attached to its NAT gateway until the gateway has
	// finished deleting, and EC2 reports an attempt to release it as an
	// AuthFailure rather than as the address being in use. Deletion is requested
	// for every gateway first so the waits overlap.
	if err := waitForNATGatewaysDeleted(ctx, service, natGatewayIDs); err != nil {
		return err
	}

	for _, elasticIPID := range elasticIPIDs {
		_, err := service.ReleaseAddress(ctx, &ec2.ReleaseAddressInput{
			AllocationId: aws.String(elasticIPID),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func waitForNATGatewaysDeleted(
	ctx context.Context,
	service ec2service.Service,
	natGatewayIDs []string,
) error {
	if len(natGatewayIDs) == 0 {
		return nil
	}

	waiter := ec2.NewNatGatewayDeletedWaiter(service)

	return waiter.Wait(
		ctx,
		&ec2.DescribeNatGatewaysInput{NatGatewayIds: natGatewayIDs},
		natGatewayWaitTimeout,
	)
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

// Unlike an internet gateway, an egress-only gateway is not attached to the VPC as a
// separate step, so there is nothing to detach before deleting it.
func (l *vpcResourceActions) deleteEgressOnlyInternetGateway(
	ctx context.Context,
	service ec2service.Service,
	resourceSpecData *core.MappingNode,
) error {
	egressOnlyInternetGatewayID, hasEgressOnlyInternetGatewayID := pluginutils.GetValueByPath(
		"$.gateways.egressOnlyInternetGatewayId",
		resourceSpecData,
	)
	if !hasEgressOnlyInternetGatewayID {
		return nil
	}

	_, err := service.DeleteEgressOnlyInternetGateway(ctx, &ec2.DeleteEgressOnlyInternetGatewayInput{
		EgressOnlyInternetGatewayId: aws.String(core.StringValue(egressOnlyInternetGatewayID)),
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
	if !hasSubnets || len(subnets.Fields) == 0 {
		return nil
	}

	// Subnets are held in state as a map keyed by the preset's subnet name, so
	// they are iterated by field rather than as a list. Deletion order is fixed
	// so a partial failure leaves the same resources behind each time.
	for _, subnetName := range sortedFieldNames(subnets) {
		subnetID, hasSubnetID := pluginutils.GetValueByPath("$.id", subnets.Fields[subnetName])
		if !hasSubnetID {
			continue
		}

		_, err := service.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
			SubnetId: aws.String(core.StringValue(subnetID)),
		})
		if err != nil {
			return err
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

func sortedFieldNames(node *core.MappingNode) []string {
	names := make([]string, 0, len(node.Fields))
	for name := range node.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
