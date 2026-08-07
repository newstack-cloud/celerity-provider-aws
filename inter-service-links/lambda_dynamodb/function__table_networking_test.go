//go:build unit

package lambdadynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/stretchr/testify/require"
)

// DynamoDB is reached over a gateway endpoint, not an interface endpoint.
//
// The distinction is important as an interface endpoint would provision a security
// group and elastic network interfaces per subnet, and pair the caller with that group.
// A gateway endpoint has none of that, so the caller's egress goes to the service's
// managed prefix list instead. Getting the type wrong produces a working-looking
// endpoint that costs money and carries no DynamoDB traffic.
func TestDynamoDBActivationUsesAGatewayEndpoint(t *testing.T) {
	setupCtx := &linkutils.LambdaLinkSetupContext{
		LambdaOutput: &lambdatypes.FunctionConfiguration{
			VpcConfig: &lambdatypes.VpcConfigResponse{
				VpcId:            aws.String("vpc-1"),
				SubnetIds:        []string{"subnet-1"},
				SecurityGroupIds: []string{"sg-caller"},
			},
		},
	}

	activation := dynamoDBNetworkingActivation(setupCtx, "us-west-2")

	require.Equal(t, ec2types.VpcEndpointTypeGateway, activation.EndpointType)
	require.Equal(t, "dynamodb", activation.AWSService)
	require.Equal(t, "us-west-2", activation.Region)
	require.Equal(t, "vpc-1", activation.Caller.VPCID)
	require.Equal(t, []string{"sg-caller"}, activation.Caller.SecurityGroupIDs)

	// A gateway endpoint has no security group of its own, so nothing is paired with
	// the caller directly.
	require.Empty(t, activation.TargetSecurityGroupIDs)
}

// A function that is not placed in a VPC reaches DynamoDB over the public endpoint, so
// the activation carries no attachment and provisions nothing.
func TestDynamoDBActivationIsInertForANonVPCFunction(t *testing.T) {
	setupCtx := &linkutils.LambdaLinkSetupContext{
		LambdaOutput: &lambdatypes.FunctionConfiguration{},
	}

	activation := dynamoDBNetworkingActivation(setupCtx, "us-west-2")

	require.Empty(t, activation.Caller.VPCID)
	require.Empty(t, activation.Caller.SubnetIDs)
	require.Empty(t, activation.Caller.SecurityGroupIDs)
}
