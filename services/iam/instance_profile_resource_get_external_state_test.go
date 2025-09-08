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

type IAMInstanceProfileResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *IAMInstanceProfileResourceGetExternalStateSuite) Test_get_external_state_iam_instance_profile() {
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
		getExternalStateInstanceProfileTestCase(providerCtx, loader),
		getExternalStateInstanceProfileNotFoundTestCase(providerCtx, loader),
		getExternalStateInstanceProfileServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		InstanceProfileResource,
		&s.Suite,
	)
}

func getExternalStateInstanceProfileTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithGetInstanceProfileOutput(&iam.GetInstanceProfileOutput{
			InstanceProfile: &types.InstanceProfile{
				Arn:                 aws.String("arn:aws:iam::123456789012:instance-profile/MyInstanceProfile"),
				InstanceProfileName: aws.String("MyInstanceProfile"),
				Path:                aws.String("/"),
				Roles: []types.Role{
					{
						Arn:      aws.String("arn:aws:iam::123456789012:role/MyRole"),
						RoleName: aws.String("MyRole"),
					},
				},
			},
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM instance profile",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
					"path":                core.MappingNodeFromString("/"),
					"role":                core.MappingNodeFromString("MyRole"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
					"path":                core.MappingNodeFromString("/"),
					"role":                core.MappingNodeFromString("MyRole"),
					"arn":                 core.MappingNodeFromString("arn:aws:iam::123456789012:instance-profile/MyInstanceProfile"),
				},
			},
		},
	}
}

func getExternalStateInstanceProfileNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithGetInstanceProfileError(&types.NoSuchEntityException{
			Message: aws.String("The instance profile MyInstanceProfile cannot be found"),
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM instance profile not found",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
					"path":                core.MappingNodeFromString("/"),
					"role":                core.MappingNodeFromString("MyRole"),
				},
			},
		},
		ExpectError: true,
	}
}

func getExternalStateInstanceProfileServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithGetInstanceProfileError(fmt.Errorf("service error")),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM instance profile with service error",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
					"path":                core.MappingNodeFromString("/"),
					"role":                core.MappingNodeFromString("MyRole"),
				},
			},
		},
		ExpectError: true,
	}
}

func TestIAMInstanceProfileResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(IAMInstanceProfileResourceGetExternalStateSuite))
}
