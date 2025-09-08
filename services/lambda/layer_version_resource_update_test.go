//go:build unit

package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaLayerVersionResourceUpdateSuite struct {
	suite.Suite
}

func (s *LambdaLayerVersionResourceUpdateSuite) Test_update_lambda_layer_version() {
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
		createLayerVersionDescriptionUpdateTestCase(providerCtx, loader),
		createLayerVersionContentUpdateTestCase(providerCtx, loader),
		createLayerVersionCompatibleRuntimesUpdateTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		LayerVersionResource,
		&s.Suite,
	)
}

func createLayerVersionDescriptionUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	layerVersionArn := "arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"

	// No service mock calls needed since updates should fail immediately
	service := lambdamock.CreateLambdaServiceMock()

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerName":       core.MappingNodeFromString("test-layer"),
			"description":     core.MappingNodeFromString("Original description"),
			"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
			"version":         core.MappingNodeFromInt(1),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerName":       core.MappingNodeFromString("test-layer"),
			"description":     core.MappingNodeFromString("Updated description"),
			"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
			"version":         core.MappingNodeFromInt(1),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "layer version description update should fail",
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
			InstanceID:      "test-instance-id",
			ResourceID:      "test-layer-version-id",
			ProviderContext: providerCtx,
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-layer-version-id",
					ResourceName: "TestLayerVersion",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-layer-version-id",
						Name:       "TestLayerVersion",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/layerVersion",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.description",
					},
				},
			},
		},
		ExpectError: true,
		// No update actions should be called since layer versions are immutable
	}
}

func createLayerVersionContentUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	layerVersionArn := "arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"

	// No service mock calls needed since updates should fail immediately
	service := lambdamock.CreateLambdaServiceMock()

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerName":       core.MappingNodeFromString("test-layer"),
			"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
			"version":         core.MappingNodeFromInt(1),
			"content": {
				Fields: map[string]*core.MappingNode{
					"s3Bucket": core.MappingNodeFromString("old-bucket"),
					"s3Key":    core.MappingNodeFromString("old-layer.zip"),
				},
			},
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerName":       core.MappingNodeFromString("test-layer"),
			"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
			"version":         core.MappingNodeFromInt(1),
			"content": {
				Fields: map[string]*core.MappingNode{
					"s3Bucket": core.MappingNodeFromString("new-bucket"),
					"s3Key":    core.MappingNodeFromString("new-layer.zip"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "layer version content update should fail",
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
			InstanceID:      "test-instance-id",
			ResourceID:      "test-layer-version-id",
			ProviderContext: providerCtx,
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-layer-version-id",
					ResourceName: "TestLayerVersion",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-layer-version-id",
						Name:       "TestLayerVersion",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/layerVersion",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.content.s3Bucket",
					},
					{
						FieldPath: "spec.content.s3Key",
					},
				},
			},
		},
		ExpectError: true,
		// No update actions should be called since layer versions are immutable
	}
}

func createLayerVersionCompatibleRuntimesUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	layerVersionArn := "arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"

	// No service mock calls needed since updates should fail immediately
	service := lambdamock.CreateLambdaServiceMock()

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerName":       core.MappingNodeFromString("test-layer"),
			"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
			"version":         core.MappingNodeFromInt(1),
			"compatibleRuntimes": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("python3.9"),
					core.MappingNodeFromString("python3.10"),
				},
			},
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerName":       core.MappingNodeFromString("test-layer"),
			"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
			"version":         core.MappingNodeFromInt(1),
			"compatibleRuntimes": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("python3.9"),
					core.MappingNodeFromString("python3.10"),
					core.MappingNodeFromString("python3.11"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "layer version compatible runtimes update should fail",
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
			InstanceID:      "test-instance-id",
			ResourceID:      "test-layer-version-id",
			ProviderContext: providerCtx,
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-layer-version-id",
					ResourceName: "TestLayerVersion",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-layer-version-id",
						Name:       "TestLayerVersion",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/layerVersion",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.compatibleRuntimes",
					},
				},
			},
		},
		ExpectError: true,
		// No update actions should be called since layer versions are immutable
	}
}

func TestLambdaLayerVersionResourceUpdate(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionResourceUpdateSuite))
}
