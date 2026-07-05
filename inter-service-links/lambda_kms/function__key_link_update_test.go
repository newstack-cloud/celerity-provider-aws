//go:build unit

package lambdakms

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	kmsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/kms_mock"
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
	fkFunctionARN  = "arn:aws:lambda:us-west-2:123456789012:function:encrypt"
	fkRoleARN      = "arn:aws:iam::123456789012:role/encrypt-role"
	fkEnvVarName   = "KMS_KEY_dataKey"
	fkExecRole     = "encryptFunctionExecutionRole"
	fkAccessSID    = "KMSAccessdataKey"
	fkRoleName     = "encrypt-role"
	fkRoleResource = "encryptFunctionRole"
)

type FunctionKeyLinkUpdateSuite struct {
	suite.Suite
}

func fkFunctionInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "encryptFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("arn", core.MappingNodeFromString(fkFunctionARN)),
		},
	}
}

func fkKeyInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "dataKey",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(testKeyARN),
			),
		},
	}
}

func fkGetFunctionOutput(vars map[string]string) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(fkFunctionARN),
			Role:        aws.String(fkRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{Variables: vars},
		},
	}
}

func (s *FunctionKeyLinkUpdateSuite) Test_update_resource_a_env_vars() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		functionKeyAddEnvVarTestCase(loader),
		functionKeyRemoveEnvVarTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		functionKeyLinkFactory(iammock.CreateIamServiceMock(), defaultKMSMock()),
		&s.Suite,
	)
}

func functionKeyAddEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fkGetFunctionOutput(map[string]string{"EXISTING": "val"})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "populates the key ARN env var on the function",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      fkFunctionInfo(),
			OtherResourceInfo: fkKeyInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"encryptFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						fkEnvVarName, core.MappingNodeFromString(testKeyARN),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				fmt.Sprintf(
					"encryptFunction::spec.environment.variables[\"%s\"]", fkEnvVarName,
				): fmt.Sprintf(
					"encryptFunction.environmentVariables[\"%s\"]", fkEnvVarName,
				),
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fkFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{
						"EXISTING":   "val",
						fkEnvVarName: testKeyARN,
					},
				},
			},
		},
	}
}

func functionKeyRemoveEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fkGetFunctionOutput(map[string]string{
			"EXISTING":   "val",
			fkEnvVarName: testKeyARN,
		})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "removes the key ARN env var on destroy",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      fkFunctionInfo(),
			OtherResourceInfo: fkKeyInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields("encryptFunction", core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fkFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{"EXISTING": "val"},
				},
			},
		},
	}
}

func fkRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: fkRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(fkRoleName),
			"arn", core.MappingNodeFromString(fkRoleARN),
		),
	}
}

func fkLambdaSvc() lambdaservice.Service {
	return lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fkGetFunctionOutput(map[string]string{})),
	)
}

func fkRoleService() provider.ResourceService {
	return resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(fkRoleState()))
}

func matchDecryptAccessPolicy(arg any) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != fkRoleName ||
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
		if statement.Sid != fkAccessSID {
			continue
		}
		return hasAll(statement.Action, []string{"kms:Decrypt", "kms:DescribeKey"}) &&
			hasAll(statement.Resource, []string{testKeyARN})
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
		fkRoleResource,
		linkutils.InlineAccessPolicyName(),
		fkAccessSID,
	)
	summary := map[string]any{}
	if actual != nil {
		summary["mappingValue"] = actual.ResourceDataMappings[mappingKey]
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[fkExecRole] != nil &&
			actual.LinkData.Fields[fkExecRole].Fields[linkutils.PermissionFieldName] != nil
	}
	expected := map[string]any{
		"mappingValue": linkutils.PermissionFieldPath(fkExecRole),
		"hasStatement": true,
	}
	return plugintestutils.EqualityCheckValues{Expected: expected, Actual: summary}, nil
}

func (s *FunctionKeyLinkUpdateSuite) Test_update_intermediary_resources_grants_decrypt() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "grants decrypt KMS access via a new inline allocator policy",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fkLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    fkFunctionInfo(),
			ResourceBInfo:    fkKeyInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  fkRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchDecryptAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionKeyLinkFactory(iamSvc, defaultKMSMock()),
		&s.Suite,
	)
}

func fkFunctionInfoWithGrant() *provider.ResourceInfo {
	info := fkFunctionInfo()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"aws.lambda.kms.dataKey.manageKeyGrant": core.MappingNodeFromBool(true),
				},
			},
		},
	}
	return info
}

// End-to-end: the manageKeyGrant annotation wires through UpdateIntermediaryResources to a
// KMS grant for the execution role, in addition to the IAM role policy.
func (s *FunctionKeyLinkUpdateSuite) Test_update_intermediary_resources_creates_grant_when_managed() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	kmsSvc := kmsmock.CreateKMSServiceMock(
		kmsmock.WithListGrantsOutput(&kms.ListGrantsOutput{}),
		kmsmock.WithCreateGrantOutput(&kms.CreateGrantOutput{GrantId: aws.String(fkGrantID)}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "creates a KMS grant when manageKeyGrant is enabled",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fkLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &kmsSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    fkFunctionInfoWithGrant(),
			ResourceBInfo:    fkKeyInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  fkRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"CreateGrant": func(arg any) bool {
				in, ok := arg.(*kms.CreateGrantInput)
				return ok &&
					aws.ToString(in.KeyId) == testKeyARN &&
					aws.ToString(in.GranteePrincipal) == fkRoleARN &&
					aws.ToString(in.Name) == fkGrantName
			},
		},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionKeyLinkFactory(iamSvc, kmsSvc),
		&s.Suite,
	)
}

func TestFunctionKeyLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionKeyLinkUpdateSuite))
}
