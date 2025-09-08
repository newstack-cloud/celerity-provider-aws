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
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaLayerVersionPermissionsResourceCreateSuite struct {
	suite.Suite
}

func (s *LambdaLayerVersionPermissionsResourceCreateSuite) Test_create_lambda_layer_version_permissions() {
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
		createBasicLayerVersionPermissionTestCase(providerCtx, loader),
		createLayerVersionPermissionWithOrganizationTestCase(providerCtx, loader),
		createLayerVersionPermissionFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		LayerVersionPermissionResource,
		&s.Suite,
	)
}

func createBasicLayerVersionPermissionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	statementJson := `{"Sid":"test-statement","Effect":"Allow","Principal":{"AWS":"arn:aws:iam::123456789012:root"},"Action":"lambda:GetLayerVersion","Resource":"arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"}`

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithAddLayerVersionPermissionOutput(&lambda.AddLayerVersionPermissionOutput{
			Statement:  aws.String(statementJson),
			RevisionId: aws.String("revision-123"),
		}),
	)

	// Create test data for layer version permission creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"),
			"statementId":     core.MappingNodeFromString("test-statement"),
			"action":          core.MappingNodeFromString("lambda:GetLayerVersion"),
			"principal":       core.MappingNodeFromString("123456789012"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create basic layer version permission",
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
			ResourceID: "test-layer-version-permission-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-layer-version-permission-id",
					ResourceName: "TestLayerVersionPermission",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/layerVersionPermission",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.layerVersionArn",
					},
					{
						FieldPath: "spec.statementId",
					},
					{
						FieldPath: "spec.action",
					},
					{
						FieldPath: "spec.principal",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"id": core.MappingNodeFromString("test-layer:1#test-statement"),
			},
		},
		SaveActionsCalled: map[string]any{
			"AddLayerVersionPermission": &lambda.AddLayerVersionPermissionInput{
				LayerName:     aws.String("test-layer"),
				VersionNumber: aws.Int64(1),
				StatementId:   aws.String("test-statement"),
				Action:        aws.String("lambda:GetLayerVersion"),
				Principal:     aws.String("123456789012"),
			},
		},
	}
}

func createLayerVersionPermissionWithOrganizationTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	statementJson := `{"Sid":"org-statement","Effect":"Allow","Principal":"*","Action":"lambda:GetLayerVersion","Resource":"arn:aws:lambda:us-west-2:123456789012:layer:my-layer:2","Condition":{"StringEquals":{"aws:PrincipalOrgID":"o-abc123defg"}}}`

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithAddLayerVersionPermissionOutput(&lambda.AddLayerVersionPermissionOutput{
			Statement:  aws.String(statementJson),
			RevisionId: aws.String("revision-456"),
		}),
	)

	// Create test data for layer version permission with organization
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:2"),
			"statementId":     core.MappingNodeFromString("org-statement"),
			"action":          core.MappingNodeFromString("lambda:GetLayerVersion"),
			"principal":       core.MappingNodeFromString("*"),
			"organizationId":  core.MappingNodeFromString("o-abc123defg"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create layer version permission with organization",
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
			ResourceID: "test-layer-version-permission-org-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-layer-version-permission-org-id",
					ResourceName: "TestLayerVersionPermissionOrg",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/layerVersionPermission",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.layerVersionArn",
					},
					{
						FieldPath: "spec.statementId",
					},
					{
						FieldPath: "spec.action",
					},
					{
						FieldPath: "spec.principal",
					},
					{
						FieldPath: "spec.organizationId",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"id": core.MappingNodeFromString("test-layer:2#org-statement"),
			},
		},
		SaveActionsCalled: map[string]any{
			"AddLayerVersionPermission": &lambda.AddLayerVersionPermissionInput{
				LayerName:      aws.String("test-layer"),
				VersionNumber:  aws.Int64(2),
				StatementId:    aws.String("org-statement"),
				Action:         aws.String("lambda:GetLayerVersion"),
				Principal:      aws.String("*"),
				OrganizationId: aws.String("o-abc123defg"),
			},
		},
	}
}

func createLayerVersionPermissionFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithAddLayerVersionPermissionError(fmt.Errorf("failed to add layer version permission")),
	)

	// Create test data for layer version permission creation failure
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:nonexistent-layer:1"),
			"statementId":     core.MappingNodeFromString("fail-statement"),
			"action":          core.MappingNodeFromString("lambda:GetLayerVersion"),
			"principal":       core.MappingNodeFromString("123456789012"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create layer version permission failure",
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
			ResourceID: "test-layer-version-permission-fail-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-layer-version-permission-fail-id",
					ResourceName: "TestLayerVersionPermissionFail",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/layerVersionPermission",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.layerVersionArn",
					},
					{
						FieldPath: "spec.statementId",
					},
					{
						FieldPath: "spec.action",
					},
					{
						FieldPath: "spec.principal",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"AddLayerVersionPermission": &lambda.AddLayerVersionPermissionInput{
				LayerName:     aws.String("nonexistent-layer"),
				VersionNumber: aws.Int64(1),
				StatementId:   aws.String("fail-statement"),
				Action:        aws.String("lambda:GetLayerVersion"),
				Principal:     aws.String("123456789012"),
			},
		},
	}
}

func TestLambdaLayerVersionPermissionsResourceCreate(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionPermissionsResourceCreateSuite))
}
