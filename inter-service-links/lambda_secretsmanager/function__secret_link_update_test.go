//go:build unit

package lambdasecretsmanager

import (
	"encoding/json"
	"fmt"
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

const (
	fsFunctionARN  = "arn:aws:lambda:us-west-2:123456789012:function:api"
	fsRoleARN      = "arn:aws:iam::123456789012:role/api-role"
	fsEnvVarName   = "SECRET_dbCredentials"
	fsExecRole     = "apiFunctionExecutionRole"
	fsAccessSID    = "SecretsManagerAccessdbCredentials"
	fsRoleName     = "api-role"
	fsRoleResource = "apiFunctionRole"
)

type FunctionSecretLinkUpdateSuite struct {
	suite.Suite
}

func fsFunctionInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "apiFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("arn", core.MappingNodeFromString(fsFunctionARN)),
		},
	}
}

func fsSecretInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "dbCredentials",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"id", core.MappingNodeFromString(testSecretARN),
			),
		},
	}
}

func fsGetFunctionOutput(vars map[string]string) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(fsFunctionARN),
			Role:        aws.String(fsRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{Variables: vars},
		},
	}
}

func (s *FunctionSecretLinkUpdateSuite) Test_update_resource_a_env_vars() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		functionSecretAddEnvVarTestCase(loader),
		functionSecretRemoveEnvVarTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		functionSecretLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

func functionSecretAddEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fsGetFunctionOutput(map[string]string{"EXISTING": "val"})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "populates the secret ARN env var on the function",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      fsFunctionInfo(),
			OtherResourceInfo: fsSecretInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"apiFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						fsEnvVarName, core.MappingNodeFromString(testSecretARN),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				fmt.Sprintf(
					"apiFunction::spec.environment.variables[\"%s\"]", fsEnvVarName,
				): fmt.Sprintf(
					"apiFunction.environmentVariables[\"%s\"]", fsEnvVarName,
				),
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fsFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{
						"EXISTING":   "val",
						fsEnvVarName: testSecretARN,
					},
				},
			},
		},
	}
}

func functionSecretRemoveEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fsGetFunctionOutput(map[string]string{
			"EXISTING":   "val",
			fsEnvVarName: testSecretARN,
		})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "removes the secret ARN env var on destroy",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      fsFunctionInfo(),
			OtherResourceInfo: fsSecretInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields("apiFunction", core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fsFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{"EXISTING": "val"},
				},
			},
		},
	}
}

func fsRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: fsRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(fsRoleName),
			"arn", core.MappingNodeFromString(fsRoleARN),
		),
	}
}

func fsLambdaSvc() lambdaservice.Service {
	return lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fsGetFunctionOutput(map[string]string{})),
	)
}

func fsRoleService() provider.ResourceService {
	return resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(fsRoleState()))
}

func matchReadAccessPolicy(arg any) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != fsRoleName ||
		aws.ToString(input.PolicyName) != linkutils.InlineAccessPolicyName() {
		return false
	}
	var doc struct {
		Statement []struct {
			Sid      string
			Action   []string
			Resource []string
		}
	}
	if err := json.Unmarshal([]byte(aws.ToString(input.PolicyDocument)), &doc); err != nil {
		return false
	}
	for _, statement := range doc.Statement {
		if statement.Sid != fsAccessSID {
			continue
		}
		return hasAll(statement.Action, []string{"secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"}) &&
			hasAll(statement.Resource, []string{testSecretARN})
	}
	return false
}

func hasAll(have, want []string) bool {
	set := map[string]bool{}
	for _, v := range have {
		set[v] = true
	}
	for _, v := range want {
		if !set[v] {
			return false
		}
	}
	return len(have) == len(want)
}

func matchAccessLinkOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	mappingKey := fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		fsRoleResource,
		linkutils.InlineAccessPolicyName(),
		fsAccessSID,
	)
	summary := map[string]any{}
	if actual != nil {
		summary["mappingValue"] = actual.ResourceDataMappings[mappingKey]
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[fsExecRole] != nil &&
			actual.LinkData.Fields[fsExecRole].Fields[linkutils.PermissionFieldName] != nil
	}
	expected := map[string]any{
		"mappingValue": linkutils.PermissionFieldPath(fsExecRole),
		"hasStatement": true,
	}
	return plugintestutils.EqualityCheckValues{Expected: expected, Actual: summary}, nil
}

func (s *FunctionSecretLinkUpdateSuite) Test_update_intermediary_resources_grants_read() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "grants read Secrets Manager access via a new inline allocator policy",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fsLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    fsFunctionInfo(),
			ResourceBInfo:    fsSecretInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  fsRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchReadAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionSecretLinkFactory(iamSvc),
		&s.Suite,
	)
}

func TestFunctionSecretLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionSecretLinkUpdateSuite))
}
