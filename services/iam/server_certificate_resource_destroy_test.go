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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type ServerCertificateResourceDestroySuite struct {
	suite.Suite
}

func (s *ServerCertificateResourceDestroySuite) Test_destroy_iam_server_certificate() {
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
		destroyServerCertificateTestCase(providerCtx, loader),
		destroyServerCertificateMissingServerCertNameTestCase(providerCtx, loader),
		destroyServerCertificateServiceErrorTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	serverCertificateResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return ServerCertificateResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		serverCertificateResourceWrapper,
		&s.Suite,
	)
}

func destroyServerCertificateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteServerCertificateOutput(&iam.DeleteServerCertificateOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM server certificate",
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
						"serverCertificateName": core.MappingNodeFromString("MyServerCertificate"),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func destroyServerCertificateMissingServerCertNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM server certificate missing server certificate name",
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
						"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:server-certificate/MyServerCertificate"),
					},
				},
			},
		},
		ExpectError: true,
	}
}

func destroyServerCertificateServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteServerCertificateError(fmt.Errorf("failed to delete server certificate")),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM server certificate service error",
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
						"serverCertificateName": core.MappingNodeFromString("MyServerCertificate"),
					},
				},
			},
		},
		ExpectError: true,
	}
}

func TestServerCertificateResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(ServerCertificateResourceDestroySuite))
}
