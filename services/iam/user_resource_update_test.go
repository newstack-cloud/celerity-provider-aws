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

type IAMUserResourceUpdateSuite struct {
	suite.Suite
}

func (s *IAMUserResourceUpdateSuite) Test_update_iam_user() {
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
		createBasicUserUpdateTestCase(providerCtx, loader),
		createUserNoUpdatesTestCase(providerCtx, loader),
		createUserTagsUpdateTestCase(providerCtx, loader),
		createUserLoginProfileUpdateTestCase(providerCtx, loader),
		createUserPoliciesUpdateTestCase(providerCtx, loader),
		createUserGroupsUpdateTestCase(providerCtx, loader),
		createUserPermissionsBoundaryUpdateTestCase(providerCtx, loader),
		createUserUpdateFailureTestCase(providerCtx, loader),
		recreateUserOnUserNameChangeTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		UserResource,
		&s.Suite,
	)
}

func createBasicUserUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateUserOutput(&iam.UpdateUserOutput{}),
		iammock.WithGetUserOutput(&iam.GetUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user"),
				Path:     aws.String("/updated/"),
			},
		}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString(resourceARN),
			"userId":   core.MappingNodeFromString(userId),
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/updated/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update user path",
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
			ResourceID: "test-user-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-user-id",
					ResourceName: "TestUser",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-user-id",
						Name:       "TestUser",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
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
				"spec.arn":    core.MappingNodeFromString(resourceARN),
				"spec.userId": core.MappingNodeFromString(userId),
			},
		},
		SaveActionsCalled: map[string]any{
			"UpdateUser": &iam.UpdateUserInput{
				UserName: aws.String("test-user"),
				NewPath:  aws.String("/updated/"),
			},
			"GetUser": &iam.GetUserInput{
				UserName: aws.String("test-user"),
			},
		},
		SaveActionsNotCalled: []string{
			"CreateLoginProfile",
			"UpdateLoginProfile",
			"DeleteLoginProfile",
			"AttachUserPolicy",
			"DetachUserPolicy",
			"PutUserPolicy",
			"DeleteUserPolicy",
			"AddUserToGroup",
			"RemoveUserFromGroup",
			"PutUserPermissionsBoundary",
			"DeleteUserPermissionsBoundary",
			"TagUser",
			"UntagUser",
		},
	}
}

func createUserNoUpdatesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithGetUserOutput(&iam.GetUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		}),
	)

	// Current state matches updated state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString(resourceARN),
			"userId":   core.MappingNodeFromString(userId),
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
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
			ResourceID: "test-user-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-user-id",
					ResourceName: "TestUser",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-user-id",
						Name:       "TestUser",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
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
				"spec.arn":    core.MappingNodeFromString(resourceARN),
				"spec.userId": core.MappingNodeFromString(userId),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetUser": &iam.GetUserInput{
				UserName: aws.String("test-user"),
			},
		},
		SaveActionsNotCalled: []string{
			"UpdateUser",
			"CreateLoginProfile",
			"UpdateLoginProfile",
			"DeleteLoginProfile",
			"AttachUserPolicy",
			"DetachUserPolicy",
			"PutUserPolicy",
			"DeleteUserPolicy",
			"AddUserToGroup",
			"RemoveUserFromGroup",
			"PutUserPermissionsBoundary",
			"DeleteUserPermissionsBoundary",
			"TagUser",
			"UntagUser",
		},
	}
}

func createUserTagsUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithUntagUserOutput(&iam.UntagUserOutput{}),
		iammock.WithTagUserOutput(&iam.TagUserOutput{}),
		iammock.WithGetUserOutput(&iam.GetUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		}),
	)

	// Current state with existing tags
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString(resourceARN),
			"userId":   core.MappingNodeFromString(userId),
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("test"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Team"),
							"value": core.MappingNodeFromString("backend"),
						},
					},
				},
			},
		},
	}

	// Updated state with modified tags
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("production"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Department"),
							"value": core.MappingNodeFromString("engineering"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update user tags",
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
			ResourceID: "test-user-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-user-id",
					ResourceName: "TestUser",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-user-id",
						Name:       "TestUser",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
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
				"spec.arn":    core.MappingNodeFromString(resourceARN),
				"spec.userId": core.MappingNodeFromString(userId),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetUser": &iam.GetUserInput{
				UserName: aws.String("test-user"),
			},
			"UntagUser": &iam.UntagUserInput{
				UserName: aws.String("test-user"),
				TagKeys:  []string{"Team"},
			},
			"TagUser": &iam.TagUserInput{
				UserName: aws.String("test-user"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Department"),
						Value: aws.String("engineering"),
					},
					{
						Key:   aws.String("Environment"),
						Value: aws.String("production"),
					},
				},
			},
		},
	}
}

func createUserLoginProfileUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateLoginProfileOutput(&iam.CreateLoginProfileOutput{}),
		iammock.WithGetUserOutput(&iam.GetUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		}),
	)

	// Current state without login profile
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString(resourceARN),
			"userId":   core.MappingNodeFromString(userId),
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
		},
	}

	// Updated state with login profile
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
			"loginProfile": {
				Fields: map[string]*core.MappingNode{
					"password":              core.MappingNodeFromString("NewPassword123!"),
					"passwordResetRequired": core.MappingNodeFromBool(false),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "add login profile",
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
			ResourceID: "test-user-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-user-id",
					ResourceName: "TestUser",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-user-id",
						Name:       "TestUser",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: updatedSpecData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.loginProfile",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString(resourceARN),
				"spec.userId": core.MappingNodeFromString(userId),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetUser": &iam.GetUserInput{
				UserName: aws.String("test-user"),
			},
			"CreateLoginProfile": &iam.CreateLoginProfileInput{
				UserName:              aws.String("test-user"),
				Password:              aws.String("NewPassword123!"),
				PasswordResetRequired: false,
			},
		},
	}
}

func createUserPoliciesUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithDetachUserPolicyOutput(&iam.DetachUserPolicyOutput{}),
		iammock.WithAttachUserPolicyOutput(&iam.AttachUserPolicyOutput{}),
		iammock.WithDeleteUserPolicyOutput(&iam.DeleteUserPolicyOutput{}),
		iammock.WithPutUserPolicyOutput(&iam.PutUserPolicyOutput{}),
		iammock.WithGetUserOutput(&iam.GetUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		}),
	)

	// Current state with existing policies
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString(resourceARN),
			"userId":   core.MappingNodeFromString(userId),
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
				},
			},
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
															core.MappingNodeFromString("arn:aws:s3:::old-bucket/*"),
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

	// Updated state with modified policies
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/PowerUserAccess"),
				},
			},
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
															core.MappingNodeFromString("s3:GetObject"),
															core.MappingNodeFromString("s3:PutObject"),
														},
													},
													"Resource": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("arn:aws:s3:::new-bucket/*"),
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
		Name: "update user policies",
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
			ResourceID: "test-user-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-user-id",
					ResourceName: "TestUser",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-user-id",
						Name:       "TestUser",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.managedPolicyArns",
					},
					{
						FieldPath: "spec.policies",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString(resourceARN),
				"spec.userId": core.MappingNodeFromString(userId),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetUser": &iam.GetUserInput{
				UserName: aws.String("test-user"),
			},
			"DetachUserPolicy": &iam.DetachUserPolicyInput{
				UserName:  aws.String("test-user"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
			},
			"AttachUserPolicy": &iam.AttachUserPolicyInput{
				UserName:  aws.String("test-user"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
			},
			"DeleteUserPolicy": &iam.DeleteUserPolicyInput{
				UserName:   aws.String("test-user"),
				PolicyName: aws.String("OldPolicy"),
			},
			"PutUserPolicy": &iam.PutUserPolicyInput{
				UserName:       aws.String("test-user"),
				PolicyName:     aws.String("NewPolicy"),
				PolicyDocument: aws.String(`{"Statement":[{"Action":["s3:GetObject","s3:PutObject"],"Effect":"Allow","Resource":["arn:aws:s3:::new-bucket/*"]}],"Version":"2012-10-17"}`),
			},
		},
	}
}

func createUserGroupsUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithRemoveUserFromGroupOutput(&iam.RemoveUserFromGroupOutput{}),
		iammock.WithAddUserToGroupOutput(&iam.AddUserToGroupOutput{}),
		iammock.WithGetUserOutput(&iam.GetUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		}),
	)

	// Current state with existing groups
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString(resourceARN),
			"userId":   core.MappingNodeFromString(userId),
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
			"groups": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("developers"),
					core.MappingNodeFromString("testers"),
				},
			},
		},
	}

	// Updated state with modified groups
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
			"groups": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("developers"),
					core.MappingNodeFromString("admins"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update user groups",
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
			ResourceID: "test-user-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-user-id",
					ResourceName: "TestUser",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-user-id",
						Name:       "TestUser",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.groups",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString(resourceARN),
				"spec.userId": core.MappingNodeFromString(userId),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetUser": &iam.GetUserInput{
				UserName: aws.String("test-user"),
			},
			"RemoveUserFromGroup": &iam.RemoveUserFromGroupInput{
				UserName:  aws.String("test-user"),
				GroupName: aws.String("testers"),
			},
			"AddUserToGroup": &iam.AddUserToGroupInput{
				UserName:  aws.String("test-user"),
				GroupName: aws.String("admins"),
			},
		},
	}
}

func createUserPermissionsBoundaryUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithPutUserPermissionsBoundaryOutput(&iam.PutUserPermissionsBoundaryOutput{}),
		iammock.WithGetUserOutput(&iam.GetUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		}),
	)

	// Current state without permissions boundary
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString(resourceARN),
			"userId":   core.MappingNodeFromString(userId),
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
		},
	}

	// Updated state with permissions boundary
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName":            core.MappingNodeFromString("test-user"),
			"path":                core.MappingNodeFromString("/"),
			"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "add permissions boundary",
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
			ResourceID: "test-user-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-user-id",
					ResourceName: "TestUser",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-user-id",
						Name:       "TestUser",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: updatedSpecData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.permissionsBoundary",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString(resourceARN),
				"spec.userId": core.MappingNodeFromString(userId),
			},
		},
		SaveActionsCalled: map[string]any{
			"GetUser": &iam.GetUserInput{
				UserName: aws.String("test-user"),
			},
			"PutUserPermissionsBoundary": &iam.PutUserPermissionsBoundaryInput{
				UserName:            aws.String("test-user"),
				PermissionsBoundary: aws.String("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
			},
		},
	}
}

func createUserUpdateFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateUserError(fmt.Errorf("failed to update user")),
		iammock.WithGetUserOutput(&iam.GetUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString(resourceARN),
			"userId":   core.MappingNodeFromString(userId),
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/updated/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update user failure",
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
			ResourceID: "test-user-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-user-id",
					ResourceName: "TestUser",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-user-id",
						Name:       "TestUser",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
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
		SaveActionsCalled: map[string]any{
			"UpdateUser": &iam.UpdateUserInput{
				UserName: aws.String("test-user"),
				NewPath:  aws.String("/updated/"),
			},
		},
		ExpectError: true,
	}
}

func recreateUserOnUserNameChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/OldUser"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteUserOutput(&iam.DeleteUserOutput{}),
		iammock.WithCreateUserOutput(&iam.CreateUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("NewUser"),
				Path:     aws.String("/"),
			},
		}),
		iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
			Groups: []types.Group{},
		}),
		iammock.WithRemoveUserFromGroupOutput(&iam.RemoveUserFromGroupOutput{}),
		iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{},
		}),
		iammock.WithDetachUserPolicyOutput(&iam.DetachUserPolicyOutput{}),
		iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
			PolicyNames: []string{},
		}),
		iammock.WithDeleteUserPolicyOutput(&iam.DeleteUserPolicyOutput{}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString(resourceARN),
			"userId":   core.MappingNodeFromString(userId),
			"userName": core.MappingNodeFromString("OldUser"),
			"path":     core.MappingNodeFromString("/"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("NewUser"),
			"path":     core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "recreate user on userName change",
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
			ResourceID: "test-user-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-user-id",
					ResourceName: "TestUser",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-user-id",
						Name:       "TestUser",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
						PrevValue: core.MappingNodeFromString("OldUser"),
						NewValue:  core.MappingNodeFromString("NewUser"),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString(resourceARN),
				"spec.userId": core.MappingNodeFromString(userId),
			},
		},
		SaveActionsCalled: map[string]any{
			"DeleteUser": &iam.DeleteUserInput{
				UserName: aws.String("OldUser"),
			},
			"CreateUser": &iam.CreateUserInput{
				UserName: aws.String("NewUser"),
				Path:     aws.String("/"),
			},
		},
	}
}

func TestIAMUserResourceUpdate(t *testing.T) {
	suite.Run(t, new(IAMUserResourceUpdateSuite))
}
