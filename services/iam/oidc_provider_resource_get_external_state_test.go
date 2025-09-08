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

type IAMOIDCProviderResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *IAMOIDCProviderResourceGetExternalStateSuite) Test_get_external_state_iam_oidc_provider() {
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
		getExternalStateOIDCProviderTestCase(providerCtx, loader),
		getExternalStateOIDCProviderWithTagsTestCase(providerCtx, loader),
		getExternalStateOIDCProviderNotFoundTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		OIDCProviderResource,
		&s.Suite,
	)
}

func getExternalStateOIDCProviderTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM OIDC provider",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetOpenIDConnectProviderOutput(&iam.GetOpenIDConnectProviderOutput{
				ClientIDList: []string{"sts.amazonaws.com"},
				ThumbprintList: []string{
					"cf23df2207d99a74fbe169e3eba035e633b65d94",
					"9e99a48a9960b14926bb7f3b02e22da2b0ab7280",
				},
				Url: aws.String("https://token.actions.githubusercontent.com"),
			}),
			iammock.WithListOpenIDConnectProviderTagsOutput(&iam.ListOpenIDConnectProviderTagsOutput{
				Tags: []types.Tag{},
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
			ResourceID: "test-oidc-provider-id",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(oidcProviderArn),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(oidcProviderArn),
					"url": core.MappingNodeFromString("https://token.actions.githubusercontent.com"),
					"clientIdList": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("sts.amazonaws.com"),
						},
					},
					"thumbprintList": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
							core.MappingNodeFromString("9e99a48a9960b14926bb7f3b02e22da2b0ab7280"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func getExternalStateOIDCProviderWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/accounts.google.com"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM OIDC provider with tags",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetOpenIDConnectProviderOutput(&iam.GetOpenIDConnectProviderOutput{
				ClientIDList: []string{"123456789012-abcdef.apps.googleusercontent.com"},
				ThumbprintList: []string{
					"cf23df2207d99a74fbe169e3eba035e633b65d94",
				},
				Url: aws.String("https://accounts.google.com"),
			}),
			iammock.WithListOpenIDConnectProviderTagsOutput(&iam.ListOpenIDConnectProviderTagsOutput{
				Tags: []types.Tag{
					{
						Key:   aws.String("Environment"),
						Value: aws.String("Production"),
					},
					{
						Key:   aws.String("Service"),
						Value: aws.String("Authentication"),
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
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(oidcProviderArn),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(oidcProviderArn),
					"url": core.MappingNodeFromString("https://accounts.google.com"),
					"clientIdList": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("123456789012-abcdef.apps.googleusercontent.com"),
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
							{
								Fields: map[string]*core.MappingNode{
									"key":   core.MappingNodeFromString("Service"),
									"value": core.MappingNodeFromString("Authentication"),
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

func getExternalStateOIDCProviderNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/notfound.example.com"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM OIDC provider not found",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetOpenIDConnectProviderError(fmt.Errorf("OIDC provider not found")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(oidcProviderArn),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
	}
}

func TestIAMOIDCProviderResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(IAMOIDCProviderResourceGetExternalStateSuite))
}
