//go:build unit

package lambda

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
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

type LambdaLayerVersionResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaLayerVersionResourceGetExternalStateSuite) Test_get_external_state() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		getLayerVersionBasicStateTestCase(providerCtx, loader),
		getLayerVersionFullStateTestCase(providerCtx, loader),
		getLayerVersionNotFoundTestCase(providerCtx, loader),
		getLayerVersionErrorTestCase(providerCtx, loader),
		getLayerVersionInvalidArnTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		LayerVersionResource,
		&s.Suite,
	)
}

func TestLambdaLayerVersionResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionResourceGetExternalStateSuite))
}

func getLayerVersionBasicStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	layerVersionArn := "arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets basic layer version state",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionOutput(&lambda.GetLayerVersionOutput{
				LayerArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:layer:test-layer"),
				LayerVersionArn: aws.String(layerVersionArn),
				Version:         1,
				CreatedDate:     aws.String("2023-12-01T12:00:00.000Z"),
				Description:     aws.String("Basic test layer"),
				Content: &types.LayerVersionContentOutput{
					CodeSha256: aws.String("abc123def456"),
					CodeSize:   1024,
					Location:   aws.String("https://s3.amazonaws.com/test-bucket/layer.zip"),
				},
			}),
		),
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
					"layerName": core.MappingNodeFromString("test-layer"),
					"version":   core.MappingNodeFromInt(1),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"layerArn":        core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer"),
					"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
					"version":         core.MappingNodeFromInt(1),
					"createdDate":     core.MappingNodeFromString("2023-12-01T12:00:00.000Z"),
					"description":     core.MappingNodeFromString("Basic test layer"),
					"content": {
						Fields: map[string]*core.MappingNode{
							"codeSha256": core.MappingNodeFromString("abc123def456"),
							"codeSize":   core.MappingNodeFromInt(1024),
							"location":   core.MappingNodeFromString("https://s3.amazonaws.com/test-bucket/layer.zip"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func getLayerVersionFullStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	layerVersionArn := "arn:aws:lambda:us-west-2:123456789012:layer:comprehensive-layer:2"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets comprehensive layer version state",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionOutput(&lambda.GetLayerVersionOutput{
				LayerArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:layer:comprehensive-layer"),
				LayerVersionArn: aws.String(layerVersionArn),
				Version:         2,
				CreatedDate:     aws.String("2023-12-01T12:30:00.000Z"),
				Description:     aws.String("Comprehensive test layer"),
				LicenseInfo:     aws.String("MIT"),
				CompatibleRuntimes: []types.Runtime{
					types.Runtime("python3.9"),
					types.Runtime("python3.10"),
					types.Runtime("nodejs18.x"),
				},
				CompatibleArchitectures: []types.Architecture{
					types.Architecture("x86_64"),
					types.Architecture("arm64"),
				},
				Content: &types.LayerVersionContentOutput{
					CodeSha256:               aws.String("xyz789abc123"),
					CodeSize:                 2048,
					Location:                 aws.String("https://s3.amazonaws.com/test-bucket/comprehensive-layer.zip"),
					SigningJobArn:            aws.String("arn:aws:signer:us-west-2:123456789012:signing-job/test-job"),
					SigningProfileVersionArn: aws.String("arn:aws:signer:us-west-2:123456789012:signing-profile/test-profile/1"),
				},
			}),
		),
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
					"layerName": core.MappingNodeFromString("comprehensive-layer"),
					"version":   core.MappingNodeFromInt(2),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"layerArn":        core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:comprehensive-layer"),
					"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
					"version":         core.MappingNodeFromInt(2),
					"createdDate":     core.MappingNodeFromString("2023-12-01T12:30:00.000Z"),
					"description":     core.MappingNodeFromString("Comprehensive test layer"),
					"licenseInfo":     core.MappingNodeFromString("MIT"),
					"compatibleRuntimes": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("python3.9"),
							core.MappingNodeFromString("python3.10"),
							core.MappingNodeFromString("nodejs18.x"),
						},
					},
					"compatibleArchitectures": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("x86_64"),
							core.MappingNodeFromString("arm64"),
						},
					},
					"content": {
						Fields: map[string]*core.MappingNode{
							"codeSha256":               core.MappingNodeFromString("xyz789abc123"),
							"codeSize":                 core.MappingNodeFromInt(2048),
							"location":                 core.MappingNodeFromString("https://s3.amazonaws.com/test-bucket/comprehensive-layer.zip"),
							"signingJobArn":            core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:signing-job/test-job"),
							"signingProfileVersionArn": core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:signing-profile/test-profile/1"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func getLayerVersionNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	// Create a mock error that implements smithy.APIError
	notFoundError := &smithy.GenericAPIError{
		Code:    "ResourceNotFoundException",
		Message: "The resource you requested does not exist.",
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "handles layer version not found",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionError(notFoundError),
		),
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
					"layerName": core.MappingNodeFromString("non-existent-layer"),
					"version":   core.MappingNodeFromInt(1),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{Fields: make(map[string]*core.MappingNode)},
		},
		ExpectError: false,
	}
}

func getLayerVersionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "handles get layer version error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionError(errors.New("internal server error")),
		),
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
					"layerName": core.MappingNodeFromString("test-layer"),
					"version":   core.MappingNodeFromInt(1),
				},
			},
		},
		ExpectError: true,
	}
}

func getLayerVersionInvalidArnTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "handles missing version field in spec",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionOutput(&lambda.GetLayerVersionOutput{
				LayerArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:layer:test-layer"),
				LayerVersionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:0"),
				Version:         0,
				CreatedDate:     aws.String("2023-12-01T12:00:00.000Z"),
				Description:     aws.String("Test layer"),
				Content: &types.LayerVersionContentOutput{
					CodeSha256: aws.String("abc123def456"),
					CodeSize:   1024,
					Location:   aws.String("https://s3.amazonaws.com/test-bucket/layer.zip"),
				},
			}),
		),
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
					"layerName": core.MappingNodeFromString("test-layer"),
					// Missing version field - will default to 0
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"layerArn":        core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer"),
					"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:0"),
					"version":         core.MappingNodeFromInt(0),
					"createdDate":     core.MappingNodeFromString("2023-12-01T12:00:00.000Z"),
					"description":     core.MappingNodeFromString("Test layer"),
					"content": {
						Fields: map[string]*core.MappingNode{
							"codeSha256": core.MappingNodeFromString("abc123def456"),
							"codeSize":   core.MappingNodeFromInt(1024),
							"location":   core.MappingNodeFromString("https://s3.amazonaws.com/test-bucket/layer.zip"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}
