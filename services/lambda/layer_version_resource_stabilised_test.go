//go:build unit

package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaLayerVersionResourceStabilisedSuite struct {
	suite.Suite
}

func (s *LambdaLayerVersionResourceStabilisedSuite) Test_stabilised() {
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

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		createLayerVersionStabilisedTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		LayerVersionResource,
		&s.Suite,
	)
}

func createLayerVersionStabilisedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service] {
	// No mock service calls needed since layer versions are immediately stable
	service := lambdamock.CreateLambdaServiceMock()

	layerVersionArn := "arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		Name: "layer version is immediately stabilised",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			InstanceID:      "test-instance-id",
			ResourceID:      "test-resource-id",
			ResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"layerName":       core.MappingNodeFromString("test-layer"),
					"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
					"version":         core.MappingNodeFromInt(1),
					"description":     core.MappingNodeFromString("Test layer version"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
		ExpectError: false,
		// No stabilised actions should be called since layer versions are immediately stable
	}
}

func TestLambdaLayerVersionResourceStabilisedSuite(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionResourceStabilisedSuite))
}
