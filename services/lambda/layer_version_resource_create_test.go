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
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaLayerVersionResourceCreateSuite struct {
	suite.Suite
}

func (s *LambdaLayerVersionResourceCreateSuite) Test_create_lambda_layer_version() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		createBasicLayerVersionTestCase(providerCtx, loader),
		createLayerVersionWithAllOptionsTestCase(providerCtx, loader),
		createLayerVersionFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		LayerVersionResource,
		&s.Suite,
	)
}

func createBasicLayerVersionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	layerArn := "arn:aws:lambda:us-west-2:123456789012:layer:test-layer"
	layerVersionArn := layerArn + ":1"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPublishLayerVersionOutput(&lambda.PublishLayerVersionOutput{
			LayerArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:layer:test-layer"),
			LayerVersionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"),
			Version:         1,
			CreatedDate:     aws.String("2023-12-01T12:00:00.000Z"),
		}),
	)

	// Create test data for layer version creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerName": core.MappingNodeFromString("test-layer"),
			"content": {
				Fields: map[string]*core.MappingNode{
					"s3Bucket": core.MappingNodeFromString("my-bucket"),
					"s3Key":    core.MappingNodeFromString("my-layer.zip"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create basic layer version",
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
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-layer-version-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-layer-version-id",
					ResourceName: "TestLayerVersion",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/layerVersion",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.layerName",
					},
					{
						FieldPath: "spec.content",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.layerArn":        core.MappingNodeFromString(layerArn),
				"spec.layerVersionArn": core.MappingNodeFromString(layerVersionArn),
				"spec.version":         core.MappingNodeFromInt(1),
				"spec.createdDate":     core.MappingNodeFromString("2023-12-01T12:00:00.000Z"),
			},
		},
		SaveActionsCalled: map[string]any{
			"PublishLayerVersion": &lambda.PublishLayerVersionInput{
				LayerName: aws.String("test-layer"),
				Content: &types.LayerVersionContentInput{
					S3Bucket: aws.String("my-bucket"),
					S3Key:    aws.String("my-layer.zip"),
				},
			},
		},
	}
}

func createLayerVersionWithAllOptionsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	layerArn := "arn:aws:lambda:us-west-2:123456789012:layer:comprehensive-layer"
	layerVersionArn := layerArn + ":2"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPublishLayerVersionOutput(&lambda.PublishLayerVersionOutput{
			LayerArn:        aws.String("arn:aws:lambda:us-west-2:123456789012:layer:comprehensive-layer"),
			LayerVersionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:layer:comprehensive-layer:2"),
			Version:         2,
			CreatedDate:     aws.String("2023-12-01T12:30:00.000Z"),
		}),
	)

	// Create test data for layer version creation with all options
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerName":   core.MappingNodeFromString("comprehensive-layer"),
			"description": core.MappingNodeFromString("A comprehensive test layer"),
			"licenseInfo": core.MappingNodeFromString("MIT"),
			"content": {
				Fields: map[string]*core.MappingNode{
					"s3Bucket":        core.MappingNodeFromString("my-bucket"),
					"s3Key":           core.MappingNodeFromString("comprehensive-layer.zip"),
					"s3ObjectVersion": core.MappingNodeFromString("version123"),
				},
			},
			"compatibleRuntimes": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("python3.9"),
					core.MappingNodeFromString("python3.10"),
				},
			},
			"compatibleArchitectures": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("x86_64"),
					core.MappingNodeFromString("arm64"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create layer version with all options",
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
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-layer-version-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-layer-version-id",
					ResourceName: "TestLayerVersion",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/layerVersion",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.layerName",
					},
					{
						FieldPath: "spec.description",
					},
					{
						FieldPath: "spec.licenseInfo",
					},
					{
						FieldPath: "spec.content",
					},
					{
						FieldPath: "spec.compatibleRuntimes",
					},
					{
						FieldPath: "spec.compatibleArchitectures",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.layerArn":        core.MappingNodeFromString(layerArn),
				"spec.layerVersionArn": core.MappingNodeFromString(layerVersionArn),
				"spec.version":         core.MappingNodeFromInt(2),
				"spec.createdDate":     core.MappingNodeFromString("2023-12-01T12:30:00.000Z"),
			},
		},
		SaveActionsCalled: map[string]any{
			"PublishLayerVersion": &lambda.PublishLayerVersionInput{
				LayerName:   aws.String("comprehensive-layer"),
				Description: aws.String("A comprehensive test layer"),
				LicenseInfo: aws.String("MIT"),
				Content: &types.LayerVersionContentInput{
					S3Bucket:        aws.String("my-bucket"),
					S3Key:           aws.String("comprehensive-layer.zip"),
					S3ObjectVersion: aws.String("version123"),
				},
				CompatibleRuntimes: []types.Runtime{
					types.Runtime("python3.9"),
					types.Runtime("python3.10"),
				},
				CompatibleArchitectures: []types.Architecture{
					types.Architecture("x86_64"),
					types.Architecture("arm64"),
				},
			},
		},
	}
}

func createLayerVersionFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPublishLayerVersionError(fmt.Errorf("failed to publish layer version")),
	)

	// Create test data for layer version creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerName": core.MappingNodeFromString("test-layer"),
			"content": {
				Fields: map[string]*core.MappingNode{
					"s3Bucket": core.MappingNodeFromString("my-bucket"),
					"s3Key":    core.MappingNodeFromString("my-layer.zip"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create layer version failure",
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
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-layer-version-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-layer-version-id",
					ResourceName: "TestLayerVersion",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/layerVersion",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.layerName",
					},
					{
						FieldPath: "spec.content",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"PublishLayerVersion": &lambda.PublishLayerVersionInput{
				LayerName: aws.String("test-layer"),
				Content: &types.LayerVersionContentInput{
					S3Bucket: aws.String("my-bucket"),
					S3Key:    aws.String("my-layer.zip"),
				},
			},
		},
	}
}

func TestLambdaLayerVersionResourceCreate(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionResourceCreateSuite))
}
