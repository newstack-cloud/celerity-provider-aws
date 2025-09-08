//go:build unit

package iam

import (
	"errors"
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
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMUserResourceDestroySuite struct {
	suite.Suite
}

func (s *IAMUserResourceDestroySuite) Test_destroy() {
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

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		createSuccessfulDestroyTestCase(providerCtx, loader),
		createDestroyUserWithPoliciesTestCase(providerCtx, loader),
		createDestroyUserWithGroupsTestCase(providerCtx, loader),
		createDestroyUserWithLoginProfileTestCase(providerCtx, loader),
		createDestroyUserWithPermissionsBoundaryTestCase(providerCtx, loader),
		createDestroyUserComplexTestCase(providerCtx, loader),
		createFailingDestroyTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		UserResource,
		&s.Suite,
	)
}

func createSuccessfulDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
			Groups: []types.Group{},
		}),
		iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{},
		}),
		iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
			PolicyNames: []string{},
		}),
		iammock.WithDeleteUserOutput(&iam.DeleteUserOutput{}),
	)

	userARN := "arn:aws:iam::123456789012:user/test-user"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "successfully deletes basic user",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn":      core.MappingNodeFromString(userARN),
						"userName": core.MappingNodeFromString("test-user"),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"ListGroupsForUser": &iam.ListGroupsForUserInput{
				UserName: aws.String("test-user"),
			},
			"ListAttachedUserPolicies": &iam.ListAttachedUserPoliciesInput{
				UserName: aws.String("test-user"),
			},
			"ListUserPolicies": &iam.ListUserPoliciesInput{
				UserName: aws.String("test-user"),
			},
			"DeleteUser": &iam.DeleteUserInput{
				UserName: aws.String("test-user"),
			},
		},
	}
}

func createDestroyUserWithPoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
			Groups: []types.Group{},
		}),
		iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{
				{
					PolicyArn:  aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					PolicyName: aws.String("ReadOnlyAccess"),
				},
			},
		}),
		iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
			PolicyNames: []string{"InlinePolicy1", "InlinePolicy2"},
		}),
		iammock.WithDetachUserPolicyOutput(&iam.DetachUserPolicyOutput{}),
		iammock.WithDeleteUserPolicyOutput(&iam.DeleteUserPolicyOutput{}),
		iammock.WithDeleteUserOutput(&iam.DeleteUserOutput{}),
	)

	userARN := "arn:aws:iam::123456789012:user/test-user-with-policies"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "successfully deletes user with policies",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn":      core.MappingNodeFromString(userARN),
						"userName": core.MappingNodeFromString("test-user-with-policies"),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"ListGroupsForUser": &iam.ListGroupsForUserInput{
				UserName: aws.String("test-user-with-policies"),
			},
			"ListAttachedUserPolicies": &iam.ListAttachedUserPoliciesInput{
				UserName: aws.String("test-user-with-policies"),
			},
			"ListUserPolicies": &iam.ListUserPoliciesInput{
				UserName: aws.String("test-user-with-policies"),
			},
			"DetachUserPolicy": &iam.DetachUserPolicyInput{
				UserName:  aws.String("test-user-with-policies"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
			},
			"DeleteUserPolicy": []any{
				&iam.DeleteUserPolicyInput{
					UserName:   aws.String("test-user-with-policies"),
					PolicyName: aws.String("InlinePolicy1"),
				},
				&iam.DeleteUserPolicyInput{
					UserName:   aws.String("test-user-with-policies"),
					PolicyName: aws.String("InlinePolicy2"),
				},
			},
			"DeleteUser": &iam.DeleteUserInput{
				UserName: aws.String("test-user-with-policies"),
			},
		},
	}
}

func createDestroyUserWithGroupsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{},
		}),
		iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
			PolicyNames: []string{},
		}),
		iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
			Groups: []types.Group{
				{
					GroupName: aws.String("developers"),
				},
				{
					GroupName: aws.String("admins"),
				},
			},
		}),
		iammock.WithRemoveUserFromGroupOutput(&iam.RemoveUserFromGroupOutput{}),
		iammock.WithDeleteUserOutput(&iam.DeleteUserOutput{}),
	)

	userARN := "arn:aws:iam::123456789012:user/test-user-with-groups"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "successfully deletes user with groups",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn":      core.MappingNodeFromString(userARN),
						"userName": core.MappingNodeFromString("test-user-with-groups"),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"ListAttachedUserPolicies": &iam.ListAttachedUserPoliciesInput{
				UserName: aws.String("test-user-with-groups"),
			},
			"ListUserPolicies": &iam.ListUserPoliciesInput{
				UserName: aws.String("test-user-with-groups"),
			},
			"ListGroupsForUser": &iam.ListGroupsForUserInput{
				UserName: aws.String("test-user-with-groups"),
			},
			"RemoveUserFromGroup": []any{
				&iam.RemoveUserFromGroupInput{
					UserName:  aws.String("test-user-with-groups"),
					GroupName: aws.String("developers"),
				},
				&iam.RemoveUserFromGroupInput{
					UserName:  aws.String("test-user-with-groups"),
					GroupName: aws.String("admins"),
				},
			},
			"DeleteUser": &iam.DeleteUserInput{
				UserName: aws.String("test-user-with-groups"),
			},
		},
	}
}

func createDestroyUserWithLoginProfileTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{},
		}),
		iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
			PolicyNames: []string{},
		}),
		iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
			Groups: []types.Group{},
		}),
		iammock.WithDeleteLoginProfileOutput(&iam.DeleteLoginProfileOutput{}),
		iammock.WithDeleteUserOutput(&iam.DeleteUserOutput{}),
	)

	userARN := "arn:aws:iam::123456789012:user/test-user-with-login"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "successfully deletes user with login profile",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn":      core.MappingNodeFromString(userARN),
						"userName": core.MappingNodeFromString("test-user-with-login"),
						"loginProfile": {
							Fields: map[string]*core.MappingNode{
								"password":              core.MappingNodeFromString("TempPassword123!"),
								"passwordResetRequired": core.MappingNodeFromBool(true),
							},
						},
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"ListAttachedUserPolicies": &iam.ListAttachedUserPoliciesInput{
				UserName: aws.String("test-user-with-login"),
			},
			"ListUserPolicies": &iam.ListUserPoliciesInput{
				UserName: aws.String("test-user-with-login"),
			},
			"ListGroupsForUser": &iam.ListGroupsForUserInput{
				UserName: aws.String("test-user-with-login"),
			},
			"DeleteLoginProfile": &iam.DeleteLoginProfileInput{
				UserName: aws.String("test-user-with-login"),
			},
			"DeleteUser": &iam.DeleteUserInput{
				UserName: aws.String("test-user-with-login"),
			},
		},
	}
}

func createDestroyUserWithPermissionsBoundaryTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{},
		}),
		iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
			PolicyNames: []string{},
		}),
		iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
			Groups: []types.Group{},
		}),
		iammock.WithDeleteUserPermissionsBoundaryOutput(&iam.DeleteUserPermissionsBoundaryOutput{}),
		iammock.WithDeleteUserOutput(&iam.DeleteUserOutput{}),
	)

	userARN := "arn:aws:iam::123456789012:user/test-user-with-boundary"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "successfully deletes user with permissions boundary",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn":                 core.MappingNodeFromString(userARN),
						"userName":            core.MappingNodeFromString("test-user-with-boundary"),
						"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"ListAttachedUserPolicies": &iam.ListAttachedUserPoliciesInput{
				UserName: aws.String("test-user-with-boundary"),
			},
			"ListUserPolicies": &iam.ListUserPoliciesInput{
				UserName: aws.String("test-user-with-boundary"),
			},
			"ListGroupsForUser": &iam.ListGroupsForUserInput{
				UserName: aws.String("test-user-with-boundary"),
			},
			"DeleteUserPermissionsBoundary": &iam.DeleteUserPermissionsBoundaryInput{
				UserName: aws.String("test-user-with-boundary"),
			},
			"DeleteUser": &iam.DeleteUserInput{
				UserName: aws.String("test-user-with-boundary"),
			},
		},
	}
}

func createDestroyUserComplexTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{
				{
					PolicyArn:  aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					PolicyName: aws.String("ReadOnlyAccess"),
				},
				{
					PolicyArn:  aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
					PolicyName: aws.String("PowerUserAccess"),
				},
			},
		}),
		iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
			PolicyNames: []string{"S3Access", "DynamoDBAccess"},
		}),
		iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
			Groups: []types.Group{
				{
					GroupName: aws.String("developers"),
				},
				{
					GroupName: aws.String("admins"),
				},
			},
		}),
		iammock.WithDetachUserPolicyOutput(&iam.DetachUserPolicyOutput{}),
		iammock.WithDeleteUserPolicyOutput(&iam.DeleteUserPolicyOutput{}),
		iammock.WithRemoveUserFromGroupOutput(&iam.RemoveUserFromGroupOutput{}),
		iammock.WithDeleteLoginProfileOutput(&iam.DeleteLoginProfileOutput{}),
		iammock.WithDeleteUserPermissionsBoundaryOutput(&iam.DeleteUserPermissionsBoundaryOutput{}),
		iammock.WithDeleteUserOutput(&iam.DeleteUserOutput{}),
	)

	userARN := "arn:aws:iam::123456789012:user/complex-user"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "successfully deletes complex user with all features",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn":      core.MappingNodeFromString(userARN),
						"userName": core.MappingNodeFromString("complex-user"),
						"loginProfile": {
							Fields: map[string]*core.MappingNode{
								"password":              core.MappingNodeFromString("TempPassword123!"),
								"passwordResetRequired": core.MappingNodeFromBool(true),
							},
						},
						"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"ListAttachedUserPolicies": &iam.ListAttachedUserPoliciesInput{
				UserName: aws.String("complex-user"),
			},
			"ListUserPolicies": &iam.ListUserPoliciesInput{
				UserName: aws.String("complex-user"),
			},
			"ListGroupsForUser": &iam.ListGroupsForUserInput{
				UserName: aws.String("complex-user"),
			},
			"DetachUserPolicy": []any{
				&iam.DetachUserPolicyInput{
					UserName:  aws.String("complex-user"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
				},
				&iam.DetachUserPolicyInput{
					UserName:  aws.String("complex-user"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
				},
			},
			"DeleteUserPolicy": []any{
				&iam.DeleteUserPolicyInput{
					UserName:   aws.String("complex-user"),
					PolicyName: aws.String("S3Access"),
				},
				&iam.DeleteUserPolicyInput{
					UserName:   aws.String("complex-user"),
					PolicyName: aws.String("DynamoDBAccess"),
				},
			},
			"RemoveUserFromGroup": []any{
				&iam.RemoveUserFromGroupInput{
					UserName:  aws.String("complex-user"),
					GroupName: aws.String("developers"),
				},
				&iam.RemoveUserFromGroupInput{
					UserName:  aws.String("complex-user"),
					GroupName: aws.String("admins"),
				},
			},
			"DeleteLoginProfile": &iam.DeleteLoginProfileInput{
				UserName: aws.String("complex-user"),
			},
			"DeleteUserPermissionsBoundary": &iam.DeleteUserPermissionsBoundaryInput{
				UserName: aws.String("complex-user"),
			},
			"DeleteUser": &iam.DeleteUserInput{
				UserName: aws.String("complex-user"),
			},
		},
	}
}

func createFailingDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{},
		}),
		iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
			PolicyNames: []string{},
		}),
		iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
			Groups: []types.Group{},
		}),
		iammock.WithDeleteUserError(errors.New("failed to delete user")),
	)

	userARN := "arn:aws:iam::123456789012:user/failing-user"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "fails to delete user",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn":      core.MappingNodeFromString(userARN),
						"userName": core.MappingNodeFromString("failing-user"),
					},
				},
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"ListAttachedUserPolicies": &iam.ListAttachedUserPoliciesInput{
				UserName: aws.String("failing-user"),
			},
			"ListUserPolicies": &iam.ListUserPoliciesInput{
				UserName: aws.String("failing-user"),
			},
			"ListGroupsForUser": &iam.ListGroupsForUserInput{
				UserName: aws.String("failing-user"),
			},
			"DeleteUser": &iam.DeleteUserInput{
				UserName: aws.String("failing-user"),
			},
		},
	}
}

func TestIAMUserResourceDestroy(t *testing.T) {
	suite.Run(t, new(IAMUserResourceDestroySuite))
}
