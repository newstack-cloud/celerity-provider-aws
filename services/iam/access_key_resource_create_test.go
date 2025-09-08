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

type IAMAccessKeyResourceCreateSuite struct {
	suite.Suite
}

func (s *IAMAccessKeyResourceCreateSuite) Test_create_iam_access_key() {
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
		createBasicAccessKeyTestCase(providerCtx, loader),
		createAccessKeyWithStatusTestCase(providerCtx, loader),
		createAccessKeyMissingUserNameTestCase(providerCtx, loader),
		createAccessKeyServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		AccessKeyResource,
		&s.Suite,
	)
}

func createBasicAccessKeyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateAccessKeyOutput(&iam.CreateAccessKeyOutput{
			AccessKey: &types.AccessKey{
				AccessKeyId:     aws.String("AKIAIOSFODNN7EXAMPLE"),
				SecretAccessKey: aws.String("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
				Status:          types.StatusTypeActive,
				UserName:        aws.String("john.doe"),
			},
		}),
		iammock.WithUpdateAccessKeyOutput(&iam.UpdateAccessKeyOutput{}),
	)

	// Create test data for access key creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("john.doe"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create basic IAM access key",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/accessKey",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.id":              core.MappingNodeFromString("AKIAIOSFODNN7EXAMPLE"),
				"spec.secretAccessKey": core.MappingNodeFromString("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateAccessKey": &iam.CreateAccessKeyInput{
				UserName: aws.String("john.doe"),
			},
		},
	}
}

func createAccessKeyWithStatusTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateAccessKeyOutput(&iam.CreateAccessKeyOutput{
			AccessKey: &types.AccessKey{
				AccessKeyId:     aws.String("AKIAIOSFODNN7EXAMPLE"),
				SecretAccessKey: aws.String("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
				Status:          types.StatusTypeActive,
				UserName:        aws.String("service-account"),
			},
		}),
		iammock.WithUpdateAccessKeyOutput(&iam.UpdateAccessKeyOutput{}),
	)

	// Create test data for access key creation with status
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("service-account"),
			"status":   core.MappingNodeFromString("Inactive"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM access key with Inactive status",
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
					ResourceName: "TestAccessKeyWithStatus",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/accessKey",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
					{
						FieldPath: "spec.status",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.id":              core.MappingNodeFromString("AKIAIOSFODNN7EXAMPLE"),
				"spec.secretAccessKey": core.MappingNodeFromString("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateAccessKey": &iam.CreateAccessKeyInput{
				UserName: aws.String("service-account"),
			},
			"UpdateAccessKey": &iam.UpdateAccessKeyInput{
				AccessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE"),
				Status:      types.StatusTypeInactive,
			},
		},
	}
}

func createAccessKeyMissingUserNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	// Create test data without userName
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"status": core.MappingNodeFromString("Active"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM access key with missing userName",
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
					ResourceName: "TestAccessKeyMissingUserName",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/accessKey",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.status",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
	}
}

func createAccessKeyServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateAccessKeyError(fmt.Errorf("service error")),
	)

	// Create test data for access key creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"userName": core.MappingNodeFromString("john.doe"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM access key with service error",
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
					ResourceName: "TestAccessKeyServiceError",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/accessKey",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.userName",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
	}
}

func TestIAMAccessKeyResourceCreate(t *testing.T) {
	suite.Run(t, new(IAMAccessKeyResourceCreateSuite))
}
