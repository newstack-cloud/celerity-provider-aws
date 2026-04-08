//go:build unit

package sqs

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	sqsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/sqs_mock"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type SQSQueueResourceDestroySuite struct {
	suite.Suite
}

func (s *SQSQueueResourceDestroySuite) Test_destroy_sqs_queue() {
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

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, sqsservice.Service]{
		createSuccessfulDestroyTestCase(providerCtx, loader),
		createFailingDestroyTestCase(providerCtx, loader),
	}

	queueResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, sqsservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return QueueResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		queueResourceWrapper,
		&s.Suite,
	)
}

func createSuccessfulDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, sqsservice.Service] {
	service := sqsmock.CreateSQSServiceMock(
		sqsmock.WithDeleteQueueOutput(&sqs.DeleteQueueOutput{}),
	)

	expectedQueueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/test-queue"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, sqsservice.Service]{
		Name: "successfully deletes queue",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"queueUrl": core.MappingNodeFromString(expectedQueueURL),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"DeleteQueue": &sqs.DeleteQueueInput{
				QueueUrl: aws.String(expectedQueueURL),
			},
		},
	}
}

func createFailingDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, sqsservice.Service] {
	service := sqsmock.CreateSQSServiceMock(
		sqsmock.WithDeleteQueueError(errors.New("failed to delete queue")),
	)

	expectedQueueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/test-queue"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, sqsservice.Service]{
		Name: "fails to delete queue",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"queueUrl": core.MappingNodeFromString(expectedQueueURL),
					},
				},
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteQueue": &sqs.DeleteQueueInput{
				QueueUrl: aws.String(expectedQueueURL),
			},
		},
	}
}

func TestSQSQueueResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(SQSQueueResourceDestroySuite))
}
