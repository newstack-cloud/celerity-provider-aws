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
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMGroupResourceDestroySuite struct {
	suite.Suite
}

func (s *IAMGroupResourceDestroySuite) Test_destroy_iam_group() {
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
		createBasicGroupDestroyTestCase(providerCtx, loader),
		createGroupWithPoliciesDestroyTestCase(providerCtx, loader),
		createGroupWithManagedPoliciesDestroyTestCase(providerCtx, loader),
		createGroupWithBothPoliciesDestroyTestCase(providerCtx, loader),
		createGroupDestroyFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		GroupResource,
		&s.Suite,
	)
}

func createBasicGroupDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteGroupOutput(&iam.DeleteGroupOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy basic group",
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
						"arn":               core.MappingNodeFromString("arn:aws:iam::123456789012:group/test-group"),
						"groupId":           core.MappingNodeFromString("AGPA1234567890123456"),
						"groupName":         core.MappingNodeFromString("test-group"),
						"path":              core.MappingNodeFromString("/"),
						"managedPolicyArns": {Items: []*core.MappingNode{}},
						"policies":          {Items: []*core.MappingNode{}},
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"DeleteGroup": &iam.DeleteGroupInput{
				GroupName: aws.String("test-group"),
			},
		},
	}
}

func createGroupWithPoliciesDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListGroupPoliciesOutput(&iam.ListGroupPoliciesOutput{
			PolicyNames: []string{"TestPolicy"},
		}),
		iammock.WithDeleteGroupPolicyOutput(&iam.DeleteGroupPolicyOutput{}),
		iammock.WithDeleteGroupOutput(&iam.DeleteGroupOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy group with inline policies",
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
						"arn":               core.MappingNodeFromString("arn:aws:iam::123456789012:group/test-group-with-policies"),
						"groupId":           core.MappingNodeFromString("AGPA1234567890123457"),
						"groupName":         core.MappingNodeFromString("test-group-with-policies"),
						"path":              core.MappingNodeFromString("/"),
						"managedPolicyArns": {Items: []*core.MappingNode{}},
						"policies":          {Items: []*core.MappingNode{}},
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"ListGroupPolicies": &iam.ListGroupPoliciesInput{
				GroupName: aws.String("test-group-with-policies"),
			},
			"DeleteGroupPolicy": &iam.DeleteGroupPolicyInput{
				GroupName:  aws.String("test-group-with-policies"),
				PolicyName: aws.String("TestPolicy"),
			},
			"DeleteGroup": &iam.DeleteGroupInput{
				GroupName: aws.String("test-group-with-policies"),
			},
		},
	}
}

func createGroupWithManagedPoliciesDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListAttachedGroupPoliciesOutput(&iam.ListAttachedGroupPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{
				{
					PolicyArn:  aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					PolicyName: aws.String("ReadOnlyAccess"),
				},
			},
		}),
		iammock.WithDetachGroupPolicyOutput(&iam.DetachGroupPolicyOutput{}),
		iammock.WithDeleteGroupOutput(&iam.DeleteGroupOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy group with managed policies",
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
						"arn":               core.MappingNodeFromString("arn:aws:iam::123456789012:group/test-group-with-managed-policies"),
						"groupId":           core.MappingNodeFromString("AGPA1234567890123458"),
						"groupName":         core.MappingNodeFromString("test-group-with-managed-policies"),
						"path":              core.MappingNodeFromString("/"),
						"managedPolicyArns": {Items: []*core.MappingNode{}},
						"policies":          {Items: []*core.MappingNode{}},
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"ListAttachedGroupPolicies": &iam.ListAttachedGroupPoliciesInput{
				GroupName: aws.String("test-group-with-managed-policies"),
			},
			"DetachGroupPolicy": &iam.DetachGroupPolicyInput{
				GroupName: aws.String("test-group-with-managed-policies"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
			},
			"DeleteGroup": &iam.DeleteGroupInput{
				GroupName: aws.String("test-group-with-managed-policies"),
			},
		},
	}
}

func createGroupWithBothPoliciesDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithListGroupPoliciesOutput(&iam.ListGroupPoliciesOutput{
			PolicyNames: []string{"TestPolicy"},
		}),
		iammock.WithListAttachedGroupPoliciesOutput(&iam.ListAttachedGroupPoliciesOutput{
			AttachedPolicies: []types.AttachedPolicy{
				{
					PolicyArn:  aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					PolicyName: aws.String("ReadOnlyAccess"),
				},
			},
		}),
		iammock.WithDeleteGroupPolicyOutput(&iam.DeleteGroupPolicyOutput{}),
		iammock.WithDetachGroupPolicyOutput(&iam.DetachGroupPolicyOutput{}),
		iammock.WithDeleteGroupOutput(&iam.DeleteGroupOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy group with both inline and managed policies",
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
						"arn":               core.MappingNodeFromString("arn:aws:iam::123456789012:group/test-group-with-both-policies"),
						"groupId":           core.MappingNodeFromString("AGPA1234567890123459"),
						"groupName":         core.MappingNodeFromString("test-group-with-both-policies"),
						"path":              core.MappingNodeFromString("/"),
						"managedPolicyArns": {Items: []*core.MappingNode{}},
						"policies":          {Items: []*core.MappingNode{}},
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"ListGroupPolicies": &iam.ListGroupPoliciesInput{
				GroupName: aws.String("test-group-with-both-policies"),
			},
			"ListAttachedGroupPolicies": &iam.ListAttachedGroupPoliciesInput{
				GroupName: aws.String("test-group-with-both-policies"),
			},
			"DeleteGroupPolicy": &iam.DeleteGroupPolicyInput{
				GroupName:  aws.String("test-group-with-both-policies"),
				PolicyName: aws.String("TestPolicy"),
			},
			"DetachGroupPolicy": &iam.DetachGroupPolicyInput{
				GroupName: aws.String("test-group-with-both-policies"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
			},
			"DeleteGroup": &iam.DeleteGroupInput{
				GroupName: aws.String("test-group-with-both-policies"),
			},
		},
	}
}

func createGroupDestroyFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteGroupError(fmt.Errorf("failed to delete group")),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "destroy group failure",
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
						"arn":               core.MappingNodeFromString("arn:aws:iam::123456789012:group/test-group"),
						"groupId":           core.MappingNodeFromString("AGPA1234567890123456"),
						"groupName":         core.MappingNodeFromString("test-group"),
						"path":              core.MappingNodeFromString("/"),
						"managedPolicyArns": {Items: []*core.MappingNode{}},
						"policies":          {Items: []*core.MappingNode{}},
					},
				},
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteGroup": &iam.DeleteGroupInput{
				GroupName: aws.String("test-group"),
			},
		},
	}
}

func TestIAMGroupResourceDestroy(t *testing.T) {
	suite.Run(t, new(IAMGroupResourceDestroySuite))
}
