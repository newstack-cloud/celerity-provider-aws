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
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type LambdaLayerVersionDataSourceSuite struct {
	suite.Suite
}

// Custom test case structure for data source tests.
type LayerVersionDataSourceFetchTestCase struct {
	Name                 string
	ServiceFactory       func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service
	ConfigStore          pluginutils.ServiceConfigStore[*aws.Config]
	Input                *provider.DataSourceFetchInput
	ExpectedOutput       *provider.DataSourceFetchOutput
	ExpectError          bool
	ExpectedErrorMessage string
}

func (s *LambdaLayerVersionDataSourceSuite) Test_fetch() {
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

	testCases := []LayerVersionDataSourceFetchTestCase{
		createBasicLayerVersionFetchTestCase(providerCtx, loader),
		createMinimalLayerVersionTestCase(providerCtx, loader),
		createLayerVersionWithAllOptionalFieldsTestCase(providerCtx, loader),
		createLayerVersionWithLayerARNTestCase(providerCtx, loader),
		createLayerVersionFetchErrorTestCase(providerCtx, loader),
		createLayerVersionMissingLayerNameTestCase(providerCtx, loader),
		createLayerVersionMissingVersionNumberTestCase(providerCtx, loader),
	}

	for _, tc := range testCases {
		s.Run(tc.Name, func() {
			// Create the data source
			dataSource := LayerVersionDataSource(tc.ServiceFactory, tc.ConfigStore)

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

func TestLambdaLayerVersionDataSourceSuite(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionDataSourceSuite))
}

// Test case generator functions below.

func createBasicLayerVersionFetchTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) LayerVersionDataSourceFetchTestCase {
	return LayerVersionDataSourceFetchTestCase{
		Name: "successfully fetches basic layer version data",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionOutput(&lambda.GetLayerVersionOutput{
				LayerArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:layer:my-layer"),
				LayerVersionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:layer:my-layer:1"),
				Version:         1,
				Description:     aws.String("My test layer"),
				LicenseInfo:     aws.String("MIT"),
				CreatedDate:     aws.String("2023-01-01T00:00:00.000Z"),
				CompatibleRuntimes: []types.Runtime{
					types.RuntimePython39,
					types.RuntimePython310,
				},
				CompatibleArchitectures: []types.Architecture{
					types.ArchitectureX8664,
				},
				Content: &types.LayerVersionContentOutput{
					CodeSha256: aws.String("abc123"),
					CodeSize:   1024,
					Location:   aws.String("https://example.com/layer.zip"),
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
						pluginutils.CreateStringEqualsFilter("layerName", "my-layer").Filters[0],
						pluginutils.CreateStringEqualsFilter("versionNumber", "1").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":                     core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:my-layer"),
				"version":                 core.MappingNodeFromInt(1),
				"layerVersionArn":         core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:my-layer:1"),
				"description":             core.MappingNodeFromString("My test layer"),
				"licenseInfo":             core.MappingNodeFromString("MIT"),
				"createdDate":             core.MappingNodeFromString("2023-01-01T00:00:00.000Z"),
				"compatibleRuntimes":      {Items: []*core.MappingNode{core.MappingNodeFromString("python3.9"), core.MappingNodeFromString("python3.10")}},
				"compatibleArchitectures": {Items: []*core.MappingNode{core.MappingNodeFromString("x86_64")}},
				"content": {
					Fields: map[string]*core.MappingNode{
						"codeSha256": core.MappingNodeFromString("abc123"),
						"codeSize":   core.MappingNodeFromInt(1024),
						"location":   core.MappingNodeFromString("https://example.com/layer.zip"),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createMinimalLayerVersionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) LayerVersionDataSourceFetchTestCase {
	return LayerVersionDataSourceFetchTestCase{
		Name: "successfully fetches minimal layer version",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionOutput(&lambda.GetLayerVersionOutput{
				LayerArn: aws.String("arn:aws:lambda:us-west-2:123456789012:layer:minimal-layer"),
				Version:  2,
				Content: &types.LayerVersionContentOutput{
					CodeSize: 512,
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
						pluginutils.CreateStringEqualsFilter("layerName", "minimal-layer").Filters[0],
						pluginutils.CreateStringEqualsFilter("versionNumber", "2").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":     core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:minimal-layer"),
				"version": core.MappingNodeFromInt(2),
				"content": {
					Fields: map[string]*core.MappingNode{
						"codeSize": core.MappingNodeFromInt(512),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createLayerVersionWithAllOptionalFieldsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) LayerVersionDataSourceFetchTestCase {
	return LayerVersionDataSourceFetchTestCase{
		Name: "successfully fetches layer version with all optional fields",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionOutput(&lambda.GetLayerVersionOutput{
				LayerArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:layer:full-layer"),
				LayerVersionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:layer:full-layer:3"),
				Version:         3,
				Description:     aws.String("Full layer with all fields"),
				LicenseInfo:     aws.String("Apache-2.0"),
				CreatedDate:     aws.String("2023-06-15T12:30:00.000Z"),
				CompatibleRuntimes: []types.Runtime{
					types.RuntimeNodejs18x,
					types.RuntimePython39,
				},
				CompatibleArchitectures: []types.Architecture{
					types.ArchitectureX8664,
					types.ArchitectureArm64,
				},
				Content: &types.LayerVersionContentOutput{
					CodeSha256:               aws.String("def456"),
					CodeSize:                 2048,
					Location:                 aws.String("https://example.com/full-layer.zip"),
					SigningJobArn:            aws.String("arn:aws:signer:us-west-2:123456789012:signing-job/signing-job-id"),
					SigningProfileVersionArn: aws.String("arn:aws:signer:us-west-2:123456789012:signing-profile/profile/version"),
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
						pluginutils.CreateStringEqualsFilter("layerName", "full-layer").Filters[0],
						pluginutils.CreateStringEqualsFilter("versionNumber", "3").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":                     core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:full-layer"),
				"version":                 core.MappingNodeFromInt(3),
				"layerVersionArn":         core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:full-layer:3"),
				"description":             core.MappingNodeFromString("Full layer with all fields"),
				"licenseInfo":             core.MappingNodeFromString("Apache-2.0"),
				"createdDate":             core.MappingNodeFromString("2023-06-15T12:30:00.000Z"),
				"compatibleRuntimes":      {Items: []*core.MappingNode{core.MappingNodeFromString("nodejs18.x"), core.MappingNodeFromString("python3.9")}},
				"compatibleArchitectures": {Items: []*core.MappingNode{core.MappingNodeFromString("x86_64"), core.MappingNodeFromString("arm64")}},
				"content": {
					Fields: map[string]*core.MappingNode{
						"codeSha256":               core.MappingNodeFromString("def456"),
						"codeSize":                 core.MappingNodeFromInt(2048),
						"location":                 core.MappingNodeFromString("https://example.com/full-layer.zip"),
						"signingJobArn":            core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:signing-job/signing-job-id"),
						"signingProfileVersionArn": core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:signing-profile/profile/version"),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createLayerVersionWithLayerARNTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) LayerVersionDataSourceFetchTestCase {
	layerARN := "arn:aws:lambda:us-west-2:123456789012:layer:shared-layer"
	return LayerVersionDataSourceFetchTestCase{
		Name: "successfully fetches layer version using layer ARN",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionOutput(&lambda.GetLayerVersionOutput{
				LayerArn:        aws.String(layerARN),
				LayerVersionArn: aws.String(layerARN + ":5"),
				Version:         5,
				Content: &types.LayerVersionContentOutput{
					CodeSize: 3072,
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
						pluginutils.CreateStringEqualsFilter("layerName", layerARN).Filters[0],
						pluginutils.CreateStringEqualsFilter("versionNumber", "5").Filters[0],
					},
				},
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":             core.MappingNodeFromString(layerARN),
				"version":         core.MappingNodeFromInt(5),
				"layerVersionArn": core.MappingNodeFromString(layerARN + ":5"),
				"content": {
					Fields: map[string]*core.MappingNode{
						"codeSize": core.MappingNodeFromInt(3072),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createLayerVersionFetchErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) LayerVersionDataSourceFetchTestCase {
	return LayerVersionDataSourceFetchTestCase{
		Name: "handles get layer version error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionError(errors.New("ResourceNotFoundException")),
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
						pluginutils.CreateStringEqualsFilter("layerName", "non-existent-layer").Filters[0],
						pluginutils.CreateStringEqualsFilter("versionNumber", "1").Filters[0],
					},
				},
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "failed to get Lambda layer version",
	}
}

func createLayerVersionMissingLayerNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) LayerVersionDataSourceFetchTestCase {
	return LayerVersionDataSourceFetchTestCase{
		Name:           "handles missing layer name filter",
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
						pluginutils.CreateStringEqualsFilter("versionNumber", "1").Filters[0],
					},
				},
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "layerName filter is required",
	}
}

func createLayerVersionMissingVersionNumberTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) LayerVersionDataSourceFetchTestCase {
	return LayerVersionDataSourceFetchTestCase{
		Name:           "handles missing version number filter",
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
						pluginutils.CreateStringEqualsFilter("layerName", "my-layer").Filters[0],
					},
				},
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "versionNumber filter is required",
	}
}
