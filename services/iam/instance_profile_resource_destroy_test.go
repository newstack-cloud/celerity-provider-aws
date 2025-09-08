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
	"github.com/stretchr/testify/suite"
)

type IAMInstanceProfileResourceDestroySuite struct {
	suite.Suite
}

func (s *IAMInstanceProfileResourceDestroySuite) Test_destroy_iam_instance_profile() {
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
		destroyInstanceProfileTestCase(providerCtx, loader),
		destroyInstanceProfileServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		InstanceProfileResource,
		&s.Suite,
	)
}

func destroyInstanceProfileTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithRemoveRoleFromInstanceProfileOutput(&iam.RemoveRoleFromInstanceProfileOutput{}),
		iammock.WithDeleteInstanceProfileOutput(&iam.DeleteInstanceProfileOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM instance profile",
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
						"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
						"path":                core.MappingNodeFromString("/"),
						"role":                core.MappingNodeFromString("MyRole"),
						"arn":                 core.MappingNodeFromString("arn:aws:iam::123456789012:instance-profile/MyInstanceProfile"),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"RemoveRoleFromInstanceProfile": &iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				RoleName:            aws.String("MyRole"),
			},
			"DeleteInstanceProfile": &iam.DeleteInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
			},
		},
	}
}

func destroyInstanceProfileServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithRemoveRoleFromInstanceProfileError(fmt.Errorf("failed to remove role from instance profile")),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM instance profile with service error",
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
						"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
						"path":                core.MappingNodeFromString("/"),
						"role":                core.MappingNodeFromString("MyRole"),
						"arn":                 core.MappingNodeFromString("arn:aws:iam::123456789012:instance-profile/MyInstanceProfile"),
					},
				},
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"RemoveRoleFromInstanceProfile": &iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				RoleName:            aws.String("MyRole"),
			},
		},
	}
}

func TestIAMInstanceProfileResourceDestroy(t *testing.T) {
	suite.Run(t, new(IAMInstanceProfileResourceDestroySuite))
}
