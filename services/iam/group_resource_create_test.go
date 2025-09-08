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

type IAMGroupResourceCreateSuite struct {
	suite.Suite
}

func (s *IAMGroupResourceCreateSuite) Test_create_iam_group() {
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
		createBasicGroupCreateTestCase(providerCtx, loader),
		createGroupWithInlinePoliciesTestCase(providerCtx, loader),
		createGroupWithManagedPoliciesTestCase(providerCtx, loader),
		createGroupWithGeneratedNameTestCase(providerCtx, loader),
		createGroupFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		GroupResource,
		&s.Suite,
	)
}

func createBasicGroupCreateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group"
	groupId := "AGPA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateGroupOutput(&iam.CreateGroupOutput{
			Group: &types.Group{
				Arn:       aws.String(resourceARN),
				GroupId:   aws.String(groupId),
				GroupName: aws.String("test-group"),
				Path:      aws.String("/"),
			},
		}),
	)

	// Create test data for group creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create basic group",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.groupName",
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
				"spec.arn":     core.MappingNodeFromString(resourceARN),
				"spec.groupId": core.MappingNodeFromString(groupId),
			},
		},
		SaveActionsCalled: map[string]any{},
	}
}

func createGroupWithInlinePoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group-with-policies"
	groupId := "AGPA1234567890123457"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateGroupOutput(&iam.CreateGroupOutput{
			Group: &types.Group{
				Arn:       aws.String(resourceARN),
				GroupId:   aws.String(groupId),
				GroupName: aws.String("test-group-with-policies"),
				Path:      aws.String("/"),
			},
		}),
		iammock.WithPutGroupPolicyOutput(&iam.PutGroupPolicyOutput{}),
	)

	// Create test data for group creation with inline policies
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group-with-policies"),
			"path":      core.MappingNodeFromString("/"),
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
		Name: "create group with inline policies",
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
					ResourceName: "TestGroupWithPolicies",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.groupName",
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
				"spec.arn":     core.MappingNodeFromString(resourceARN),
				"spec.groupId": core.MappingNodeFromString(groupId),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateGroup": &iam.CreateGroupInput{
				GroupName: aws.String("test-group-with-policies"),
				Path:      aws.String("/"),
			},
			"PutGroupPolicy": &iam.PutGroupPolicyInput{
				GroupName:      aws.String("test-group-with-policies"),
				PolicyName:     aws.String("S3Access"),
				PolicyDocument: aws.String(`{"Statement":[{"Action":["s3:GetObject","s3:PutObject"],"Effect":"Allow","Resource":["arn:aws:s3:::my-bucket/*"]}],"Version":"2012-10-17"}`),
			},
		},
	}
}

func createGroupWithManagedPoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group-with-managed-policies"
	groupId := "AGPA1234567890123458"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateGroupOutput(&iam.CreateGroupOutput{
			Group: &types.Group{
				Arn:       aws.String(resourceARN),
				GroupId:   aws.String(groupId),
				GroupName: aws.String("test-group-with-managed-policies"),
				Path:      aws.String("/"),
			},
		}),
		iammock.WithAttachGroupPolicyOutput(&iam.AttachGroupPolicyOutput{}),
	)

	// Create test data for group creation with managed policies
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group-with-managed-policies"),
			"path":      core.MappingNodeFromString("/"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					core.MappingNodeFromString("arn:aws:iam::aws:policy/PowerUserAccess"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create group with managed policies",
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
					ResourceName: "TestGroupWithManagedPolicies",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.groupName",
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
				"spec.arn":     core.MappingNodeFromString(resourceARN),
				"spec.groupId": core.MappingNodeFromString(groupId),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateGroup": &iam.CreateGroupInput{
				GroupName: aws.String("test-group-with-managed-policies"),
				Path:      aws.String("/"),
			},
			"AttachGroupPolicy": []any{
				&iam.AttachGroupPolicyInput{
					GroupName: aws.String("test-group-with-managed-policies"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
				},
				&iam.AttachGroupPolicyInput{
					GroupName: aws.String("test-group-with-managed-policies"),
					PolicyArn: aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
				},
			},
		},
	}
}

func createGroupWithGeneratedNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/generated-group-name-test"
	groupId := "AGPA1234567890123459"

	// Mock name generator
	// The resource implementation should allow injection of the name generator for deterministic tests.
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateGroupOutput(&iam.CreateGroupOutput{
			Group: &types.Group{
				Arn:       aws.String(resourceARN),
				GroupId:   aws.String(groupId),
				GroupName: aws.String("generated-group-name-test"),
				Path:      aws.String("/"),
			},
		}),
	)

	// Create test data for group creation with generated name
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create group with generated name",
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
					ResourceName: "TestGroupWithGeneratedName",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{{FieldPath: "spec.path"}},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":     core.MappingNodeFromString(resourceARN),
				"spec.groupId": core.MappingNodeFromString(groupId),
			},
		},
		SaveActionsCalled: map[string]any{},
		// Inject the mock name generator into the resource actions if possible
		// (If not, this will need to be set up in the resource factory or test setup)
	}
}

func createGroupFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateGroupError(fmt.Errorf("failed to create group")),
	)

	// Create test data for group creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create group failure",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/group",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.groupName",
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
			"CreateGroup": &iam.CreateGroupInput{
				GroupName: aws.String("test-group"),
				Path:      aws.String("/"),
			},
		},
	}
}

func TestIAMGroupResourceCreate(t *testing.T) {
	suite.Run(t, new(IAMGroupResourceCreateSuite))
}
