//go:build unit

package sqslinks

import (
	"fmt"
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
	"github.com/stretchr/testify/suite"
)

type QueueQueueLinkUpdateSuite struct {
	suite.Suite
}

func (s *QueueQueueLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}
	linkCtx := plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {
				"region": core.ScalarFromString("us-west-2"),
			},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		s.createUpdateLinkSourceQueueTestCase(linkCtx, loader),
		s.createUpdateLinkDeadLetterQueueTestCase(linkCtx, loader),
		s.createUpdateLinkRemoveRedrivePolicyTestCase(linkCtx, loader),
		s.createUpdateLinkErrorSourceQueueMissingURLTestCase(linkCtx, loader),
		s.createUpdateLinkErrorDLQMissingARNTestCase(linkCtx, loader),
		s.createUpdateLinkErrorUpdateServiceErrorTestCase(linkCtx, loader),
		s.createUpdateLinkErrorRemoveServiceErrorTestCase(linkCtx, loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		QueueQueueLink,
		&s.Suite,
	)
}

func (s *QueueQueueLinkUpdateSuite) createUpdateLinkSourceQueueTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	service := sqsmock.CreateSQSServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}

	sourceQueueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/source-queue"
	dlqARN := "arn:aws:sqs:us-west-2:123456789012:dead-letter-queue"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name:                    "Updates source queue with redrive policy reference to DLQ",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "source-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"queueUrl": core.MappingNodeFromString(sourceQueueURL),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "dead-letter-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(dlqARN),
						},
					},
				},
			},
			LinkContext: linkCtx,
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"source-queue": {
						Fields: map[string]*core.MappingNode{
							"redrivePolicy": {
								Fields: map[string]*core.MappingNode{
									"deadLetterTargetArn": core.MappingNodeFromString(dlqARN),
									"maxReceiveCount":     core.MappingNodeFromInt(10),
								},
							},
						},
					},
				},
			},
			ResourceDataMappings: map[string]string{
				"source-queue::spec.redrivePolicy": "source-queue.redrivePolicy",
			},
		},
		UpdateActionsCalled: map[string]any{
			"SetQueueAttributes": &sqs.SetQueueAttributesInput{
				QueueUrl: aws.String(sourceQueueURL),
				Attributes: map[string]string{
					"RedrivePolicy": `{"deadLetterTargetArn":"` + dlqARN + `","maxReceiveCount":10}`,
				},
			},
		},
	}
}

func (s *QueueQueueLinkUpdateSuite) createUpdateLinkDeadLetterQueueTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	service := sqsmock.CreateSQSServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name:                    "Returns empty link data as DLQ resource is not updated for the link",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName:         "source-queue",
				CurrentResourceState: &state.ResourceState{},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName:         "dead-letter-queue",
				CurrentResourceState: &state.ResourceState{},
			},
			LinkContext: linkCtx,
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
	}
}

func (s *QueueQueueLinkUpdateSuite) createUpdateLinkRemoveRedrivePolicyTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	service := sqsmock.CreateSQSServiceMock(
		sqsmock.WithSetQueueAttributesOutput(
			&sqs.SetQueueAttributesOutput{},
		),
	)
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}

	sourceQueueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/source-queue"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name:                    "Removes redrive policy from source queue",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeDestroy,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "source-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"queueUrl": core.MappingNodeFromString(sourceQueueURL),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "dead-letter-queue",
			},
			LinkContext: linkCtx,
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
		UpdateActionsCalled: map[string]any{
			"SetQueueAttributes": &sqs.SetQueueAttributesInput{
				QueueUrl: aws.String(sourceQueueURL),
				Attributes: map[string]string{
					"RedrivePolicy": "",
				},
			},
		},
	}
}

func (s *QueueQueueLinkUpdateSuite) createUpdateLinkErrorSourceQueueMissingURLTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	service := sqsmock.CreateSQSServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}

	dlqARN := "arn:aws:sqs:us-west-2:123456789012:dead-letter-queue"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name:                    "Returns error if source queue URL is missing from queue resource spec",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "source-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							// Missing queueUrl field.
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "dead-letter-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(dlqARN),
						},
					},
				},
			},
			LinkContext: linkCtx,
		},
		UpdateActionsNotCalled: []string{"SetQueueAttributes"},
		ExpectError:            true,
		ExpectedErrorMessage:   "queue URL could not be retrieved from source queue \"source-queue\"",
	}
}

func (s *QueueQueueLinkUpdateSuite) createUpdateLinkErrorDLQMissingARNTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	service := sqsmock.CreateSQSServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}

	sourceQueueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/source-queue"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name:                    "Returns error if DLQ ARN is missing from queue resource spec",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "source-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"queueUrl": core.MappingNodeFromString(sourceQueueURL),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "dead-letter-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							// Missing arn field.
						},
					},
				},
			},
			LinkContext: linkCtx,
		},
		UpdateActionsNotCalled: []string{"SetQueueAttributes"},
		ExpectError:            true,
		ExpectedErrorMessage:   "queue ARN could not be retrieved from dead-letter queue \"dead-letter-queue\"",
	}
}

func (s *QueueQueueLinkUpdateSuite) createUpdateLinkErrorUpdateServiceErrorTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	service := sqsmock.CreateSQSServiceMock(
		sqsmock.WithSetQueueAttributesError(fmt.Errorf("test error")),
	)
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}

	sourceQueueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/source-queue"
	dlqARN := "arn:aws:sqs:us-west-2:123456789012:dead-letter-queue"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name:                    "Returns error if service returns error when adding redrive policy to queue",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "source-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"queueUrl": core.MappingNodeFromString(sourceQueueURL),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "dead-letter-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(dlqARN),
						},
					},
				},
			},
			LinkContext: linkCtx,
		},
		UpdateActionsCalled: map[string]any{
			"SetQueueAttributes": &sqs.SetQueueAttributesInput{
				QueueUrl: aws.String(sourceQueueURL),
				Attributes: map[string]string{
					"RedrivePolicy": `{"deadLetterTargetArn":"` + dlqARN + `","maxReceiveCount":10}`,
				},
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "test error",
	}
}

func (s *QueueQueueLinkUpdateSuite) createUpdateLinkErrorRemoveServiceErrorTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	service := sqsmock.CreateSQSServiceMock(
		sqsmock.WithSetQueueAttributesError(fmt.Errorf("test error")),
	)
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}

	sourceQueueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/source-queue"

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name:                    "Returns error if service returns error when removing redrive policy from queue",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         serviceFactory,
		ConfigStoreA:            configStore,
		ServiceFactoryB:         serviceFactory,
		ConfigStoreB:            configStore,
		CurrentServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType: provider.LinkUpdateTypeDestroy,
			ResourceInfo: &provider.ResourceInfo{
				ResourceName: "source-queue",
				CurrentResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"queueUrl": core.MappingNodeFromString(sourceQueueURL),
						},
					},
				},
			},
			OtherResourceInfo: &provider.ResourceInfo{
				ResourceName: "dead-letter-queue",
			},
			LinkContext: linkCtx,
		},
		UpdateActionsCalled: map[string]any{
			"SetQueueAttributes": &sqs.SetQueueAttributesInput{
				QueueUrl: aws.String(sourceQueueURL),
				Attributes: map[string]string{
					"RedrivePolicy": "",
				},
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "test error",
	}
}

func (s *QueueQueueLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}
	linkCtx := plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {
				"region": core.ScalarFromString("us-west-2"),
			},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		s.createUpdateLinkIntermediaryResourcesTestCase(linkCtx, loader),
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		testCases,
		QueueQueueLink,
		&s.Suite,
	)
}

func (s *QueueQueueLinkUpdateSuite) createUpdateLinkIntermediaryResourcesTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	service := sqsmock.CreateSQSServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name:                           "Returns empty link data as there are no intermediary resources to create or update",
		ServiceFactoryA:                serviceFactory,
		ConfigStoreA:                   configStore,
		ServiceFactoryB:                serviceFactory,
		ConfigStoreB:                   configStore,
		IntermediariesServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			LinkContext:    linkCtx,
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
	}
}

func TestQueueQueueLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(QueueQueueLinkUpdateSuite))
}
