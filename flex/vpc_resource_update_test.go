package flex

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FlexVPCResourceUpdateSuite struct {
	suite.Suite
}

func (s *FlexVPCResourceUpdateSuite) Test_update() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		createReferenceModeUpdateTestCase(providerCtx, loader),
		createVPCNoChangesTestCase(providerCtx, loader),
		createVPCEnableDNSSupportUpdateTestCase(providerCtx, loader),
		createVPCEnableDNSHostnamesUpdateTestCase(providerCtx, loader),
		createVPCBothDNSAttributesUpdateTestCase(providerCtx, loader),
		createVPCTagsUpdateTestCase(providerCtx, loader),
		createVPCTagsAddTestCase(providerCtx, loader),
		createVPCTagsRemoveTestCase(providerCtx, loader),
		createVPCTagsModifyTestCase(providerCtx, loader),
		createVPCUpdateWithMissingVPCIDTestCase(providerCtx, loader),
		createVPCWithMissingModeTestCase(providerCtx, loader),
		createVPCWithModifyVpcAttributeErrorTestCase(providerCtx, loader),
		createVPCWithCreateTagsErrorTestCase(providerCtx, loader),
		createVPCWithDeleteTagsErrorTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	vpcResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		// Create a mock resource group tagging service factory for testing
		mockResourceGroupTaggingServiceFactory := func(config *aws.Config, ctx provider.Context) resgrouptagservice.Service {
			return nil // Mock service for testing
		}
		return VPCResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		vpcResourceWrapper,
		&s.Suite,
	)
}

func TestFlexVPCResourceUpdate(t *testing.T) {
	suite.Run(t, new(FlexVPCResourceUpdateSuite))
}

func createReferenceModeUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock()

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("reference"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("reference"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "reference mode update returns early with computed fields",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.mode",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.gateways":       nil,
				"spec.networkAcls":    nil,
				"spec.routeTables":    nil,
				"spec.securityGroups": nil,
				"spec.subnets":        nil,
				"spec.vpcId":          core.MappingNodeFromString("vpc-12345678"),
			},
		},
		SaveActionsNotCalled: []string{
			"ModifyVpcAttribute",
			"CreateTags",
			"DeleteTags",
		},
		ExpectError: false,
	}
}

func createVPCNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock()

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "no changes to VPC",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: currentStateSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.gateways":       nil,
				"spec.networkAcls":    nil,
				"spec.routeTables":    nil,
				"spec.securityGroups": nil,
				"spec.subnets":        nil,
				"spec.vpcId":          core.MappingNodeFromString("vpc-12345678"),
			},
		},
		SaveActionsNotCalled: []string{
			"ModifyVpcAttribute",
			"CreateTags",
			"DeleteTags",
		},
		ExpectError: false,
	}
}

func createVPCEnableDNSSupportUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithModifyVpcAttributeOutput(&ec2.ModifyVpcAttributeOutput{}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId":            core.MappingNodeFromString("vpc-12345678"),
			"mode":             core.MappingNodeFromString("create"),
			"enableDNSSupport": core.MappingNodeFromBool(false),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "update VPC enableDNSSupport attribute",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.enableDNSSupport",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.gateways":       nil,
				"spec.networkAcls":    nil,
				"spec.routeTables":    nil,
				"spec.securityGroups": nil,
				"spec.subnets":        nil,
				"spec.vpcId":          core.MappingNodeFromString("vpc-12345678"),
			},
		},
		SaveActionsCalled: map[string]any{
			"ModifyVpcAttribute": &ec2.ModifyVpcAttributeInput{
				VpcId: aws.String("vpc-12345678"),
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(false),
				},
			},
		},
		ExpectError: false,
	}
}

func createVPCEnableDNSHostnamesUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithModifyVpcAttributeOutput(&ec2.ModifyVpcAttributeOutput{}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId":              core.MappingNodeFromString("vpc-12345678"),
			"mode":               core.MappingNodeFromString("create"),
			"enableDNSHostnames": core.MappingNodeFromBool(true),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "update VPC enableDNSHostnames attribute",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.enableDNSHostnames",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.gateways":       nil,
				"spec.networkAcls":    nil,
				"spec.routeTables":    nil,
				"spec.securityGroups": nil,
				"spec.subnets":        nil,
				"spec.vpcId":          core.MappingNodeFromString("vpc-12345678"),
			},
		},
		SaveActionsCalled: map[string]any{
			"ModifyVpcAttribute": &ec2.ModifyVpcAttributeInput{
				VpcId: aws.String("vpc-12345678"),
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
			},
		},
		ExpectError: false,
	}
}

func createVPCBothDNSAttributesUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithModifyVpcAttributeOutput(&ec2.ModifyVpcAttributeOutput{}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId":              core.MappingNodeFromString("vpc-12345678"),
			"mode":               core.MappingNodeFromString("create"),
			"enableDNSSupport":   core.MappingNodeFromBool(false),
			"enableDNSHostnames": core.MappingNodeFromBool(true),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "update VPC both DNS attributes",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.enableDNSSupport",
					},
					{
						FieldPath: "spec.enableDNSHostnames",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.gateways":       nil,
				"spec.networkAcls":    nil,
				"spec.routeTables":    nil,
				"spec.securityGroups": nil,
				"spec.subnets":        nil,
				"spec.vpcId":          core.MappingNodeFromString("vpc-12345678"),
			},
		},
		SaveActionsCalled: map[string]any{
			"ModifyVpcAttribute": &ec2.ModifyVpcAttributeInput{
				VpcId: aws.String("vpc-12345678"),
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(false),
				},
			},
		},
		ExpectError: false,
	}
}

func createVPCTagsUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithCreateTagsOutput(&ec2.CreateTagsOutput{}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
			"tags": core.MappingNodeItems(
				core.MappingNodeFields(
					"key", core.MappingNodeFromString("Environment"),
					"value", core.MappingNodeFromString("dev"),
				),
			),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
			"tags": core.MappingNodeItems(
				core.MappingNodeFields(
					"key", core.MappingNodeFromString("Environment"),
					"value", core.MappingNodeFromString("prod"),
				),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "update VPC tags",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.gateways":       nil,
				"spec.networkAcls":    nil,
				"spec.routeTables":    nil,
				"spec.securityGroups": nil,
				"spec.subnets":        nil,
				"spec.vpcId":          core.MappingNodeFromString("vpc-12345678"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateTags": &ec2.CreateTagsInput{
				Resources: []string{"vpc-12345678"},
				Tags: []types.Tag{
					{
						Key:   aws.String("Environment"),
						Value: aws.String("prod"),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createVPCTagsAddTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithCreateTagsOutput(&ec2.CreateTagsOutput{}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
			"tags": core.MappingNodeItems(
				core.MappingNodeFields(
					"key", core.MappingNodeFromString("Environment"),
					"value", core.MappingNodeFromString("prod"),
				),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "add VPC tags",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.gateways":       nil,
				"spec.networkAcls":    nil,
				"spec.routeTables":    nil,
				"spec.securityGroups": nil,
				"spec.subnets":        nil,
				"spec.vpcId":          core.MappingNodeFromString("vpc-12345678"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateTags": &ec2.CreateTagsInput{
				Resources: []string{"vpc-12345678"},
				Tags: []types.Tag{
					{
						Key:   aws.String("Environment"),
						Value: aws.String("prod"),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createVPCTagsRemoveTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDeleteTagsOutput(&ec2.DeleteTagsOutput{}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
			"tags": core.MappingNodeItems(
				core.MappingNodeFields(
					"key", core.MappingNodeFromString("Environment"),
					"value", core.MappingNodeFromString("prod"),
				),
			),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "remove VPC tags",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.gateways":       nil,
				"spec.networkAcls":    nil,
				"spec.routeTables":    nil,
				"spec.securityGroups": nil,
				"spec.subnets":        nil,
				"spec.vpcId":          core.MappingNodeFromString("vpc-12345678"),
			},
		},
		SaveActionsCalled: map[string]any{
			"DeleteTags": &ec2.DeleteTagsInput{
				Resources: []string{"vpc-12345678"},
				Tags: []types.Tag{
					{
						Key: aws.String("Environment"),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createVPCTagsModifyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithCreateTagsOutput(&ec2.CreateTagsOutput{}),
		ec2mock.WithDeleteTagsOutput(&ec2.DeleteTagsOutput{}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
			"tags": core.MappingNodeItems(
				core.MappingNodeFields(
					"key", core.MappingNodeFromString("Environment"),
					"value", core.MappingNodeFromString("dev"),
				),
			),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
			"tags": core.MappingNodeItems(
				core.MappingNodeFields(
					"key", core.MappingNodeFromString("Environment"),
					"value", core.MappingNodeFromString("prod"),
				),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "modify VPC tags",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.gateways":       nil,
				"spec.networkAcls":    nil,
				"spec.routeTables":    nil,
				"spec.securityGroups": nil,
				"spec.subnets":        nil,
				"spec.vpcId":          core.MappingNodeFromString("vpc-12345678"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateTags": &ec2.CreateTagsInput{
				Resources: []string{"vpc-12345678"},
				Tags: []types.Tag{
					{
						Key:   aws.String("Environment"),
						Value: aws.String("prod"),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createVPCUpdateWithMissingVPCIDTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock()

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"mode": core.MappingNodeFromString("create"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"mode": core.MappingNodeFromString("create"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when vpcId is missing",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.enableDNSSupport",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCWithMissingModeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock()

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when mode is missing",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.enableDNSSupport",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCWithModifyVpcAttributeErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithModifyVpcAttributeError(errors.New("failed to modify VPC attribute")),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId":            core.MappingNodeFromString("vpc-12345678"),
			"mode":             core.MappingNodeFromString("create"),
			"enableDNSSupport": core.MappingNodeFromBool(false),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when ModifyVpcAttribute fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.enableDNSSupport",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCWithCreateTagsErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithCreateTagsError(errors.New("failed to create tags")),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
			"tags": core.MappingNodeItems(
				core.MappingNodeFields(
					"key", core.MappingNodeFromString("Environment"),
					"value", core.MappingNodeFromString("prod"),
				),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when CreateTags fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCWithDeleteTagsErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDeleteTagsError(errors.New("failed to delete tags")),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
			"tags": core.MappingNodeItems(
				core.MappingNodeFields(
					"key", core.MappingNodeFromString("Environment"),
					"value", core.MappingNodeFromString("prod"),
				),
			),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"mode":  core.MappingNodeFromString("create"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when DeleteTags fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "vpc-12345678",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "vpc-12345678",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "vpc-12345678",
						Name:       "TestVPC",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}
