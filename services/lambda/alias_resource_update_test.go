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
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaAliasResourceUpdateSuite struct {
	suite.Suite
}

func (s *LambdaAliasResourceUpdateSuite) Test_update_lambda_alias() {
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
		updateAliasDescriptionTestCase(providerCtx, loader),
		updateAliasFunctionVersionTestCase(providerCtx, loader),
		updateAliasRoutingConfigTestCase(providerCtx, loader),
		updateAliasProvisionedConcurrencyTestCase(providerCtx, loader),
		updateAliasComplexTestCase(providerCtx, loader),
		updateAliasFailureTestCase(providerCtx, loader),
		recreateAliasOnNameOrFunctionNameChangeTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		AliasResource,
		&s.Suite,
	)
}

func updateAliasDescriptionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:PROD"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateAliasOutput(&lambda.UpdateAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("PROD"),
			FunctionVersion: aws.String("1"),
			Description:     aws.String("Updated production alias"),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("PROD"),
			"functionVersion": core.MappingNodeFromString("1"),
			"description":     core.MappingNodeFromString("Updated production alias"),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("PROD"),
			"functionVersion": core.MappingNodeFromString("1"),
			"description":     core.MappingNodeFromString("Old production alias"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update alias description",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-alias-id",
						Name:       "TestAlias",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
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
			"UpdateAlias": &lambda.UpdateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("PROD"),
				FunctionVersion: aws.String("1"),
				Description:     aws.String("Updated production alias"),
			},
		},
	}
}

func updateAliasFunctionVersionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:STAGING"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateAliasOutput(&lambda.UpdateAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("STAGING"),
			FunctionVersion: aws.String("3"),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("STAGING"),
			"functionVersion": core.MappingNodeFromString("3"),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("STAGING"),
			"functionVersion": core.MappingNodeFromString("2"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update alias function version",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-alias-id",
						Name:       "TestAlias",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
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
			"UpdateAlias": &lambda.UpdateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("STAGING"),
				FunctionVersion: aws.String("3"),
			},
		},
	}
}

func updateAliasRoutingConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:CANARY"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateAliasOutput(&lambda.UpdateAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("CANARY"),
			FunctionVersion: aws.String("4"),
			RoutingConfig: &types.AliasRoutingConfiguration{
				AdditionalVersionWeights: map[string]float64{
					"3": 0.2,
				},
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("CANARY"),
			"functionVersion": core.MappingNodeFromString("4"),
			"routingConfig": {
				Fields: map[string]*core.MappingNode{
					"additionalVersionWeights": {
						Fields: map[string]*core.MappingNode{
							"3": core.MappingNodeFromFloat(0.2),
						},
					},
				},
			},
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("CANARY"),
			"functionVersion": core.MappingNodeFromString("3"),
			"routingConfig": {
				Fields: map[string]*core.MappingNode{
					"additionalVersionWeights": {
						Fields: map[string]*core.MappingNode{
							"3": core.MappingNodeFromFloat(0.2),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update alias routing config",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-alias-id",
						Name:       "TestAlias",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
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
			"UpdateAlias": &lambda.UpdateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("CANARY"),
				FunctionVersion: aws.String("4"),
				RoutingConfig: &types.AliasRoutingConfiguration{
					AdditionalVersionWeights: map[string]float64{
						"3": 0.2,
					},
				},
			},
		},
	}
}

func updateAliasProvisionedConcurrencyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:HIGHPERF"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateAliasOutput(&lambda.UpdateAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("HIGHPERF"),
			FunctionVersion: aws.String("2"),
		}),
		lambdamock.WithPutProvisionedConcurrencyConfigOutput(&lambda.PutProvisionedConcurrencyConfigOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("HIGHPERF"),
			"functionVersion": core.MappingNodeFromString("2"),
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(75),
				},
			},
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("HIGHPERF"),
			"functionVersion": core.MappingNodeFromString("1"),
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(50),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update alias provisioned concurrency",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-alias-id",
						Name:       "TestAlias",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
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
			"UpdateAlias": &lambda.UpdateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("HIGHPERF"),
				FunctionVersion: aws.String("2"),
			},
			"PutProvisionedConcurrencyConfig": &lambda.PutProvisionedConcurrencyConfigInput{
				FunctionName:                    aws.String(aliasArn),
				ProvisionedConcurrentExecutions: aws.Int32(75),
			},
		},
	}
}

func updateAliasComplexTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	aliasArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function:COMPLEX"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateAliasOutput(&lambda.UpdateAliasOutput{
			AliasArn:        aws.String(aliasArn),
			Name:            aws.String("COMPLEX"),
			FunctionVersion: aws.String("6"),
			Description:     aws.String("Updated complex alias"),
			RoutingConfig: &types.AliasRoutingConfiguration{
				AdditionalVersionWeights: map[string]float64{
					"5": 0.3,
					"4": 0.2,
				},
			},
		}),
		lambdamock.WithPutProvisionedConcurrencyConfigOutput(&lambda.PutProvisionedConcurrencyConfigOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("COMPLEX"),
			"functionVersion": core.MappingNodeFromString("6"),
			"description":     core.MappingNodeFromString("Updated complex alias"),
			"routingConfig": {
				Fields: map[string]*core.MappingNode{
					"additionalVersionWeights": {
						Fields: map[string]*core.MappingNode{
							"5": core.MappingNodeFromFloat(0.3),
							"4": core.MappingNodeFromFloat(0.2),
						},
					},
				},
			},
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(150),
				},
			},
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("COMPLEX"),
			"functionVersion": core.MappingNodeFromString("5"),
			"description":     core.MappingNodeFromString("Old complex alias"),
			"routingConfig": {
				Fields: map[string]*core.MappingNode{
					"additionalVersionWeights": {
						Fields: map[string]*core.MappingNode{
							"5": core.MappingNodeFromFloat(0.3),
							"4": core.MappingNodeFromFloat(0.2),
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
		Name: "update complex alias with all features",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-alias-id",
						Name:       "TestAlias",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
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
			"UpdateAlias": &lambda.UpdateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("COMPLEX"),
				FunctionVersion: aws.String("6"),
				Description:     aws.String("Updated complex alias"),
				RoutingConfig: &types.AliasRoutingConfiguration{
					AdditionalVersionWeights: map[string]float64{
						"5": 0.3,
						"4": 0.2,
					},
				},
			},
			"PutProvisionedConcurrencyConfig": &lambda.PutProvisionedConcurrencyConfigInput{
				FunctionName:                    aws.String(aliasArn),
				ProvisionedConcurrentExecutions: aws.Int32(150),
			},
		},
	}
}

func updateAliasFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateAliasError(fmt.Errorf("failed to update alias")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("FAIL"),
			"functionVersion": core.MappingNodeFromString("2"),
			"description":     core.MappingNodeFromString("This should fail"),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("FAIL"),
			"functionVersion": core.MappingNodeFromString("1"),
			"description":     core.MappingNodeFromString("Old failed alias"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update alias failure",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-alias-id",
						Name:       "TestAlias",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.description",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"UpdateAlias": &lambda.UpdateAliasInput{
				FunctionName:    aws.String("test-function"),
				Name:            aws.String("FAIL"),
				FunctionVersion: aws.String("2"),
				Description:     aws.String("This should fail"),
			},
		},
	}
}

func recreateAliasOnNameOrFunctionNameChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	oldAliasArn := "arn:aws:lambda:us-west-2:123456789012:function:old-function:old-alias"
	newAliasArn := "arn:aws:lambda:us-west-2:123456789012:function:new-function:new-alias"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteAliasOutput(&lambda.DeleteAliasOutput{}),
		lambdamock.WithCreateAliasOutput(&lambda.CreateAliasOutput{
			AliasArn:        aws.String(newAliasArn),
			Name:            aws.String("new-alias"),
			FunctionVersion: aws.String("$LATEST"),
			Description:     aws.String("Updated alias"),
		}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"aliasArn":        core.MappingNodeFromString(oldAliasArn),
			"functionName":    core.MappingNodeFromString("old-function"),
			"name":            core.MappingNodeFromString("old-alias"),
			"functionVersion": core.MappingNodeFromString("$LATEST"),
			"description":     core.MappingNodeFromString("Original alias"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("new-function"),
			"name":            core.MappingNodeFromString("new-alias"),
			"functionVersion": core.MappingNodeFromString("$LATEST"),
			"description":     core.MappingNodeFromString("Updated alias"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "recreate alias on functionName or name change",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-alias-id",
						Name:       "TestAlias",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/alias",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
						PrevValue: core.MappingNodeFromString("old-function"),
						NewValue:  core.MappingNodeFromString("new-function"),
					},
					{
						FieldPath: "spec.name",
						PrevValue: core.MappingNodeFromString("old-alias"),
						NewValue:  core.MappingNodeFromString("new-alias"),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.aliasArn": core.MappingNodeFromString(newAliasArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"DeleteAlias": &lambda.DeleteAliasInput{
				FunctionName: aws.String("old-function"),
				Name:         aws.String("old-alias"),
			},
			"CreateAlias": &lambda.CreateAliasInput{
				FunctionName:    aws.String("new-function"),
				Name:            aws.String("new-alias"),
				FunctionVersion: aws.String("$LATEST"),
				Description:     aws.String("Updated alias"),
			},
		},
	}
}

func TestLambdaAliasResourceUpdate(t *testing.T) {
	suite.Run(t, new(LambdaAliasResourceUpdateSuite))
}
