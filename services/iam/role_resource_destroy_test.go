//go:build unit

package iam

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type IamRoleResourceDestroySuite struct {
	suite.Suite
}

func (s *IamRoleResourceDestroySuite) Test_destroy_iam_role() {
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
		destroyBasicRoleTestCase(providerCtx, loader),
		destroyRoleFailureTestCase(providerCtx, loader),
		destroyRoleWithInlinePoliciesTestCase(providerCtx, loader),
		destroyRoleWithManagedPoliciesTestCase(providerCtx, loader),
		destroyRoleWithBothPolicyTypesTestCase(providerCtx, loader),
		destroyRoleWithInlinePolicyFailureTestCase(providerCtx, loader),
		destroyRoleWithManagedPolicyFailureTestCase(providerCtx, loader),
		destroyRoleWithPermissionsBoundaryTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	roleResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return RoleResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		roleResourceWrapper,
		&s.Suite,
	)
}

func destroyBasicRoleTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteRoleOutput(&iam.DeleteRoleOutput{}),
	)

	// Create test data for role destruction
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
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

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy basic IAM role",
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
			InstanceID: "test-instance-id",
			ResourceID: "TestRole",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		DestroyActionsCalled: map[string]any{
			"DeleteRole": &iam.DeleteRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
	}
}

func destroyRoleFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteRoleError(fmt.Errorf("failed to delete IAM role")),
	)

	// Create test data for role destruction
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
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

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy IAM role failure",
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
			InstanceID: "test-instance-id",
			ResourceID: "TestRole",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteRole": &iam.DeleteRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
	}
}

func destroyRoleWithInlinePoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
		iammock.WithDeleteRoleOutput(&iam.DeleteRoleOutput{}),
	)

	// Create test data for role destruction with inline policies
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
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
							"policyName": core.MappingNodeFromString("S3AccessPolicy"),
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
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("CloudWatchPolicy"),
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
															core.MappingNodeFromString("logs:CreateLogGroup"),
															core.MappingNodeFromString("logs:CreateLogStream"),
															core.MappingNodeFromString("logs:PutLogEvents"),
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

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy IAM role with inline policies",
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
			InstanceID: "test-instance-id",
			ResourceID: "TestRole",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		DestroyActionsCalled: map[string]any{
			"DeleteRolePolicy": []any{
				&iam.DeleteRolePolicyInput{
					RoleName:   aws.String("TestRole"),
					PolicyName: aws.String("S3AccessPolicy"),
				},
				&iam.DeleteRolePolicyInput{
					RoleName:   aws.String("TestRole"),
					PolicyName: aws.String("CloudWatchPolicy"),
				},
			},
			"DeleteRole": &iam.DeleteRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
	}
}

func destroyRoleWithManagedPoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDetachRolePolicyOutput(&iam.DetachRolePolicyOutput{}),
		iammock.WithDeleteRoleOutput(&iam.DeleteRoleOutput{}),
	)

	// Create test data for role destruction with managed policies
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
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

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy IAM role with managed policies",
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
			InstanceID: "test-instance-id",
			ResourceID: "TestRole",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		DestroyActionsCalled: map[string]any{
			"DetachRolePolicy": []any{
				&iam.DetachRolePolicyInput{
					RoleName:  aws.String("TestRole"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				},
				&iam.DetachRolePolicyInput{
					RoleName:  aws.String("TestRole"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"),
				},
			},
			"DeleteRole": &iam.DeleteRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
	}
}

func destroyRoleWithBothPolicyTypesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
		iammock.WithDetachRolePolicyOutput(&iam.DetachRolePolicyOutput{}),
		iammock.WithDeleteRoleOutput(&iam.DeleteRoleOutput{}),
	)

	// Create test data for role destruction with both inline and managed policies
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
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
							"policyName": core.MappingNodeFromString("CustomPolicy"),
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
															core.MappingNodeFromString("arn:aws:dynamodb:*:*:table/my-table"),
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
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				},
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy IAM role with both inline and managed policies",
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
			InstanceID: "test-instance-id",
			ResourceID: "TestRole",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		DestroyActionsCalled: map[string]any{
			"DeleteRolePolicy": []any{
				&iam.DeleteRolePolicyInput{
					RoleName:   aws.String("TestRole"),
					PolicyName: aws.String("CustomPolicy"),
				},
			},
			"DetachRolePolicy": []any{
				&iam.DetachRolePolicyInput{
					RoleName:  aws.String("TestRole"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				},
			},
			"DeleteRole": &iam.DeleteRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
	}
}

func destroyRoleWithInlinePolicyFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteRolePolicyError(fmt.Errorf("failed to delete inline policy")),
	)

	// Create test data for role destruction with inline policy failure
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
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
							"policyName": core.MappingNodeFromString("FailingPolicy"),
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

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy IAM role with inline policy failure",
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
			InstanceID: "test-instance-id",
			ResourceID: "TestRole",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteRolePolicy": []any{
				&iam.DeleteRolePolicyInput{
					RoleName:   aws.String("TestRole"),
					PolicyName: aws.String("FailingPolicy"),
				},
			},
		},
	}
}

func destroyRoleWithManagedPolicyFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDetachRolePolicyError(fmt.Errorf("failed to detach managed policy")),
	)

	// Create test data for role destruction with managed policy failure
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
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
				},
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy IAM role with managed policy failure",
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
			InstanceID: "test-instance-id",
			ResourceID: "TestRole",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DetachRolePolicy": []any{
				&iam.DetachRolePolicyInput{
					RoleName:  aws.String("TestRole"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				},
			},
		},
	}
}

func destroyRoleWithPermissionsBoundaryTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteRoleOutput(&iam.DeleteRoleOutput{}),
	)

	// Create test data for role destruction with permissions boundary
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
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
			"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/PermissionsBoundary"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy IAM role with permissions boundary",
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
			InstanceID: "test-instance-id",
			ResourceID: "TestRole",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		DestroyActionsCalled: map[string]any{
			"DeleteRolePermissionsBoundary": &iam.DeleteRolePermissionsBoundaryInput{
				RoleName: aws.String("TestRole"),
			},
			"DeleteRole": &iam.DeleteRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
	}
}

func TestIamRoleResourceDestroy(t *testing.T) {
	suite.Run(t, new(IamRoleResourceDestroySuite))
}
