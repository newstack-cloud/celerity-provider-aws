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
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMSAMLProviderResourceCreateSuite struct {
	suite.Suite
}

func (s *IAMSAMLProviderResourceCreateSuite) Test_create_iam_saml_provider() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		createBasicSAMLProviderTestCase(providerCtx, loader),
		createSAMLProviderWithTagsTestCase(providerCtx, loader),
		createSAMLProviderServiceErrorTestCase(providerCtx, loader),
		createSAMLProviderMissingMetadataTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		SAMLProviderResource,
		&s.Suite,
	)
}

func createBasicSAMLProviderTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"
	samlMetadataDocument := `<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" 
                  entityID="http://www.example.com/saml">
    <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" 
                           Location="https://www.example.com/saml/sso"/>
    </IDPSSODescriptor>
</EntityDescriptor>`
	samlProviderUUID := "96dc2683-50a4-4f46-8c0f-4dedf83a8ead"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateSAMLProviderOutput(&iam.CreateSAMLProviderOutput{
			SAMLProviderArn: aws.String(samlProviderArn),
		}),
		iammock.WithGetSAMLProviderOutput(&iam.GetSAMLProviderOutput{
			SAMLProviderUUID: aws.String(samlProviderUUID),
		}),
	)

	// Create test data for SAML provider creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":                 core.MappingNodeFromString("MySAMLProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(samlMetadataDocument),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create basic IAM SAML provider",
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
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-saml-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-saml-provider-id",
					ResourceName: "TestSAMLProvider",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/samlProvider",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.name",
					},
					{
						FieldPath: "spec.samlMetadataDocument",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":              core.MappingNodeFromString(samlProviderArn),
				"spec.samlProviderUUID": core.MappingNodeFromString(samlProviderUUID),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateSAMLProvider": &iam.CreateSAMLProviderInput{
				Name:                 aws.String("MySAMLProvider"),
				SAMLMetadataDocument: aws.String(samlMetadataDocument),
				Tags:                 []types.Tag{},
			},
			"GetSAMLProvider": &iam.GetSAMLProviderInput{
				SAMLProviderArn: aws.String(samlProviderArn),
			},
		},
	}
}

func createSAMLProviderWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/ExampleCorpProvider"
	samlMetadataDocument := `<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" 
                  entityID="http://corp.example.com/saml">
    <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" 
                           Location="https://idp.corp.example.com/saml/sso"/>
    </IDPSSODescriptor>
</EntityDescriptor>`
	samlProviderUUID := "96dc2683-50a4-4f46-8c0f-4dedf83a8ead"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateSAMLProviderOutput(&iam.CreateSAMLProviderOutput{
			SAMLProviderArn: aws.String(samlProviderArn),
		}),
		iammock.WithGetSAMLProviderOutput(&iam.GetSAMLProviderOutput{
			SAMLProviderUUID: aws.String(samlProviderUUID),
		}),
	)

	// Create test data for SAML provider creation with tags
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
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
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM SAML provider with tags",
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
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-saml-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-saml-provider-id",
					ResourceName: "TestSAMLProvider",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/samlProvider",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.name",
					},
					{
						FieldPath: "spec.samlMetadataDocument",
					},
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":              core.MappingNodeFromString(samlProviderArn),
				"spec.samlProviderUUID": core.MappingNodeFromString(samlProviderUUID),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateSAMLProvider": &iam.CreateSAMLProviderInput{
				Name:                 aws.String("ExampleCorpProvider"),
				SAMLMetadataDocument: aws.String(samlMetadataDocument),
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
			},
			"GetSAMLProvider": &iam.GetSAMLProviderInput{
				SAMLProviderArn: aws.String(samlProviderArn),
			},
		},
	}
}

func createSAMLProviderServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	serviceError := fmt.Errorf("AWS service error")

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateSAMLProviderError(serviceError),
	)

	samlMetadataDocument := `<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":                 core.MappingNodeFromString("FailedProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(samlMetadataDocument),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM SAML provider with service error",
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
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-saml-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-saml-provider-id",
					ResourceName: "TestSAMLProvider",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/samlProvider",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.name",
					},
					{
						FieldPath: "spec.samlMetadataDocument",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"CreateSAMLProvider": &iam.CreateSAMLProviderInput{
				Name:                 aws.String("FailedProvider"),
				SAMLMetadataDocument: aws.String(samlMetadataDocument),
				Tags:                 []types.Tag{},
			},
		},
	}
}

func createSAMLProviderMissingMetadataTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	// Create test data without SAML metadata document
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("InvalidProvider"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM SAML provider without metadata document",
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
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-saml-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-saml-provider-id",
					ResourceName: "TestSAMLProvider",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/samlProvider",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.name",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
	}
}

func TestIAMSAMLProviderResourceCreateSuite(t *testing.T) {
	suite.Run(t, new(IAMSAMLProviderResourceCreateSuite))
}
