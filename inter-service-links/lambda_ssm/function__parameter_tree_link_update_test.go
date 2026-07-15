//go:build unit

package lambdassm

import (
	"encoding/json"
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
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

const (
	fptPath = "/my-app/config"
	// The partition and account derive from the function ARN, the region from the
	// provider config.
	fptPathARN = "arn:aws:ssm:us-west-2:123456789012:parameter/my-app/config"
	// The resource name "config-store" contains a "-", which the default env var name
	// sanitises to "_" since Lambda rejects it in env var keys. The prefix matches the
	// parameter path link's convention so runtimes can consume either resource type.
	fptDefaultEnvVarName = "SSM_PARAMETER_PATH_config_store"
	fptAccessSID         = "SSMTreeAccessconfigstore"
)

type FunctionParameterTreeLinkUpdateSuite struct {
	suite.Suite
}

func functionParameterTreeLinkFactory(
	iamSvc iamservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
) provider.Link {
	build := FunctionParameterTreeLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
	) provider.Link {
		return build(FunctionToParameterLinkDeps(deps))
	}
}

func fptParameterTreeInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "config-store",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"path", core.MappingNodeFromString(fptPath),
				"values", core.MappingNodeFields(
					"logLevel", core.MappingNodeFromString("info"),
				),
			),
		},
	}
}

func (s *FunctionParameterTreeLinkUpdateSuite) Test_update_resource_a_env_vars() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		functionParameterTreeAddEnvVarTestCase(loader),
		functionParameterTreeEnvVarsDisabledTestCase(loader),
		functionParameterTreeMissingPathTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		functionParameterTreeLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

func functionParameterTreeAddEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fpGetFunctionOutput(map[string]string{"EXISTING": "val"})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "populates the tree path prefix env var with a sanitised default name",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopSSMServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      fpFunctionInfo(),
			OtherResourceInfo: fptParameterTreeInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"apiFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						fptDefaultEnvVarName, core.MappingNodeFromString(fptPath),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				fmt.Sprintf(
					"apiFunction::spec.environment.variables[\"%s\"]", fptDefaultEnvVarName,
				): fmt.Sprintf(
					"apiFunction.environmentVariables[\"%s\"]", fptDefaultEnvVarName,
				),
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fpFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{
						"EXISTING":           "val",
						fptDefaultEnvVarName: fptPath,
					},
				},
			},
		},
	}
}

func functionParameterTreeEnvVarsDisabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:            "does nothing when env var population is disabled",
		Resource:        plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA: func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:    testConfigStore(loader),
		ServiceFactoryB: noopSSMServiceFactory,
		ConfigStoreB:    testConfigStore(loader),

		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: fppFunctionInfoWithAnnotations(map[string]*core.MappingNode{
				"aws.lambda.ssm.config-store.populateEnvVars": core.MappingNodeFromBool(false),
			}),
			OtherResourceInfo: fptParameterTreeInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields(),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsNotCalled: []string{"GetFunction", "UpdateFunctionConfiguration"},
	}
}

func functionParameterTreeMissingPathTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fpGetFunctionOutput(map[string]string{})),
	)

	parameterTreeInfo := &provider.ResourceInfo{
		ResourceName: "config-store",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(),
		},
	}

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "returns error when the path is missing from the target resource state",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopSSMServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      fpFunctionInfo(),
			OtherResourceInfo: parameterTreeInfo,
			LinkContext:       testLinkContext(),
		},
		UpdateActionsNotCalled: []string{"UpdateFunctionConfiguration"},
		ExpectError:            true,
	}
}

// Verifies the inline policy carries the tree link's statement with the three read ssm
// actions over both the path ARN and its wildcard child ARN.
func matchTreeAccessPolicy(arg any, expectedActions []string) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != fpRoleName ||
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
		if statement.Sid != fptAccessSID {
			continue
		}
		return hasAll(statement.Action, expectedActions) &&
			hasAll(statement.Resource, []string{fptPathARN, fptPathARN + "/*"})
	}
	return false
}

func matchTreeAccessLinkOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	mappingKey := fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		fpRoleResource,
		linkutils.InlineAccessPolicyName(),
		fptAccessSID,
	)
	summary := map[string]any{}
	if actual != nil {
		summary["mappingValue"] = actual.ResourceDataMappings[mappingKey]
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[fpExecRole] != nil &&
			actual.LinkData.Fields[fpExecRole].Fields[linkutils.PermissionFieldName] != nil
	}
	expected := map[string]any{
		"mappingValue": linkutils.PermissionFieldPath(fpExecRole),
		"hasStatement": true,
	}
	return plugintestutils.EqualityCheckValues{Expected: expected, Actual: summary}, nil
}

func (s *FunctionParameterTreeLinkUpdateSuite) Test_update_intermediary_resources_grants_read() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                           "grants read access over the tree's path prefix and everything beneath it",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fpLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopSSMServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    fpFunctionInfo(),
			ResourceBInfo:    fptParameterTreeInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  fpRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchTreeAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool {
				return matchTreeAccessPolicy(arg, []string{
					"ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath",
				})
			},
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
		]{testCase},
		functionParameterTreeLinkFactory(iamSvc),
		&s.Suite,
	)
}

func (s *FunctionParameterTreeLinkUpdateSuite) Test_update_intermediary_resources_grants_readwrite() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	functionInfo := fppFunctionInfoWithAnnotations(map[string]*core.MappingNode{
		"aws.lambda.ssm.config-store.accessLevel": core.MappingNodeFromString("readwrite"),
	})

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                           "grants readwrite access so runtime config writes are possible",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fpLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopSSMServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    functionInfo,
			ResourceBInfo:    fptParameterTreeInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  fpRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchTreeAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool {
				return matchTreeAccessPolicy(arg, []string{
					"ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath", "ssm:PutParameter",
				})
			},
		},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
		]{testCase},
		functionParameterTreeLinkFactory(iamSvc),
		&s.Suite,
	)
}

func (s *FunctionParameterTreeLinkUpdateSuite) Test_update_intermediary_resources_destroy_revokes_grant() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                           "revokes the grant on destroy without touching the tree resource",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fpLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopSSMServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeDestroy,
			InstanceName:     "test-instance",
			ResourceAInfo:    fpFunctionInfo(),
			ResourceBInfo:    fptParameterTreeInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  fpRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
		]{testCase},
		functionParameterTreeLinkFactory(iamSvc),
		&s.Suite,
	)
}

func TestFunctionParameterTreeLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionParameterTreeLinkUpdateSuite))
}
