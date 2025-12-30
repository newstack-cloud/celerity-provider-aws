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
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
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
	}

	// Create a wrapper function that matches the expected signature
	queueResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, sqsservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return QueueResource(serviceFactory, resgrouptagservice.NewService, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		queueResourceWrapper,
		&s.Suite,
	)
}

func (s *SQSQueueResourceCreateIntegrationSuite) createBasicQueueIntegrationTestCase(
	providerCtx provider.Context,
	loader utils.AWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service] {
	queueName := "bluelink-provider-test-queue"
	tags := map[string]string{
		"Environment": "test",
		"Project":     "test-project",
	}

	config := CreateIntegrationTestCaseConfig(queueName, tags, loader, providerCtx)
	testCase := CreateBasicQueueTestCase(config)

	// Add integration test specific configurations
	testCase.ConfigStore = utils.NewAWSConfigStore(
		os.Environ(),
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	testCase.ExpectedOutputMatcher = func(actual *provider.ResourceDeployOutput) (plugintestutils.EqualityCheckValues, error) {
		ctx := context.Background()
		awsAccountInfo, err := integration.GetAWSAccountInfo(ctx, loader)
		if err != nil {
			return plugintestutils.EqualityCheckValues{}, fmt.Errorf("failed to get AWS account info: %w", err)
		}

		expectedQueueARN := fmt.Sprintf(
			"arn:aws:sqs:%s:%s:%s",
			awsAccountInfo.Region,
			awsAccountInfo.AccountID,
			queueName,
		)
		expectedQueueURL := fmt.Sprintf(
			"https://sqs.%s.amazonaws.com/%s/%s",
			awsAccountInfo.Region,
			awsAccountInfo.AccountID,
			queueName,
		)
		return plugintestutils.EqualityCheckValues{
			Expected: map[string]*core.MappingNode{
				"spec.arn":      core.MappingNodeFromString(expectedQueueARN),
				"spec.queueUrl": core.MappingNodeFromString(expectedQueueURL),
			},
			Actual: actual.ComputedFieldValues,
		}, nil
	}

	testCase.ExtraAssertions = func(
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

		specData := CreateBasicQueueSpecData(queueName, tags)
		testutils.AssertTags(
			suite,
			specData.Fields["tags"].Items,
			tagsOutput.Tags,
		)
	}

	testCase.Cleanup = func(ctx context.Context, output *provider.ResourceDeployOutput) error {
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
	}

	return testCase
}

func TestSQSQueueResourceCreateIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SQSQueueResourceCreateIntegrationSuite))
}
