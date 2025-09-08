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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMGroupResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *IAMGroupResourceGetExternalStateSuite) Test_get_external_state_iam_group() {
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
		createBasicGroupGetExternalStateTestCase(providerCtx, loader),
		createGroupWithPoliciesGetExternalStateTestCase(providerCtx, loader),
		createGroupWithManagedPoliciesGetExternalStateTestCase(providerCtx, loader),
		createGroupGetExternalStateFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		GroupResource,
		&s.Suite,
	)
}

func createBasicGroupGetExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group"
	groupId := "AGPA1234567890123456"

	// Create test data for group get external state
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	// Expected output with computed fields
	expectedResourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "get external state basic group",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetGroupOutput(&iam.GetGroupOutput{
				Group: &types.Group{
					Arn:       aws.String(resourceARN),
					GroupId:   aws.String(groupId),
					GroupName: aws.String("test-group"),
					Path:      aws.String("/"),
				},
			}),
			iammock.WithListGroupPoliciesOutput(&iam.ListGroupPoliciesOutput{
				PolicyNames: []string{},
			}),
			iammock.WithListAttachedGroupPoliciesOutput(&iam.ListAttachedGroupPoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{},
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
			ResourceID:          "test-group-id",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expectedResourceSpecState,
		},
	}
}

func createGroupWithPoliciesGetExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group-with-policies"
	groupId := "AGPA1234567890123457"

	// Create test data for group get external state with policies
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group-with-policies"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	// Expected output with computed fields
	expectedResourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group-with-policies"),
			"path":      core.MappingNodeFromString("/"),
			"policies": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("TestPolicy"),
							"policyDocument": {
								Fields: map[string]*core.MappingNode{
									"Version": core.MappingNodeFromString("2012-10-17"),
									"Statement": {
										Items: []*core.MappingNode{
											{
												Fields: map[string]*core.MappingNode{
													"Effect":   core.MappingNodeFromString("Allow"),
													"Action":   {Items: []*core.MappingNode{core.MappingNodeFromString("s3:GetObject")}},
													"Resource": {Items: []*core.MappingNode{core.MappingNodeFromString("*")}},
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

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "get external state group with policies",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetGroupOutput(&iam.GetGroupOutput{
				Group: &types.Group{
					Arn:       aws.String(resourceARN),
					GroupId:   aws.String(groupId),
					GroupName: aws.String("test-group-with-policies"),
					Path:      aws.String("/"),
				},
			}),
			iammock.WithListGroupPoliciesOutput(&iam.ListGroupPoliciesOutput{
				PolicyNames: []string{"TestPolicy"},
			}),
			iammock.WithGetGroupPolicyOutput(&iam.GetGroupPolicyOutput{
				PolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject"],"Resource":["*"]}]}`),
			}),
			iammock.WithListAttachedGroupPoliciesOutput(&iam.ListAttachedGroupPoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{},
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
			ResourceID:          "test-group-id",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expectedResourceSpecState,
		},
	}
}

func createGroupWithManagedPoliciesGetExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group-with-managed-policies"
	groupId := "AGPA1234567890123458"

	// Create test data for group get external state with managed policies
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group-with-managed-policies"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	// Expected output with computed fields
	expectedResourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group-with-managed-policies"),
			"path":      core.MappingNodeFromString("/"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
				},
			},
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "get external state group with managed policies",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetGroupOutput(&iam.GetGroupOutput{
				Group: &types.Group{
					Arn:       aws.String(resourceARN),
					GroupId:   aws.String(groupId),
					GroupName: aws.String("test-group-with-managed-policies"),
					Path:      aws.String("/"),
				},
			}),
			iammock.WithListGroupPoliciesOutput(&iam.ListGroupPoliciesOutput{
				PolicyNames: []string{},
			}),
			iammock.WithListAttachedGroupPoliciesOutput(&iam.ListAttachedGroupPoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
						PolicyName: aws.String("ReadOnlyAccess"),
					},
				},
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
			ResourceID:          "test-group-id",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expectedResourceSpecState,
		},
	}
}

func createGroupGetExternalStateFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:group/test-group"
	groupId := "AGPA1234567890123456"
	// Create test data for group get external state
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":       core.MappingNodeFromString(resourceARN),
			"groupId":   core.MappingNodeFromString(groupId),
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "get external state group failure",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetGroupError(fmt.Errorf("failed to get group")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-group-id",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectError: true,
	}
}

func TestIAMGroupResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(IAMGroupResourceGetExternalStateSuite))
}
