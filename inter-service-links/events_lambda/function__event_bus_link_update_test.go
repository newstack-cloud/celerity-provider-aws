//go:build unit

package eventslambda

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
	eventsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/events_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FunctionEventBusLinkUpdateSuite struct {
	suite.Suite
}

const (
	febFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:publish-order"
	febRoleARN     = "arn:aws:iam::123456789012:role/publish-order-role"
)

func functionResourceInfoA() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "publishOrderFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(febFunctionARN),
			),
		},
	}
}

func eventBusResourceInfoB() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "orderEventBus",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"name", core.MappingNodeFromString("order-events"),
				"arn", core.MappingNodeFromString("arn:aws:events:us-west-2:123456789012:event-bus/order-events"),
			),
		},
	}
}

func getFunctionWithEnvOutput() *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(febFunctionARN),
			Role:        aws.String(febRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{
				Variables: map[string]string{"EXISTING": "val"},
			},
		},
	}
}

func (s *FunctionEventBusLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		functionEventBusAddEnvVarTestCase(loader),
		functionEventBusRemoveEnvVarTestCase(loader),
		functionEventBusUpdateErrorTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(testCases, functionEventBusLinkFactory(), &s.Suite)
}

func functionEventBusAddEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	eventsservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(getFunctionWithEnvOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		Name:                    "populates the event bus env var on the function",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoA(),
			OtherResourceInfo: eventBusResourceInfoB(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"publishOrderFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						"EVENT_BUS_orderEventBus", core.MappingNodeFromString("order-events"),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				"publishOrderFunction::spec.environment.variables[\"EVENT_BUS_orderEventBus\"]": "publishOrderFunction.environmentVariables[\"EVENT_BUS_orderEventBus\"]",
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(febFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{
						"EXISTING":                "val",
						"EVENT_BUS_orderEventBus": "order-events",
					},
				},
			},
		},
	}
}

func functionEventBusRemoveEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	eventsservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(febFunctionARN),
				Role:        aws.String(febRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{
					Variables: map[string]string{
						"EXISTING":                "val",
						"EVENT_BUS_orderEventBus": "order-events",
					},
				},
			},
		}),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		Name:                    "removes the event bus env var on destroy",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      functionResourceInfoA(),
			OtherResourceInfo: eventBusResourceInfoB(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"publishOrderFunction",
				core.MappingNodeFields(),
			),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(febFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{"EXISTING": "val"},
				},
			},
		},
	}
}

func functionEventBusUpdateErrorTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	eventsservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(getFunctionWithEnvOutput()),
		lambdamock.WithUpdateFunctionConfigurationError(errors.New("update failed")),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		Name:                    "returns error when the function update fails",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoA(),
			OtherResourceInfo: eventBusResourceInfoB(),
			LinkContext:       testLinkContext(),
		},
		ExpectError: true,
	}
}

const (
	febRoleName         = "publish-order-role"
	febRoleResourceName = "publishOrderFunctionRole"
	febEventBusARN      = "arn:aws:events:us-west-2:123456789012:event-bus/order-events"
	febExecRole         = "publishOrderFunctionExecutionRole"
	febPutEventsSID     = "EventBridgePutEventsorderEventBus"
)

func functionEventBusLinkFactoryWithIam(
	iamSvc iamservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service],
) provider.Link {
	build := FunctionEventBusLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service],
	) provider.Link {
		return build(FunctionToEventBusLinkDeps(deps))
	}
}

func febRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: febRoleResourceName,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(febRoleName),
			"arn", core.MappingNodeFromString(febRoleARN),
		),
	}
}

func febLambdaSvc() lambdaservice.Service {
	return lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(febFunctionARN),
				Role:        aws.String(febRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
	)
}

func febRoleService() provider.ResourceService {
	return resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(febRoleState()))
}

func matchPutInlineAccessPolicy(arg any) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != febRoleName ||
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
		if statement.Sid == febPutEventsSID {
			return true
		}
	}
	return false
}

// Asserts the link records its statement in link data and
// maps it onto the role's spec by Sid (so the role does not strip the grant).
func matchAccessLinkOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	mappingKey := fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		febRoleResourceName,
		linkutils.InlineAccessPolicyName(),
		febPutEventsSID,
	)
	summary := map[string]any{}
	if actual != nil {
		summary["mappingValue"] = actual.ResourceDataMappings[mappingKey]
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[febExecRole] != nil &&
			actual.LinkData.Fields[febExecRole].Fields[linkutils.PermissionFieldName] != nil
	}
	expected := map[string]any{
		"mappingValue": linkutils.PermissionFieldPath(febExecRole),
		"hasStatement": true,
	}
	return plugintestutils.EqualityCheckValues{Expected: expected, Actual: summary}, nil
}

// Test_link_update_intermediary_resources exercises the create, update and
// destroy paths of UpdateIntermediaryResources, which grant the function's
// execution role permission to publish events via the IAM permission allocator
// (a statement packed into the role's allocator-managed inline policy).
func (s *FunctionEventBusLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	createCase, createIam := functionEventBusCreateIntermediaryTestCase(loader)
	updateCase, updateIam := functionEventBusUpdateIntermediaryTestCase(loader)
	destroyCase, destroyIam := functionEventBusDestroyIntermediaryTestCase(loader)

	cases := []struct {
		testCase plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
		]
		iamSvc iamservice.Service
	}{
		{createCase, createIam},
		{updateCase, updateIam},
		{destroyCase, destroyIam},
	}

	for _, c := range cases {
		plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
			[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
				*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
			]{c.testCase},
			functionEventBusLinkFactoryWithIam(c.iamSvc),
			&s.Suite,
		)
	}
}

// functionEventBusCreateIntermediaryTestCase exercises the create path: the role
// has no allocator policy yet, so the statement is packed into a new inline policy.
func functionEventBusCreateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
], iamservice.Service) {
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
	]{
		Name:                           "grants put-events access via a new inline allocator policy on create",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return febLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    functionResourceInfoA(),
			ResourceBInfo:    eventBusResourceInfoB(),
			LinkContext:      testLinkContext(),
			ResourceService:  febRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutInlineAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

func functionEventBusUpdateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
], iamservice.Service) {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"` + febPutEventsSID + `","Effect":"Allow","Action":["events:PutEvents"],"Resource":"arn:old"}]}`
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
	]{
		Name:                           "replaces the put-events statement in the inline allocator policy on update",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return febLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeUpdate,
			InstanceName:     "test-instance",
			ResourceAInfo:    functionResourceInfoA(),
			ResourceBInfo:    eventBusResourceInfoB(),
			LinkContext:      testLinkContext(),
			ResourceService:  febRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutInlineAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

func functionEventBusDestroyIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
], iamservice.Service) {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"` + febPutEventsSID + `","Effect":"Allow","Action":["events:PutEvents"],"Resource":"` + febEventBusARN + `"}]}`
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
	]{
		Name:                           "removes the put-events statement and deletes the empty inline policy on destroy",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return febLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeDestroy,
			InstanceName:     "test-instance",
			ResourceAInfo:    functionResourceInfoA(),
			ResourceBInfo:    eventBusResourceInfoB(),
			LinkContext:      testLinkContext(),
			ResourceService:  febRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		},
		UpdateActionsCalled: map[string]any{
			"DeleteRolePolicy": func(arg any) bool {
				input, ok := arg.(*iam.DeleteRolePolicyInput)
				return ok &&
					aws.ToString(input.RoleName) == febRoleName &&
					aws.ToString(input.PolicyName) == linkutils.InlineAccessPolicyName()
			},
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
	}, iamSvc
}

func TestFunctionEventBusLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionEventBusLinkUpdateSuite))
}
