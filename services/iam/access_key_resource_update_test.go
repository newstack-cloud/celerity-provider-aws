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

type IAMAccessKeyResourceUpdateSuite struct {
	suite.Suite
}

func (s *IAMAccessKeyResourceUpdateSuite) Test_update_iam_access_key() {
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
		updateAccessKeyStatusTestCase(providerCtx, loader),
		updateAccessKeyNoChangesTestCase(providerCtx, loader),
		updateAccessKeyServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		AccessKeyResource,
		&s.Suite,
	)
}

func updateAccessKeyStatusTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateAccessKeyOutput(&iam.UpdateAccessKeyOutput{}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"id":       core.MappingNodeFromString("AKIAIOSFODNN7EXAMPLE"),
			"userName": core.MappingNodeFromString("john.doe"),
			"status":   core.MappingNodeFromString("Active"),
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("john.doe"),
			"status":   core.MappingNodeFromString("Inactive"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM access key status",
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
			ResourceID: "test-access-key-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-access-key-id",
					ResourceName: "TestAccessKey",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-access-key-id",
						Name:       "TestAccessKey",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/accessKey",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.status",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{},
		SaveActionsCalled: map[string]any{
			"UpdateAccessKey": &iam.UpdateAccessKeyInput{
				AccessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE"),
				Status:      types.StatusTypeInactive,
			},
		},
	}
}

func updateAccessKeyNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	// Current state (same as updated state)
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"id":       core.MappingNodeFromString("AKIAIOSFODNN7EXAMPLE"),
			"userName": core.MappingNodeFromString("john.doe"),
			"status":   core.MappingNodeFromString("Active"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("john.doe"),
			"status":   core.MappingNodeFromString("Active"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM access key with no changes",
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
			ResourceID: "test-access-key-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-access-key-id",
					ResourceName: "TestAccessKey",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-access-key-id",
						Name:       "TestAccessKey",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/accessKey",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput:    &provider.ResourceDeployOutput{},
		SaveActionsCalled: map[string]any{},
	}
}

func updateAccessKeyServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateAccessKeyError(fmt.Errorf("service error")),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"id":       core.MappingNodeFromString("AKIAIOSFODNN7EXAMPLE"),
			"userName": core.MappingNodeFromString("john.doe"),
			"status":   core.MappingNodeFromString("Active"),
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("john.doe"),
			"status":   core.MappingNodeFromString("Inactive"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM access key with service error",
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
			ResourceID: "test-access-key-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-access-key-id",
					ResourceName: "TestAccessKey",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-access-key-id",
						Name:       "TestAccessKey",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/accessKey",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.status",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"UpdateAccessKey": &iam.UpdateAccessKeyInput{
				AccessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE"),
				Status:      types.StatusTypeInactive,
			},
		},
	}
}

func TestIAMAccessKeyResourceUpdate(t *testing.T) {
	suite.Run(t, new(IAMAccessKeyResourceUpdateSuite))
}
