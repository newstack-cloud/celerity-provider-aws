package linkutils

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Provisions (or ref-count-aware removes) a gateway VPC endpoint
// so a VPC-isolated caller can reach S3 or DynamoDB over the private network. Unlike an
// interface endpoint, a gateway endpoint attaches to the caller's route tables and needs
// no security group or network interface: access is via the route-table prefix list.
func activateGatewayEndpoint(
	ctx context.Context,
	ec2Service ec2service.Service,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	activation NetworkingActivation,
	flexVPCResourceState *state.ResourceState,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	flexVPCNameValue, hasFlexVPCName := pluginutils.GetValueByPath("$.name", flexVPCResourceState.SpecData)
	if !hasFlexVPCName {
		return nil, fmt.Errorf("flex VPC name could not be retrieved from the flex VPC resource")
	}
	flexVPCName := core.StringValue(flexVPCNameValue)

	serviceName := serviceEndpointName(activation.Region, activation.AWSService)
	vpcID := activation.Caller.VPCID

	routeTableIDs, err := callerRouteTableIDs(ctx, ec2Service, vpcID, activation.Caller.SubnetIDs)
	if err != nil {
		return nil, err
	}

	vpcEndpointsOutput, err := ec2Service.DescribeVpcEndpoints(
		ctx,
		&ec2.DescribeVpcEndpointsInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{vpcID}},
				{Name: aws.String("service-name"), Values: []string{serviceName}},
				utils.CreateTagFilterFlexVPCNameForLink(flexVPCName),
			},
		},
	)
	if err != nil {
		return nil, err
	}

	info := &gatewayEndpointInfo{
		flexVPCName:   flexVPCName,
		vpcID:         vpcID,
		serviceName:   serviceName,
		routeTableIDs: routeTableIDs,
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy && len(vpcEndpointsOutput.VpcEndpoints) > 0 {
		return removeGatewayEndpoint(ctx, ec2Service, &vpcEndpointsOutput.VpcEndpoints[0], input, output)
	}

	if len(vpcEndpointsOutput.VpcEndpoints) == 0 {
		return createGatewayEndpoint(ctx, ec2Service, info, input, output)
	}

	return updateGatewayEndpoint(ctx, ec2Service, info, &vpcEndpointsOutput.VpcEndpoints[0], input, output)
}

func createGatewayEndpoint(
	ctx context.Context,
	ec2Service ec2service.Service,
	info *gatewayEndpointInfo,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	vpcEndpointOutput, err := ec2Service.CreateVpcEndpoint(
		ctx,
		&ec2.CreateVpcEndpointInput{
			VpcId:           aws.String(info.vpcID),
			VpcEndpointType: ec2types.VpcEndpointTypeGateway,
			ServiceName:     aws.String(info.serviceName),
			RouteTableIds:   info.routeTableIDs,
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeVpcEndpoint,
					Tags: []ec2types.Tag{
						utils.CreateTagLinkVPCEndpoint(),
						utils.CreateTagBlueprintInstanceName(input.InstanceName),
						utils.CreateTagBlueprintLinkID(input.LinkID),
						utils.CreateTagFlexVPCNameForLink(info.flexVPCName),
					},
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	vpcEndpointName := fmt.Sprintf("%sVPCEndpoint", input.ResourceAInfo.ResourceName)
	vpcEndpointLinkData := core.MappingNodeFields(
		vpcEndpointName,
		core.MappingNodeFields(
			"id",
			core.MappingNodeFromString(aws.ToString(vpcEndpointOutput.VpcEndpoint.VpcEndpointId)),
		),
	)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: output.IntermediaryResourceStates,
		LinkData:                   core.MergeMaps(output.LinkData, vpcEndpointLinkData),
		ResourceDataMappings:       output.ResourceDataMappings,
	}, nil
}

func updateGatewayEndpoint(
	ctx context.Context,
	ec2Service ec2service.Service,
	info *gatewayEndpointInfo,
	endpoint *ec2types.VpcEndpoint,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Tag the VPC endpoint with the current link, for reference counting so it is only
	// removed when the last link that references it is destroyed.
	if !utils.HasVPCEndpointTagForLink(endpoint, input.LinkID) {
		_, err := ec2Service.CreateTags(
			ctx,
			&ec2.CreateTagsInput{
				Resources: []string{aws.ToString(endpoint.VpcEndpointId)},
				Tags:      []ec2types.Tag{utils.CreateTagBlueprintLinkID(input.LinkID)},
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Make sure the gateway endpoint is associated with the caller's route tables.
	missing := missingRouteTableIDs(endpoint, info.routeTableIDs)
	if len(missing) > 0 {
		_, err := ec2Service.ModifyVpcEndpoint(
			ctx,
			&ec2.ModifyVpcEndpointInput{
				VpcEndpointId:    endpoint.VpcEndpointId,
				AddRouteTableIds: missing,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

func removeGatewayEndpoint(
	ctx context.Context,
	ec2Service ec2service.Service,
	endpoint *ec2types.VpcEndpoint,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	if !utils.HasVPCEndpointTagForLink(endpoint, input.LinkID) {
		return output, nil
	}

	// Remove the gateway endpoint when the current link is the last one holding a
	// reference, otherwise just drop this link's tag. A gateway endpoint has no security
	// group or network interface, so there is nothing else to clean up.
	otherLinkTags := utils.GetOtherLinkTagsFromVPCEndpoint(endpoint, input.LinkID)
	if len(otherLinkTags) == 0 {
		_, err := ec2Service.DeleteVpcEndpoints(
			ctx,
			&ec2.DeleteVpcEndpointsInput{
				VpcEndpointIds: []string{aws.ToString(endpoint.VpcEndpointId)},
			},
		)
		if err != nil {
			return nil, err
		}
		return output, nil
	}

	_, err := ec2Service.DeleteTags(
		ctx,
		&ec2.DeleteTagsInput{
			Resources: []string{aws.ToString(endpoint.VpcEndpointId)},
			Tags:      []ec2types.Tag{utils.CreateTagBlueprintLinkID(input.LinkID)},
		},
	)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// Resolves the route tables associated with the caller's subnets.
// Subnets without an explicit association use the VPC main route table, so that is
// included as a fallback when an explicit association is absent.
func callerRouteTableIDs(
	ctx context.Context,
	ec2Service ec2service.Service,
	vpcID string,
	subnetIDs []string,
) ([]string, error) {
	routeTablesOutput, err := ec2Service.DescribeRouteTables(
		ctx,
		&ec2.DescribeRouteTablesInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{vpcID}},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	explicit := map[string]string{} // subnetID -> routeTableID
	var mainRouteTableID string
	for i := range routeTablesOutput.RouteTables {
		routeTable := routeTablesOutput.RouteTables[i]
		for _, association := range routeTable.Associations {
			if aws.ToBool(association.Main) {
				mainRouteTableID = aws.ToString(routeTable.RouteTableId)
			}
			if association.SubnetId != nil {
				explicit[aws.ToString(association.SubnetId)] = aws.ToString(routeTable.RouteTableId)
			}
		}
	}

	ids := map[string]struct{}{}
	for _, subnetID := range subnetIDs {
		if routeTableID, ok := explicit[subnetID]; ok {
			ids[routeTableID] = struct{}{}
		} else if mainRouteTableID != "" {
			ids[mainRouteTableID] = struct{}{}
		}
	}

	if len(ids) == 0 {
		return nil, errors.New(
			"no route tables found for the caller's subnets to associate with the gateway VPC endpoint",
		)
	}

	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	return result, nil
}

func missingRouteTableIDs(endpoint *ec2types.VpcEndpoint, wanted []string) []string {
	present := map[string]struct{}{}
	for _, id := range endpoint.RouteTableIds {
		present[id] = struct{}{}
	}
	var missing []string
	for _, id := range wanted {
		if _, ok := present[id]; !ok {
			missing = append(missing, id)
		}
	}
	return missing
}

type gatewayEndpointInfo struct {
	flexVPCName   string
	vpcID         string
	serviceName   string
	routeTableIDs []string
}
