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

type IAMSAMLProviderResourceStabilisedSuite struct {
	suite.Suite
}

func (s *IAMSAMLProviderResourceStabilisedSuite) Test_stabilised_iam_saml_provider() {
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
		stabilisedSAMLProviderTestCase(providerCtx, loader),
		stabilisedSAMLProviderWithAllFieldsTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		SAMLProviderResource,
		&s.Suite,
	)
}

func stabilisedSAMLProviderTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name:           "IAM SAML provider is always stabilised",
		ServiceFactory: iammock.CreateIamServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-saml-provider-id",
			ResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"name":                 core.MappingNodeFromString("MySAMLProvider"),
					"samlMetadataDocument": core.MappingNodeFromString(`<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func stabilisedSAMLProviderWithAllFieldsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name:           "IAM SAML provider with all fields is always stabilised",
		ServiceFactory: iammock.CreateIamServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-saml-provider-id",
			ResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"name":                 core.MappingNodeFromString("MySAMLProvider"),
					"samlMetadataDocument": core.MappingNodeFromString(`<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`),
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

func TestIAMSAMLProviderResourceStabilised(t *testing.T) {
	suite.Run(t, new(IAMSAMLProviderResourceStabilisedSuite))
}
