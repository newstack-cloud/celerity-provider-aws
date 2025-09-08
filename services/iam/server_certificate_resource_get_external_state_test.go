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

type ServerCertificateResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *ServerCertificateResourceGetExternalStateSuite) Test_get_external_state_iam_server_certificate() {
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
		getExternalStateServerCertificateTestCase(providerCtx, loader),
		getExternalStateServerCertificateNotFoundTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		ServerCertificateResource,
		&s.Suite,
	)
}

func getExternalStateServerCertificateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	arn := "arn:aws:iam::123456789012:server-certificate/MyServerCertificate"
	iamTags := []types.Tag{
		{
			Key:   aws.String("exampleKey"),
			Value: aws.String("exampleValue"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM server certificate",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetServerCertificateOutput(&iam.GetServerCertificateOutput{
				ServerCertificate: &types.ServerCertificate{
					CertificateBody:  aws.String("certificateBody"),
					CertificateChain: aws.String("certificateChain"),
					ServerCertificateMetadata: &types.ServerCertificateMetadata{
						Arn:                   aws.String(arn),
						ServerCertificateName: aws.String("serverCertificateName"),
						Path:                  aws.String("path"),
					},
					Tags: iamTags,
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
			ResourceID: "test-saml-provider-id",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"serverCertificateName": core.MappingNodeFromString("serverCertificateName"),
					"privateKey":            core.MappingNodeFromString("privateKey"),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":                   core.MappingNodeFromString(arn),
					"serverCertificateName": core.MappingNodeFromString("serverCertificateName"),
					"certificateBody":       core.MappingNodeFromString("certificateBody"),
					"certificateChain":      core.MappingNodeFromString("certificateChain"),
					"path":                  core.MappingNodeFromString("path"),
					"privateKey":            core.MappingNodeFromString("privateKey"),
					"tags":                  extractIAMTags(iamTags),
				},
			},
		},
		ExpectError: false,
	}
}

func getExternalStateServerCertificateNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	serverCertificateArn := "arn:aws:iam::123456789012:server-certificate/NonExistentServerCertificate"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for non-existent IAM server certificate",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetServerCertificateError(fmt.Errorf("server certificate not found")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-server-certificate-id",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(serverCertificateArn),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
	}
}

func TestServerCertificateResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(ServerCertificateResourceGetExternalStateSuite))
}
