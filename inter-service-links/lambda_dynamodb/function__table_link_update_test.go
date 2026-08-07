//go:build unit

package lambdadynamodb

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
	dynamodbmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/dynamodb_mock"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
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
	fdtFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:process-orders"
	fdtRoleARN     = "arn:aws:iam::123456789012:role/process-orders-role"
	fdtTableARN    = "arn:aws:dynamodb:us-west-2:123456789012:table/orders"
	fdtTableName   = "orders-table"
	fdtEnvVarName  = "DYNAMODB_TABLE_ordersTable"
	fdtExecRole    = "processOrdersFunctionExecutionRole"
	fdtAccessSID   = "DynamoDBAccessordersTable"
)

type FunctionTableLinkUpdateSuite struct {
	suite.Suite
}

func functionTableLinkFactory(
	iamSvc iamservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service],
) provider.Link {
	build := FunctionDynamoDBTableLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service],
	) provider.Link {
		return build(FunctionToDynamoDBTableLinkDeps(deps))
	}
}

func testLinkContext() provider.LinkContext {
	return plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {"region": core.ScalarFromString("us-west-2")},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func testConfigStore(loader *testutils.MockAWSConfigLoader) pluginutils.ServiceConfigStore[*aws.Config] {
	return utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
}

func functionResourceInfoA() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "processOrdersFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(fdtFunctionARN),
			),
		},
	}
}

func tableResourceInfoB() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersTable",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"tableName", core.MappingNodeFromString(fdtTableName),
				"arn", core.MappingNodeFromString(fdtTableARN),
			),
		},
	}
}

func functionWithEnvOutput(vars map[string]string) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(fdtFunctionARN),
			Role:        aws.String(fdtRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{Variables: vars},
		},
	}
}

func (s *FunctionTableLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		functionTableAddEnvVarTestCase(loader),
		functionTableAddEnvVarFromARNTestCase(loader),
		functionTableRemoveEnvVarTestCase(loader),
		functionTableUpdateErrorTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		functionTableLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

func functionTableAddEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(functionWithEnvOutput(map[string]string{"EXISTING": "val"})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name:                    "populates the table name env var on the function",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoA(),
			OtherResourceInfo: tableResourceInfoB(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"processOrdersFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						fdtEnvVarName, core.MappingNodeFromString(fdtTableName),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				fmt.Sprintf(
					"processOrdersFunction::spec.environment.variables[\"%s\"]",
					fdtEnvVarName,
				): fmt.Sprintf(
					"processOrdersFunction.environmentVariables[\"%s\"]",
					fdtEnvVarName,
				),
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fdtFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{
						"EXISTING":    "val",
						fdtEnvVarName: fdtTableName,
					},
				},
			},
		},
	}
}

// A name-less (auto-named) table has no tableName in state at link-update
// time; the link must derive the physical name from the table ARN instead.
func functionTableAddEnvVarFromARNTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(functionWithEnvOutput(map[string]string{})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)
	arnDerivedTableName := "orders"

	namelessTableInfo := &provider.ResourceInfo{
		ResourceName: "ordersTable",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(fdtTableARN),
			),
		},
	}

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name:                    "derives the table name from the ARN for a name-less table",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoA(),
			OtherResourceInfo: namelessTableInfo,
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"processOrdersFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						fdtEnvVarName, core.MappingNodeFromString(arnDerivedTableName),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				fmt.Sprintf(
					"processOrdersFunction::spec.environment.variables[\"%s\"]",
					fdtEnvVarName,
				): fmt.Sprintf(
					"processOrdersFunction.environmentVariables[\"%s\"]",
					fdtEnvVarName,
				),
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fdtFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{
						fdtEnvVarName: arnDerivedTableName,
					},
				},
			},
		},
	}
}

func functionTableRemoveEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(functionWithEnvOutput(map[string]string{
			"EXISTING":    "val",
			fdtEnvVarName: fdtTableName,
		})),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name:     "removes the table name env var on destroy",
		Resource: plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA: func(c *aws.Config, pc provider.Context) lambdaservice.Service {
			return lambdaSvc
		},
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      functionResourceInfoA(),
			OtherResourceInfo: tableResourceInfoB(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"processOrdersFunction",
				core.MappingNodeFields(),
			),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(fdtFunctionARN),
				Environment: &lambdatypes.Environment{
					Variables: map[string]string{"EXISTING": "val"},
				},
			},
		},
	}
}

func functionTableUpdateErrorTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(functionWithEnvOutput(map[string]string{"EXISTING": "val"})),
		lambdamock.WithUpdateFunctionConfigurationError(errors.New("update failed")),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name:                    "returns error when the function update fails",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoA(),
			OtherResourceInfo: tableResourceInfoB(),
			LinkContext:       testLinkContext(),
		},
		ExpectError: true,
	}
}

func (s *FunctionTableLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	createCase, createIam := functionTableCreateIntermediaryTestCase(loader)
	updateCase, updateIam := functionTableUpdateIntermediaryTestCase(loader)
	destroyCase, destroyIam := functionTableDestroyIntermediaryTestCase(loader)

	cases := []struct {
		testCase plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service,
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
				*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service,
			]{c.testCase},
			functionTableLinkFactory(c.iamSvc),
			&s.Suite,
		)
	}
}

func fdtRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: fdtRoleResourceName,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(fdtRoleName),
			"arn", core.MappingNodeFromString(fdtRoleARN),
		),
	}
}

func fdtLambdaSvc() lambdaservice.Service {
	return lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(functionWithEnvOutput(map[string]string{})),
	)
}

const (
	fdtRoleName         = "process-orders-role"
	fdtRoleResourceName = "processOrdersFunctionRole"
)

func fdtRoleService() provider.ResourceService {
	return resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(fdtRoleState()))
}

func matchPutInlineAccessPolicy(arg any) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != fdtRoleName ||
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
		if statement.Sid == fdtAccessSID {
			return true
		}
	}
	return false
}

func matchAccessLinkOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	mappingKey := fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		fdtRoleResourceName,
		linkutils.InlineAccessPolicyName(),
		fdtAccessSID,
	)
	summary := map[string]any{}
	if actual != nil {
		summary["mappingValue"] = actual.ResourceDataMappings[mappingKey]
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[fdtExecRole] != nil &&
			actual.LinkData.Fields[fdtExecRole].Fields[linkutils.PermissionFieldName] != nil
	}
	expected := map[string]any{
		"mappingValue": linkutils.PermissionFieldPath(fdtExecRole),
		"hasStatement": true,
	}
	return plugintestutils.EqualityCheckValues{Expected: expected, Actual: summary}, nil
}

func functionTableCreateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service,
], iamservice.Service) {
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service,
	]{
		Name:                           "grants access via a new inline allocator policy on create",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fdtLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    functionResourceInfoA(),
			ResourceBInfo:    tableResourceInfoB(),
			LinkContext:      testLinkContext(),
			ResourceService:  fdtRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool {
				return matchPutInlineAccessPolicy(arg)
			},
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

func functionTableUpdateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service,
], iamservice.Service) {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"` + fdtAccessSID + `","Effect":"Allow","Action":["dynamodb:GetItem"],"Resource":"arn:old"}]}`
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service,
	]{
		Name:                           "replaces the access statement in the inline allocator policy on update",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fdtLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeUpdate,
			InstanceName:     "test-instance",
			ResourceAInfo:    functionResourceInfoA(),
			ResourceBInfo:    tableResourceInfoB(),
			LinkContext:      testLinkContext(),
			ResourceService:  fdtRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchAccessLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutInlineAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

func functionTableDestroyIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service,
], iamservice.Service) {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"` + fdtAccessSID + `","Effect":"Allow","Action":["dynamodb:GetItem"],"Resource":"` + fdtTableARN + `"}]}`
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service,
	]{
		Name:                           "removes the access statement and deletes the empty inline policy on destroy",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return fdtLambdaSvc() },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeDestroy,
			InstanceName:     "test-instance",
			ResourceAInfo:    functionResourceInfoA(),
			ResourceBInfo:    tableResourceInfoB(),
			LinkContext:      testLinkContext(),
			ResourceService:  fdtRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		},
		UpdateActionsCalled: map[string]any{
			"DeleteRolePolicy": func(arg any) bool {
				input, ok := arg.(*iam.DeleteRolePolicyInput)
				return ok &&
					aws.ToString(input.RoleName) == fdtRoleName &&
					aws.ToString(input.PolicyName) == linkutils.InlineAccessPolicyName()
			},
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
	}, iamSvc
}

func TestFunctionTableLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionTableLinkUpdateSuite))
}
