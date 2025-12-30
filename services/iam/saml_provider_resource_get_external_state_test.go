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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type IAMSAMLProviderResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *IAMSAMLProviderResourceGetExternalStateSuite) Test_get_external_state_iam_saml_provider() {
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
		getExternalStateSAMLProviderTestCase(providerCtx, loader),
		getExternalStateSAMLProviderWithTagsTestCase(providerCtx, loader),
		getExternalStateSAMLProviderNotFoundTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	samlProviderResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return SAMLProviderResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		samlProviderResourceWrapper,
		&s.Suite,
	)
}

func getExternalStateSAMLProviderTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"
	samlMetadataDocument := `<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" 
                  entityID="http://www.example.com/saml">
    <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" 
                           Location="https://www.example.com/saml/sso"/>
    </IDPSSODescriptor>
</EntityDescriptor>`

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM SAML provider",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetSAMLProviderOutput(&iam.GetSAMLProviderOutput{
				SAMLMetadataDocument: aws.String(samlMetadataDocument),
			}),
			iammock.WithListSAMLProviderTagsOutput(&iam.ListSAMLProviderTagsOutput{
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
			ResourceID: "test-saml-provider-id",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(samlProviderArn),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":                  core.MappingNodeFromString(samlProviderArn),
					"name":                 core.MappingNodeFromString("MySAMLProvider"),
					"samlMetadataDocument": core.MappingNodeFromString(samlMetadataDocument),
				},
			},
		},
		ExpectError: false,
	}
}

func getExternalStateSAMLProviderWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/ExampleCorpProvider"
	samlMetadataDocument := `<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" 
                  entityID="http://corp.example.com/saml">
    <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" 
                           Location="https://idp.corp.example.com/saml/sso"/>
    </IDPSSODescriptor>
</EntityDescriptor>`

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM SAML provider with tags",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetSAMLProviderOutput(&iam.GetSAMLProviderOutput{
				SAMLMetadataDocument: aws.String(samlMetadataDocument),
			}),
			iammock.WithListSAMLProviderTagsOutput(&iam.ListSAMLProviderTagsOutput{
				Tags: []types.Tag{
					{
						Key:   aws.String("Environment"),
						Value: aws.String("Production"),
					},
					{
						Key:   aws.String("Service"),
						Value: aws.String("SSO"),
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
			ResourceID: "test-saml-provider-id",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(samlProviderArn),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":                  core.MappingNodeFromString(samlProviderArn),
					"name":                 core.MappingNodeFromString("ExampleCorpProvider"),
					"samlMetadataDocument": core.MappingNodeFromString(samlMetadataDocument),
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
									"value": core.MappingNodeFromString("SSO"),
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

func getExternalStateSAMLProviderNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/NonExistentProvider"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for non-existent IAM SAML provider",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithGetSAMLProviderError(fmt.Errorf("SAML provider not found")),
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
					"arn": core.MappingNodeFromString(samlProviderArn),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
	}
}

func TestIAMSAMLProviderResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(IAMSAMLProviderResourceGetExternalStateSuite))
}
