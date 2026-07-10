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
	fppPath = "/my-app/config"
	// The partition and account derive from the function ARN, the region from the
	// provider config.
	fppPathARN = "arn:aws:ssm:us-west-2:123456789012:parameter/my-app/config"
	// The resource name "app-config" contains a "-", which the default env var name
	// sanitises to "_" since Lambda rejects it in env var keys.
	fppDefaultEnvVarName = "SSM_PARAMETER_PATH_app_config"
	fppCustomEnvVarName  = "APP_CONFIG_STORE_PATH"
	fppAccessSID         = "SSMPathAccessappconfig"
)

type FunctionParameterPathLinkUpdateSuite struct {
	suite.Suite
}

func functionParameterPathLinkFactory(
	iamSvc iamservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
) provider.Link {
	build := FunctionParameterPathLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
	) provider.Link {
		return build(FunctionToParameterLinkDeps(deps))
	}
}

func fppParameterPathInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "app-config",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"path", core.MappingNodeFromString(fppPath),
			),
		},
	}
}

func fppFunctionInfoWithAnnotations(annotations map[string]*core.MappingNode) *provider.ResourceInfo {
	info := fpFunctionInfo()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: &core.MappingNode{Fields: annotations},
		},
	}
	return info
}

func (s *FunctionParameterPathLinkUpdateSuite) Test_update_resource_a_env_vars() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		functionParameterPathAddEnvVarTestCase(loader),
		functionParameterPathCustomEnvVarNameTestCase(loader),
		functionParameterPathEnvVarsDisabledTestCase(loader),
		functionParameterPathRemoveEnvVarTestCase(loader),
		functionParameterPathMissingPathTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		functionParameterPathLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

func functionParameterPathAddEnvVarTestCase(
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
		Name:                    "populates the path prefix env var with a sanitised default name",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopSSMServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      fpFunctionInfo(),
			OtherResourceInfo: fppParameterPathInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"apiFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						fppDefaultEnvVarName, core.MappingNodeFromString(fppPath),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				fmt.Sprintf(
					"apiFunction::spec.environment.variables[\"%s\"]", fppDefaultEnvVarName,
				): fmt.Sprintf(
					"apiFunction.environmentVariables[\"%s\"]", fppDefaultEnvVarName,
				),
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fpFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{
						"EXISTING":           "val",
						fppDefaultEnvVarName: fppPath,
					},
				},
			},
		},
	}
}

func functionParameterPathCustomEnvVarNameTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fpGetFunctionOutput(map[string]string{})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "populates the path prefix env var with a custom name",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopSSMServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: fppFunctionInfoWithAnnotations(map[string]*core.MappingNode{
				"aws.lambda.ssm.app-config.envVarName": core.MappingNodeFromString(fppCustomEnvVarName),
			}),
			OtherResourceInfo: fppParameterPathInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"apiFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						fppCustomEnvVarName, core.MappingNodeFromString(fppPath),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				fmt.Sprintf(
					"apiFunction::spec.environment.variables[\"%s\"]", fppCustomEnvVarName,
				): fmt.Sprintf(
					"apiFunction.environmentVariables[\"%s\"]", fppCustomEnvVarName,
				),
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fpFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{fppCustomEnvVarName: fppPath},
				},
			},
		},
	}
}

func functionParameterPathEnvVarsDisabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "does nothing when env var population is disabled",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopSSMServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: fppFunctionInfoWithAnnotations(map[string]*core.MappingNode{
				"aws.lambda.ssm.app-config.populateEnvVars": core.MappingNodeFromBool(false),
			}),
			OtherResourceInfo: fppParameterPathInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields(),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsNotCalled: []string{"GetFunction", "UpdateFunctionConfiguration"},
	}
}

func functionParameterPathRemoveEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fpGetFunctionOutput(map[string]string{
			"EXISTING":           "val",
			fppDefaultEnvVarName: fppPath,
		})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "removes the path prefix env var on destroy",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopSSMServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      fpFunctionInfo(),
			OtherResourceInfo: fppParameterPathInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields("apiFunction", core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fpFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{"EXISTING": "val"},
				},
			},
		},
	}
}

func functionParameterPathMissingPathTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fpGetFunctionOutput(map[string]string{})),
	)

	parameterPathInfo := &provider.ResourceInfo{
		ResourceName: "app-config",
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
			OtherResourceInfo: parameterPathInfo,
			LinkContext:       testLinkContext(),
		},
		UpdateActionsNotCalled: []string{"UpdateFunctionConfiguration"},
		ExpectError:            true,
	}
}

// Verifies the inline policy carries the link's statement with
// the three read ssm actions over both the path ARN and its wildcard child ARN.
func matchPathReadAccessPolicy(arg any) bool {
	return matchPathAccessPolicy(arg, []string{
		"ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath",
	})
}

func matchPathAccessPolicy(arg any, expectedActions []string) bool {
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
		if statement.Sid != fppAccessSID {
			continue
		}
		return hasAll(statement.Action, expectedActions) &&
			hasAll(statement.Resource, []string{fppPathARN, fppPathARN + "/*"})
	}
	return false
}

func matchPathAccessLinkOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	mappingKey := fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		fpRoleResource,
		linkutils.InlineAccessPolicyName(),
		fppAccessSID,
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

func (s *FunctionParameterPathLinkUpdateSuite) Test_update_intermediary_resources_grants_read() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                           "grants read access over the path prefix and everything beneath it",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fpLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopSSMServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    fpFunctionInfo(),
			ResourceBInfo:    fppParameterPathInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  fpRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchPathAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPathReadAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
		]{testCase},
		functionParameterPathLinkFactory(iamSvc),
		&s.Suite,
	)
}

func (s *FunctionParameterPathLinkUpdateSuite) Test_update_intermediary_resources_grants_readwrite() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	functionInfo := fppFunctionInfoWithAnnotations(map[string]*core.MappingNode{
		"aws.lambda.ssm.app-config.accessLevel": core.MappingNodeFromString("readwrite"),
	})

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                           "grants readwrite access over the path prefix",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fpLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopSSMServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    functionInfo,
			ResourceBInfo:    fppParameterPathInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  fpRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchPathAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool {
				return matchPathAccessPolicy(arg, []string{
					"ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath", "ssm:PutParameter",
				})
			},
		},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
		]{testCase},
		functionParameterPathLinkFactory(iamSvc),
		&s.Suite,
	)
}

func (s *FunctionParameterPathLinkUpdateSuite) Test_update_intermediary_resources_destroy_revokes_grant() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                           "revokes the grant on destroy without touching the path resource",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fpLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopSSMServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeDestroy,
			InstanceName:     "test-instance",
			ResourceAInfo:    fpFunctionInfo(),
			ResourceBInfo:    fppParameterPathInfo(),
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
		functionParameterPathLinkFactory(iamSvc),
		&s.Suite,
	)
}

func (s *FunctionParameterPathLinkUpdateSuite) Test_update_intermediary_resources_malformed_function_arn() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
	)

	functionInfo := &provider.ResourceInfo{
		ResourceName: "apiFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("arn", core.MappingNodeFromString("arn:aws")),
		},
	}

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                           "returns error when the partition and account cannot be derived",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fpLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopSSMServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    functionInfo,
			ResourceBInfo:    fppParameterPathInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  fpRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
		ExpectError:            true,
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
		]{testCase},
		functionParameterPathLinkFactory(iamSvc),
		&s.Suite,
	)
}

func TestFunctionParameterPathLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionParameterPathLinkUpdateSuite))
}
