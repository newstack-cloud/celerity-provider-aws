//go:build unit

package lambdalinks

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
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionCodeSigningConfigLinkUpdateSuite struct {
	suite.Suite
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}
	linkCtx := plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {
				"region": core.ScalarFromString("us-west-2"),
			},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		s.createUpdateLinkFunctionTestCase(linkCtx, loader),
		s.createUpdateLinkCSCTestCase(linkCtx, loader),
		s.createUpdateLinkRemoveCSCFromFunctionTestCase(linkCtx, loader),
		s.createUpdateLinkErrorFuncMissingARNTestCase(linkCtx, loader),
		s.createUpdateLinkErrorCSCMissingARNTestCase(linkCtx, loader),
		s.createUpdateLinkErrorUpdateServiceErrorTestCase(linkCtx, loader),
		s.createUpdateLinkErrorRemoveServiceErrorTestCase(linkCtx, loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		FunctionCodeSigningConfigLink,
		&s.Suite,
	)
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) createUpdateLinkFunctionTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	service := lambdamock.CreateLambdaServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return service
	}

	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	cscARN := "arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-csc"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "Updates function with code signing config ARN reference",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-function",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(functionARN),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-csc",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"codeSigningConfigArn": core.MappingNodeFromString(cscARN),
						},
					},
				},
			},
			LinkContext: linkCtx,
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"test-function": {
						Fields: map[string]*core.MappingNode{
							"codeSigningConfigArn": core.MappingNodeFromString(cscARN),
						},
					},
				},
			},
			ResourceDataMappings: map[string]string{
				"test-function::spec.codeSigningConfigArn": "test-function.codeSigningConfigArn",
			},
		},
		UpdateActionsCalled: map[string]any{
			"PutFunctionCodeSigningConfig": &lambda.PutFunctionCodeSigningConfigInput{
				FunctionName:         aws.String(functionARN),
				CodeSigningConfigArn: aws.String(cscARN),
			},
		},
	}
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) createUpdateLinkCSCTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	service := lambdamock.CreateLambdaServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return service
	}

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "Returns empty link data as CSC resource is not updated for the link",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName:         "test-function",
				CurrentResourceState: &state.ResourceState{},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName:         "test-csc",
				CurrentResourceState: &state.ResourceState{},
			},
			LinkContext: linkCtx,
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
	}
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) createUpdateLinkRemoveCSCFromFunctionTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionCodeSigningConfigOutput(
			&lambda.DeleteFunctionCodeSigningConfigOutput{},
		),
	)
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return service
	}

	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "Removes code signing config from function",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeDestroy,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-function",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(functionARN),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-csc",
			},
			LinkContext: linkCtx,
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
		UpdateActionsCalled: map[string]any{
			"DeleteFunctionCodeSigningConfig": &lambda.DeleteFunctionCodeSigningConfigInput{
				FunctionName: aws.String(functionARN),
			},
		},
	}
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) createUpdateLinkErrorFuncMissingARNTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	service := lambdamock.CreateLambdaServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return service
	}

	cscARN := "arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-csc"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "Returns error if function ARN is missing from function resource spec",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-function",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							// Missing ARN field.
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-csc",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"codeSigningConfigArn": core.MappingNodeFromString(cscARN),
						},
					},
				},
			},
			LinkContext: linkCtx,
		},
		UpdateActionsNotCalled: []string{"PutFunctionCodeSigningConfig"},
		ExpectError:            true,
		ExpectedErrorMessage:   "function ARN could not be retrieved from function",
	}
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) createUpdateLinkErrorCSCMissingARNTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	service := lambdamock.CreateLambdaServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return service
	}

	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "Returns error if code signing config ARN is missing from CSC resource spec",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-function",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(functionARN),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-csc",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							// Missing codeSigningConfigArn field.
						},
					},
				},
			},
			LinkContext: linkCtx,
		},
		UpdateActionsNotCalled: []string{"PutFunctionCodeSigningConfig"},
		ExpectError:            true,
		ExpectedErrorMessage:   "code signing config ARN could not be retrieved from code signing config",
	}
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) createUpdateLinkErrorUpdateServiceErrorTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPutFunctionCodeSigningConfigError(fmt.Errorf("test error")),
	)
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return service
	}

	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	cscARN := "arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-csc"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "Returns error if service returns error when adding code signing config to function",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-function",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(functionARN),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-csc",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"codeSigningConfigArn": core.MappingNodeFromString(cscARN),
						},
					},
				},
			},
			LinkContext: linkCtx,
		},
		UpdateActionsCalled: map[string]any{
			"PutFunctionCodeSigningConfig": &lambda.PutFunctionCodeSigningConfigInput{
				FunctionName:         aws.String(functionARN),
				CodeSigningConfigArn: aws.String(cscARN),
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "test error",
	}
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) createUpdateLinkErrorRemoveServiceErrorTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionCodeSigningConfigError(fmt.Errorf("test error")),
	)
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return service
	}

	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "Returns error if service returns error when removing code signing config from function",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeDestroy,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-function",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(functionARN),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "test-csc",
			},
			LinkContext: linkCtx,
		},
		UpdateActionsCalled: map[string]any{
			"DeleteFunctionCodeSigningConfig": &lambda.DeleteFunctionCodeSigningConfigInput{
				FunctionName: aws.String(functionARN),
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "test error",
	}
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}
	linkCtx := plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {
				"region": core.ScalarFromString("us-west-2"),
			},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		s.createUpdateLinkIntermediaryResourcesTestCase(linkCtx, loader),
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		testCases,
		FunctionCodeSigningConfigLink,
		&s.Suite,
	)
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) createUpdateLinkIntermediaryResourcesTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	service := lambdamock.CreateLambdaServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return service
	}

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "Returns empty link data as there are no intermediary resources to create or update",
		ServiceFactoryA:                serviceFactory,
		ConfigStoreA:                   configStore,
		ServiceFactoryB:                serviceFactory,
		ConfigStoreB:                   configStore,
		IntermediariesServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			LinkContext:    linkCtx,
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
	}
}

func TestFunctionCodeSigningConfigLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionCodeSigningConfigLinkUpdateSuite))
}
