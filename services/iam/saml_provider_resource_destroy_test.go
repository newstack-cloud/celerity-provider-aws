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

type IAMSAMLProviderResourceDestroySuite struct {
	suite.Suite
}

func (s *IAMSAMLProviderResourceDestroySuite) Test_destroy_iam_saml_provider() {
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
		destroySAMLProviderTestCase(providerCtx, loader),
		destroySAMLProviderMissingArnTestCase(providerCtx, loader),
		destroySAMLProviderServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		SAMLProviderResource,
		&s.Suite,
	)
}

func destroySAMLProviderTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"

	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteSAMLProviderOutput(&iam.DeleteSAMLProviderOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM SAML provider",
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
						"arn":                  core.MappingNodeFromString(samlProviderArn),
						"name":                 core.MappingNodeFromString("MySAMLProvider"),
						"samlMetadataDocument": core.MappingNodeFromString(`<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func destroySAMLProviderMissingArnTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM SAML provider missing ARN",
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
						"name":                 core.MappingNodeFromString("MySAMLProvider"),
						"samlMetadataDocument": core.MappingNodeFromString(`<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`),
					},
				},
			},
		},
		ExpectError: true,
	}
}

func destroySAMLProviderServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"

	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteSAMLProviderError(fmt.Errorf("failed to delete SAML provider")),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM SAML provider service error",
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
						"arn":                  core.MappingNodeFromString(samlProviderArn),
						"name":                 core.MappingNodeFromString("MySAMLProvider"),
						"samlMetadataDocument": core.MappingNodeFromString(`<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`),
					},
				},
			},
		},
		ExpectError: true,
	}
}

func TestIAMSAMLProviderResourceDestroy(t *testing.T) {
	suite.Run(t, new(IAMSAMLProviderResourceDestroySuite))
}
