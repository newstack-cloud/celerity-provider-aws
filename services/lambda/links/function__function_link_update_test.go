//go:build unit

package lambdalinks

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

const (
	ffLinkCallerARN    = "arn:aws:lambda:us-west-2:123456789012:function:caller"
	ffLinkCalleeARN    = "arn:aws:lambda:us-west-2:123456789012:function:callee"
	ffLinkRoleARN      = "arn:aws:iam::123456789012:role/caller-role"
	ffLinkRoleName     = "caller-role"
	ffLinkRoleResource = "callerFunctionRole"
	ffLinkExecRole     = "callerFunctionExecutionRole"
	ffLinkEnvVarName   = "INVOKE_LAMBDA_FUNCTION_calleeFunction"
	ffLinkInvokeSID    = "InvokeLambdaFunctioncalleeFunction"
)

type FunctionFunctionLinkUpdateSuite struct {
	suite.Suite
}

// ffLinkFactory adapts the curried link constructor to the underlying
// LinkServiceDeps type expected by the test harness, binding a specific IAM
// service for the case under test. The link uses lambda + iam + ec2 services;
// the ec2 service is only exercised when the caller function has a VPC config,
// which none of these cases set, so a bare ec2 mock is fine.
func ffLinkFactory(
	iamSvc iamservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, lambdaservice.Service],
) provider.Link {
	build := FunctionFunctionLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
		func(c *aws.Config, pc provider.Context) ec2service.Service {
			return ec2mock.CreateEc2ServiceMock()
		},
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, lambdaservice.Service],
	) provider.Link {
		return build(FunctionToFunctionLinkDeps(deps))
	}
}

func ffLinkTestLinkContext() provider.LinkContext {
	return plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {"region": core.ScalarFromString("us-west-2")},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func ffLinkConfigStore(loader *testutils.MockAWSConfigLoader) pluginutils.ServiceConfigStore[*aws.Config] {
	return utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
}

func callerFunctionResourceInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "callerFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(ffLinkCallerARN),
			),
		},
	}
}

func calleeFunctionResourceInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "calleeFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(ffLinkCalleeARN),
			),
		},
	}
}

func calleeFunctionNoARNResourceInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName:         "calleeFunction",
		CurrentResourceState: &state.ResourceState{SpecData: core.MappingNodeFields()},
	}
}

func ffLinkFunctionWithEnvOutput(vars map[string]string) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(ffLinkCallerARN),
			Role:        aws.String(ffLinkRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{Variables: vars},
		},
	}
}

func ffLinkRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: ffLinkRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(ffLinkRoleName),
			"arn", core.MappingNodeFromString(ffLinkRoleARN),
		),
	}
}

func ffLinkLambdaSvc() lambdaservice.Service {
	return lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(ffLinkFunctionWithEnvOutput(map[string]string{})),
	)
}

func (s *FunctionFunctionLinkUpdateSuite) Test_link_update_linked_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		ffLinkAddEnvVarTestCase(loader),
		ffLinkDisableEnvVarPopulationTestCase(loader),
		ffLinkUpdateResourceBNoOpTestCase(loader),
		ffLinkRemoveEnvVarTestCase(loader),
		ffLinkMissingCalleeARNErrorTestCase(loader),
		ffLinkUpdateServiceErrorTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		ffLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

func ffLinkAddEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(ffLinkFunctionWithEnvOutput(map[string]string{"EXISTING": "val"})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "populates the invoke env var on the caller function when enabled",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            ffLinkConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            ffLinkConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      callerFunctionResourceInfo(),
			OtherResourceInfo: calleeFunctionResourceInfo(),
			LinkContext:       ffLinkTestLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"callerFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						// The default env var name is the empty string when no
						// envVarName annotation is provided (the implementation
						// stores the user-provided name, not the derived one).
						"", core.MappingNodeFromString(ffLinkCalleeARN),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				fmt.Sprintf(
					"callerFunction::spec.environment.variables[\"%s\"]",
					ffLinkEnvVarName,
				): fmt.Sprintf(
					"callerFunction.environmentVariables[\"%s\"]",
					ffLinkEnvVarName,
				),
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(ffLinkCallerARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{
						"EXISTING":       "val",
						ffLinkEnvVarName: ffLinkCalleeARN,
					},
				},
			},
		},
	}
}

func ffLinkDisableEnvVarPopulationTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	resourceInfo := callerFunctionResourceInfo()
	// Disable env var population for the callee function via the per-target
	// annotation, which gates UpdateResourceA.
	resourceInfo.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"aws.lambda.invoke.calleeFunction.populateEnvVars": core.MappingNodeFromBool(false),
				},
			},
		},
	}

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "is a no-op on the caller function when populateEnvVars is false",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            ffLinkConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            ffLinkConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      resourceInfo,
			OtherResourceInfo: calleeFunctionResourceInfo(),
			LinkContext:       ffLinkTestLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields(),
			ResourceDataMappings: map[string]string{},
		},
	}
}

func ffLinkUpdateResourceBNoOpTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "returns empty link data as the callee function is not updated for the link",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            ffLinkConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            ffLinkConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      calleeFunctionResourceInfo(),
			OtherResourceInfo: callerFunctionResourceInfo(),
			LinkContext:       ffLinkTestLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
	}
}

func ffLinkRemoveEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(ffLinkFunctionWithEnvOutput(map[string]string{
			"EXISTING":       "val",
			ffLinkEnvVarName: ffLinkCalleeARN,
		})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "removes the invoke env var on destroy",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            ffLinkConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            ffLinkConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      callerFunctionResourceInfo(),
			OtherResourceInfo: calleeFunctionResourceInfo(),
			LinkContext:       ffLinkTestLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"callerFunction",
				core.MappingNodeFields(),
			),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(ffLinkCallerARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{"EXISTING": "val"},
				},
			},
		},
	}
}

func ffLinkMissingCalleeARNErrorTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(ffLinkFunctionWithEnvOutput(map[string]string{})),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "returns error when the callee function ARN is missing",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            ffLinkConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            ffLinkConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      callerFunctionResourceInfo(),
			OtherResourceInfo: calleeFunctionNoARNResourceInfo(),
			LinkContext:       ffLinkTestLinkContext(),
		},
		ExpectError: true,
	}
}

func ffLinkUpdateServiceErrorTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(ffLinkFunctionWithEnvOutput(map[string]string{"EXISTING": "val"})),
		lambdamock.WithUpdateFunctionConfigurationError(errors.New("update failed")),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "returns error when the caller function update fails",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            ffLinkConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            ffLinkConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      callerFunctionResourceInfo(),
			OtherResourceInfo: calleeFunctionResourceInfo(),
			LinkContext:       ffLinkTestLinkContext(),
		},
		ExpectError: true,
	}
}

// UpdateIntermediaryResources manages the lambda:InvokeFunction permission on
// the caller function's execution role policy.
func (s *FunctionFunctionLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	createCase, createIam := ffLinkCreateIntermediaryTestCase(loader)
	updateCase, updateIam := ffLinkUpdateIntermediaryTestCase(loader)
	destroyCase, destroyIam := ffLinkDestroyIntermediaryTestCase(loader)
	errCase, errIam := ffLinkMissingRoleIntermediaryErrorTestCase(loader)

	cases := []struct {
		testCase plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			lambdaservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]
		iamSvc iamservice.Service
	}{
		{createCase, createIam},
		{updateCase, updateIam},
		{destroyCase, destroyIam},
		{errCase, errIam},
	}

	for _, c := range cases {
		plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
			[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
				*aws.Config,
				lambdaservice.Service,
				*aws.Config,
				lambdaservice.Service,
			]{c.testCase},
			ffLinkFactory(c.iamSvc),
			&s.Suite,
		)
	}
}

// matchPutInlineAccessPolicy verifies a PutRolePolicy targets the role's shared
// allocator inline policy and its document grants the link's invoke statement.
func matchPutInlineAccessPolicy(arg any) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != ffLinkRoleName ||
		aws.ToString(input.PolicyName) != linkutils.InlineAccessPolicyName() {
		return false
	}
	var doc struct {
		Statement []struct{ Sid string }
	}
	if err := json.Unmarshal([]byte(aws.ToString(input.PolicyDocument)), &doc); err != nil {
		return false
	}
	for _, statement := range doc.Statement {
		if statement.Sid == ffLinkInvokeSID {
			return true
		}
	}
	return false
}

// ffLinkMatchInvokeAccessOutput asserts the link records its statement in link
// data and maps it onto the role's spec by Sid (so the role does not strip the
// grant).
func ffLinkMatchInvokeAccessOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	mappingKey := fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		ffLinkRoleResource,
		linkutils.InlineAccessPolicyName(),
		ffLinkInvokeSID,
	)
	summary := map[string]any{}
	if actual != nil {
		summary["mappingValue"] = actual.ResourceDataMappings[mappingKey]
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[ffLinkExecRole] != nil &&
			actual.LinkData.Fields[ffLinkExecRole].Fields[linkutils.PermissionFieldName] != nil
	}
	expected := map[string]any{
		"mappingValue": linkutils.PermissionFieldPath(ffLinkExecRole),
		"hasStatement": true,
	}
	return plugintestutils.EqualityCheckValues{Expected: expected, Actual: summary}, nil
}

// ffLinkCreateIntermediaryTestCase exercises the create path: the role has no
// allocator policy yet, so the statement is packed into a new inline policy.
func ffLinkCreateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "grants invoke access via a new inline allocator policy on create",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return ffLinkLambdaSvc() },
		ConfigStoreA:                   ffLinkConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return ffLinkLambdaSvc() },
		ConfigStoreB:                   ffLinkConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    callerFunctionResourceInfo(),
			ResourceBInfo:    calleeFunctionResourceInfo(),
			LinkContext:      ffLinkTestLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(ffLinkRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: ffLinkMatchInvokeAccessOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutInlineAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

// ffLinkUpdateIntermediaryTestCase exercises the update path: the inline policy
// already holds the statement, which is replaced in place.
func ffLinkUpdateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"` + ffLinkInvokeSID + `","Effect":"Allow","Action":"lambda:InvokeFunction","Resource":"arn:old"}]}`
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "replaces the invoke statement in the inline allocator policy on update",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return ffLinkLambdaSvc() },
		ConfigStoreA:                   ffLinkConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return ffLinkLambdaSvc() },
		ConfigStoreB:                   ffLinkConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeUpdate,
			InstanceName:     "test-instance",
			ResourceAInfo:    callerFunctionResourceInfo(),
			ResourceBInfo:    calleeFunctionResourceInfo(),
			LinkContext:      ffLinkTestLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(ffLinkRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: ffLinkMatchInvokeAccessOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutInlineAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

// ffLinkDestroyIntermediaryTestCase exercises the destroy path: the link's
// statement is the only one in the inline policy, so the policy is deleted.
func ffLinkDestroyIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"` + ffLinkInvokeSID + `","Effect":"Allow","Action":"lambda:InvokeFunction","Resource":"` + ffLinkCalleeARN + `"}]}`
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "removes the invoke statement and deletes the empty inline policy on destroy",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return ffLinkLambdaSvc() },
		ConfigStoreA:                   ffLinkConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return ffLinkLambdaSvc() },
		ConfigStoreB:                   ffLinkConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeDestroy,
			InstanceName:     "test-instance",
			ResourceAInfo:    callerFunctionResourceInfo(),
			ResourceBInfo:    calleeFunctionResourceInfo(),
			LinkContext:      ffLinkTestLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(ffLinkRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		},
		UpdateActionsCalled: map[string]any{
			"DeleteRolePolicy": func(arg any) bool {
				input, ok := arg.(*iam.DeleteRolePolicyInput)
				return ok &&
					aws.ToString(input.RoleName) == ffLinkRoleName &&
					aws.ToString(input.PolicyName) == linkutils.InlineAccessPolicyName()
			},
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
	}, iamSvc
}

func ffLinkMissingRoleIntermediaryErrorTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	iamSvc := iammock.CreateIamServiceMock()

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "returns error when the execution role is missing from the blueprint",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return ffLinkLambdaSvc() },
		ConfigStoreA:                   ffLinkConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return ffLinkLambdaSvc() },
		ConfigStoreB:                   ffLinkConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    callerFunctionResourceInfo(),
			ResourceBInfo:    calleeFunctionResourceInfo(),
			LinkContext:      ffLinkTestLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(nil)),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectError: true,
	}, iamSvc
}

func TestFunctionFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionFunctionLinkUpdateSuite))
}
