//go:build integration

package sqs

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/integration"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type SQSQueueResourceCreateIntegrationSuite struct {
	suite.Suite
}

func (s *SQSQueueResourceCreateIntegrationSuite) Test_create_sqs_queue() {
	loader := &utils.DefaultAWSConfigLoader{}
	region := os.Getenv("AWS_REGION")
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString(region),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service]{
		s.createBasicQueueIntegrationTestCase(providerCtx, loader),
		// createAdvancedQueueTestCase(providerCtx, loader),
		// createFIFOQueueTestCase(providerCtx, loader),
		// createSQSFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		QueueResource,
		&s.Suite,
	)
}

func (s *SQSQueueResourceCreateIntegrationSuite) createBasicQueueIntegrationTestCase(
	providerCtx provider.Context,
	loader utils.AWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service] {
	specData := core.MappingNodeFields(
		"queueName",
		core.MappingNodeFromString("bluelink-provider-test-queue"),
		"tags",
		core.MappingNodeItems(
			core.MappingNodeFields(
				"key",
				core.MappingNodeFromString("Environment"),
				"value",
				core.MappingNodeFromString("test"),
			),
			core.MappingNodeFields(
				"key",
				core.MappingNodeFromString("Project"),
				"value",
				core.MappingNodeFromString("test-project"),
			),
		),
	)

	return plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service]{
		Name:           "create basic queue",
		ServiceFactory: sqsservice.NewService,
		ConfigStore: utils.NewAWSConfigStore(
			os.Environ(),
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-queue-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-queue-id",
					ResourceName: "TestQueue",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/sqs/queue",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.queueName",
					},
					{
						FieldPath: "spec.tags[0].key",
					},
					{
						FieldPath: "spec.tags[0].value",
					},
					{
						FieldPath: "spec.tags[1].key",
					},
					{
						FieldPath: "spec.tags[1].value",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutputMatcher: func(actual *provider.ResourceDeployOutput) (plugintestutils.EqualityCheckValues, error) {
			ctx := context.Background()
			awsAccountInfo, err := integration.GetAWSAccountInfo(ctx, loader)
			if err != nil {
				return plugintestutils.EqualityCheckValues{}, fmt.Errorf("failed to get AWS account info: %w", err)
			}

			expectedQueueARN := fmt.Sprintf(
				"arn:aws:sqs:%s:%s:bluelink-provider-test-queue",
				awsAccountInfo.Region,
				awsAccountInfo.AccountID,
			)
			expectedQueueURL := fmt.Sprintf(
				"https://sqs.%s.amazonaws.com/%s/bluelink-provider-test-queue",
				awsAccountInfo.Region,
				awsAccountInfo.AccountID,
			)
			return plugintestutils.EqualityCheckValues{
				Expected: map[string]*core.MappingNode{
					"spec.arn":      core.MappingNodeFromString(expectedQueueARN),
					"spec.queueUrl": core.MappingNodeFromString(expectedQueueURL),
				},
				Actual: actual.ComputedFieldValues,
			}, nil
		},
		ExtraAssertions: func(
			ctx context.Context,
			suite *suite.Suite,
			output *provider.ResourceDeployOutput,
		) {
			config, err := loader.LoadDefaultConfig(ctx)
			if err != nil {
				suite.FailNow("failed to load AWS config: %w", err)
			}

			sqsService := sqs.NewFromConfig(config)
			tagsOutput, err := sqsService.ListQueueTags(
				ctx,
				&sqs.ListQueueTagsInput{
					QueueUrl: aws.String(
						core.StringValue(output.ComputedFieldValues["spec.queueUrl"]),
					),
				},
			)
			if err != nil {
				suite.FailNow("failed to get queue attributes: %w", err)
			}

			testutils.AssertTags(
				suite,
				specData.Fields["tags"].Items,
				tagsOutput.Tags,
			)
		},
		Cleanup: func(ctx context.Context, output *provider.ResourceDeployOutput) error {
			config, err := loader.LoadDefaultConfig(ctx)
			if err != nil {
				return fmt.Errorf("failed to load AWS config: %w", err)
			}

			sqsService := sqs.NewFromConfig(config)
			_, err = sqsService.DeleteQueue(ctx, &sqs.DeleteQueueInput{
				QueueUrl: aws.String(
					core.StringValue(output.ComputedFieldValues["spec.queueUrl"]),
				),
			})
			if err != nil {
				return fmt.Errorf("failed to delete queue: %w", err)
			}

			return nil
		},
	}
}

func TestSQSQueueResourceCreateIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SQSQueueResourceCreateIntegrationSuite))
}
