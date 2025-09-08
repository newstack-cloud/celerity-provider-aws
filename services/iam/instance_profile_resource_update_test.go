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

type IAMInstanceProfileResourceUpdateSuite struct {
	suite.Suite
}

func (s *IAMInstanceProfileResourceUpdateSuite) Test_update_iam_instance_profile() {
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
		updateInstanceProfileRoleTestCase(providerCtx, loader),
		updateInstanceProfileNoChangesTestCase(providerCtx, loader),
		updateInstanceProfileServiceErrorTestCase(providerCtx, loader),
		recreateInstanceProfileOnNameOrPathChangeTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		InstanceProfileResource,
		&s.Suite,
	)
}

func updateInstanceProfileRoleTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithRemoveRoleFromInstanceProfileOutput(&iam.RemoveRoleFromInstanceProfileOutput{}),
		iammock.WithAddRoleToInstanceProfileOutput(&iam.AddRoleToInstanceProfileOutput{}),
	)

	// Create test data for instance profile update
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
			"path":                core.MappingNodeFromString("/"),
			"role":                core.MappingNodeFromString("NewRole"),
		},
	}

	// Create current state data
	currentState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
			"path":                core.MappingNodeFromString("/"),
			"role":                core.MappingNodeFromString("OldRole"),
			"arn":                 core.MappingNodeFromString("arn:aws:iam::123456789012:instance-profile/MyInstanceProfile"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM instance profile role",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-instance-profile-id",
						Name:       "TestInstanceProfile",
						InstanceID: "test-instance-id",
						SpecData:   currentState,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/instanceProfile",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.role",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{},
		SaveActionsCalled: map[string]any{
			"RemoveRoleFromInstanceProfile": &iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				RoleName:            aws.String("OldRole"),
			},
			"AddRoleToInstanceProfile": &iam.AddRoleToInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				RoleName:            aws.String("NewRole"),
			},
		},
	}
}

func updateInstanceProfileNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	// Create test data for instance profile with no changes
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
			"path":                core.MappingNodeFromString("/"),
			"role":                core.MappingNodeFromString("MyRole"),
		},
	}

	// Create current state data (same as spec)
	currentState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
			"path":                core.MappingNodeFromString("/"),
			"role":                core.MappingNodeFromString("MyRole"),
			"arn":                 core.MappingNodeFromString("arn:aws:iam::123456789012:instance-profile/MyInstanceProfile"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM instance profile with no changes",
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
					ResourceName: "TestInstanceProfileNoChanges",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-instance-profile-id",
						Name:       "TestInstanceProfileNoChanges",
						InstanceID: "test-instance-id",
						SpecData:   currentState,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/instanceProfile",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectError:       true,
		SaveActionsCalled: map[string]any{},
	}
}

func updateInstanceProfileServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithRemoveRoleFromInstanceProfileError(fmt.Errorf("failed to remove role from instance profile")),
	)

	// Create test data for instance profile update
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
			"path":                core.MappingNodeFromString("/"),
			"role":                core.MappingNodeFromString("NewRole"),
		},
	}

	// Create current state data
	currentState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
			"path":                core.MappingNodeFromString("/"),
			"role":                core.MappingNodeFromString("OldRole"),
			"arn":                 core.MappingNodeFromString("arn:aws:iam::123456789012:instance-profile/MyInstanceProfile"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM instance profile with service error",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-instance-profile-id",
						Name:       "TestInstanceProfileError",
						InstanceID: "test-instance-id",
						SpecData:   currentState,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/instanceProfile",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.role",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"RemoveRoleFromInstanceProfile": &iam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: aws.String("MyInstanceProfile"),
				RoleName:            aws.String("OldRole"),
			},
		},
	}
}

func recreateInstanceProfileOnNameOrPathChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:instance-profile/OldProfile"
	instanceProfileId := "AIPA1234567890123456"

	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteInstanceProfileOutput(&iam.DeleteInstanceProfileOutput{}),
		iammock.WithCreateInstanceProfileOutput(&iam.CreateInstanceProfileOutput{
			InstanceProfile: &types.InstanceProfile{
				Arn:                 aws.String(resourceARN),
				InstanceProfileId:   aws.String(instanceProfileId),
				InstanceProfileName: aws.String("NewProfile"),
				Path:                aws.String("/newpath/"),
			},
		}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("OldProfile"),
			"path":                core.MappingNodeFromString("/oldpath/"),
			"role":                core.MappingNodeFromString("MyRole"),
			"arn":                 core.MappingNodeFromString(resourceARN),
		},
	}
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"instanceProfileName": core.MappingNodeFromString("NewProfile"),
			"path":                core.MappingNodeFromString("/newpath/"),
			"role":                core.MappingNodeFromString("MyRole"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "recreate instance profile on instanceProfileName or path change",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-instance-profile-id",
						Name:       "TestInstanceProfile",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/instanceProfile",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.instanceProfileName",
						PrevValue: core.MappingNodeFromString("OldProfile"),
						NewValue:  core.MappingNodeFromString("NewProfile"),
					},
					{
						FieldPath: "spec.path",
						PrevValue: core.MappingNodeFromString("/oldpath/"),
						NewValue:  core.MappingNodeFromString("/newpath/"),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(resourceARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"DeleteInstanceProfile": &iam.DeleteInstanceProfileInput{
				InstanceProfileName: aws.String("OldProfile"),
			},
			"CreateInstanceProfile": &iam.CreateInstanceProfileInput{
				InstanceProfileName: aws.String("NewProfile"),
				Path:                aws.String("/newpath/"),
			},
		},
	}
}

func TestIAMInstanceProfileResourceUpdate(t *testing.T) {
	suite.Run(t, new(IAMInstanceProfileResourceUpdateSuite))
}
