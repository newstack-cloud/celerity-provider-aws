//go:build unit

package iam

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

type IamRoleResourceStabilisedSuite struct {
	suite.Suite
}

func (s *IamRoleResourceStabilisedSuite) Test_stabilised_iam_role() {
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

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		stabilisedBasicRoleTestCase(providerCtx, loader),
		stabilisedCompleteRoleTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	roleResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return RoleResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		roleResourceWrapper,
		&s.Suite,
	)
}

func stabilisedBasicRoleTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	// Create test data for role stabilisation check
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
			"description": core.MappingNodeFromString("Test role for Lambda execution"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name: "IAM role is always stabilised",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID:      "test-instance-id",
			ResourceID:      "TestRole",
			ResourceSpec:    specData,
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func stabilisedCompleteRoleTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	// Create test data for complete role stabilisation check
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                core.MappingNodeFromString("arn:aws:iam::123456789012:role/test/CompleteTestRole"),
			"description":        core.MappingNodeFromString("A complete test role"),
			"maxSessionDuration": core.MappingNodeFromInt(7200),
			"path":               core.MappingNodeFromString("/test/"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				},
			},
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("Test"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, iamservice.Service]{
		Name: "Complete IAM role is always stabilised",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID:      "test-instance-id",
			ResourceID:      "CompleteTestRole",
			ResourceSpec:    specData,
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func TestIamRoleResourceStabilised(t *testing.T) {
	suite.Run(t, new(IamRoleResourceStabilisedSuite))
}
