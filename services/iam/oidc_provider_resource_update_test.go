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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type IAMOIDCProviderResourceUpdateSuite struct {
	suite.Suite
}

func (s *IAMOIDCProviderResourceUpdateSuite) Test_update_iam_oidc_provider() {
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
		updateOIDCProviderClientIdsTestCase(providerCtx, loader),
		updateOIDCProviderThumbprintsTestCase(providerCtx, loader),
		updateOIDCProviderTagsTestCase(providerCtx, loader),
		updateOIDCProviderNoChangesTestCase(providerCtx, loader),
		updateOIDCProviderServiceErrorTestCase(providerCtx, loader),
		recreateOIDCProviderOnUrlChangeTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	oidcProviderResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return OIDCProviderResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		oidcProviderResourceWrapper,
		&s.Suite,
	)
}

func updateOIDCProviderClientIdsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithAddClientIDToOpenIDConnectProviderOutput(&iam.AddClientIDToOpenIDConnectProviderOutput{}),
		iammock.WithRemoveClientIDFromOpenIDConnectProviderOutput(&iam.RemoveClientIDFromOpenIDConnectProviderOutput{}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
			"clientIdList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("old-client-id"),
					core.MappingNodeFromString("keep-client-id"),
				},
			},
			"thumbprintList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
			"clientIdList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("keep-client-id"),
					core.MappingNodeFromString("new-client-id"),
				},
			},
			"thumbprintList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider client IDs",
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
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOIDCProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOIDCProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.clientIdList",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(oidcProviderArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"RemoveClientIDFromOpenIDConnectProvider": &iam.RemoveClientIDFromOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: aws.String("arn:aws:iam::123456789012:oidc-provider/example.com"),
				ClientID:                 aws.String("old-client-id"),
			},
			"AddClientIDToOpenIDConnectProvider": &iam.AddClientIDToOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: aws.String("arn:aws:iam::123456789012:oidc-provider/example.com"),
				ClientID:                 aws.String("new-client-id"),
			},
		},
	}
}

func updateOIDCProviderThumbprintsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateOpenIDConnectProviderThumbprintOutput(&iam.UpdateOpenIDConnectProviderThumbprintOutput{}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
			"thumbprintList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
			"thumbprintList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
					core.MappingNodeFromString("9e99a48a9960b14926bb7f3b02e22da2b0ab7280"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider thumbprints",
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
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOIDCProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOIDCProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.thumbprintList",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(oidcProviderArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"UpdateOpenIDConnectProviderThumbprint": &iam.UpdateOpenIDConnectProviderThumbprintInput{
				OpenIDConnectProviderArn: aws.String("arn:aws:iam::123456789012:oidc-provider/example.com"),
				ThumbprintList:           []string{"cf23df2207d99a74fbe169e3eba035e633b65d94", "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"},
			},
		},
	}
}

func updateOIDCProviderTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithTagOpenIDConnectProviderOutput(&iam.TagOpenIDConnectProviderOutput{}),
		iammock.WithUntagOpenIDConnectProviderOutput(&iam.UntagOpenIDConnectProviderOutput{}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
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
							"key":   core.MappingNodeFromString("ToRemove"),
							"value": core.MappingNodeFromString("OldValue"),
						},
					},
				},
			},
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
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
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider tags",
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
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOIDCProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOIDCProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(oidcProviderArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"TagOpenIDConnectProvider": &iam.TagOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: aws.String("arn:aws:iam::123456789012:oidc-provider/example.com"),
				Tags: []types.Tag{
					{Key: aws.String("Environment"), Value: aws.String("Production")},
					{Key: aws.String("Service"), Value: aws.String("Authentication")},
				},
			},
			"UntagOpenIDConnectProvider": &iam.UntagOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: aws.String("arn:aws:iam::123456789012:oidc-provider/example.com"),
				TagKeys:                  []string{"ToRemove"},
			},
		},
	}
}

func updateOIDCProviderNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock()

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
		},
	}

	// Updated state - same as current
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider no changes",
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
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOIDCProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOIDCProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
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
				"spec.arn": core.MappingNodeFromString(oidcProviderArn),
			},
		},
		SaveActionsCalled: map[string]any{},
	}
}

func updateOIDCProviderServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateOpenIDConnectProviderThumbprintError(fmt.Errorf("failed to update thumbprints")),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
			"thumbprintList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
			"thumbprintList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("9e99a48a9960b14926bb7f3b02e22da2b0ab7280"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider service error",
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
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOIDCProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOIDCProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.thumbprintList",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
	}
}

func recreateOIDCProviderOnUrlChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oldArn := "arn:aws:iam::123456789012:oidc-provider/old.example.com"
	newArn := "arn:aws:iam::123456789012:oidc-provider/new.example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteOpenIDConnectProviderOutput(&iam.DeleteOpenIDConnectProviderOutput{}),
		iammock.WithCreateOpenIDConnectProviderOutput(&iam.CreateOpenIDConnectProviderOutput{
			OpenIDConnectProviderArn: aws.String(newArn),
		}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oldArn),
			"url": core.MappingNodeFromString("https://old.example.com"),
			"clientIdList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("sts.amazonaws.com"),
				},
			},
			"thumbprintList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://new.example.com"),
			"clientIdList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("sts.amazonaws.com"),
				},
			},
			"thumbprintList": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "recreate OIDC provider on url change",
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
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOIDCProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOIDCProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.url",
						PrevValue: core.MappingNodeFromString("https://old.example.com"),
						NewValue:  core.MappingNodeFromString("https://new.example.com"),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(newArn),
			},
		},
		SaveActionsCalled: map[string]any{
			"DeleteOpenIDConnectProvider": &iam.DeleteOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: aws.String(oldArn),
			},
			"CreateOpenIDConnectProvider": &iam.CreateOpenIDConnectProviderInput{
				Url:            aws.String("https://new.example.com"),
				ClientIDList:   []string{"sts.amazonaws.com"},
				ThumbprintList: []string{"cf23df2207d99a74fbe169e3eba035e633b65d94"},
				Tags:           []types.Tag{},
			},
		},
	}
}

func TestIAMOIDCProviderResourceUpdate(t *testing.T) {
	suite.Run(t, new(IAMOIDCProviderResourceUpdateSuite))
}
