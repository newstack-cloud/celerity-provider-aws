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

type IAMUserResourceStabilisedSuite struct {
	suite.Suite
}

func (s *IAMUserResourceStabilisedSuite) Test_stabilised() {
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
		{
			Name: "returns stabilised when user exists",
			ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
				return iammock.CreateIamServiceMock()
			},
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceHasStabilisedInput{
				ProviderContext: providerCtx,
				ResourceSpec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/test-user"),
						"userName": core.MappingNodeFromString("test-user"),
					},
				},
			},
			ExpectedOutput: &provider.ResourceHasStabilisedOutput{
				Stabilised: true,
			},
			ExpectError: false,
		},
		{
			Name: "returns stabilised when user exists with all features",
			ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
				return iammock.CreateIamServiceMock()
			},
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceHasStabilisedInput{
				ProviderContext: providerCtx,
				ResourceSpec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:user/complex-user"),
						"userName": core.MappingNodeFromString("complex-user"),
						"path":     core.MappingNodeFromString("/engineering/"),
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
										"key":   core.MappingNodeFromString("Department"),
										"value": core.MappingNodeFromString("engineering"),
									},
								},
							},
						},
						"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/UserPermissionsBoundary"),
					},
				},
			},
			ExpectedOutput: &provider.ResourceHasStabilisedOutput{
				Stabilised: true,
			},
			ExpectError: false,
		},
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		UserResource,
		&s.Suite,
	)
}

func TestIAMUserResourceStabilised(t *testing.T) {
	suite.Run(t, new(IAMUserResourceStabilisedSuite))
}
