//go:build unit

package lambda

import (
	"fmt"
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

type LambdaAliasResourceDestroySuite struct {
	suite.Suite
}

func (s *LambdaAliasResourceDestroySuite) Test_destroy_lambda_alias() {
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
		destroyBasicAliasTestCase(providerCtx, loader),
		destroyAliasWithComplexConfigTestCase(providerCtx, loader),
		destroyAliasFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		AliasResource,
		&s.Suite,
	)
}

func destroyBasicAliasTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteAliasOutput(&lambda.DeleteAliasOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("PROD"),
			"functionVersion": core.MappingNodeFromString("1"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy basic alias",
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
				SpecData: specData,
			},
		},
		DestroyActionsCalled: map[string]any{
			"DeleteAlias": &lambda.DeleteAliasInput{
				FunctionName: aws.String("test-function"),
				Name:         aws.String("PROD"),
			},
		},
	}
}

func destroyAliasWithComplexConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteAliasOutput(&lambda.DeleteAliasOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("COMPLEX"),
			"functionVersion": core.MappingNodeFromString("5"),
			"description":     core.MappingNodeFromString("Complex alias with all features"),
			"routingConfig": {
				Fields: map[string]*core.MappingNode{
					"additionalVersionWeights": {
						Fields: map[string]*core.MappingNode{
							"4": core.MappingNodeFromFloat(0.2),
							"3": core.MappingNodeFromFloat(0.1),
						},
					},
				},
			},
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(100),
				},
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy alias with complex config",
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
				SpecData: specData,
			},
		},
		DestroyActionsCalled: map[string]any{
			"DeleteAlias": &lambda.DeleteAliasInput{
				FunctionName: aws.String("test-function"),
				Name:         aws.String("COMPLEX"),
			},
		},
	}
}

func destroyAliasFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteAliasError(fmt.Errorf("failed to delete alias")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("FAIL"),
			"functionVersion": core.MappingNodeFromString("1"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy alias failure",
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
				SpecData: specData,
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteAlias": &lambda.DeleteAliasInput{
				FunctionName: aws.String("test-function"),
				Name:         aws.String("FAIL"),
			},
		},
	}
}

func TestLambdaAliasResourceDestroy(t *testing.T) {
	suite.Run(t, new(LambdaAliasResourceDestroySuite))
}
