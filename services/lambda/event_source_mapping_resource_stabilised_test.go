//go:build unit

package lambda

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaEventSourceMappingResourceStabilisedSuite struct {
	suite.Suite
}

func (s *LambdaEventSourceMappingResourceStabilisedSuite) Test_stabilised_lambda_event_source_mapping() {
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
		stabilisedBasicEventSourceMappingEnabledTestCase(providerCtx, loader),
		stabilisedBasicEventSourceMappingDisabledTestCase(providerCtx, loader),
		stabilisedBasicEventSourceMappingErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		EventSourceMappingResource,
		&s.Suite,
	)
}

func stabilisedBasicEventSourceMappingEnabledTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetEventSourceMappingOutput(&lambda.GetEventSourceMappingOutput{
			State: aws.String(string("Enabled")),
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"id": core.MappingNodeFromString("123"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		Name: "basic event source mapping is stabilised when in an enabled state",
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
			ResourceSpec:    resourceSpecState,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func stabilisedBasicEventSourceMappingDisabledTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetEventSourceMappingOutput(&lambda.GetEventSourceMappingOutput{
			State: aws.String(string("Disabled")),
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"id": core.MappingNodeFromString("123"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		Name: "basic event source mapping is stabilised when in a disabled state",
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
			ResourceSpec:    resourceSpecState,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func stabilisedBasicEventSourceMappingErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetEventSourceMappingError(errors.New("event source mapping error")),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"id": core.MappingNodeFromString("123"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		Name: "basic event source mapping is not stabilised when there is an error",
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
			ResourceSpec:    resourceSpecState,
		},
		ExpectError: true,
	}
}

func TestLambdaEventSourceMappingResourceStabilised(t *testing.T) {
	suite.Run(t, new(LambdaEventSourceMappingResourceStabilisedSuite))
}
