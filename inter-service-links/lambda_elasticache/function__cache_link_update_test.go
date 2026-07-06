//go:build unit

package lambdaelasticache

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

type FunctionCacheLinkUpdateSuite struct {
	suite.Suite
}

func lcEnvVarAnnotations() map[string]*core.MappingNode {
	return map[string]*core.MappingNode{
		"aws.lambda.elasticache.sessionCache.envVarPrefix": core.MappingNodeFromString(lcPrefix),
	}
}

func (s *FunctionCacheLinkUpdateSuite) Test_update_resource_a_env_vars() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		lcAddEnvVarsTestCase(loader),
		lcRemoveEnvVarsTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		functionCacheLinkFactory(iammock.CreateIamServiceMock(), noopEC2ServiceFactory()),
		&s.Suite,
	)
}

func lcAddEnvVarsTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(lcGetFunctionOutput(map[string]string{"EXISTING": "val"}, nil)),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "populates connection env vars from the cache endpoint",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      lcFunctionInfo(lcEnvVarAnnotations()),
			OtherResourceInfo: lcCacheInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"apiFunction",
				core.MappingNodeFields(
					"environmentVariables",
					core.MappingNodeFields(
						lcPrefix+"_HOST", core.MappingNodeFromString(testCacheEndpoint),
						lcPrefix+"_PORT", core.MappingNodeFromString("6379"),
					),
				),
			),
			ResourceDataMappings: map[string]string{
				"apiFunction::spec.environment.variables[\"" + lcPrefix + "_HOST\"]": "apiFunction.environmentVariables[\"" + lcPrefix + "_HOST\"]",
				"apiFunction::spec.environment.variables[\"" + lcPrefix + "_PORT\"]": "apiFunction.environmentVariables[\"" + lcPrefix + "_PORT\"]",
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
					v[lcPrefix+"_HOST"] == testCacheEndpoint &&
					v[lcPrefix+"_PORT"] == "6379"
			},
		},
	}
}

func lcRemoveEnvVarsTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(lcGetFunctionOutput(map[string]string{
			"EXISTING":         "val",
			lcPrefix + "_HOST": testCacheEndpoint,
			lcPrefix + "_PORT": "6379",
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
			ResourceInfo:      lcFunctionInfo(lcEnvVarAnnotations()),
			OtherResourceInfo: lcCacheInfo(),
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
				_, hasHost := v[lcPrefix+"_HOST"]
				return v["EXISTING"] == "val" && !hasHost
			},
		},
	}
}

// Test_update_intermediary_resources_sg_pair: a VPC-attached function opens the SG-pair rule to the
// cache. The cache's SG comes from the resolved spec (securityGroupIds is write-only).
func (s *FunctionCacheLinkUpdateSuite) Test_update_intermediary_resources_sg_pair() {
	loader := &testutils.MockAWSConfigLoader{}

	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(lcGetFunctionOutput(map[string]string{}, &lambdatypes.VpcConfigResponse{
			VpcId:            aws.String("vpc-1"),
			SubnetIds:        []string{"subnet-1"},
			SecurityGroupIds: []string{"sg-caller"},
		})),
	)
	rs := resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(&state.ResourceState{
		Name: "appVpc",
		SpecData: core.MappingNodeFields(
			"name", core.MappingNodeFromString("app-vpc"),
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
			ResourceAInfo:    lcFunctionInfo(map[string]*core.MappingNode{}),
			ResourceBInfo:    lcCacheInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  rs,
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
		UpdateActionsCalled: map[string]any{
			"AuthorizeSecurityGroupIngress": func(arg any) bool {
				in, ok := arg.(*ec2.AuthorizeSecurityGroupIngressInput)
				return ok && aws.ToString(in.GroupId) == testCacheSGID && len(in.IpPermissions) == 1 &&
					aws.ToInt32(in.IpPermissions[0].FromPort) == 6379
			},
			"AuthorizeSecurityGroupEgress": func(arg any) bool {
				in, ok := arg.(*ec2.AuthorizeSecurityGroupEgressInput)
				return ok && aws.ToString(in.GroupId) == "sg-caller"
			},
		},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionCacheLinkFactory(
			iammock.CreateIamServiceMock(),
			func(c *aws.Config, pc provider.Context) ec2service.Service { return ec2Svc },
		),
		&s.Suite,
	)
}

// Test_update_intermediary_resources_iam_grant: authMode=iam on a non-VPC function grants
// elasticache:Connect scoped to the replication group and user (networking is a no-op when the
// function is not VPC-attached).
func (s *FunctionCacheLinkUpdateSuite) Test_update_intermediary_resources_iam_grant() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(lcGetFunctionOutput(map[string]string{}, nil)),
	)
	rs := resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(lcRoleState()))

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "grants elasticache:Connect scoped to the cache and user",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			InstanceName:   "test-instance",
			ResourceAInfo: lcFunctionInfo(map[string]*core.MappingNode{
				"aws.lambda.elasticache.sessionCache.authMode": core.MappingNodeFromString("iam"),
				"aws.lambda.elasticache.sessionCache.userId":   core.MappingNodeFromString("app-user"),
			}),
			ResourceBInfo:    lcCacheInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  rs,
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: lcMatchConnectOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return lcMatchConnectPolicy(arg, "app-user") },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionCacheLinkFactory(iamSvc, noopEC2ServiceFactory()),
		&s.Suite,
	)
}

func lcMatchConnectPolicy(arg any, userId string) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != lcRoleName ||
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
	wantRG := "arn:aws:elasticache:us-west-2:123456789012:replicationgroup:" + testCacheRGId
	wantUser := "arn:aws:elasticache:us-west-2:123456789012:user:" + userId
	for _, statement := range doc.Statement {
		if statement.Sid != lcConnectSID {
			continue
		}
		return len(statement.Action) == 1 && statement.Action[0] == "elasticache:Connect" &&
			len(statement.Resource) == 2 && statement.Resource[0] == wantRG && statement.Resource[1] == wantUser
	}
	return false
}

func lcMatchConnectOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	summary := map[string]any{}
	if actual != nil {
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[lcExecRole] != nil &&
			actual.LinkData.Fields[lcExecRole].Fields[linkutils.PermissionFieldName] != nil
	}
	return plugintestutils.EqualityCheckValues{
		Expected: map[string]any{"hasStatement": true},
		Actual:   summary,
	}, nil
}

func TestFunctionCacheLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionCacheLinkUpdateSuite))
}
