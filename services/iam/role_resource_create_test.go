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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type IAMRoleResourceCreateSuite struct {
	suite.Suite
}

func (s *IAMRoleResourceCreateSuite) Test_create_iam_role() {
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
		createBasicRoleCreateTestCase(providerCtx, loader),
		createRoleWithTagsTestCase(providerCtx, loader),
		createRoleFailureTestCase(providerCtx, loader),
		createRoleWithInlinePoliciesTestCase(providerCtx, loader),
		createRoleWithManagedPoliciesTestCase(providerCtx, loader),
		createRoleWithGeneratedNameTestCase(providerCtx, loader),
		createRoleWithPermissionsBoundaryTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	roleResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return RoleResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		roleResourceWrapper,
		&s.Suite,
	)
}

func createBasicRoleCreateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:role/test-role"
	roleId := "AROA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateRoleOutput(&iam.CreateRoleOutput{
			Role: &types.Role{
				Arn:                      aws.String(resourceARN),
				RoleId:                   aws.String(roleId),
				RoleName:                 aws.String("test-role"),
				AssumeRolePolicyDocument: aws.String(`{"Statement":[{"Action":["sts:AssumeRole"],"Effect":"Allow","Principal":{"Service":["lambda.amazonaws.com"]}}],"Version":"2012-10-17"}`),
			},
		}),
	)

	// Create test data for role creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("lambda.amazonaws.com"),
												},
											},
										},
									},
									"Action": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("sts:AssumeRole"),
										},
									},
								},
							},
						},
					},
				},
			},
			"description": core.MappingNodeFromString("Test role for Lambda execution"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create basic role",
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
			ResourceID: "test-role-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-id",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.roleName",
					},
					{
						FieldPath: "spec.assumeRolePolicyDocument",
					},
					{
						FieldPath: "spec.description",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString(resourceARN),
				"spec.roleId": core.MappingNodeFromString(roleId),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateRole": &iam.CreateRoleInput{
				RoleName:                 aws.String("test-role"),
				AssumeRolePolicyDocument: aws.String(`{"Statement":[{"Action":["sts:AssumeRole"],"Effect":"Allow","Principal":{"Service":["lambda.amazonaws.com"]}}],"Version":"2012-10-17"}`),
				Description:              aws.String("Test role for Lambda execution"),
			},
		},
	}
}

func createRoleWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:role/test-role-with-tags"
	roleId := "AROA1234567890123457"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateRoleOutput(&iam.CreateRoleOutput{
			Role: &types.Role{
				Arn:                      aws.String(resourceARN),
				RoleId:                   aws.String(roleId),
				RoleName:                 aws.String("test-role-with-tags"),
				AssumeRolePolicyDocument: aws.String(`{"Statement":[{"Action":["sts:AssumeRole"],"Effect":"Allow","Principal":{"Service":["ec2.amazonaws.com"]}}],"Version":"2012-10-17"}`),
			},
		}),
	)

	// Create test data for role creation with tags
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role-with-tags"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("ec2.amazonaws.com"),
												},
											},
										},
									},
									"Action": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("sts:AssumeRole"),
										},
									},
								},
							},
						},
					},
				},
			},
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
							"key":   core.MappingNodeFromString("Project"),
							"value": core.MappingNodeFromString("test-project"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create role with tags",
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
			ResourceID: "test-role-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-id",
					ResourceName: "TestRoleWithTags",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.roleName",
					},
					{
						FieldPath: "spec.assumeRolePolicyDocument",
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
				"spec.roleId": core.MappingNodeFromString(roleId),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateRole": &iam.CreateRoleInput{
				RoleName:                 aws.String("test-role-with-tags"),
				AssumeRolePolicyDocument: aws.String(`{"Statement":[{"Action":["sts:AssumeRole"],"Effect":"Allow","Principal":{"Service":["ec2.amazonaws.com"]}}],"Version":"2012-10-17"}`),
				Tags: []types.Tag{
					{
						Key:   aws.String("Environment"),
						Value: aws.String("test"),
					},
					{
						Key:   aws.String("Project"),
						Value: aws.String("test-project"),
					},
				},
			},
		},
	}
}

func createRoleFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateRoleError(fmt.Errorf("failed to create role")),
	)

	// Create test data for role creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("lambda.amazonaws.com"),
												},
											},
										},
									},
									"Action": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("sts:AssumeRole"),
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
		Name: "create role failure",
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
			ResourceID: "test-role-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-id",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.roleName",
					},
					{
						FieldPath: "spec.assumeRolePolicyDocument",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"CreateRole": &iam.CreateRoleInput{
				RoleName:                 aws.String("test-role"),
				AssumeRolePolicyDocument: aws.String(`{"Statement":[{"Action":["sts:AssumeRole"],"Effect":"Allow","Principal":{"Service":["lambda.amazonaws.com"]}}],"Version":"2012-10-17"}`),
			},
		},
	}
}

func createRoleWithInlinePoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:role/test-role-with-policies"
	roleId := "AROA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateRoleOutput(&iam.CreateRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("test-role-with-policies"),
				Arn:      aws.String(resourceARN),
				RoleId:   aws.String(roleId),
			},
		}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role-with-policies"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("lambda.amazonaws.com"),
												},
											},
										},
									},
									"Action": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("sts:AssumeRole"),
										},
									},
								},
							},
						},
					},
				},
			},
			"policies": {
				Items: []*core.MappingNode{
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
													"Effect": core.MappingNodeFromString("Allow"),
													"Action": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("dynamodb:GetItem"),
															core.MappingNodeFromString("dynamodb:PutItem"),
														},
													},
													"Resource": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("arn:aws:dynamodb:*:*:table/MyTable"),
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
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("SQSAccess"),
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
															core.MappingNodeFromString("sqs:SendMessage"),
															core.MappingNodeFromString("sqs:ReceiveMessage"),
														},
													},
													"Resource": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("arn:aws:sqs:*:*:my-queue"),
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
		Name: "create role with inline policies",
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
			ResourceID: "test-role-with-policies-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-with-policies-id",
					ResourceName: "test-role-with-policies",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.roleName",
					},
					{
						FieldPath: "spec.assumeRolePolicyDocument",
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
				"spec.roleId": core.MappingNodeFromString(roleId),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateRole": &iam.CreateRoleInput{
				RoleName:                 aws.String("test-role-with-policies"),
				AssumeRolePolicyDocument: aws.String(`{"Statement":[{"Action":["sts:AssumeRole"],"Effect":"Allow","Principal":{"Service":["lambda.amazonaws.com"]}}],"Version":"2012-10-17"}`),
			},
			"PutRolePolicy": []any{
				&iam.PutRolePolicyInput{
					RoleName:       aws.String("test-role-with-policies"),
					PolicyName:     aws.String("DynamoDBAccess"),
					PolicyDocument: aws.String(`{"Statement":[{"Action":["dynamodb:GetItem","dynamodb:PutItem"],"Effect":"Allow","Resource":["arn:aws:dynamodb:*:*:table/MyTable"]}],"Version":"2012-10-17"}`),
				},
				&iam.PutRolePolicyInput{
					RoleName:       aws.String("test-role-with-policies"),
					PolicyName:     aws.String("SQSAccess"),
					PolicyDocument: aws.String(`{"Statement":[{"Action":["sqs:SendMessage","sqs:ReceiveMessage"],"Effect":"Allow","Resource":["arn:aws:sqs:*:*:my-queue"]}],"Version":"2012-10-17"}`),
				},
			},
		},
	}
}

func createRoleWithManagedPoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:role/test-role-with-managed-policies"
	roleId := "AROA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateRoleOutput(&iam.CreateRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("test-role-with-managed-policies"),
				Arn:      aws.String(resourceARN),
				RoleId:   aws.String(roleId),
			},
		}),
		iammock.WithAttachRolePolicyOutput(&iam.AttachRolePolicyOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role-with-managed-policies"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("lambda.amazonaws.com"),
												},
											},
										},
									},
									"Action": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("sts:AssumeRole"),
										},
									},
								},
							},
						},
					},
				},
			},
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
					core.MappingNodeFromString("arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create role with managed policies",
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
			ResourceID: "test-role-with-managed-policies-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-with-managed-policies-id",
					ResourceName: "test-role-with-managed-policies",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.roleName",
					},
					{
						FieldPath: "spec.assumeRolePolicyDocument",
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
				"spec.roleId": core.MappingNodeFromString(roleId),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateRole": &iam.CreateRoleInput{
				RoleName:                 aws.String("test-role-with-managed-policies"),
				AssumeRolePolicyDocument: aws.String(`{"Statement":[{"Action":["sts:AssumeRole"],"Effect":"Allow","Principal":{"Service":["lambda.amazonaws.com"]}}],"Version":"2012-10-17"}`),
			},
			"AttachRolePolicy": []any{
				&iam.AttachRolePolicyInput{
					RoleName:  aws.String("test-role-with-managed-policies"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				},
				&iam.AttachRolePolicyInput{
					RoleName:  aws.String("test-role-with-managed-policies"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"),
				},
			},
		},
	}
}

func createRoleWithGeneratedNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	// We'll use a mock that captures the actual role name that gets generated
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateRoleOutput(&iam.CreateRoleOutput{
			Role: &types.Role{
				Arn:                      aws.String("arn:aws:iam::123456789012:role/test-instance-id-TestRoleAutoGenerated-test123"),
				RoleId:                   aws.String("AROA1234567890123456"),
				RoleName:                 aws.String("test-instance-id-TestRoleAutoGenerated-test123"),
				AssumeRolePolicyDocument: aws.String(`{"Statement":[{"Action":["sts:AssumeRole"],"Effect":"Allow","Principal":{"Service":["lambda.amazonaws.com"]}}],"Version":"2012-10-17"}`),
			},
		}),
	)

	// Create test data for role creation WITHOUT roleName (to test auto-generation)
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			// Note: roleName is intentionally omitted to test auto-generation
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("lambda.amazonaws.com"),
												},
											},
										},
									},
									"Action": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("sts:AssumeRole"),
										},
									},
								},
							},
						},
					},
				},
			},
			"description": core.MappingNodeFromString("Test role with auto-generated name"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create role with auto-generated name",
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
			ResourceID: "test-role-auto-generated-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-auto-generated-id",
					ResourceName: "TestRoleAutoGenerated",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.assumeRolePolicyDocument",
					},
					{
						FieldPath: "spec.description",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-instance-id-TestRoleAutoGenerated-test123"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
		// Note: We can't predict the exact role name due to nanoid generation,
		// so we'll omit SaveActionsCalled for this test case
		// The important thing is that the role gets created successfully
	}
}

func createRoleWithPermissionsBoundaryTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:role/test-role-with-permissions-boundary"
	roleId := "AROA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateRoleOutput(&iam.CreateRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("test-role-with-permissions-boundary"),
				Arn:      aws.String(resourceARN),
				RoleId:   aws.String(roleId),
			},
		}),
		iammock.WithPutRolePermissionsBoundaryOutput(&iam.PutRolePermissionsBoundaryOutput{}),
		iammock.WithAttachRolePolicyOutput(&iam.AttachRolePolicyOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role-with-permissions-boundary"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("lambda.amazonaws.com"),
												},
											},
										},
									},
									"Action": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("sts:AssumeRole"),
										},
									},
								},
							},
						},
					},
				},
			},
			"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/test-permissions-boundary"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
					core.MappingNodeFromString("arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create role with permissions boundary",
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
			ResourceID: "test-role-with-permissions-boundary-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-with-permissions-boundary-id",
					ResourceName: "test-role-with-permissions-boundary",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.roleName",
					},
					{
						FieldPath: "spec.assumeRolePolicyDocument",
					},
					{
						FieldPath: "spec.permissionsBoundary",
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
				"spec.roleId": core.MappingNodeFromString(roleId),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateRole": &iam.CreateRoleInput{
				RoleName:                 aws.String("test-role-with-permissions-boundary"),
				AssumeRolePolicyDocument: aws.String(`{"Statement":[{"Action":["sts:AssumeRole"],"Effect":"Allow","Principal":{"Service":["lambda.amazonaws.com"]}}],"Version":"2012-10-17"}`),
			},
			"PutRolePermissionsBoundary": &iam.PutRolePermissionsBoundaryInput{
				RoleName:            aws.String("test-role-with-permissions-boundary"),
				PermissionsBoundary: aws.String("arn:aws:iam::123456789012:policy/test-permissions-boundary"),
			},
			"AttachRolePolicy": []any{
				&iam.AttachRolePolicyInput{
					RoleName:  aws.String("test-role-with-permissions-boundary"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				},
				&iam.AttachRolePolicyInput{
					RoleName:  aws.String("test-role-with-permissions-boundary"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"),
				},
			},
		},
	}
}

func TestIAMRoleResourceCreate(t *testing.T) {
	suite.Run(t, new(IAMRoleResourceCreateSuite))
}
