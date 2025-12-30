//go:build unit

package iam

import (
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

type IamRoleResourceGetExternalStateSuite struct {
	suite.Suite
}

// PolicyDocument represents the structure of an IAM policy document.
type PolicyDocument struct {
	Version   string            `json:"Version"`
	Statement []PolicyStatement `json:"Statement"`
}

type PolicyStatement struct {
	Effect    string                 `json:"Effect"`
	Principal map[string]interface{} `json:"Principal,omitempty"`
	Action    interface{}            `json:"Action,omitempty"`
	Resource  interface{}            `json:"Resource,omitempty"`
}

func (s *IamRoleResourceGetExternalStateSuite) Test_get_external_state_iam_role() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		getExternalStateBasicRoleTestCase(providerCtx, loader),
		getExternalStateRoleNotFoundTestCase(providerCtx, loader),
		getExternalStateCompleteRoleTestCase(providerCtx, loader),
		createGetExternalStateWithInlinePoliciesTestCase(providerCtx, loader),
		createGetExternalStateWithTagsAndManagedPoliciesTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	roleResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return RoleResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		roleResourceWrapper,
		&s.Suite,
	)
}

func getExternalStateBasicRoleTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	// Create test data for role get external state
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
			"assumeRolePolicyDocument": core.MappingNodeFromString(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": {"Service": "lambda.amazonaws.com"},
					"Action": "sts:AssumeRole"
				}]
			}`),
		},
	}

	// Expected output with computed fields
	expectedResourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("TestRole"),
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
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
			"arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
			"roleId": core.MappingNodeFromString("AROA1234567890123456"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "get external state basic IAM role",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetRoleOutput(&iam.GetRoleOutput{
				Role: &types.Role{
					RoleName: aws.String("TestRole"),
					Arn:      aws.String("arn:aws:iam::123456789012:role/TestRole"),
					RoleId:   aws.String("AROA1234567890123456"),
					AssumeRolePolicyDocument: aws.String(`{
						"Version": "2012-10-17",
						"Statement": [{
							"Effect": "Allow",
							"Principal": {"Service": "lambda.amazonaws.com"},
							"Action": "sts:AssumeRole"
						}]
					}`),
				},
			}),
			iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{},
			}),
			iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{
				PolicyNames: []string{},
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
			ResourceID:          "TestRole",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expectedResourceSpecState,
		},
	}
}

func getExternalStateRoleNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "get external state IAM role not found",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetRoleError(&smithy.GenericAPIError{
				Code:    "NoSuchEntity",
				Message: "The role with name NonexistentRole cannot be found.",
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID: "test-instance-id",
			ResourceID: "NonexistentRole",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/NonexistentRole"),
					"assumeRolePolicyDocument": core.MappingNodeFromString(`{
						"Version": "2012-10-17",
						"Statement": [{
							"Effect": "Allow",
							"Principal": {"Service": "lambda.amazonaws.com"},
							"Action": "sts:AssumeRole"
						}]
					}`),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{Fields: make(map[string]*core.MappingNode)},
		},
	}
}

func getExternalStateCompleteRoleTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	// Create test data for complete role get external state
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                core.MappingNodeFromString("arn:aws:iam::123456789012:role/test/CompleteTestRole"),
			"description":        core.MappingNodeFromString("A complete test role"),
			"maxSessionDuration": core.MappingNodeFromInt(7200),
			"path":               core.MappingNodeFromString("/test/"),
			"assumeRolePolicyDocument": core.MappingNodeFromString(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": {"Service": "lambda.amazonaws.com"},
					"Action": "sts:AssumeRole"
				}]
			}`),
		},
	}

	// Expected output with all fields
	expectedResourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName":           core.MappingNodeFromString("CompleteTestRole"),
			"description":        core.MappingNodeFromString("A complete test role"),
			"maxSessionDuration": core.MappingNodeFromInt(7200),
			"path":               core.MappingNodeFromString("/test/"),
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
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
			"arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/test/CompleteTestRole"),
			"roleId": core.MappingNodeFromString("AROA1234567890123456"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "get external state complete IAM role",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetRoleOutput(&iam.GetRoleOutput{
				Role: &types.Role{
					RoleName:           aws.String("CompleteTestRole"),
					Arn:                aws.String("arn:aws:iam::123456789012:role/test/CompleteTestRole"),
					RoleId:             aws.String("AROA1234567890123456"),
					Description:        aws.String("A complete test role"),
					MaxSessionDuration: aws.Int32(7200),
					Path:               aws.String("/test/"),
					AssumeRolePolicyDocument: aws.String(`{
						"Version": "2012-10-17",
						"Statement": [{
							"Effect": "Allow",
							"Principal": {"Service": "lambda.amazonaws.com"},
							"Action": "sts:AssumeRole"
						}]
					}`),
				},
			}),
			iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{},
			}),
			iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{
				PolicyNames: []string{},
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
			ResourceID:          "CompleteTestRole",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expectedResourceSpecState,
		},
	}
}

func createGetExternalStateWithInlinePoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("TestRole"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/TestRole"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{},
		}),
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{
			PolicyNames: []string{"DynamoDBAccess"},
		}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{
			RoleName:       aws.String("TestRole"),
			PolicyName:     aws.String("DynamoDBAccess"),
			PolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["dynamodb:GetItem","dynamodb:PutItem"],"Resource":["arn:aws:dynamodb:*:*:table/MyTable"]}]}`),
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "successfully gets role state with inline policies",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
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
					"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"roleName": core.MappingNodeFromString("TestRole"),
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
					"roleId":   core.MappingNodeFromString("AROA1234567890123456"),
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
													"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
												},
											},
											"Action": core.MappingNodeFromString("sts:AssumeRole"),
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
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createGetExternalStateWithTagsAndManagedPoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("TestRoleWithExtras"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/TestRoleWithExtras"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
				PermissionsBoundary: &types.AttachedPermissionsBoundary{
					PermissionsBoundaryArn:  aws.String("arn:aws:iam::123456789012:policy/MyPermissionsBoundary"),
					PermissionsBoundaryType: types.PermissionsBoundaryAttachmentTypePolicy,
				},
				Tags: []types.Tag{
					{Key: aws.String("Environment"), Value: aws.String("test")},
					{Key: aws.String("Team"), Value: aws.String("platform")},
				},
			},
		}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{
				{PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"), PolicyName: aws.String("AmazonS3ReadOnlyAccess")},
				{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/CustomPolicy"), PolicyName: aws.String("CustomPolicy")},
			},
		}),
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{
			PolicyNames: []string{},
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "successfully gets role state with tags and managed policies",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
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
					"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRoleWithExtras"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"roleName": core.MappingNodeFromString("TestRoleWithExtras"),
					"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRoleWithExtras"),
					"roleId":   core.MappingNodeFromString("AROA1234567890123456"),
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
													"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
												},
											},
											"Action": core.MappingNodeFromString("sts:AssumeRole"),
										},
									},
								},
							},
						},
					},
					"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/MyPermissionsBoundary"),
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
									"value": core.MappingNodeFromString("platform"),
								},
							},
						},
					},
					"managedPolicyArns": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"),
							core.MappingNodeFromString("arn:aws:iam::123456789012:policy/CustomPolicy"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func TestIamRoleResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(IamRoleResourceGetExternalStateSuite))
}
