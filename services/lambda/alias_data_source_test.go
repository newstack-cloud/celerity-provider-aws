//go:build unit

package lambda

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type LambdaAliasDataSourceSuite struct {
	suite.Suite
}

// Custom test case structure for data source tests.
type AliasDataSourceFetchTestCase struct {
	Name                 string
	ServiceFactory       func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service
	ConfigStore          pluginutils.ServiceConfigStore[*aws.Config]
	Input                *provider.DataSourceFetchInput
	ExpectedOutput       *provider.DataSourceFetchOutput
	ExpectError          bool
	ExpectedErrorMessage string
}

func (s *LambdaAliasDataSourceSuite) Test_fetch() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			pluginutils.SessionIDKey: core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []AliasDataSourceFetchTestCase{
		createBasicAliasFetchTestCase(providerCtx, loader),
		createDataSourceAliasWithDescriptionTestCase(providerCtx, loader),
		createDataSourceAliasWithRoutingConfigTestCase(providerCtx, loader),
		createAliasWithAllOptionalConfigsTestCase(providerCtx, loader),
		createAliasFetchErrorTestCase(providerCtx, loader),
		createAliasMissingFunctionNameTestCase(providerCtx, loader),
		createAliasMissingNameTestCase(providerCtx, loader),
	}

	for _, tc := range testCases {
		s.Run(tc.Name, func() {
			// Create the data source
			dataSource := AliasDataSource(tc.ServiceFactory, tc.ConfigStore)

			// Execute the fetch
			output, err := dataSource.Fetch(context.Background(), tc.Input)

			// Assert results
			if tc.ExpectError {
				s.Error(err)
				if tc.ExpectedErrorMessage != "" {
					s.ErrorContains(err, tc.ExpectedErrorMessage)
				}
				s.Nil(output)
			} else {
				s.NoError(err)
				s.NotNil(output)
				s.Equal(tc.ExpectedOutput.Data, output.Data)
			}
		})
	}
}

func TestLambdaAliasDataSourceSuite(t *testing.T) {
	suite.Run(t, new(LambdaAliasDataSourceSuite))
}

// Test case generator functions below.

func createBasicAliasFetchTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) AliasDataSourceFetchTestCase {
	return AliasDataSourceFetchTestCase{
		Name: "successfully fetches basic alias data",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetAliasOutput(&lambda.GetAliasOutput{
				AliasArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				FunctionVersion: aws.String("1"),
				Name:            aws.String("test-alias"),
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: &provider.ResolvedDataSourceFilters{
					Filters: []*provider.ResolvedDataSourceFilter{
						pluginutils.CreateStringEqualsFilter("functionName", "test-function").Filters[0],
						pluginutils.CreateStringEqualsFilter("name", "test-alias").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":             core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				"functionName":    core.MappingNodeFromString("test-function"),
				"functionVersion": core.MappingNodeFromString("1"),
				"invokeArn":       core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				"name":            core.MappingNodeFromString("test-alias"),
			},
		},
		ExpectError: false,
	}
}

func createDataSourceAliasWithDescriptionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) AliasDataSourceFetchTestCase {
	return AliasDataSourceFetchTestCase{
		Name: "successfully fetches alias with description",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetAliasOutput(&lambda.GetAliasOutput{
				AliasArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				Description:     aws.String("Test alias description"),
				FunctionVersion: aws.String("1"),
				Name:            aws.String("test-alias"),
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: &provider.ResolvedDataSourceFilters{
					Filters: []*provider.ResolvedDataSourceFilter{
						pluginutils.CreateStringEqualsFilter("functionName", "test-function").Filters[0],
						pluginutils.CreateStringEqualsFilter("name", "test-alias").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":             core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				"description":     core.MappingNodeFromString("Test alias description"),
				"functionName":    core.MappingNodeFromString("test-function"),
				"functionVersion": core.MappingNodeFromString("1"),
				"invokeArn":       core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				"name":            core.MappingNodeFromString("test-alias"),
			},
		},
		ExpectError: false,
	}
}

func createDataSourceAliasWithRoutingConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) AliasDataSourceFetchTestCase {
	routingConfig := map[string]float64{
		"1": 0.7,
		"2": 0.3,
	}
	routingConfigJSON, _ := json.Marshal(routingConfig)

	return AliasDataSourceFetchTestCase{
		Name: "successfully fetches alias with routing config",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetAliasOutput(&lambda.GetAliasOutput{
				AliasArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				FunctionVersion: aws.String("1"),
				Name:            aws.String("test-alias"),
				RoutingConfig: &types.AliasRoutingConfiguration{
					AdditionalVersionWeights: routingConfig,
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: &provider.ResolvedDataSourceFilters{
					Filters: []*provider.ResolvedDataSourceFilter{
						pluginutils.CreateStringEqualsFilter("functionName", "test-function").Filters[0],
						pluginutils.CreateStringEqualsFilter("name", "test-alias").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":                                    core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				"functionName":                           core.MappingNodeFromString("test-function"),
				"functionVersion":                        core.MappingNodeFromString("1"),
				"invokeArn":                              core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				"name":                                   core.MappingNodeFromString("test-alias"),
				"routingConfig.additionalVersionWeights": core.MappingNodeFromString(string(routingConfigJSON)),
			},
		},
		ExpectError: false,
	}
}

func createAliasWithAllOptionalConfigsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) AliasDataSourceFetchTestCase {
	routingConfig := map[string]float64{
		"1": 0.6,
		"2": 0.4,
	}
	routingConfigJSON, _ := json.Marshal(routingConfig)

	return AliasDataSourceFetchTestCase{
		Name: "successfully fetches alias with all optional configurations",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetAliasOutput(&lambda.GetAliasOutput{
				AliasArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				Description:     aws.String("Production alias with routing"),
				FunctionVersion: aws.String("1"),
				Name:            aws.String("test-alias"),
				RoutingConfig: &types.AliasRoutingConfiguration{
					AdditionalVersionWeights: routingConfig,
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: &provider.ResolvedDataSourceFilters{
					Filters: []*provider.ResolvedDataSourceFilter{
						pluginutils.CreateStringEqualsFilter("functionName", "test-function").Filters[0],
						pluginutils.CreateStringEqualsFilter("name", "test-alias").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":                                    core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				"description":                            core.MappingNodeFromString("Production alias with routing"),
				"functionName":                           core.MappingNodeFromString("test-function"),
				"functionVersion":                        core.MappingNodeFromString("1"),
				"invokeArn":                              core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:test-alias"),
				"name":                                   core.MappingNodeFromString("test-alias"),
				"routingConfig.additionalVersionWeights": core.MappingNodeFromString(string(routingConfigJSON)),
			},
		},
		ExpectError: false,
	}
}

func createAliasFetchErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) AliasDataSourceFetchTestCase {
	return AliasDataSourceFetchTestCase{
		Name: "handles get alias error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetAliasError(errors.New("ResourceNotFoundException")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: &provider.ResolvedDataSourceFilters{
					Filters: []*provider.ResolvedDataSourceFilter{
						pluginutils.CreateStringEqualsFilter("functionName", "test-function").Filters[0],
						pluginutils.CreateStringEqualsFilter("name", "test-alias").Filters[0],
					},
				},
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "failed to get Lambda alias",
	}
}

func createAliasMissingFunctionNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) AliasDataSourceFetchTestCase {
	return AliasDataSourceFetchTestCase{
		Name:           "handles missing function name filter",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: &provider.ResolvedDataSourceFilters{
					Filters: []*provider.ResolvedDataSourceFilter{
						pluginutils.CreateStringEqualsFilter("name", "test-alias").Filters[0],
					},
				},
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "function_name filter is required",
	}
}

func createAliasMissingNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) AliasDataSourceFetchTestCase {
	return AliasDataSourceFetchTestCase{
		Name:           "handles missing name filter",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: &provider.ResolvedDataSourceFilters{
					Filters: []*provider.ResolvedDataSourceFilter{
						pluginutils.CreateStringEqualsFilter("functionName", "test-function").Filters[0],
					},
				},
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "name filter is required",
	}
}
