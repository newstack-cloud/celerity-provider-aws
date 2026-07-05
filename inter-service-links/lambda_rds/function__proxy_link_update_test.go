//go:build unit

package lambdards

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

const (
	fpFunctionARN  = "arn:aws:lambda:us-west-2:123456789012:function:api"
	fpRoleARN      = "arn:aws:iam::123456789012:role/api-role"
	fpRoleName     = "api-role"
	fpRoleResource = "apiFunctionRole"
	fpExecRole     = "apiFunctionExecutionRole"
	fpConnectSID   = "RDSConnectordersProxy"
	fpPrefix       = "ORDERS_DB"
)

type FunctionProxyLinkUpdateSuite struct {
	suite.Suite
}

func fpFunctionInfo(annotations map[string]*core.MappingNode) *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "apiFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("arn", core.MappingNodeFromString(fpFunctionARN)),
		},
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: annotations},
			},
		},
	}
}

func fpEnvVarAnnotations() map[string]*core.MappingNode {
	return map[string]*core.MappingNode{
		"aws.lambda.rds.ordersProxy.envVarPrefix": core.MappingNodeFromString(fpPrefix),
		"aws.lambda.rds.ordersProxy.databaseName": core.MappingNodeFromString("orders"),
	}
}

func fpProxyInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersProxy",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"endpoint", core.MappingNodeFromString(testProxyEndpoint),
				"dbProxyArn", core.MappingNodeFromString(testProxyARN),
				"vpcSecurityGroupIds", &core.MappingNode{Items: []*core.MappingNode{
					core.MappingNodeFromString(testProxySGID),
				}},
			),
		},
	}
}

func fpGetFunctionOutput(vars map[string]string, vpcConfig *lambdatypes.VpcConfigResponse) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(fpFunctionARN),
			Role:        aws.String(fpRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{Variables: vars},
			VpcConfig:   vpcConfig,
		},
	}
}

func (s *FunctionProxyLinkUpdateSuite) Test_update_resource_a_env_vars() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		fpAddEnvVarsTestCase(loader),
		fpRemoveEnvVarsTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		functionProxyLinkFactory(iammock.CreateIamServiceMock(), noopEC2ServiceFactory()),
		&s.Suite,
	)
}

func fpAddEnvVarsTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fpGetFunctionOutput(map[string]string{"EXISTING": "val"}, nil)),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "populates connection env vars from the proxy endpoint",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      fpFunctionInfo(fpEnvVarAnnotations()),
			OtherResourceInfo: fpProxyInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"apiFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						fpPrefix+"_DATABASE", core.MappingNodeFromString("orders"),
						fpPrefix+"_HOST", core.MappingNodeFromString(testProxyEndpoint),
						fpPrefix+"_PORT", core.MappingNodeFromString("5432"),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				"apiFunction::spec.environment.variables[\"" + fpPrefix + "_DATABASE\"]": "apiFunction.environmentVariables[\"" + fpPrefix + "_DATABASE\"]",
				"apiFunction::spec.environment.variables[\"" + fpPrefix + "_HOST\"]":     "apiFunction.environmentVariables[\"" + fpPrefix + "_HOST\"]",
				"apiFunction::spec.environment.variables[\"" + fpPrefix + "_PORT\"]":     "apiFunction.environmentVariables[\"" + fpPrefix + "_PORT\"]",
			},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": func(arg any) bool {
				in, ok := arg.(*lambda.UpdateFunctionConfigurationInput)
				if !ok || in.Environment == nil {
					return false
				}
				v := in.Environment.Variables
				return v["EXISTING"] == "val" &&
					v[fpPrefix+"_HOST"] == testProxyEndpoint &&
					v[fpPrefix+"_PORT"] == "5432" &&
					v[fpPrefix+"_DATABASE"] == "orders"
			},
		},
	}
}

func fpRemoveEnvVarsTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fpGetFunctionOutput(map[string]string{
			"EXISTING":             "val",
			fpPrefix + "_HOST":     testProxyEndpoint,
			fpPrefix + "_PORT":     "5432",
			fpPrefix + "_DATABASE": "orders",
		}, nil)),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "removes connection env vars on destroy",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      fpFunctionInfo(fpEnvVarAnnotations()),
			OtherResourceInfo: fpProxyInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields("apiFunction", core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": func(arg any) bool {
				in, ok := arg.(*lambda.UpdateFunctionConfigurationInput)
				if !ok || in.Environment == nil {
					return false
				}
				v := in.Environment.Variables
				_, hasHost := v[fpPrefix+"_HOST"]
				return v["EXISTING"] == "val" && !hasHost
			},
		},
	}
}

func fpRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: fpRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(fpRoleName),
			"arn", core.MappingNodeFromString(fpRoleARN),
		),
	}
}

// Test_update_intermediary_resources_iam_grant: authMode=iam on a non-VPC function grants
// rds-db:connect (networking is a no-op when the function is not VPC-attached).
func (s *FunctionProxyLinkUpdateSuite) Test_update_intermediary_resources_iam_grant() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fpGetFunctionOutput(map[string]string{}, nil)),
	)
	rs := resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(fpRoleState()))

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "grants rds-db:connect scoped to the proxy and db user",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			InstanceName:   "test-instance",
			ResourceAInfo: fpFunctionInfo(map[string]*core.MappingNode{
				"aws.lambda.rds.ordersProxy.authMode": core.MappingNodeFromString("iam"),
				"aws.lambda.rds.ordersProxy.dbUser":   core.MappingNodeFromString("orders_app"),
			}),
			ResourceBInfo:    fpProxyInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  rs,
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: fpMatchConnectOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return fpMatchConnectPolicy(arg, "orders_app") },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionProxyLinkFactory(iamSvc, noopEC2ServiceFactory()),
		&s.Suite,
	)
}

// Test_update_intermediary_resources_sg_pair: authMode=password on a VPC-attached function
// opens the SG-pair rule (no IAM grant).
func (s *FunctionProxyLinkUpdateSuite) Test_update_intermediary_resources_sg_pair() {
	loader := &testutils.MockAWSConfigLoader{}

	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fpGetFunctionOutput(map[string]string{}, &lambdatypes.VpcConfigResponse{
			VpcId:            aws.String("vpc-1"),
			SubnetIds:        []string{"subnet-1"},
			SecurityGroupIds: []string{"sg-caller"},
		})),
	)
	// Password mode does not look up the role, so the single lookup resolves the flex VPC.
	rs := resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(&state.ResourceState{
		Name: "appVpc",
		SpecData: core.MappingNodeFields(
			"name", core.MappingNodeFromString("orders-vpc"),
			"enableDNSSupport", core.MappingNodeFromBool(true),
			"enableDNSHostnames", core.MappingNodeFromBool(true),
		),
	}))

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "opens the SG-pair rule for a VPC-attached function",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &ec2Svc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    fpFunctionInfo(map[string]*core.MappingNode{}),
			ResourceBInfo:    fpProxyInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  rs,
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
		UpdateActionsCalled: map[string]any{
			"AuthorizeSecurityGroupIngress": func(arg any) bool {
				in, ok := arg.(*ec2.AuthorizeSecurityGroupIngressInput)
				return ok && aws.ToString(in.GroupId) == testProxySGID && len(in.IpPermissions) == 1 &&
					aws.ToInt32(in.IpPermissions[0].FromPort) == 5432
			},
			"AuthorizeSecurityGroupEgress": func(arg any) bool {
				in, ok := arg.(*ec2.AuthorizeSecurityGroupEgressInput)
				return ok && aws.ToString(in.GroupId) == "sg-caller"
			},
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
	}

	// The asserted service here is ec2 (SG-pair); the iam service is unused in password mode.
	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionProxyLinkFactory(
			iammock.CreateIamServiceMock(),
			func(c *aws.Config, pc provider.Context) ec2service.Service { return ec2Svc },
		),
		&s.Suite,
	)
}

func fpMatchConnectPolicy(arg any, dbUser string) bool {
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
	wantResource := "arn:aws:rds-db:us-west-2:123456789012:dbuser:" + testProxyID + "/" + dbUser
	for _, statement := range doc.Statement {
		if statement.Sid != fpConnectSID {
			continue
		}
		return len(statement.Action) == 1 && statement.Action[0] == "rds-db:connect" &&
			len(statement.Resource) == 1 && statement.Resource[0] == wantResource
	}
	return false
}

func fpMatchConnectOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	summary := map[string]any{}
	if actual != nil {
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[fpExecRole] != nil &&
			actual.LinkData.Fields[fpExecRole].Fields[linkutils.PermissionFieldName] != nil
	}
	return plugintestutils.EqualityCheckValues{
		Expected: map[string]any{"hasStatement": true},
		Actual:   summary,
	}, nil
}

func TestFunctionProxyLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionProxyLinkUpdateSuite))
}
