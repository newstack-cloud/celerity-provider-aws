//go:build unit

package iam

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMInstanceProfileResourceStabilisedSuite struct {
	suite.Suite
}

func (s *IAMInstanceProfileResourceStabilisedSuite) Test_stabilised_iam_instance_profile() {
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

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		stabilisedInstanceProfileTestCase(providerCtx, loader),
		stabilisedInstanceProfileWithErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		InstanceProfileResource,
		&s.Suite,
	)
}

func stabilisedInstanceProfileTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name:           "IAM instance profile is stabilised",
		ServiceFactory: iammock.CreateIamServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-instance-profile-id",
			ResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
					"path":                core.MappingNodeFromString("/"),
					"role":                core.MappingNodeFromString("MyRole"),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func stabilisedInstanceProfileWithErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name:           "IAM instance profile is stabilised even on error",
		ServiceFactory: iammock.CreateIamServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-instance-profile-id",
			ResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
					"path":                core.MappingNodeFromString("/"),
					"role":                core.MappingNodeFromString("MyRole"),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func TestIAMInstanceProfileResourceStabilised(t *testing.T) {
	suite.Run(t, new(IAMInstanceProfileResourceStabilisedSuite))
}
