//go:build unit

package lambdasqs

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionQueueLinkUpdateSuite struct {
	suite.Suite
}

func lqGetFunctionOutput(vars map[string]string) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(lqFunctionARN),
			Role:        aws.String(lqRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{Variables: vars},
		},
	}
}

func (s *FunctionQueueLinkUpdateSuite) Test_update_resource_a_env_vars() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		lqAddEnvVarTestCase(loader),
		lqRemoveEnvVarTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases, functionQueueLinkFactory(iammock.CreateIamServiceMock()), &s.Suite,
	)
}

func lqAddEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(lqGetFunctionOutput(map[string]string{"EXISTING": "val"})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "populates the queue URL env var on the function",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      lqFunctionInfo(nil),
			OtherResourceInfo: lqQueueInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"submitOrderFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						"SQS_QUEUE_ordersQueue", core.MappingNodeFromString(lqQueueURL),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				"submitOrderFunction::spec.environment.variables[\"SQS_QUEUE_ordersQueue\"]": "submitOrderFunction.environmentVariables[\"SQS_QUEUE_ordersQueue\"]",
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": func(arg any) bool {
				in, ok := arg.(*lambda.UpdateFunctionConfigurationInput)
				return ok && in.Environment != nil &&
					in.Environment.Variables["EXISTING"] == "val" &&
					in.Environment.Variables["SQS_QUEUE_ordersQueue"] == lqQueueURL
			},
		},
	}
}

func lqRemoveEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(lqGetFunctionOutput(map[string]string{
			"EXISTING": "val", "SQS_QUEUE_ordersQueue": lqQueueURL,
		})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "removes the queue env var on destroy",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      lqFunctionInfo(nil),
			OtherResourceInfo: lqQueueInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields("submitOrderFunction", core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": func(arg any) bool {
				in, ok := arg.(*lambda.UpdateFunctionConfigurationInput)
				if !ok || in.Environment == nil {
					return false
				}
				_, hasQueue := in.Environment.Variables["SQS_QUEUE_ordersQueue"]
				return in.Environment.Variables["EXISTING"] == "val" && !hasQueue
			},
		},
	}
}

// The default access level grants sqs:SendMessage.
func (s *FunctionQueueLinkUpdateSuite) Test_update_intermediary_resources_grants_send() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(lqGetFunctionOutput(map[string]string{})),
	)
	rs := resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(lqRoleState()))

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "grants sqs:SendMessage scoped to the queue",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    lqFunctionInfo(nil),
			ResourceBInfo:    lqQueueInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  rs,
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: lqMatchAccessOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return lqMatchAccessPolicy(arg, []string{"sqs:SendMessage"}) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionQueueLinkFactory(iamSvc),
		&s.Suite,
	)
}

// Checks the queue access statement contains the wanted actions (a subset)
// scoped to the queue ARN.
func lqMatchAccessPolicy(arg any, wantActions []string) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok || aws.ToString(input.RoleName) != lqRoleName {
		return false
	}
	var doc struct {
		Statement []struct {
			Sid      string
			Action   []string
			Resource string
		}
	}
	if err := json.Unmarshal([]byte(aws.ToString(input.PolicyDocument)), &doc); err != nil {
		return false
	}
	for _, statement := range doc.Statement {
		if statement.Sid != "SQSAccessordersQueue" {
			continue
		}
		if statement.Resource != lqQueueARN {
			return false
		}
		actions := map[string]bool{}
		for _, a := range statement.Action {
			actions[a] = true
		}
		for _, want := range wantActions {
			if !actions[want] {
				return false
			}
		}
		return true
	}
	return false
}

func lqMatchAccessOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	summary := map[string]any{}
	if actual != nil {
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields["submitOrderFunctionExecutionRole"] != nil &&
			actual.LinkData.Fields["submitOrderFunctionExecutionRole"].Fields[linkutils.PermissionFieldName] != nil
	}
	return plugintestutils.EqualityCheckValues{
		Expected: map[string]any{"hasStatement": true},
		Actual:   summary,
	}, nil
}

func TestFunctionQueueLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionQueueLinkUpdateSuite))
}
