//go:build unit

package lambda

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaAliasResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaAliasResourceGetExternalStateSuite) Test_get_external_state_lambda_alias() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		getBasicAliasExternalStateTestCase(providerCtx, loader),
		getAliasWithDescriptionExternalStateTestCase(providerCtx, loader),
		getAliasWithRoutingConfigExternalStateTestCase(providerCtx, loader),
		getComplexAliasExternalStateTestCase(providerCtx, loader),
		getAliasExternalStateFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		AliasResource,
		&s.Suite,
	)
}

func getBasicAliasExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:PROD"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetAliasOutput(&lambda.GetAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("PROD"),
			FunctionVersion: aws.String("1"),
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get basic alias external state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"name":         core.MappingNodeFromString("PROD"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName":    core.MappingNodeFromString("test-function"),
					"name":            core.MappingNodeFromString("PROD"),
					"functionVersion": core.MappingNodeFromString("1"),
					"aliasArn":        core.MappingNodeFromString(aliasArn),
				},
			},
		},
	}
}

func getAliasWithDescriptionExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:STAGING"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetAliasOutput(&lambda.GetAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("STAGING"),
			FunctionVersion: aws.String("2"),
			Description:     aws.String("Staging environment alias"),
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get alias with description external state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"name":         core.MappingNodeFromString("STAGING"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName":    core.MappingNodeFromString("test-function"),
					"name":            core.MappingNodeFromString("STAGING"),
					"functionVersion": core.MappingNodeFromString("2"),
					"description":     core.MappingNodeFromString("Staging environment alias"),
					"aliasArn":        core.MappingNodeFromString(aliasArn),
				},
			},
		},
	}
}

func getAliasWithRoutingConfigExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:CANARY"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetAliasOutput(&lambda.GetAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("CANARY"),
			FunctionVersion: aws.String("3"),
			RoutingConfig: &types.AliasRoutingConfiguration{
				AdditionalVersionWeights: map[string]float64{
					"2": 0.1,
				},
			},
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get alias with routing config external state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"name":         core.MappingNodeFromString("CANARY"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName":    core.MappingNodeFromString("test-function"),
					"name":            core.MappingNodeFromString("CANARY"),
					"functionVersion": core.MappingNodeFromString("3"),
					"aliasArn":        core.MappingNodeFromString(aliasArn),
					"routingConfig": {
						Fields: map[string]*core.MappingNode{
							"additionalVersionWeights": {
								Fields: map[string]*core.MappingNode{
									"2": core.MappingNodeFromFloat(0.1),
								},
							},
						},
					},
				},
			},
		},
	}
}

func getComplexAliasExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:COMPLEX"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetAliasOutput(&lambda.GetAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("COMPLEX"),
			FunctionVersion: aws.String("4"),
			Description:     aws.String("Complex alias with all features"),
			RoutingConfig: &types.AliasRoutingConfiguration{
				AdditionalVersionWeights: map[string]float64{
					"3": 0.2,
					"2": 0.1,
				},
			},
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get complex alias external state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"name":         core.MappingNodeFromString("COMPLEX"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName":    core.MappingNodeFromString("test-function"),
					"name":            core.MappingNodeFromString("COMPLEX"),
					"functionVersion": core.MappingNodeFromString("4"),
					"description":     core.MappingNodeFromString("Complex alias with all features"),
					"aliasArn":        core.MappingNodeFromString(aliasArn),
					"routingConfig": {
						Fields: map[string]*core.MappingNode{
							"additionalVersionWeights": {
								Fields: map[string]*core.MappingNode{
									"3": core.MappingNodeFromFloat(0.2),
									"2": core.MappingNodeFromFloat(0.1),
								},
							},
						},
					},
				},
			},
		},
	}
}

func getAliasExternalStateFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("NOTFOUND"),
			"functionVersion": core.MappingNodeFromString("1"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get alias external state failure",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetAliasError(fmt.Errorf("failed to get alias")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext:     providerCtx,
			CurrentResourceSpec: currentResourceSpec,
		},
		ExpectError: true,
	}
}

func TestLambdaAliasResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(LambdaAliasResourceGetExternalStateSuite))
}
