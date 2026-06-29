//go:build unit

package lambdasns

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

const (
	ltFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:publish-order"
	ltRoleARN     = "arn:aws:iam::123456789012:role/publish-order-role"
	ltTopicARN    = "arn:aws:sns:us-west-2:123456789012:orders"
)

type FunctionTopicLinkUpdateSuite struct {
	suite.Suite
}

func ltFunctionInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "publishOrderFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("arn", core.MappingNodeFromString(ltFunctionARN)),
		},
	}
}

func ltTopicInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersTopic",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("topicArn", core.MappingNodeFromString(ltTopicARN)),
		},
	}
}

func ltGetFunctionOutput() *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(ltFunctionARN),
			Role:        aws.String(ltRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{
				Variables: map[string]string{"EXISTING": "val"},
			},
		},
	}
}

func (s *FunctionTopicLinkUpdateSuite) Test_update_resource_a_env_vars() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		functionTopicAddEnvVarTestCase(loader),
		functionTopicRemoveEnvVarTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(testCases, functionTopicLinkFactory(), &s.Suite)
}

func functionTopicAddEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(ltGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "populates the topic ARN env var on the function",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      ltFunctionInfo(),
			OtherResourceInfo: ltTopicInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"publishOrderFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						"SNS_TOPIC_ordersTopic", core.MappingNodeFromString(ltTopicARN),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				"publishOrderFunction::spec.environment.variables[\"SNS_TOPIC_ordersTopic\"]": "publishOrderFunction.environmentVariables[\"SNS_TOPIC_ordersTopic\"]",
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(ltFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{
						"EXISTING":              "val",
						"SNS_TOPIC_ordersTopic": ltTopicARN,
					},
				},
			},
		},
	}
}

func functionTopicRemoveEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(ltFunctionARN),
				Role:        aws.String(ltRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{
					Variables: map[string]string{"EXISTING": "val", "SNS_TOPIC_ordersTopic": ltTopicARN},
				},
			},
		}),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "removes the topic env var on destroy",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      ltFunctionInfo(),
			OtherResourceInfo: ltTopicInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields("publishOrderFunction", core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(ltFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{"EXISTING": "val"},
				},
			},
		},
	}
}

func TestFunctionTopicLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionTopicLinkUpdateSuite))
}
