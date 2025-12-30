//go:build unit

package iam

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type IAMUserResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *IAMUserResourceGetExternalStateSuite) Test_get_external_state() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"sessionID": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		createBasicUserStateTestCase(providerCtx, loader),
		createUserWithAllFeaturesTestCase(providerCtx, loader),
		createUserWithTagsStateTestCase(providerCtx, loader),
		createGetUserErrorTestCase(providerCtx, loader),
		createGetUserGroupsErrorTestCase(providerCtx, loader),
		createUserWithLoginProfileStateTestCase(providerCtx, loader),
		createUserWithPermissionsBoundaryStateTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	userResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return UserResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		userResourceWrapper,
		&s.Suite,
	)
}

func createBasicUserStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	// Create test data for user get external state
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/test-user"),
			"userName": core.MappingNodeFromString("test-user"),
		},
	}

	// Expected output with computed fields
	expectedResourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/test-user"),
			"userId":   core.MappingNodeFromString("AIDA1234567890123456"),
			"userName": core.MappingNodeFromString("test-user"),
			"path":     core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "successfully gets basic user state",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetUserOutput(&iam.GetUserOutput{
				User: &types.User{
					Arn:      aws.String("arn:aws:iam::123456789012:user/test-user"),
					UserId:   aws.String("AIDA1234567890123456"),
					UserName: aws.String("test-user"),
					Path:     aws.String("/"),
				},
			}),
			iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
				Groups: []types.Group{},
			}),
			iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{},
			}),
			iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
				PolicyNames: []string{},
			}),
			iammock.WithListUserTagsOutput(&iam.ListUserTagsOutput{
				Tags: []types.Tag{},
			}),
			iammock.WithGetLoginProfileError(&smithy.GenericAPIError{
				Code: "NoSuchEntity",
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-user",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expectedResourceSpecState,
		},
	}
}

func createUserWithAllFeaturesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	// Create sorted tag items for expected output
	tagItems := []*core.MappingNode{
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Department"),
				"value": core.MappingNodeFromString("engineering"),
			},
		},
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Environment"),
				"value": core.MappingNodeFromString("test"),
			},
		},
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Project"),
				"value": core.MappingNodeFromString("bluelink"),
			},
		},
	}

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
		return iammock.CreateIamServiceMock(
			iammock.WithGetUserOutput(&iam.GetUserOutput{
				User: &types.User{
					Arn:      aws.String("arn:aws:iam::123456789012:user/complex-user"),
					UserId:   aws.String("AIDA1234567890123456"),
					UserName: aws.String("complex-user"),
					Path:     aws.String("/engineering/"),
					Tags: []types.Tag{
						{Key: aws.String("Environment"), Value: aws.String("test")},
						{Key: aws.String("Department"), Value: aws.String("engineering")},
						{Key: aws.String("Project"), Value: aws.String("bluelink")},
					},
					PermissionsBoundary: &types.AttachedPermissionsBoundary{
						PermissionsBoundaryArn:  aws.String("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
						PermissionsBoundaryType: types.PermissionsBoundaryAttachmentTypePolicy,
					},
				},
			}),
			iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
						PolicyName: aws.String("PowerUserAccess"),
					},
					{
						PolicyArn:  aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
						PolicyName: aws.String("ReadOnlyAccess"),
					},
				},
			}),
			iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
				PolicyNames: []string{"S3Access", "DynamoDBAccess"},
			}),
			iammock.WithGetUserPolicyOutput(&iam.GetUserPolicyOutput{
				UserName:       aws.String("complex-user"),
				PolicyName:     aws.String("S3Access"),
				PolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject","s3:PutObject"],"Resource":["arn:aws:s3:::my-bucket/*"]}]}`),
			}),
			iammock.WithGetUserPolicyOutput(&iam.GetUserPolicyOutput{
				UserName:       aws.String("complex-user"),
				PolicyName:     aws.String("DynamoDBAccess"),
				PolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["dynamodb:PutItem","dynamodb:GetItem"],"Resource":["arn:aws:dynamodb:::table/my-table"]}]}`),
			}),
			iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
				Groups: []types.Group{
					{GroupName: aws.String("admins")},
					{GroupName: aws.String("developers")},
				},
			}),
			iammock.WithGetLoginProfileOutput(&iam.GetLoginProfileOutput{
				LoginProfile: &types.LoginProfile{
					UserName:              aws.String("complex-user"),
					PasswordResetRequired: false,
				},
			}),
			iammock.WithListUserTagsOutput(&iam.ListUserTagsOutput{
				Tags: []types.Tag{
					{Key: aws.String("Environment"), Value: aws.String("test")},
					{Key: aws.String("Department"), Value: aws.String("engineering")},
					{Key: aws.String("Project"), Value: aws.String("bluelink")},
				},
			}),
		)
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name:           "successfully gets user state with all features",
		ServiceFactory: serviceFactory,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/complex-user"),
					"userName": core.MappingNodeFromString("complex-user"),
				},
			},
		},
		CheckTags: true,
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/complex-user"),
					"userId":   core.MappingNodeFromString("AIDA1234567890123456"),
					"userName": core.MappingNodeFromString("complex-user"),
					"path":     core.MappingNodeFromString("/engineering/"),
					"tags": {
						Items: tagItems,
					},
					"managedPolicyArns": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:iam::aws:policy/PowerUserAccess"),
							core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
						},
					},
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
															"Effect":   core.MappingNodeFromString("Allow"),
															"Action":   {Items: []*core.MappingNode{core.MappingNodeFromString("s3:GetObject"), core.MappingNodeFromString("s3:PutObject")}},
															"Resource": {Items: []*core.MappingNode{core.MappingNodeFromString("arn:aws:s3:::my-bucket/*")}},
														},
														Items:                   nil,
														FieldsSourceMeta:        nil,
														StringWithSubstitutions: nil,
														SourceMeta:              nil,
													},
												},
											},
										},
									},
								},
							},
							{
								Fields: map[string]*core.MappingNode{
									"policyName": core.MappingNodeFromString("DynamoDBAccess"),
									"policyDocument": {
										Fields: map[string]*core.MappingNode{
											"Version": core.MappingNodeFromString("2012-10-17"),
											"Statement": {
												Items: []*core.MappingNode{
													{
														Fields: map[string]*core.MappingNode{
															"Effect":   core.MappingNodeFromString("Allow"),
															"Action":   {Items: []*core.MappingNode{core.MappingNodeFromString("dynamodb:PutItem"), core.MappingNodeFromString("dynamodb:GetItem")}},
															"Resource": {Items: []*core.MappingNode{core.MappingNodeFromString("arn:aws:dynamodb:::table/my-table")}},
														},
														Items:                   nil,
														FieldsSourceMeta:        nil,
														StringWithSubstitutions: nil,
														SourceMeta:              nil,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					"groups": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("admins"),
							core.MappingNodeFromString("developers"),
						},
					},
					"loginProfile": {
						Fields: map[string]*core.MappingNode{
							"password":              core.MappingNodeFromString("<hidden>"),
							"passwordResetRequired": core.MappingNodeFromBool(false),
						},
					},
					"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
				},
			},
		},
		ExpectError: false,
	}
}

func createUserWithTagsStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	// Create sorted tag items for expected output
	tagItems := []*core.MappingNode{
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Environment"),
				"value": core.MappingNodeFromString("production"),
			},
		},
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Team"),
				"value": core.MappingNodeFromString("backend"),
			},
		},
	}

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
		return iammock.CreateIamServiceMock(
			iammock.WithGetUserOutput(&iam.GetUserOutput{
				User: &types.User{
					Arn:      aws.String("arn:aws:iam::123456789012:user/tagged-user"),
					UserId:   aws.String("AIDA1234567890123456"),
					UserName: aws.String("tagged-user"),
					Path:     aws.String("/"),
					Tags: []types.Tag{
						{Key: aws.String("Environment"), Value: aws.String("production")},
						{Key: aws.String("Team"), Value: aws.String("backend")},
					},
				},
			}),
			iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{},
			}),
			iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
				PolicyNames: []string{},
			}),
			iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
				Groups: []types.Group{},
			}),
			iammock.WithListUserTagsOutput(&iam.ListUserTagsOutput{
				Tags: []types.Tag{
					{Key: aws.String("Environment"), Value: aws.String("production")},
					{Key: aws.String("Team"), Value: aws.String("backend")},
				},
			}),
			iammock.WithGetLoginProfileError(&smithy.GenericAPIError{
				Code: "NoSuchEntity",
			}),
		)
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name:           "successfully gets user state with tags",
		ServiceFactory: serviceFactory,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/tagged-user"),
					"userName": core.MappingNodeFromString("tagged-user"),
				},
			},
		},
		CheckTags: true,
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/tagged-user"),
					"userId":   core.MappingNodeFromString("AIDA1234567890123456"),
					"userName": core.MappingNodeFromString("tagged-user"),
					"path":     core.MappingNodeFromString("/"),
					"tags": {
						Items: tagItems,
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createGetUserErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
		return iammock.CreateIamServiceMock(
			iammock.WithGetUserError(errors.New("failed to get user")),
		)
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name:           "handles get user error",
		ServiceFactory: serviceFactory,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/test-user"),
					"userName": core.MappingNodeFromString("test-user"),
				},
			},
		},
		ExpectError: true,
	}
}

func createGetUserGroupsErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
		return iammock.CreateIamServiceMock(
			iammock.WithGetUserOutput(&iam.GetUserOutput{
				User: &types.User{
					Arn:      aws.String("arn:aws:iam::123456789012:user/test-user"),
					UserId:   aws.String("AIDA1234567890123456"),
					UserName: aws.String("test-user"),
					Path:     aws.String("/"),
				},
			}),
			iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{},
			}),
			iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
				PolicyNames: []string{},
			}),
			iammock.WithListGroupsForUserError(errors.New("failed to get user groups")),
			iammock.WithListUserTagsOutput(&iam.ListUserTagsOutput{
				Tags: []types.Tag{},
			}),
			iammock.WithGetLoginProfileError(&smithy.GenericAPIError{
				Code: "NoSuchEntity",
			}),
		)
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name:           "handles get user groups error",
		ServiceFactory: serviceFactory,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/test-user"),
					"userName": core.MappingNodeFromString("test-user"),
				},
			},
		},
		ExpectError: true,
	}
}

func createUserWithLoginProfileStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
		return iammock.CreateIamServiceMock(
			iammock.WithGetUserOutput(&iam.GetUserOutput{
				User: &types.User{
					Arn:      aws.String("arn:aws:iam::123456789012:user/console-user"),
					UserId:   aws.String("AIDA1234567890123456"),
					UserName: aws.String("console-user"),
					Path:     aws.String("/"),
				},
			}),
			iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{},
			}),
			iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
				PolicyNames: []string{},
			}),
			iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
				Groups: []types.Group{},
			}),
			iammock.WithGetLoginProfileOutput(&iam.GetLoginProfileOutput{
				LoginProfile: &types.LoginProfile{
					UserName:              aws.String("console-user"),
					PasswordResetRequired: true,
				},
			}),
			iammock.WithListUserTagsOutput(&iam.ListUserTagsOutput{
				Tags: []types.Tag{},
			}),
		)
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name:           "successfully gets user state with login profile",
		ServiceFactory: serviceFactory,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/console-user"),
					"userName": core.MappingNodeFromString("console-user"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/console-user"),
					"userId":   core.MappingNodeFromString("AIDA1234567890123456"),
					"userName": core.MappingNodeFromString("console-user"),
					"path":     core.MappingNodeFromString("/"),
					"loginProfile": {
						Fields: map[string]*core.MappingNode{
							"password":              core.MappingNodeFromString("<hidden>"),
							"passwordResetRequired": core.MappingNodeFromBool(false),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createUserWithPermissionsBoundaryStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
		return iammock.CreateIamServiceMock(
			iammock.WithGetUserOutput(&iam.GetUserOutput{
				User: &types.User{
					Arn:      aws.String("arn:aws:iam::123456789012:user/bounded-user"),
					UserId:   aws.String("AIDA1234567890123456"),
					UserName: aws.String("bounded-user"),
					Path:     aws.String("/"),
					PermissionsBoundary: &types.AttachedPermissionsBoundary{
						PermissionsBoundaryArn:  aws.String("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
						PermissionsBoundaryType: types.PermissionsBoundaryAttachmentTypePolicy,
					},
				},
			}),
			iammock.WithListAttachedUserPoliciesOutput(&iam.ListAttachedUserPoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{},
			}),
			iammock.WithListUserPoliciesOutput(&iam.ListUserPoliciesOutput{
				PolicyNames: []string{},
			}),
			iammock.WithListGroupsForUserOutput(&iam.ListGroupsForUserOutput{
				Groups: []types.Group{},
			}),
			iammock.WithListUserTagsOutput(&iam.ListUserTagsOutput{
				Tags: []types.Tag{},
			}),
			iammock.WithGetLoginProfileError(&smithy.GenericAPIError{
				Code: "NoSuchEntity",
			}),
		)
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name:           "successfully gets user state with permissions boundary",
		ServiceFactory: serviceFactory,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/bounded-user"),
					"userName": core.MappingNodeFromString("bounded-user"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":                 core.MappingNodeFromString("arn:aws:iam::123456789012:user/bounded-user"),
					"userId":              core.MappingNodeFromString("AIDA1234567890123456"),
					"userName":            core.MappingNodeFromString("bounded-user"),
					"path":                core.MappingNodeFromString("/"),
					"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
				},
			},
		},
		ExpectError: false,
	}
}

func TestIAMUserResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(IAMUserResourceGetExternalStateSuite))
}
