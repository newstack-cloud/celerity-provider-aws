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
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type ServerCertificateResourceUpdateSuite struct {
	suite.Suite
}

func (s *ServerCertificateResourceUpdateSuite) Test_update_iam_server_certificate() {
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
		updateServerCertificateMetadataTestCase(providerCtx, loader),
		updateServerCertificateTagsTestCase(providerCtx, loader),
		recreateServerCertificateOnPrivateKeyChangeTestCase(providerCtx, loader),
		updateServerCertificateNoChangesTestCase(providerCtx, loader),
		updateServerCertificateServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		ServerCertificateResource,
		&s.Suite,
	)
}

func updateServerCertificateMetadataTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oldServerCertificateName := "old-server-certificate-name"
	newServerCertificateName := "new-server-certificate-name"
	newServerCertificateArn := "arn:aws:iam::123456789012:server-certificate/new-server-certificate-name"

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateServerCertificateOutput(&iam.UpdateServerCertificateOutput{}),
		iammock.WithGetServerCertificateOutput(&iam.GetServerCertificateOutput{
			ServerCertificate: &types.ServerCertificate{
				ServerCertificateMetadata: &types.ServerCertificateMetadata{
					ServerCertificateName: aws.String(oldServerCertificateName),
					Arn:                   aws.String(newServerCertificateArn),
				},
			},
		}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                   core.MappingNodeFromString(newServerCertificateArn),
			"serverCertificateName": core.MappingNodeFromString(oldServerCertificateName),
			"path":                  core.MappingNodeFromString("old-path"),
			"privateKey":            core.MappingNodeFromString("private-key"),
			"certificateBody":       core.MappingNodeFromString("certificate-body"),
			"certificateChain":      core.MappingNodeFromString("certificate-chain"),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{},
					},
				},
			},
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"serverCertificateName": core.MappingNodeFromString(newServerCertificateName),
			"path":                  core.MappingNodeFromString("new-path"),
			"privateKey":            core.MappingNodeFromString("private-key"),
			"certificateBody":       core.MappingNodeFromString("certificate-body"),
			"certificateChain":      core.MappingNodeFromString("certificate-chain"),
			"tags": {
				Items: []*core.MappingNode{},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM server certificate metadata",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-server-certificate-id",
						Name:       "TestServerCertificate",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.serverCertificateName",
						PrevValue: core.MappingNodeFromString(oldServerCertificateName),
						NewValue:  core.MappingNodeFromString(newServerCertificateName),
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(newServerCertificateArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"UpdateServerCertificate": &iam.UpdateServerCertificateInput{
				ServerCertificateName:    aws.String(oldServerCertificateName),
				NewServerCertificateName: aws.String(newServerCertificateName),
			},
			"GetServerCertificate": &iam.GetServerCertificateInput{
				ServerCertificateName: aws.String(newServerCertificateName),
			},
		},
	}
}

func updateServerCertificateTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	serverCertificateName := "test-server-certificate"
	serverCertificateArn := "arn:aws:iam::123456789012:server-certificate/test-server-certificate"

	service := iammock.CreateIamServiceMock(
		iammock.WithTagServerCertificateOutput(&iam.TagServerCertificateOutput{}),
		iammock.WithUntagServerCertificateOutput(&iam.UntagServerCertificateOutput{}),
		iammock.WithGetServerCertificateOutput(&iam.GetServerCertificateOutput{
			ServerCertificate: &types.ServerCertificate{
				ServerCertificateMetadata: &types.ServerCertificateMetadata{
					ServerCertificateName: aws.String(serverCertificateName),
					Arn:                   aws.String(serverCertificateArn),
				},
			},
		}),
	)

	// Current state with tags
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                   core.MappingNodeFromString(serverCertificateArn),
			"serverCertificateName": core.MappingNodeFromString(serverCertificateName),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("Development"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("OldTag"),
							"value": core.MappingNodeFromString("OldValue"),
						},
					},
				},
			},
		},
	}

	// Updated state with different tags
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"serverCertificateName": core.MappingNodeFromString(serverCertificateName),
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
							"key":   core.MappingNodeFromString("NewTag"),
							"value": core.MappingNodeFromString("NewValue"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM server certificate tags",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-server-certificate-id",
						Name:       "TestServerCertificate",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{{FieldPath: "spec.tags"}},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(serverCertificateArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"UntagServerCertificate": &iam.UntagServerCertificateInput{
				ServerCertificateName: aws.String(serverCertificateName),
				TagKeys:               []string{"OldTag"},
			},
			"TagServerCertificate": &iam.TagServerCertificateInput{
				ServerCertificateName: aws.String(serverCertificateName),
				Tags: []types.Tag{
					{Key: aws.String("Environment"), Value: aws.String("Production")},
					{Key: aws.String("NewTag"), Value: aws.String("NewValue")},
				},
			},
			"GetServerCertificate": &iam.GetServerCertificateInput{
				ServerCertificateName: aws.String(serverCertificateName),
			},
		},
	}
}

func recreateServerCertificateOnPrivateKeyChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	serverCertificateName := "old-server-certificate"
	serverCertificateArn := "arn:aws:iam::123456789012:server-certificate/old-server-certificate"
	newPrivateKey := "new-private-key"

	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteServerCertificateOutput(&iam.DeleteServerCertificateOutput{}),
		iammock.WithUploadServerCertificateOutput(&iam.UploadServerCertificateOutput{
			ServerCertificateMetadata: &types.ServerCertificateMetadata{
				ServerCertificateName: aws.String(serverCertificateName),
				Arn:                   aws.String(serverCertificateArn),
			},
		}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                   core.MappingNodeFromString(serverCertificateArn),
			"serverCertificateName": core.MappingNodeFromString(serverCertificateName),
			"privateKey":            core.MappingNodeFromString("old-private-key"),
			"certificateBody":       core.MappingNodeFromString("certificate-body"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"serverCertificateName": core.MappingNodeFromString(serverCertificateName),
			"privateKey":            core.MappingNodeFromString(newPrivateKey),
			"certificateBody":       core.MappingNodeFromString("certificate-body"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Recreate IAM server certificate on private key change",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-server-certificate-id",
						Name:       "TestServerCertificate",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.privateKey",
						PrevValue: core.MappingNodeFromString("old-private-key"),
						NewValue:  core.MappingNodeFromString(newPrivateKey),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(serverCertificateArn),
			},
		},
		SaveActionsCalled: map[string]any{
			// The old server certificate needs to be deleted before the new one is uploaded
			// as the same certificate name will be used for the new one.
			"DeleteServerCertificate": &iam.DeleteServerCertificateInput{
				ServerCertificateName: aws.String(serverCertificateName),
			},
			"UploadServerCertificate": &iam.UploadServerCertificateInput{
				ServerCertificateName: aws.String(serverCertificateName),
				PrivateKey:            aws.String(newPrivateKey),
				CertificateBody:       aws.String("certificate-body"),
			},
		},
	}
}

func updateServerCertificateNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	serverCertificateName := "test-server-certificate"
	serverCertificateArn := "arn:aws:iam::123456789012:server-certificate/test-server-certificate"

	service := iammock.CreateIamServiceMock()

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                   core.MappingNodeFromString(serverCertificateArn),
			"serverCertificateName": core.MappingNodeFromString(serverCertificateName),
			"privateKey":            core.MappingNodeFromString("private-key"),
			"certificateBody":       core.MappingNodeFromString("certificate-body"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"serverCertificateName": core.MappingNodeFromString(serverCertificateName),
			"privateKey":            core.MappingNodeFromString("private-key"),
			"certificateBody":       core.MappingNodeFromString("certificate-body"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM server certificate with no changes",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-server-certificate-id",
						Name:       "TestServerCertificate",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(serverCertificateArn),
			},
		},
	}
}

func updateServerCertificateServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	serverCertificateName := "test-server-certificate"
	newServerCertificateName := "new-server-certificate"
	serverCertificateArn := "arn:aws:iam::123456789012:server-certificate/test-server-certificate"
	serviceError := fmt.Errorf("AWS service error")

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateServerCertificateError(serviceError),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                   core.MappingNodeFromString(serverCertificateArn),
			"serverCertificateName": core.MappingNodeFromString(serverCertificateName),
			"privateKey":            core.MappingNodeFromString("private-key"),
			"certificateBody":       core.MappingNodeFromString("certificate-body"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			// Add a new certificate to ensure the update call the AWS IAM service is attempted.
			"serverCertificateName": core.MappingNodeFromString(newServerCertificateName),
			"privateKey":            core.MappingNodeFromString("private-key"),
			"certificateBody":       core.MappingNodeFromString("certificate-body"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM server certificate with service error",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-server-certificate-id",
						Name:       "TestServerCertificate",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/serverCertificate",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.serverCertificateName",
						PrevValue: core.MappingNodeFromString(serverCertificateName),
						NewValue:  core.MappingNodeFromString(newServerCertificateName),
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"UpdateServerCertificate": &iam.UpdateServerCertificateInput{
				ServerCertificateName: aws.String(serverCertificateName),
			},
		},
	}
}

// func updateSAMLProviderTagsTestCase(
// 	providerCtx provider.Context,
// 	loader *testutils.MockAWSConfigLoader,
// ) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
// 	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"
// 	metadata := `<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`
// 	samlProviderUUID := "96dc2683-50a4-4f46-8c0f-4dedf83a8ead"

// 	service := iammock.CreateIamServiceMock(
// 		iammock.WithTagSAMLProviderOutput(&iam.TagSAMLProviderOutput{}),
// 		iammock.WithUntagSAMLProviderOutput(&iam.UntagSAMLProviderOutput{}),
// 		iammock.WithGetSAMLProviderOutput(&iam.GetSAMLProviderOutput{
// 			SAMLProviderUUID: aws.String(samlProviderUUID),
// 		}),
// 	)

// 	// Current state with tags
// 	currentStateSpecData := &core.MappingNode{
// 		Fields: map[string]*core.MappingNode{
// 			"arn":                  core.MappingNodeFromString(samlProviderArn),
// 			"name":                 core.MappingNodeFromString("MySAMLProvider"),
// 			"samlMetadataDocument": core.MappingNodeFromString(metadata),
// 			"tags": {
// 				Items: []*core.MappingNode{
// 					{
// 						Fields: map[string]*core.MappingNode{
// 							"key":   core.MappingNodeFromString("Environment"),
// 							"value": core.MappingNodeFromString("Development"),
// 						},
// 					},
// 					{
// 						Fields: map[string]*core.MappingNode{
// 							"key":   core.MappingNodeFromString("OldTag"),
// 							"value": core.MappingNodeFromString("OldValue"),
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	// Updated state with different tags
// 	updatedSpecData := &core.MappingNode{
// 		Fields: map[string]*core.MappingNode{
// 			"name":                 core.MappingNodeFromString("MySAMLProvider"),
// 			"samlMetadataDocument": core.MappingNodeFromString(metadata),
// 			"tags": {
// 				Items: []*core.MappingNode{
// 					{
// 						Fields: map[string]*core.MappingNode{
// 							"key":   core.MappingNodeFromString("Environment"),
// 							"value": core.MappingNodeFromString("Production"),
// 						},
// 					},
// 					{
// 						Fields: map[string]*core.MappingNode{
// 							"key":   core.MappingNodeFromString("NewTag"),
// 							"value": core.MappingNodeFromString("NewValue"),
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
// 		Name: "Update IAM SAML provider tags",
// 		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
// 			return service
// 		},
// 		ServiceMockCalls: &service.MockCalls,
// 		ConfigStore: utils.NewAWSConfigStore(
// 			[]string{},
// 			utils.AWSConfigFromProviderContext,
// 			loader,
// 			utils.AWSConfigCacheKey,
// 		),
// 		Input: &provider.ResourceDeployInput{
// 			InstanceID: "test-instance-id",
// 			ResourceID: "test-saml-provider-id",
// 			Changes: &provider.Changes{
// 				AppliedResourceInfo: provider.ResourceInfo{
// 					ResourceID:   "test-saml-provider-id",
// 					ResourceName: "TestSAMLProvider",
// 					InstanceID:   "test-instance-id",
// 					CurrentResourceState: &state.ResourceState{
// 						ResourceID: "test-saml-provider-id",
// 						Name:       "TestSAMLProvider",
// 						InstanceID: "test-instance-id",
// 						SpecData:   currentStateSpecData,
// 					},
// 					ResourceWithResolvedSubs: &provider.ResolvedResource{
// 						Type: &schema.ResourceTypeWrapper{
// 							Value: "aws/iam/samlProvider",
// 						},
// 						Spec: updatedSpecData,
// 					},
// 				},
// 				ModifiedFields: []provider.FieldChange{
// 					{
// 						FieldPath: "spec.tags",
// 					},
// 				},
// 			},
// 			ProviderContext: providerCtx,
// 		},
// 		ExpectedOutput: &provider.ResourceDeployOutput{
// 			ComputedFieldValues: map[string]*core.MappingNode{
// 				"spec.arn":              core.MappingNodeFromString(samlProviderArn),
// 				"spec.samlProviderUUID": core.MappingNodeFromString(samlProviderUUID),
// 			},
// 		},
// 		SaveActionsCalled: map[string]any{
// 			"UntagSAMLProvider": &iam.UntagSAMLProviderInput{
// 				SAMLProviderArn: aws.String(samlProviderArn),
// 				TagKeys:         []string{"OldTag"},
// 			},
// 			"TagSAMLProvider": &iam.TagSAMLProviderInput{
// 				SAMLProviderArn: aws.String(samlProviderArn),
// 				Tags: []types.Tag{
// 					{
// 						Key:   aws.String("Environment"),
// 						Value: aws.String("Production"),
// 					},
// 					{
// 						Key:   aws.String("NewTag"),
// 						Value: aws.String("NewValue"),
// 					},
// 				},
// 			},
// 			"GetSAMLProvider": &iam.GetSAMLProviderInput{
// 				SAMLProviderArn: aws.String(samlProviderArn),
// 			},
// 		},
// 	}
// }

// func updateSAMLProviderNoChangesTestCase(
// 	providerCtx provider.Context,
// 	loader *testutils.MockAWSConfigLoader,
// ) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
// 	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"
// 	metadata := `<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`
// 	samlProviderUUID := "96dc2683-50a4-4f46-8c0f-4dedf83a8ead"

// 	service := iammock.CreateIamServiceMock()

// 	// Current state
// 	currentStateSpecData := &core.MappingNode{
// 		Fields: map[string]*core.MappingNode{
// 			"arn":                  core.MappingNodeFromString(samlProviderArn),
// 			"name":                 core.MappingNodeFromString("MySAMLProvider"),
// 			"samlMetadataDocument": core.MappingNodeFromString(metadata),
// 			"samlProviderUUID":     core.MappingNodeFromString(samlProviderUUID),
// 		},
// 	}

// 	// No changes in updated state
// 	updatedSpecData := &core.MappingNode{
// 		Fields: map[string]*core.MappingNode{
// 			"name":                 core.MappingNodeFromString("MySAMLProvider"),
// 			"samlMetadataDocument": core.MappingNodeFromString(metadata),
// 			"samlProviderUUID":     core.MappingNodeFromString(samlProviderUUID),
// 		},
// 	}

// 	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
// 		Name: "Update IAM SAML provider with no changes",
// 		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
// 			return service
// 		},
// 		ServiceMockCalls: &service.MockCalls,
// 		ConfigStore: utils.NewAWSConfigStore(
// 			[]string{},
// 			utils.AWSConfigFromProviderContext,
// 			loader,
// 			utils.AWSConfigCacheKey,
// 		),
// 		Input: &provider.ResourceDeployInput{
// 			InstanceID: "test-instance-id",
// 			ResourceID: "test-saml-provider-id",
// 			Changes: &provider.Changes{
// 				AppliedResourceInfo: provider.ResourceInfo{
// 					ResourceID:   "test-saml-provider-id",
// 					ResourceName: "TestSAMLProvider",
// 					InstanceID:   "test-instance-id",
// 					CurrentResourceState: &state.ResourceState{
// 						ResourceID: "test-saml-provider-id",
// 						Name:       "TestSAMLProvider",
// 						InstanceID: "test-instance-id",
// 						SpecData:   currentStateSpecData,
// 					},
// 					ResourceWithResolvedSubs: &provider.ResolvedResource{
// 						Type: &schema.ResourceTypeWrapper{
// 							Value: "aws/iam/samlProvider",
// 						},
// 						Spec: updatedSpecData,
// 					},
// 				},
// 				ModifiedFields: []provider.FieldChange{},
// 			},
// 			ProviderContext: providerCtx,
// 		},
// 		ExpectedOutput: &provider.ResourceDeployOutput{
// 			ComputedFieldValues: map[string]*core.MappingNode{
// 				"spec.arn":              core.MappingNodeFromString(samlProviderArn),
// 				"spec.samlProviderUUID": core.MappingNodeFromString(samlProviderUUID),
// 			},
// 		},
// 	}
// }

// func updateSAMLProviderServiceErrorTestCase(
// 	providerCtx provider.Context,
// 	loader *testutils.MockAWSConfigLoader,
// ) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
// 	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"
// 	oldMetadata := `<?xml version="1.0"?><EntityDescriptor>old</EntityDescriptor>`
// 	newMetadata := `<?xml version="1.0"?><EntityDescriptor>new</EntityDescriptor>`

// 	serviceError := fmt.Errorf("AWS service error")

// 	service := iammock.CreateIamServiceMock(
// 		iammock.WithUpdateSAMLProviderError(serviceError),
// 	)

// 	// Current state
// 	currentStateSpecData := &core.MappingNode{
// 		Fields: map[string]*core.MappingNode{
// 			"arn":                  core.MappingNodeFromString(samlProviderArn),
// 			"name":                 core.MappingNodeFromString("MySAMLProvider"),
// 			"samlMetadataDocument": core.MappingNodeFromString(oldMetadata),
// 		},
// 	}

// 	// Updated state
// 	updatedSpecData := &core.MappingNode{
// 		Fields: map[string]*core.MappingNode{
// 			"name":                 core.MappingNodeFromString("MySAMLProvider"),
// 			"samlMetadataDocument": core.MappingNodeFromString(newMetadata),
// 		},
// 	}

// 	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
// 		Name: "Update IAM SAML provider with service error",
// 		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
// 			return service
// 		},
// 		ServiceMockCalls: &service.MockCalls,
// 		ConfigStore: utils.NewAWSConfigStore(
// 			[]string{},
// 			utils.AWSConfigFromProviderContext,
// 			loader,
// 			utils.AWSConfigCacheKey,
// 		),
// 		Input: &provider.ResourceDeployInput{
// 			InstanceID: "test-instance-id",
// 			ResourceID: "test-saml-provider-id",
// 			Changes: &provider.Changes{
// 				AppliedResourceInfo: provider.ResourceInfo{
// 					ResourceID:   "test-saml-provider-id",
// 					ResourceName: "TestSAMLProvider",
// 					InstanceID:   "test-instance-id",
// 					CurrentResourceState: &state.ResourceState{
// 						ResourceID: "test-saml-provider-id",
// 						Name:       "TestSAMLProvider",
// 						InstanceID: "test-instance-id",
// 						SpecData:   currentStateSpecData,
// 					},
// 					ResourceWithResolvedSubs: &provider.ResolvedResource{
// 						Type: &schema.ResourceTypeWrapper{
// 							Value: "aws/iam/samlProvider",
// 						},
// 						Spec: updatedSpecData,
// 					},
// 				},
// 				ModifiedFields: []provider.FieldChange{
// 					{
// 						FieldPath: "spec.samlMetadataDocument",
// 					},
// 				},
// 			},
// 			ProviderContext: providerCtx,
// 		},
// 		ExpectError: true,
// 		SaveActionsCalled: map[string]any{
// 			"UpdateSAMLProvider": &iam.UpdateSAMLProviderInput{
// 				SAMLProviderArn:      aws.String(samlProviderArn),
// 				SAMLMetadataDocument: aws.String(newMetadata),
// 			},
// 		},
// 	}
// }

func TestServerCertificateResourceUpdateSuite(t *testing.T) {
	suite.Run(t, new(ServerCertificateResourceUpdateSuite))
}
