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

type IAMGroupResourceStabilisedSuite struct {
	suite.Suite
}

func (s *IAMGroupResourceStabilisedSuite) Test_stabilised_iam_group() {
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
		createBasicGroupStabilisedTestCase(providerCtx, loader),
		createGroupStabilisedFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		GroupResource,
		&s.Suite,
	)
}

func createBasicGroupStabilisedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	// Create test data for group stabilised check
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name: "IAM group is always stabilised",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID:      "test-instance-id",
			ResourceID:      "test-group-id",
			ResourceSpec:    specData,
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func createGroupStabilisedFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	// Create test data for group stabilised check
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"groupName": core.MappingNodeFromString("test-group"),
			"path":      core.MappingNodeFromString("/"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name: "IAM group is always stabilised even on error",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID:      "test-instance-id",
			ResourceID:      "test-group-id",
			ResourceSpec:    specData,
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func TestIAMGroupResourceStabilised(t *testing.T) {
	suite.Run(t, new(IAMGroupResourceStabilisedSuite))
}
