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

type LambdaAliasResourceCreateSuite struct {
	suite.Suite
}

func (s *LambdaAliasResourceCreateSuite) Test_create_lambda_alias() {
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
		createBasicAliasTestCase(providerCtx, loader),
		createAliasWithDescriptionTestCase(providerCtx, loader),
		createAliasWithRoutingConfigTestCase(providerCtx, loader),
		createAliasWithProvisionedConcurrencyTestCase(providerCtx, loader),
		createComplexAliasTestCase(providerCtx, loader),
		createAliasFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		AliasResource,
		&s.Suite,
	)
}

func createBasicAliasTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:PROD"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateAliasOutput(&lambda.CreateAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("PROD"),
			FunctionVersion: aws.String("1"),
			Description:     aws.String("Production alias"),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("PROD"),
			"functionVersion": core.MappingNodeFromString("1"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create basic alias",
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
			ResourceID: "test-alias-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-alias-id",
					ResourceName: "TestAlias",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.name",
					},
					{
						FieldPath: "spec.functionVersion",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.aliasArn": core.MappingNodeFromString(aliasArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateAlias": &lambda.CreateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("PROD"),
				FunctionVersion: aws.String("1"),
			},
		},
	}
}

func createAliasWithDescriptionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:STAGING"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateAliasOutput(&lambda.CreateAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("STAGING"),
			FunctionVersion: aws.String("2"),
			Description:     aws.String("Staging environment alias"),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("STAGING"),
			"functionVersion": core.MappingNodeFromString("2"),
			"description":     core.MappingNodeFromString("Staging environment alias"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create alias with description",
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
			ResourceID: "test-alias-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-alias-id",
					ResourceName: "TestAlias",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.name",
					},
					{
						FieldPath: "spec.functionVersion",
					},
					{
						FieldPath: "spec.description",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.aliasArn": core.MappingNodeFromString(aliasArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateAlias": &lambda.CreateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("STAGING"),
				FunctionVersion: aws.String("2"),
				Description:     aws.String("Staging environment alias"),
			},
		},
	}
}

func createAliasWithRoutingConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:CANARY"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateAliasOutput(&lambda.CreateAliasOutput{
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

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("CANARY"),
			"functionVersion": core.MappingNodeFromString("3"),
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
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create alias with routing config",
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
			ResourceID: "test-alias-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-alias-id",
					ResourceName: "TestAlias",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.name",
					},
					{
						FieldPath: "spec.functionVersion",
					},
					{
						FieldPath: "spec.routingConfig",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.aliasArn": core.MappingNodeFromString(aliasArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateAlias": &lambda.CreateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("CANARY"),
				FunctionVersion: aws.String("3"),
				RoutingConfig: &types.AliasRoutingConfiguration{
					AdditionalVersionWeights: map[string]float64{
						"2": 0.1,
					},
				},
			},
		},
	}
}

func createAliasWithProvisionedConcurrencyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:HIGHPERF"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateAliasOutput(&lambda.CreateAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("HIGHPERF"),
			FunctionVersion: aws.String("4"),
		}),
		lambdamock.WithPutProvisionedConcurrencyConfigOutput(&lambda.PutProvisionedConcurrencyConfigOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("HIGHPERF"),
			"functionVersion": core.MappingNodeFromString("4"),
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(50),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create alias with provisioned concurrency",
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
			ResourceID: "test-alias-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-alias-id",
					ResourceName: "TestAlias",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.name",
					},
					{
						FieldPath: "spec.functionVersion",
					},
					{
						FieldPath: "spec.provisionedConcurrencyConfig",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.aliasArn": core.MappingNodeFromString(aliasArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateAlias": &lambda.CreateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("HIGHPERF"),
				FunctionVersion: aws.String("4"),
			},
			"PutProvisionedConcurrencyConfig": &lambda.PutProvisionedConcurrencyConfigInput{
				FunctionName:                    aws.String(aliasArn),
				ProvisionedConcurrentExecutions: aws.Int32(50),
			},
		},
	}
}

func createComplexAliasTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:COMPLEX"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateAliasOutput(&lambda.CreateAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("COMPLEX"),
			FunctionVersion: aws.String("5"),
			Description:     aws.String("Complex alias with all features"),
			RoutingConfig: &types.AliasRoutingConfiguration{
				AdditionalVersionWeights: map[string]float64{
					"4": 0.2,
					"3": 0.1,
				},
			},
		}),
		lambdamock.WithPutProvisionedConcurrencyConfigOutput(&lambda.PutProvisionedConcurrencyConfigOutput{}),
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

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create complex alias with all features",
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
			ResourceID: "test-alias-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-alias-id",
					ResourceName: "TestAlias",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.name",
					},
					{
						FieldPath: "spec.functionVersion",
					},
					{
						FieldPath: "spec.description",
					},
					{
						FieldPath: "spec.routingConfig",
					},
					{
						FieldPath: "spec.provisionedConcurrencyConfig",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.aliasArn": core.MappingNodeFromString(aliasArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateAlias": &lambda.CreateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("COMPLEX"),
				FunctionVersion: aws.String("5"),
				Description:     aws.String("Complex alias with all features"),
				RoutingConfig: &types.AliasRoutingConfiguration{
					AdditionalVersionWeights: map[string]float64{
						"4": 0.2,
						"3": 0.1,
					},
				},
			},
			"PutProvisionedConcurrencyConfig": &lambda.PutProvisionedConcurrencyConfigInput{
				FunctionName:                    aws.String(aliasArn),
				ProvisionedConcurrentExecutions: aws.Int32(100),
			},
		},
	}
}

func createAliasFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateAliasError(fmt.Errorf("failed to create alias")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("FAIL"),
			"functionVersion": core.MappingNodeFromString("1"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create alias failure",
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
			ResourceID: "test-alias-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-alias-id",
					ResourceName: "TestAlias",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.name",
					},
					{
						FieldPath: "spec.functionVersion",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"CreateAlias": &lambda.CreateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("FAIL"),
				FunctionVersion: aws.String("1"),
			},
		},
	}
}

func TestLambdaAliasResourceCreate(t *testing.T) {
	suite.Run(t, new(LambdaAliasResourceCreateSuite))
}
