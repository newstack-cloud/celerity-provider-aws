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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMUserResourceCreateSuite struct {
	suite.Suite
}

func (s *IAMUserResourceCreateSuite) Test_create_iam_user() {
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
		createBasicUserCreateTestCase(providerCtx, loader),
		createUserWithTagsTestCase(providerCtx, loader),
		createUserWithLoginProfileTestCase(providerCtx, loader),
		createUserWithManagedPoliciesTestCase(providerCtx, loader),
		createUserWithInlinePoliciesTestCase(providerCtx, loader),
		createUserWithGroupsTestCase(providerCtx, loader),
		createUserWithPermissionsBoundaryTestCase(providerCtx, loader),
		createUserWithGeneratedNameTestCase(providerCtx, loader),
		createUserFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		UserResource,
		&s.Suite,
	)
}

func createBasicUserCreateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user"
	userId := "AIDA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateUserOutput(&iam.CreateUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		}),
	)

	// Create test data for user creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create basic user",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
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
			"CreateUser": &iam.CreateUserInput{
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		},
	}
}

func createUserWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user-with-tags"
	userId := "AIDA1234567890123457"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateUserOutput(&iam.CreateUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user-with-tags"),
				Path:     aws.String("/"),
			},
		}),
	)

	// Create test data for user creation with tags
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user-with-tags"),
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
							"key":   core.MappingNodeFromString("Department"),
							"value": core.MappingNodeFromString("engineering"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create user with tags",
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
					ResourceName: "TestUserWithTags",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
					{
						FieldPath: "spec.path",
					},
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
			"CreateUser": &iam.CreateUserInput{
				UserName: aws.String("test-user-with-tags"),
				Path:     aws.String("/"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Department"),
						Value: aws.String("engineering"),
					},
					{
						Key:   aws.String("Environment"),
						Value: aws.String("test"),
					},
				},
			},
		},
	}
}

func createUserWithLoginProfileTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user-with-login"
	userId := "AIDA1234567890123458"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateUserOutput(&iam.CreateUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user-with-login"),
				Path:     aws.String("/"),
			},
		}),
		iammock.WithCreateLoginProfileOutput(&iam.CreateLoginProfileOutput{
			LoginProfile: &types.LoginProfile{
				UserName:              aws.String("test-user-with-login"),
				PasswordResetRequired: true,
			},
		}),
	)

	// Create test data for user creation with login profile
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user-with-login"),
			"path":     core.MappingNodeFromString("/"),
			"loginProfile": {
				Fields: map[string]*core.MappingNode{
					"password":              core.MappingNodeFromString("TempPassword123!"),
					"passwordResetRequired": core.MappingNodeFromBool(true),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create user with login profile",
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
					ResourceName: "TestUserWithLogin",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
					{
						FieldPath: "spec.path",
					},
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
			"CreateUser": &iam.CreateUserInput{
				UserName: aws.String("test-user-with-login"),
				Path:     aws.String("/"),
			},
			"CreateLoginProfile": &iam.CreateLoginProfileInput{
				UserName:              aws.String("test-user-with-login"),
				Password:              aws.String("TempPassword123!"),
				PasswordResetRequired: true,
			},
		},
	}
}

func createUserWithManagedPoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user-with-managed-policies"
	userId := "AIDA1234567890123459"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateUserOutput(&iam.CreateUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user-with-managed-policies"),
				Path:     aws.String("/"),
			},
		}),
		iammock.WithAttachUserPolicyOutput(&iam.AttachUserPolicyOutput{}),
	)

	// Create test data for user creation with managed policies
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user-with-managed-policies"),
			"path":     core.MappingNodeFromString("/"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					core.MappingNodeFromString("arn:aws:iam::aws:policy/PowerUserAccess"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create user with managed policies",
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
					ResourceName: "TestUserWithManagedPolicies",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
					{
						FieldPath: "spec.path",
					},
					{
						FieldPath: "spec.managedPolicyArns",
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
			"CreateUser": &iam.CreateUserInput{
				UserName: aws.String("test-user-with-managed-policies"),
				Path:     aws.String("/"),
			},
			"AttachUserPolicy": []any{
				&iam.AttachUserPolicyInput{
					UserName:  aws.String("test-user-with-managed-policies"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
				},
				&iam.AttachUserPolicyInput{
					UserName:  aws.String("test-user-with-managed-policies"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
				},
			},
		},
	}
}

func createUserWithInlinePoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user-with-inline-policies"
	userId := "AIDA1234567890123460"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateUserOutput(&iam.CreateUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user-with-inline-policies"),
				Path:     aws.String("/"),
			},
		}),
		iammock.WithPutUserPolicyOutput(&iam.PutUserPolicyOutput{}),
	)

	// Create test data for user creation with inline policies
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user-with-inline-policies"),
			"path":     core.MappingNodeFromString("/"),
			"policies": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("S3Access"),
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
															core.MappingNodeFromString("arn:aws:s3:::my-bucket/*"),
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
		Name: "create user with inline policies",
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
					ResourceName: "TestUserWithInlinePolicies",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
					{
						FieldPath: "spec.path",
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
			"CreateUser": &iam.CreateUserInput{
				UserName: aws.String("test-user-with-inline-policies"),
				Path:     aws.String("/"),
			},
			"PutUserPolicy": &iam.PutUserPolicyInput{
				UserName:       aws.String("test-user-with-inline-policies"),
				PolicyName:     aws.String("S3Access"),
				PolicyDocument: aws.String(`{"Statement":[{"Action":["s3:GetObject","s3:PutObject"],"Effect":"Allow","Resource":["arn:aws:s3:::my-bucket/*"]}],"Version":"2012-10-17"}`),
			},
		},
	}
}

func createUserWithGroupsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user-with-groups"
	userId := "AIDA1234567890123461"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateUserOutput(&iam.CreateUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user-with-groups"),
				Path:     aws.String("/"),
			},
		}),
		iammock.WithAddUserToGroupOutput(&iam.AddUserToGroupOutput{}),
	)

	// Create test data for user creation with groups
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user-with-groups"),
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
		Name: "create user with groups",
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
					ResourceName: "TestUserWithGroups",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
					{
						FieldPath: "spec.path",
					},
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
			"CreateUser": &iam.CreateUserInput{
				UserName: aws.String("test-user-with-groups"),
				Path:     aws.String("/"),
			},
			"AddUserToGroup": []any{
				&iam.AddUserToGroupInput{
					UserName:  aws.String("test-user-with-groups"),
					GroupName: aws.String("developers"),
				},
				&iam.AddUserToGroupInput{
					UserName:  aws.String("test-user-with-groups"),
					GroupName: aws.String("admins"),
				},
			},
		},
	}
}

func createUserWithPermissionsBoundaryTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:user/test-user-with-boundary"
	userId := "AIDA1234567890123462"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateUserOutput(&iam.CreateUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String("test-user-with-boundary"),
				Path:     aws.String("/"),
			},
		}),
		iammock.WithPutUserPermissionsBoundaryOutput(&iam.PutUserPermissionsBoundaryOutput{}),
	)

	// Create test data for user creation with permissions boundary
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName":            core.MappingNodeFromString("test-user-with-boundary"),
			"path":                core.MappingNodeFromString("/"),
			"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create user with permissions boundary",
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
					ResourceName: "TestUserWithBoundary",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
					{
						FieldPath: "spec.path",
					},
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
			"CreateUser": &iam.CreateUserInput{
				UserName: aws.String("test-user-with-boundary"),
				Path:     aws.String("/"),
			},
			"PutUserPermissionsBoundary": &iam.PutUserPermissionsBoundaryInput{
				UserName:            aws.String("test-user-with-boundary"),
				PermissionsBoundary: aws.String("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
			},
		},
	}
}

func createUserWithGeneratedNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	generatedName := "bluelink-generated-user-abcd1234"
	resourceARN := "arn:aws:iam::123456789012:user/" + generatedName
	userId := "AIDA1234567890123463"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateUserOutput(&iam.CreateUserOutput{
			User: &types.User{
				Arn:      aws.String(resourceARN),
				UserId:   aws.String(userId),
				UserName: aws.String(generatedName),
				Path:     aws.String("/"),
			},
		}),
	)

	// Create test data for user creation with generated name
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create user with generated name",
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
					ResourceName: "TestUserGenerated",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
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
		// Note: We can't predict the exact user name due to nanoid generation,
		// so we'll omit SaveActionsCalled for this test case
		// The important thing is that the user gets created successfully
	}
}

func createUserFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateUserError(fmt.Errorf("failed to create user")),
	)

	// Create test data for user creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create user failure",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/user",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
					{
						FieldPath: "spec.path",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"CreateUser": &iam.CreateUserInput{
				UserName: aws.String("test-user"),
				Path:     aws.String("/"),
			},
		},
	}
}

func TestIAMUserResourceCreate(t *testing.T) {
	suite.Run(t, new(IAMUserResourceCreateSuite))
}
