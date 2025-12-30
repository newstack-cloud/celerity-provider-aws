//go:build unit

package sqs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	sqsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/sqs_mock"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type SQSQueueResourceStabilisedSuite struct {
	suite.Suite
}

func (s *SQSQueueResourceStabilisedSuite) Test_stabilised_sqs_queue() {
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

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, sqsservice.Service]{
		stabilisedBasicAliasTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	queueResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, sqsservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return QueueResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		queueResourceWrapper,
		&s.Suite,
	)
}

func stabilisedBasicAliasTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, sqsservice.Service] {
	service := sqsmock.CreateSQSServiceMock()

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"queueName": core.MappingNodeFromString("test-queue"),
			"queueUrl":  core.MappingNodeFromString("https://sqs.us-west-2.amazonaws.com/123456789012/test-queue"),
			"arn":       core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:test-queue"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, sqsservice.Service]{
		Name: "basic queue is stabilised",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
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

func TestSQSQueueResourceStabilisedSuite(t *testing.T) {
	suite.Run(t, new(SQSQueueResourceStabilisedSuite))
}
