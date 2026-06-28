//go:build unit

package flexlambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

const vfFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:get-order"

func vpcFunctionLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, ec2service.Service, *aws.Config, lambdaservice.Service],
) provider.Link {
	build := VPCFunctionLink()
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, ec2service.Service, *aws.Config, lambdaservice.Service],
	) provider.Link {
		return build(VPCToFunctionLinkDeps(deps))
	}
}

type VPCFunctionLinkUpdateSuite struct {
	suite.Suite
}

func functionResourceInfoB() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "getOrderFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(vfFunctionARN),
			),
		},
	}
}

func functionResourceInfoBWithSubnetType(subnetType string) *provider.ResourceInfo {
	info := functionResourceInfoB()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"aws.flexvpc.lambda.subnetType": core.MappingNodeFromString(subnetType),
				},
			},
		},
	}
	return info
}

func expectVPCConfigOutput(
	functionName string,
	subnetIDs, securityGroupIDs []string,
) *provider.LinkUpdateResourceOutput {
	subnetItems := make([]*core.MappingNode, len(subnetIDs))
	for i, id := range subnetIDs {
		subnetItems[i] = core.MappingNodeFromString(id)
	}
	sgItems := make([]*core.MappingNode, len(securityGroupIDs))
	for i, id := range securityGroupIDs {
		sgItems[i] = core.MappingNodeFromString(id)
	}
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			functionName,
			core.MappingNodeFields(
				"vpcConfig",
				core.MappingNodeFields(
					"subnetIds", &core.MappingNode{Items: subnetItems},
					"securityGroupIds", &core.MappingNode{Items: sgItems},
				),
			),
		),
		ResourceDataMappings: map[string]string{
			functionName + "::spec.vpcConfig.subnetIds":        functionName + ".vpcConfig.subnetIds",
			functionName + "::spec.vpcConfig.securityGroupIds": functionName + ".vpcConfig.securityGroupIds",
		},
	}
}

func flexVPCResourceInfoA() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "appVpc",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"name", core.MappingNodeFromString("orders-vpc"),
				"subnets", core.MappingNodeFields(
					"private-az-1", core.MappingNodeFields(
						"id", core.MappingNodeFromString("subnet-priv-b"),
						"subnetType", core.MappingNodeFromString("private"),
					),
					"private-az-2", core.MappingNodeFields(
						"id", core.MappingNodeFromString("subnet-priv-a"),
						"subnetType", core.MappingNodeFromString("private"),
					),
					"public-az-1", core.MappingNodeFields(
						"id", core.MappingNodeFromString("subnet-pub-a"),
						"subnetType", core.MappingNodeFromString("public"),
					),
				),
				"securityGroups", &core.MappingNode{
					Items: []*core.MappingNode{core.MappingNodeFromString("sg-123")},
				},
			),
		},
	}
}

func (s *VPCFunctionLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		vpcFunctionPlaceTestCase(loader),
		vpcFunctionPlacePublicTestCase(loader),
		vpcFunctionDetachTestCase(loader),
		vpcFunctionNoMatchingTierTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(testCases, vpcFunctionLinkFactory(), &s.Suite)
}

func vpcFunctionPlaceTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "places the function in the VPC's private subnets",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         ec2mock.CreateEc2ServiceMockFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoB(),
			OtherResourceInfo: flexVPCResourceInfoA(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: expectVPCConfigOutput(
			"getOrderFunction",
			[]string{"subnet-priv-a", "subnet-priv-b"},
			[]string{"sg-123"},
		),
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(vfFunctionARN),
				VpcConfig: &lambdatypes.VpcConfig{
					SubnetIds:        []string{"subnet-priv-a", "subnet-priv-b"},
					SecurityGroupIds: []string{"sg-123"},
				},
			},
		},
	}
}

func vpcFunctionPlacePublicTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "places the function in the VPC's public subnets when subnetType is public",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         ec2mock.CreateEc2ServiceMockFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoBWithSubnetType("public"),
			OtherResourceInfo: flexVPCResourceInfoA(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: expectVPCConfigOutput(
			"getOrderFunction",
			[]string{"subnet-pub-a"},
			[]string{"sg-123"},
		),
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(vfFunctionARN),
				VpcConfig: &lambdatypes.VpcConfig{
					SubnetIds:        []string{"subnet-pub-a"},
					SecurityGroupIds: []string{"sg-123"},
				},
			},
		},
	}
}

func vpcFunctionNoMatchingTierTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "returns an error when the VPC has no subnets in the requested tier",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         ec2mock.CreateEc2ServiceMockFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoBWithSubnetType("isolated"),
			OtherResourceInfo: flexVPCResourceInfoA(),
			LinkContext:       testLinkContext(),
		},
		ExpectError:            true,
		ExpectedErrorMessage:   "no \"isolated\" subnets",
		UpdateActionsNotCalled: []string{"UpdateFunctionConfiguration"},
	}
}

func vpcFunctionDetachTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "detaches the function from the VPC on destroy",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         ec2mock.CreateEc2ServiceMockFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      functionResourceInfoB(),
			OtherResourceInfo: flexVPCResourceInfoA(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields("getOrderFunction", core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(vfFunctionARN),
				VpcConfig: &lambdatypes.VpcConfig{
					SubnetIds:        []string{},
					SecurityGroupIds: []string{},
				},
			},
		},
	}
}

func TestVPCFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(VPCFunctionLinkUpdateSuite))
}
