package flex

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Stabilised checks if the VPC is available and all NAT gateways are available as
// these are the only resources that are created asynchronously and will need time to stabilise.
func (l *vpcResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	service, _, err := l.getEC2ServiceWithRegion(ctx, input.ProviderContext, nil)
	if err != nil {
		return nil, err
	}

	vpcID, hasVPCID := pluginutils.GetValueByPath("$.vpcId", input.ResourceSpec)
	if !hasVPCID {
		return nil, errors.New("vpcId is not set in resource spec data")
	}

	vpc, err := service.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{core.StringValue(vpcID)},
	})
	if err != nil {
		return nil, err
	}

	if len(vpc.Vpcs) == 0 || vpc.Vpcs[0].State != types.VpcStateAvailable {
		return &provider.ResourceHasStabilisedOutput{
			Stabilised: false,
		}, nil
	}

	natGateways, hasNatGateways := pluginutils.GetValueByPath("$.gateways.natGateways", input.ResourceSpec)
	if hasNatGateways && core.IsArrayMappingNode(natGateways) {
		for _, natGateway := range natGateways.Items {
			natGatewayID, hasID := pluginutils.GetValueByPath("$.id", natGateway)
			if hasID {
				natGatewayDesc, err := service.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{
					NatGatewayIds: []string{core.StringValue(natGatewayID)},
				})
				if err != nil {
					return nil, err
				}

				if len(natGatewayDesc.NatGateways) == 0 ||
					natGatewayDesc.NatGateways[0].State != types.NatGatewayStateAvailable {
					return &provider.ResourceHasStabilisedOutput{
						Stabilised: false,
					}, nil
				}
			}
		}
	}

	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
