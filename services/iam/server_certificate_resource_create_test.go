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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type ServerCertificateResourceCreateSuite struct {
	suite.Suite
}

func TestServerCertificateResourceCreateSuite(t *testing.T) {
	suite.Run(t, new(ServerCertificateResourceCreateSuite))
}

func (s *ServerCertificateResourceCreateSuite) Test_create_iam_server_certificate() {
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
		createBasicServerCertificateTestCase(providerCtx, loader),
		createServerCertificateWithTagsTestCase(providerCtx, loader),
		createServerCertificateWithChainAndPathTestCase(providerCtx, loader),
		createServerCertificateWithGeneratedNameTestCase(providerCtx, loader),
		createServerCertificateFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		func(iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service], awsConfigStore pluginutils.ServiceConfigStore[*aws.Config]) provider.Resource {
			return serverCertificateResourceWithNameGen(iamServiceFactory, awsConfigStore, func(input *provider.ResourceDeployInput) (string, error) {
				return "generated-server-certificate", nil
			})
		},
		&s.Suite,
	)
}

func createBasicServerCertificateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:server-certificate/test-server-certificate"
	service := iammock.CreateIamServiceMock(
		iammock.WithUploadServerCertificateOutput(&iam.UploadServerCertificateOutput{
			ServerCertificateMetadata: &types.ServerCertificateMetadata{
				Arn:                   aws.String(resourceARN),
				ServerCertificateName: aws.String("test-server-certificate"),
			},
		}),
	)
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"serverCertificateName": core.MappingNodeFromString("test-server-certificate"),
			"certificateBody":       core.MappingNodeFromString("-----BEGIN CERTIFICATE-----..."),
			"privateKey":            core.MappingNodeFromString("-----BEGIN PRIVATE KEY-----..."),
		},
	}
	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create basic server certificate",
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
			ResourceID: "test-server-certificate-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-server-certificate-id",
					ResourceName: "TestServerCertificate",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.serverCertificateName"},
					{FieldPath: "spec.certificateBody"},
					{FieldPath: "spec.privateKey"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(resourceARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"UploadServerCertificate": &iam.UploadServerCertificateInput{
				ServerCertificateName: aws.String("test-server-certificate"),
				CertificateBody:       aws.String("-----BEGIN CERTIFICATE-----..."),
				PrivateKey:            aws.String("-----BEGIN PRIVATE KEY-----..."),
			},
		},
	}
}

func createServerCertificateWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:server-certificate/test-server-certificate-tags"
	service := iammock.CreateIamServiceMock(
		iammock.WithUploadServerCertificateOutput(&iam.UploadServerCertificateOutput{
			ServerCertificateMetadata: &types.ServerCertificateMetadata{
				Arn:                   aws.String(resourceARN),
				ServerCertificateName: aws.String("test-server-certificate-tags"),
			},
		}),
	)
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"serverCertificateName": core.MappingNodeFromString("test-server-certificate-tags"),
			"certificateBody":       core.MappingNodeFromString("-----BEGIN CERTIFICATE-----..."),
			"privateKey":            core.MappingNodeFromString("-----BEGIN PRIVATE KEY-----..."),
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
	}
	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create server certificate with tags",
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
			ResourceID: "test-server-certificate-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-server-certificate-id",
					ResourceName: "TestServerCertificateWithTags",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.serverCertificateName"},
					{FieldPath: "spec.certificateBody"},
					{FieldPath: "spec.privateKey"},
					{FieldPath: "spec.tags"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(resourceARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"UploadServerCertificate": &iam.UploadServerCertificateInput{
				ServerCertificateName: aws.String("test-server-certificate-tags"),
				CertificateBody:       aws.String("-----BEGIN CERTIFICATE-----..."),
				PrivateKey:            aws.String("-----BEGIN PRIVATE KEY-----..."),
				Tags: []types.Tag{
					{Key: aws.String("Environment"), Value: aws.String("Production")},
				},
			},
		},
	}
}

func createServerCertificateWithChainAndPathTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:server-certificate/test-server-certificate-chain"
	service := iammock.CreateIamServiceMock(
		iammock.WithUploadServerCertificateOutput(&iam.UploadServerCertificateOutput{
			ServerCertificateMetadata: &types.ServerCertificateMetadata{
				Arn:                   aws.String(resourceARN),
				ServerCertificateName: aws.String("test-server-certificate-chain"),
			},
		}),
	)
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"serverCertificateName": core.MappingNodeFromString("test-server-certificate-chain"),
			"certificateBody":       core.MappingNodeFromString("-----BEGIN CERTIFICATE-----..."),
			"privateKey":            core.MappingNodeFromString("-----BEGIN PRIVATE KEY-----..."),
			"certificateChain":      core.MappingNodeFromString("-----BEGIN CERTIFICATE-----..."),
			"path":                  core.MappingNodeFromString("/cloudfront/test/"),
		},
	}
	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create server certificate with chain and path",
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
			ResourceID: "test-server-certificate-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-server-certificate-id",
					ResourceName: "TestServerCertificateWithChain",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.serverCertificateName"},
					{FieldPath: "spec.certificateBody"},
					{FieldPath: "spec.privateKey"},
					{FieldPath: "spec.certificateChain"},
					{FieldPath: "spec.path"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(resourceARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"UploadServerCertificate": &iam.UploadServerCertificateInput{
				ServerCertificateName: aws.String("test-server-certificate-chain"),
				CertificateBody:       aws.String("-----BEGIN CERTIFICATE-----..."),
				PrivateKey:            aws.String("-----BEGIN PRIVATE KEY-----..."),
				CertificateChain:      aws.String("-----BEGIN CERTIFICATE-----..."),
				Path:                  aws.String("/cloudfront/test/"),
			},
		},
	}
}

func createServerCertificateWithGeneratedNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:server-certificate/generated-server-certificate"
	generatedName := "generated-server-certificate"
	service := iammock.CreateIamServiceMock(
		iammock.WithUploadServerCertificateOutput(&iam.UploadServerCertificateOutput{
			ServerCertificateMetadata: &types.ServerCertificateMetadata{
				Arn:                   aws.String(resourceARN),
				ServerCertificateName: aws.String(generatedName),
			},
		}),
	)
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"certificateBody": core.MappingNodeFromString("-----BEGIN CERTIFICATE-----..."),
			"privateKey":      core.MappingNodeFromString("-----BEGIN PRIVATE KEY-----..."),
		},
	}
	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create server certificate with generated name",
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
			ResourceID: "test-server-certificate-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-server-certificate-id",
					ResourceName: "TestServerCertificateGeneratedName",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.certificateBody"},
					{FieldPath: "spec.privateKey"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(resourceARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"UploadServerCertificate": &iam.UploadServerCertificateInput{
				ServerCertificateName: aws.String(generatedName),
				CertificateBody:       aws.String("-----BEGIN CERTIFICATE-----..."),
				PrivateKey:            aws.String("-----BEGIN PRIVATE KEY-----..."),
			},
		},
	}
}

func createServerCertificateFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	serviceError := fmt.Errorf("failed to create server certificate")
	service := iammock.CreateIamServiceMock(
		iammock.WithUploadServerCertificateError(serviceError),
	)
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"serverCertificateName": core.MappingNodeFromString("fail-server-certificate"),
			"certificateBody":       core.MappingNodeFromString("-----BEGIN CERTIFICATE-----..."),
			"privateKey":            core.MappingNodeFromString("-----BEGIN PRIVATE KEY-----..."),
		},
	}
	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "create server certificate failure",
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
			ResourceID: "test-server-certificate-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-server-certificate-id",
					ResourceName: "TestServerCertificateFailure",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.serverCertificateName"},
					{FieldPath: "spec.certificateBody"},
					{FieldPath: "spec.privateKey"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"UploadServerCertificate": &iam.UploadServerCertificateInput{
				ServerCertificateName: aws.String("fail-server-certificate"),
				CertificateBody:       aws.String("-----BEGIN CERTIFICATE-----..."),
				PrivateKey:            aws.String("-----BEGIN PRIVATE KEY-----..."),
			},
		},
	}
}
