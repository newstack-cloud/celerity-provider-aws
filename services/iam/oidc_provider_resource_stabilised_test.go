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

type IAMOIDCProviderResourceStabilisedSuite struct {
	suite.Suite
}

func (s *IAMOIDCProviderResourceStabilisedSuite) Test_stabilised_iam_oidc_provider() {
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
		stabilisedOIDCProviderTestCase(providerCtx, loader),
		stabilisedOIDCProviderWithAllFieldsTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		OIDCProviderResource,
		&s.Suite,
	)
}

func stabilisedOIDCProviderTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name:           "IAM OIDC provider is always stabilised",
		ServiceFactory: iammock.CreateIamServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			ResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"url": core.MappingNodeFromString("https://example.com"),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func stabilisedOIDCProviderWithAllFieldsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name:           "IAM OIDC provider with all fields is always stabilised",
		ServiceFactory: iammock.CreateIamServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			ResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"url": core.MappingNodeFromString("https://example.com"),
					"clientIdList": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("my-app-id"),
						},
					},
					"thumbprintList": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
						},
					},
					"tags": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"key":   core.MappingNodeFromString("Environment"),
									"value": core.MappingNodeFromString("Production"),
								},
							},
						},
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func TestIAMOIDCProviderResourceStabilised(t *testing.T) {
	suite.Run(t, new(IAMOIDCProviderResourceStabilisedSuite))
}
