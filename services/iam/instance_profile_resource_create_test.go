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

type IAMInstanceProfileResourceCreateSuite struct {
	suite.Suite
}

func (s *IAMInstanceProfileResourceCreateSuite) Test_create_iam_instance_profile() {
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
		createBasicInstanceProfileTestCase(providerCtx, loader),
		createInstanceProfileWithPathTestCase(providerCtx, loader),
		createInstanceProfileWithGeneratedNameTestCase(providerCtx, loader),
		createInstanceProfileServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		InstanceProfileResource,
		&s.Suite,
	)
}

func createBasicInstanceProfileTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateInstanceProfileOutput(&iam.CreateInstanceProfileOutput{
			InstanceProfile: &types.InstanceProfile{
				Arn:                 aws.String("arn:aws:iam::123456789012:instance-profile/MyInstanceProfile"),
				InstanceProfileName: aws.String("MyInstanceProfile"),
				Path:                aws.String("/"),
			},
		}),
		iammock.WithAddRoleToInstanceProfileOutput(&iam.AddRoleToInstanceProfileOutput{}),
	)

	// Create test data for instance profile creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
			"path":                core.MappingNodeFromString("/"),
			"role":                core.MappingNodeFromString("MyRole"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create basic IAM instance profile",
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
			ResourceID: "test-instance-profile-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-instance-profile-id",
					ResourceName: "TestInstanceProfile",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/instanceProfile",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.instanceProfileName",
					},
					{
						FieldPath: "spec.path",
					},
					{
						FieldPath: "spec.role",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString("arn:aws:iam::123456789012:instance-profile/MyInstanceProfile"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateInstanceProfile": &iam.CreateInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				Path:                aws.String("/"),
			},
			"AddRoleToInstanceProfile": &iam.AddRoleToInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				RoleName:            aws.String("MyRole"),
			},
		},
	}
}

func createInstanceProfileWithPathTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateInstanceProfileOutput(&iam.CreateInstanceProfileOutput{
			InstanceProfile: &types.InstanceProfile{
				Arn:                 aws.String("arn:aws:iam::123456789012:instance-profile/application/MyInstanceProfile"),
				InstanceProfileName: aws.String("MyInstanceProfile"),
				Path:                aws.String("/application/"),
			},
		}),
		iammock.WithAddRoleToInstanceProfileOutput(&iam.AddRoleToInstanceProfileOutput{}),
	)

	// Create test data for instance profile creation with custom path
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
			"path":                core.MappingNodeFromString("/application/"),
			"role":                core.MappingNodeFromString("MyRole"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM instance profile with custom path",
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
			ResourceID: "test-instance-profile-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-instance-profile-id",
					ResourceName: "TestInstanceProfileWithPath",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/instanceProfile",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.instanceProfileName",
					},
					{
						FieldPath: "spec.path",
					},
					{
						FieldPath: "spec.role",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString("arn:aws:iam::123456789012:instance-profile/application/MyInstanceProfile"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateInstanceProfile": &iam.CreateInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				Path:                aws.String("/application/"),
			},
			"AddRoleToInstanceProfile": &iam.AddRoleToInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				RoleName:            aws.String("MyRole"),
			},
		},
	}
}

func createInstanceProfileWithGeneratedNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateInstanceProfileOutput(&iam.CreateInstanceProfileOutput{
			InstanceProfile: &types.InstanceProfile{
				Arn:                 aws.String("arn:aws:iam::123456789012:instance-profile/generated-instance-profile-123"),
				InstanceProfileName: aws.String("generated-instance-profile-123"),
				Path:                aws.String("/"),
			},
		}),
		iammock.WithAddRoleToInstanceProfileOutput(&iam.AddRoleToInstanceProfileOutput{}),
	)

	// Create test data for instance profile creation with generated name
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString("/"),
			"role": core.MappingNodeFromString("MyRole"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM instance profile with generated name",
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
			ResourceID: "test-instance-profile-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-instance-profile-id",
					ResourceName: "TestInstanceProfileGenerated",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/instanceProfile",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.path",
					},
					{
						FieldPath: "spec.role",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString("arn:aws:iam::123456789012:instance-profile/generated-instance-profile-123"),
			},
		},
		// Note: We can't predict the exact instance profile name due to nanoid generation,
		// so we'll omit SaveActionsCalled for this test case
		// The important thing is that the instance profile gets created successfully
	}
}

func createInstanceProfileServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateInstanceProfileError(fmt.Errorf("failed to create instance profile")),
	)

	// Create test data for instance profile creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
			"path":                core.MappingNodeFromString("/"),
			"role":                core.MappingNodeFromString("MyRole"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM instance profile with service error",
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
			ResourceID: "test-instance-profile-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-instance-profile-id",
					ResourceName: "TestInstanceProfileError",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/instanceProfile",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.instanceProfileName",
					},
					{
						FieldPath: "spec.path",
					},
					{
						FieldPath: "spec.role",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"CreateInstanceProfile": &iam.CreateInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				Path:                aws.String("/"),
			},
		},
	}
}

func TestIAMInstanceProfileResourceCreate(t *testing.T) {
	suite.Run(t, new(IAMInstanceProfileResourceCreateSuite))
}
