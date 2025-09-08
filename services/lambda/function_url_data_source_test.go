//go:build unit

package lambda

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionUrlDataSourceSuite struct {
	suite.Suite
}

func (s *LambdaFunctionUrlDataSourceSuite) Test_fetch() {
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

	testCases := []DataSourceFetchTestCase{
		createBasicFunctionUrlDataSourceFetchTestCase(providerCtx, loader),
		createFunctionUrlDataSourceWithCorsTestCase(providerCtx, loader),
		createFunctionUrlDataSourceWithInvokeModeTestCase(providerCtx, loader),
		createFunctionUrlDataSourceWithQualifierTestCase(providerCtx, loader),
		createFunctionUrlDataSourceWithRegionFilterTestCase(providerCtx, loader),
		createFunctionUrlDataSourceFetchErrorTestCase(providerCtx, loader),
		createFunctionUrlDataSourceMissingNameTestCase(providerCtx, loader),
	}

	for _, tc := range testCases {
		s.Run(tc.Name, func() {
			// Create the data source
			dataSource := FunctionUrlDataSource(tc.ServiceFactory, tc.ConfigStore)

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

func TestLambdaFunctionUrlDataSourceSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionUrlDataSourceSuite))
}

// Test case generator functions

func createBasicFunctionUrlDataSourceFetchTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches basic function URL data",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
				FunctionUrl:      aws.String("https://abc123.lambda-url.us-west-2.on.aws/"),
				FunctionArn:      aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:         types.FunctionUrlAuthTypeNone,
				CreationTime:     aws.String("2023-01-01T00:00:00.000Z"),
				LastModifiedTime: aws.String("2023-01-01T00:00:00.000Z"),
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
				Filter: pluginutils.CreateStringEqualsFilter("functionName", "test-function"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"functionUrl":      core.MappingNodeFromString("https://abc123.lambda-url.us-west-2.on.aws/"),
				"functionArn":      core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				"authType":         core.MappingNodeFromString("NONE"),
				"creationTime":     core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
				"lastModifiedTime": core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
			},
		},
		ExpectError: false,
	}
}

func createFunctionUrlDataSourceWithCorsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches function URL with CORS configuration",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
				FunctionUrl:      aws.String("https://def456.lambda-url.us-west-2.on.aws/"),
				FunctionArn:      aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:         types.FunctionUrlAuthTypeAwsIam,
				CreationTime:     aws.String("2023-01-01T00:00:00.000Z"),
				LastModifiedTime: aws.String("2023-01-01T00:00:00.000Z"),
				Cors: &types.Cors{
					AllowCredentials: aws.Bool(true),
					AllowHeaders:     []string{"Content-Type", "X-Amz-Date"},
					AllowMethods:     []string{"GET", "POST"},
					AllowOrigins:     []string{"https://example.com"},
					ExposeHeaders:    []string{"X-Custom-Header"},
					MaxAge:           aws.Int32(300),
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
				Filter: pluginutils.CreateStringEqualsFilter("functionName", "test-function"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"functionUrl":           core.MappingNodeFromString("https://def456.lambda-url.us-west-2.on.aws/"),
				"functionArn":           core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				"authType":              core.MappingNodeFromString("AWS_IAM"),
				"creationTime":          core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
				"lastModifiedTime":      core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
				"cors.allowCredentials": core.MappingNodeFromBool(true),
				"cors.allowHeaders": {
					Items: []*core.MappingNode{
						core.MappingNodeFromString("Content-Type"),
						core.MappingNodeFromString("X-Amz-Date"),
					},
				},
				"cors.allowMethods": {
					Items: []*core.MappingNode{
						core.MappingNodeFromString("GET"),
						core.MappingNodeFromString("POST"),
					},
				},
				"cors.allowOrigins": {
					Items: []*core.MappingNode{
						core.MappingNodeFromString("https://example.com"),
					},
				},
				"cors.exposeHeaders": {
					Items: []*core.MappingNode{
						core.MappingNodeFromString("X-Custom-Header"),
					},
				},
				"cors.maxAge": core.MappingNodeFromInt(300),
			},
		},
		ExpectError: false,
	}
}

func createFunctionUrlDataSourceWithInvokeModeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches function URL with invoke mode",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
				FunctionUrl:      aws.String("https://ghi789.lambda-url.us-west-2.on.aws/"),
				FunctionArn:      aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:         types.FunctionUrlAuthTypeNone,
				CreationTime:     aws.String("2023-01-01T00:00:00.000Z"),
				LastModifiedTime: aws.String("2023-01-01T00:00:00.000Z"),
				InvokeMode:       types.InvokeModeResponseStream,
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
				Filter: pluginutils.CreateStringEqualsFilter("functionName", "test-function"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"functionUrl":      core.MappingNodeFromString("https://ghi789.lambda-url.us-west-2.on.aws/"),
				"functionArn":      core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				"authType":         core.MappingNodeFromString("NONE"),
				"creationTime":     core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
				"lastModifiedTime": core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
				"invokeMode":       core.MappingNodeFromString("RESPONSE_STREAM"),
			},
		},
		ExpectError: false,
	}
}

func createFunctionUrlDataSourceWithQualifierTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches function URL with qualifier",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
				FunctionUrl:      aws.String("https://jkl012.lambda-url.us-west-2.on.aws/"),
				FunctionArn:      aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function:prod"),
				AuthType:         types.FunctionUrlAuthTypeNone,
				CreationTime:     aws.String("2023-01-01T00:00:00.000Z"),
				LastModifiedTime: aws.String("2023-01-01T00:00:00.000Z"),
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
						pluginutils.CreateStringEqualsFilter("qualifier", "prod").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"functionUrl":      core.MappingNodeFromString("https://jkl012.lambda-url.us-west-2.on.aws/"),
				"functionArn":      core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:prod"),
				"authType":         core.MappingNodeFromString("NONE"),
				"creationTime":     core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
				"lastModifiedTime": core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
			},
		},
		ExpectError: false,
	}
}

func createFunctionUrlDataSourceWithRegionFilterTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches function URL with region filter",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
				FunctionUrl:      aws.String("https://mno345.lambda-url.eu-west-1.on.aws/"),
				FunctionArn:      aws.String("arn:aws:lambda:eu-west-1:123456789012:function:test-function"),
				AuthType:         types.FunctionUrlAuthTypeNone,
				CreationTime:     aws.String("2023-01-01T00:00:00.000Z"),
				LastModifiedTime: aws.String("2023-01-01T00:00:00.000Z"),
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
						pluginutils.CreateStringEqualsFilter("region", "eu-west-1").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"functionUrl":      core.MappingNodeFromString("https://mno345.lambda-url.eu-west-1.on.aws/"),
				"functionArn":      core.MappingNodeFromString("arn:aws:lambda:eu-west-1:123456789012:function:test-function"),
				"authType":         core.MappingNodeFromString("NONE"),
				"creationTime":     core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
				"lastModifiedTime": core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
			},
		},
		ExpectError: false,
	}
}

func createFunctionUrlDataSourceFetchErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "handles get function URL config error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionUrlConfigError(errors.New("Function URL not found")),
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
				Filter: pluginutils.CreateStringEqualsFilter("functionName", "non-existent-function"),
			},
		},
		ExpectedOutput:       nil,
		ExpectError:          true,
		ExpectedErrorMessage: "Function URL not found",
	}
}

func createFunctionUrlDataSourceMissingNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "handles missing function name",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{}),
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
					Filters: []*provider.ResolvedDataSourceFilter{},
				},
			},
		},
		ExpectedOutput:       nil,
		ExpectError:          true,
		ExpectedErrorMessage: "function name is required",
	}
}
