//go:build unit

package lambda

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaCodeSigningConfigResourceDestroySuite struct {
	suite.Suite
}

func (s *LambdaCodeSigningConfigResourceDestroySuite) Test_destroy() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		createSuccessfulCodeSigningConfigDestroyTestCase(providerCtx, loader),
		createFailingCodeSigningConfigDestroyTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		CodeSigningConfigResource,
		&s.Suite,
	)
}

func createSuccessfulCodeSigningConfigDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteCodeSigningConfigOutput(&lambda.DeleteCodeSigningConfigOutput{}),
	)

	expectedCodeSigningConfigARN := "arn:aws:lambda:us-east-1:123456789012:code-signing-config:test-config"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully deletes code signing config",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"codeSigningConfigArn": core.MappingNodeFromString(expectedCodeSigningConfigARN),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"DeleteCodeSigningConfig": &lambda.DeleteCodeSigningConfigInput{
				CodeSigningConfigArn: aws.String(expectedCodeSigningConfigARN),
			},
		},
	}
}

func createFailingCodeSigningConfigDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteCodeSigningConfigError(errors.New("failed to delete code signing config")),
	)

	expectedCodeSigningConfigARN := "arn:aws:lambda:us-east-1:123456789012:code-signing-config:test-config"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "fails to delete code signing config",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"codeSigningConfigArn": core.MappingNodeFromString(expectedCodeSigningConfigARN),
					},
				},
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteCodeSigningConfig": &lambda.DeleteCodeSigningConfigInput{
				CodeSigningConfigArn: aws.String(expectedCodeSigningConfigARN),
			},
		},
	}
}

func TestLambdaCodeSigningConfigResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(LambdaCodeSigningConfigResourceDestroySuite))
}
