//go:build unit

package flexlambda

import (
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
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

const (
	vfFunctionARN   = "arn:aws:lambda:us-west-2:123456789012:function:get-order"
	vfRoleARN       = "arn:aws:iam::123456789012:role/get-order-role"
	vfRoleName      = "get-order-role"
	vfRoleResource  = "getOrderFunctionRole"
	vfLinkID        = "link-vpc-function-1"
	vfENIStatement  = "VPCNetworkInterfacesgetOrderFunction"
	vfExecRoleField = "getOrderFunctionExecutionRole"
)

// The link resolves the function's execution role from the live function configuration,
// so every case has to be able to answer GetFunction.
func vfGetFunctionOutput() *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(vfFunctionARN),
			Role:        aws.String(vfRoleARN),
		},
	}
}

func vfRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: vfRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(vfRoleName),
			"arn", core.MappingNodeFromString(vfRoleARN),
		),
	}
}

// A role with no Bluelink-managed policies yet, which is the state every one of these
// cases starts from.
func vfEmptyRoleIamMock() iamservice.Service {
	return iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
	)
}

func vfRoleService() provider.ResourceService {
	return resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(vfRoleState()),
	)
}

func vpcFunctionLinkFactory(
	iamSvc iamservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, ec2service.Service, *aws.Config, lambdaservice.Service],
) provider.Link {
	build := VPCFunctionLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, ec2service.Service, *aws.Config, lambdaservice.Service],
	) provider.Link {
		return build(VPCToFunctionLinkDeps(deps))
	}
}

type VPCFunctionLinkUpdateSuite struct {
	suite.Suite
}

func functionResourceInfoB() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "getOrderFunction",
		InstanceID:   "instance-1",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(vfFunctionARN),
			),
		},
	}
}

func functionResourceInfoBWithSubnetType(subnetType string) *provider.ResourceInfo {
	info := functionResourceInfoB()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"aws.flexvpc.lambda.subnetType": core.MappingNodeFromString(subnetType),
				},
			},
		},
	}
	return info
}

func expectVPCConfigOutput(
	functionName string,
	subnetIDs, securityGroupIDs []string,
	ipv6AllowedForDualStack bool,
) *provider.LinkUpdateResourceOutput {
	subnetItems := make([]*core.MappingNode, len(subnetIDs))
	for i, id := range subnetIDs {
		subnetItems[i] = core.MappingNodeFromString(id)
	}
	sgItems := make([]*core.MappingNode, len(securityGroupIDs))
	for i, id := range securityGroupIDs {
		sgItems[i] = core.MappingNodeFromString(id)
	}
	// Placing a function also grants its execution role the network interface
	// permissions Lambda requires, and attributes that statement to this link so the
	// role does not report drift for it.
	eniMapping := vfRoleResource +
		`::spec.policies[@.policyName="` + linkutils.InlineAccessPolicyName() +
		`"].policyDocument.statement[@.sid="` + vfENIStatement + `"]`

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			functionName,
			core.MappingNodeFields(
				"vpcConfig",
				core.MappingNodeFields(
					"subnetIds", &core.MappingNode{Items: subnetItems},
					"securityGroupIds", &core.MappingNode{Items: sgItems},
					"ipv6AllowedForDualStack", core.MappingNodeFromBool(ipv6AllowedForDualStack),
				),
			),
			vfExecRoleField,
			core.MappingNodeFields(
				linkutils.PermissionFieldName,
				specENIStatementNode(vfENIStatement),
			),
		),
		ResourceDataMappings: map[string]string{
			functionName + "::spec.vpcConfig.subnetIds":        functionName + ".vpcConfig.subnetIds",
			functionName + "::spec.vpcConfig.securityGroupIds": functionName + ".vpcConfig.securityGroupIds",
			functionName + "::spec.vpcConfig.ipv6AllowedForDualStack": functionName +
				".vpcConfig.ipv6AllowedForDualStack",
			eniMapping: `["` + vfExecRoleField + `"].permission`,
		},
	}
}

// The same VPC with IPv6 CIDRs on its subnets, which is what every flex VPC preset
// actually provisions. Kept separate from flexVPCResourceInfoA so the single-stack
// case stays covered.
func dualStackFlexVPCResourceInfoA() *provider.ResourceInfo {
	info := flexVPCResourceInfoA()
	subnets := info.CurrentResourceState.SpecData.Fields["subnets"]
	ipv6CIDRs := map[string]string{
		"private-az-1": "2001:db8::/64",
		"private-az-2": "2001:db8:0:1::/64",
		"public-az-1":  "2001:db8:0:2::/64",
	}
	for name, cidr := range ipv6CIDRs {
		subnets.Fields[name].Fields["ipv6CidrBlock"] = core.MappingNodeFromString(cidr)
	}
	return info
}

func flexVPCResourceInfoA() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "appVpc",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"name", core.MappingNodeFromString("orders-vpc"),
				"vpcId", core.MappingNodeFromString("vpc-1"),
				"subnets", core.MappingNodeFields(
					"private-az-1", core.MappingNodeFields(
						"id", core.MappingNodeFromString("subnet-priv-b"),
						"subnetType", core.MappingNodeFromString("private"),
					),
					"private-az-2", core.MappingNodeFields(
						"id", core.MappingNodeFromString("subnet-priv-a"),
						"subnetType", core.MappingNodeFromString("private"),
					),
					"public-az-1", core.MappingNodeFields(
						"id", core.MappingNodeFromString("subnet-pub-a"),
						"subnetType", core.MappingNodeFromString("public"),
					),
				),
				"securityGroups", &core.MappingNode{
					Items: []*core.MappingNode{core.MappingNodeFromString("sg-123")},
				},
			),
		},
	}
}

func (s *VPCFunctionLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		vpcFunctionPlaceTestCase(loader),
		vpcFunctionPlacePublicTestCase(loader),
		vpcFunctionPlaceDualStackTestCase(loader),
		vpcFunctionPlacePartiallyDualStackTestCase(loader),
		vpcFunctionDetachTestCase(loader),
		vpcFunctionNoMatchingTierTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		vpcFunctionLinkFactory(vfEmptyRoleIamMock()),
		&s.Suite,
	)
}

func vpcFunctionPlaceTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "places the function in the VPC's private subnets",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         placementEC2ServiceFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoB(),
			OtherResourceInfo: flexVPCResourceInfoA(),
			LinkContext:       testLinkContext(),
			LinkID:            vfLinkID,
			ResourceService:   vfRoleService(),
		},
		ExpectedOutput: expectVPCConfigOutput(
			"getOrderFunction",
			[]string{"subnet-priv-a", "subnet-priv-b"},
			[]string{"sg-workload"},
			false,
		),
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(vfFunctionARN),
				VpcConfig: &lambdatypes.VpcConfig{
					SubnetIds:               []string{"subnet-priv-a", "subnet-priv-b"},
					SecurityGroupIds:        []string{"sg-workload"},
					Ipv6AllowedForDualStack: aws.Bool(false),
				},
			},
		},
	}
}

// A function placed in dual-stack subnets is given an IPv6 address, which is the only
// way a VPC-attached function reaches the internet without a NAT gateway: it never
// receives a public IPv4 address, so the public subnet's internet gateway is unusable
// over IPv4 no matter which tier it sits in.
func vpcFunctionPlaceDualStackTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "allows outbound IPv6 when the VPC's subnets are dual-stack",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         placementEC2ServiceFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoB(),
			OtherResourceInfo: dualStackFlexVPCResourceInfoA(),
			LinkContext:       testLinkContext(),
			LinkID:            vfLinkID,
			ResourceService:   vfRoleService(),
		},
		ExpectedOutput: expectVPCConfigOutput(
			"getOrderFunction",
			[]string{"subnet-priv-a", "subnet-priv-b"},
			[]string{"sg-workload"},
			true,
		),
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(vfFunctionARN),
				VpcConfig: &lambdatypes.VpcConfig{
					SubnetIds:               []string{"subnet-priv-a", "subnet-priv-b"},
					SecurityGroupIds:        []string{"sg-workload"},
					Ipv6AllowedForDualStack: aws.Bool(true),
				},
			},
		},
	}
}

// A function is placed in every subnet of its tier, so if one of them cannot hand out
// an IPv6 address then whether an invocation has IPv6 depends on which subnet it lands
// in. IPv6 stays off rather than working intermittently.
func vpcFunctionPlacePartiallyDualStackTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(
			&lambda.UpdateFunctionConfigurationOutput{},
		),
	)

	vpcInfo := dualStackFlexVPCResourceInfoA()
	delete(
		vpcInfo.CurrentResourceState.SpecData.Fields["subnets"].Fields["private-az-1"].Fields,
		"ipv6CidrBlock",
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "leaves outbound IPv6 off when only some of the tier's subnets are dual-stack",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         placementEC2ServiceFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoB(),
			OtherResourceInfo: vpcInfo,
			LinkContext:       testLinkContext(),
			LinkID:            vfLinkID,
			ResourceService:   vfRoleService(),
		},
		ExpectedOutput: expectVPCConfigOutput(
			"getOrderFunction",
			[]string{"subnet-priv-a", "subnet-priv-b"},
			[]string{"sg-workload"},
			false,
		),
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(vfFunctionARN),
				VpcConfig: &lambdatypes.VpcConfig{
					SubnetIds:               []string{"subnet-priv-a", "subnet-priv-b"},
					SecurityGroupIds:        []string{"sg-workload"},
					Ipv6AllowedForDualStack: aws.Bool(false),
				},
			},
		},
	}
}

func vpcFunctionPlacePublicTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "places the function in the VPC's public subnets when subnetType is public",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         placementEC2ServiceFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoBWithSubnetType("public"),
			OtherResourceInfo: flexVPCResourceInfoA(),
			LinkContext:       testLinkContext(),
			LinkID:            vfLinkID,
			ResourceService:   vfRoleService(),
		},
		ExpectedOutput: expectVPCConfigOutput(
			"getOrderFunction",
			[]string{"subnet-pub-a"},
			[]string{"sg-workload"},
			false,
		),
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(vfFunctionARN),
				VpcConfig: &lambdatypes.VpcConfig{
					SubnetIds:               []string{"subnet-pub-a"},
					SecurityGroupIds:        []string{"sg-workload"},
					Ipv6AllowedForDualStack: aws.Bool(false),
				},
			},
		},
	}
}

func vpcFunctionNoMatchingTierTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "returns an error when the VPC has no subnets in the requested tier",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         placementEC2ServiceFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      functionResourceInfoBWithSubnetType("isolated"),
			OtherResourceInfo: flexVPCResourceInfoA(),
			LinkContext:       testLinkContext(),
			LinkID:            vfLinkID,
			ResourceService:   vfRoleService(),
		},
		ExpectError:            true,
		ExpectedErrorMessage:   "no \"isolated\" subnets",
		UpdateActionsNotCalled: []string{"UpdateFunctionConfiguration"},
	}
}

func vpcFunctionDetachTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(&lambda.UpdateFunctionConfigurationOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		ec2service.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "detaches the function from the VPC on destroy",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         placementEC2ServiceFactory(),
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      functionResourceInfoB(),
			OtherResourceInfo: flexVPCResourceInfoA(),
			LinkContext:       testLinkContext(),
			LinkID:            vfLinkID,
			ResourceService:   vfRoleService(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields("getOrderFunction", core.MappingNodeFields()),
			ResourceDataMappings: map[string]string{},
		},
		UpdateActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(vfFunctionARN),
				VpcConfig: &lambdatypes.VpcConfig{
					SubnetIds:               []string{},
					SecurityGroupIds:        []string{},
					Ipv6AllowedForDualStack: aws.Bool(false),
				},
			},
		},
	}
}

func TestVPCFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(VPCFunctionLinkUpdateSuite))
}

// The placement link prepares a security group for the function it places, so every
// case needs the EC2 calls behind find-or-create. Returning no existing group means
// each test exercises the create path.
func placementEC2ServiceFactory() func(*aws.Config, provider.Context) ec2service.Service {
	return ec2mock.CreateEc2ServiceMockFactory(
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{}),
		ec2mock.WithCreateSecurityGroupOutput(&ec2.CreateSecurityGroupOutput{
			GroupId: aws.String("sg-workload"),
		}),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
		ec2mock.WithDeleteSecurityGroupOutput(&ec2.DeleteSecurityGroupOutput{}),
	)
}
