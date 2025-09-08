//go:build unit

package iam

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMGroupResourceUpdateSuite struct {
	suite.Suite
}

func (s *IAMGroupResourceUpdateSuite) Test_update_iam_group() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		createBasicGroupUpdateTestCase(providerCtx, loader),
		createGroupNoUpdatesTestCase(providerCtx, loader),
		createGroupPoliciesUpdateTestCase(providerCtx, loader),
		createGroupManagedPoliciesUpdateTestCase(providerCtx, loader),
		createGroupUpdateFailureTestCase(providerCtx, loader),
		recreateGroupOnGroupNameChangeTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		GroupResource,
		&s.Suite,
	)
}

func createBasicGroupUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group"
	groupId := "AGPA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateGroupOutput(&iam.UpdateGroupOutput{}),
		iammock.WithGetGroupOutput(&iam.GetGroupOutput{
			Group: &types.Group{
				Arn:       aws.String(resourceARN),
				GroupId:   aws.String(groupId),
				GroupName: aws.String("test-group"),
				Path:      aws.String("/updated/"),
			},
		}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/updated/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update group path",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
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
			ResourceID: "test-group-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-group-id",
					ResourceName: "TestGroup",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-group-id",
						Name:       "TestGroup",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.path",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":       core.MappingNodeFromString(resourceARN),
				"spec.groupId":   core.MappingNodeFromString(groupId),
				"spec.groupName": core.MappingNodeFromString("test-group"),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetGroup": &iam.GetGroupInput{
				GroupName: aws.String("test-group"),
			},
			"UpdateGroup": &iam.UpdateGroupInput{
				GroupName: aws.String("test-group"),
				NewPath:   aws.String("/updated/"),
			},
		},
	}
}

func createGroupNoUpdatesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group"
	groupId := "AGPA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithGetGroupOutput(&iam.GetGroupOutput{
			Group: &types.Group{
				Arn:       aws.String(resourceARN),
				GroupId:   aws.String(groupId),
				GroupName: aws.String("test-group"),
				Path:      aws.String("/"),
			},
		}),
	)

	// Current state (same as updated state)
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "no updates",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
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
			ResourceID: "test-group-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-group-id",
					ResourceName: "TestGroup",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-group-id",
						Name:       "TestGroup",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":       core.MappingNodeFromString(resourceARN),
				"spec.groupId":   core.MappingNodeFromString(groupId),
				"spec.groupName": core.MappingNodeFromString("test-group"),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetGroup": &iam.GetGroupInput{
				GroupName: aws.String("test-group"),
			},
		},
	}
}

func createGroupPoliciesUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group"
	groupId := "AGPA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithGetGroupOutput(&iam.GetGroupOutput{
			Group: &types.Group{
				Arn:       aws.String(resourceARN),
				GroupId:   aws.String(groupId),
				GroupName: aws.String("test-group"),
				Path:      aws.String("/"),
			},
		}),
		iammock.WithListGroupPoliciesOutput(&iam.ListGroupPoliciesOutput{
			PolicyNames: []string{"OldPolicy"},
		}),
		iammock.WithDeleteGroupPolicyOutput(&iam.DeleteGroupPolicyOutput{}),
		iammock.WithPutGroupPolicyOutput(&iam.PutGroupPolicyOutput{}),
	)

	// Current state with old policies
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
			"policies": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("OldPolicy"),
							"policyDocument": {
								Fields: map[string]*core.MappingNode{
									"Version": core.MappingNodeFromString("2012-10-17"),
									"Statement": {
										Items: []*core.MappingNode{
											{
												Fields: map[string]*core.MappingNode{
													"Effect": core.MappingNodeFromString("Allow"),
													"Action": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("s3:GetObject"),
														},
													},
													"Resource": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("*"),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Updated state with new policies
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
			"policies": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("NewPolicy"),
							"policyDocument": {
								Fields: map[string]*core.MappingNode{
									"Version": core.MappingNodeFromString("2012-10-17"),
									"Statement": {
										Items: []*core.MappingNode{
											{
												Fields: map[string]*core.MappingNode{
													"Effect": core.MappingNodeFromString("Allow"),
													"Action": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("s3:PutObject"),
														},
													},
													"Resource": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("*"),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update group policies",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
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
			ResourceID: "test-group-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-group-id",
					ResourceName: "TestGroup",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-group-id",
						Name:       "TestGroup",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.policies",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":       core.MappingNodeFromString(resourceARN),
				"spec.groupId":   core.MappingNodeFromString(groupId),
				"spec.groupName": core.MappingNodeFromString("test-group"),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetGroup": &iam.GetGroupInput{
				GroupName: aws.String("test-group"),
			},
			"DeleteGroupPolicy": &iam.DeleteGroupPolicyInput{
				GroupName:  aws.String("test-group"),
				PolicyName: aws.String("OldPolicy"),
			},
			"PutGroupPolicy": &iam.PutGroupPolicyInput{
				GroupName:      aws.String("test-group"),
				PolicyName:     aws.String("NewPolicy"),
				PolicyDocument: aws.String(`{"Statement":[{"Action":["s3:PutObject"],"Effect":"Allow","Resource":["*"]}],"Version":"2012-10-17"}`),
			},
		},
	}
}

func createGroupManagedPoliciesUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group"
	groupId := "AGPA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithGetGroupOutput(&iam.GetGroupOutput{
			Group: &types.Group{
				Arn:       aws.String(resourceARN),
				GroupId:   aws.String(groupId),
				GroupName: aws.String("test-group"),
				Path:      aws.String("/"),
			},
		}),
		iammock.WithListAttachedGroupPoliciesOutput(&iam.ListAttachedGroupPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{
				{
					PolicyArn:  aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					PolicyName: aws.String("ReadOnlyAccess"),
				},
			},
		}),
		iammock.WithDetachGroupPolicyOutput(&iam.DetachGroupPolicyOutput{}),
		iammock.WithAttachGroupPolicyOutput(&iam.AttachGroupPolicyOutput{}),
	)

	// Current state with old managed policies
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
				},
			},
		},
	}

	// Updated state with new managed policies
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/PowerUserAccess"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update group managed policies",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
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
			ResourceID: "test-group-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-group-id",
					ResourceName: "TestGroup",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-group-id",
						Name:       "TestGroup",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.managedPolicyArns",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":       core.MappingNodeFromString(resourceARN),
				"spec.groupId":   core.MappingNodeFromString(groupId),
				"spec.groupName": core.MappingNodeFromString("test-group"),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetGroup": &iam.GetGroupInput{
				GroupName: aws.String("test-group"),
			},
			"DetachGroupPolicy": &iam.DetachGroupPolicyInput{
				GroupName: aws.String("test-group"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
			},
			"AttachGroupPolicy": &iam.AttachGroupPolicyInput{
				GroupName: aws.String("test-group"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
			},
		},
	}
}

func createGroupUpdateFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithGetGroupError(fmt.Errorf("failed to get group")),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString("arn:aws:iam::123456789012:group/test-group"),
			"groupId":   core.MappingNodeFromString("AGPA1234567890123456"),
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/updated/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update group failure",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
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
			ResourceID: "test-group-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-group-id",
					ResourceName: "TestGroup",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-group-id",
						Name:       "TestGroup",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.path",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"GetGroup": &iam.GetGroupInput{
				GroupName: aws.String("test-group"),
			},
		},
	}
}

func recreateGroupOnGroupNameChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/OldGroup"
	groupId := "AGPA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteGroupOutput(&iam.DeleteGroupOutput{}),
		iammock.WithCreateGroupOutput(&iam.CreateGroupOutput{
			Group: &types.Group{
				Arn:       aws.String(resourceARN),
				GroupId:   aws.String(groupId),
				GroupName: aws.String("NewGroup"),
				Path:      aws.String("/"),
			},
		}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("OldGroup"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("NewGroup"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "recreate group on groupName change",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
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
			ResourceID: "test-group-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-group-id",
					ResourceName: "TestGroup",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-group-id",
						Name:       "TestGroup",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.groupName",
						PrevValue: core.MappingNodeFromString("OldGroup"),
						NewValue:  core.MappingNodeFromString("NewGroup"),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":     core.MappingNodeFromString(resourceARN),
				"spec.groupId": core.MappingNodeFromString(groupId),
			},
		},
		SaveActionsCalled: map[string]any{
			"DeleteGroup": &iam.DeleteGroupInput{
				GroupName: aws.String("OldGroup"),
			},
			"CreateGroup": &iam.CreateGroupInput{
				GroupName: aws.String("NewGroup"),
				Path:      aws.String("/"),
			},
		},
	}
}

func TestIAMGroupResourceUpdate(t *testing.T) {
	suite.Run(t, new(IAMGroupResourceUpdateSuite))
}
