package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (s *sqsQueueResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	sqsService, err := s.getSQSService(ctx, input.ProviderContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get SQS service: %w", err)
	}

	// Try to get ARN from current spec first
	arnStr := ""
	arnValue, hasArn := pluginutils.GetValueByPath("$.arn", input.CurrentResourceSpec)
	if hasArn {
		arnStr = core.StringValue(arnValue)
	}

	// If no ARN, attempt fallback lookup by Bluelink tags
	if arnStr == "" {
		fallbackArn, err := s.lookupQueueARNByTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup queue by tags: %w", err)
		}
		if fallbackArn == "" {
			// Resource doesn't exist yet
			return &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				},
			}, nil
		}
		arnStr = fallbackArn
	}

	// Extract queue name from ARN and get queue URL
	queueName := extractQueueNameFromARN(arnStr)
	getURLOutput, err := sqsService.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue URL: %w", err)
	}
	queueURL := aws.ToString(getURLOutput.QueueUrl)

	// Get queue attributes
	attrsOutput, err := sqsService.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	// Get queue tags
	tagsOutput, err := sqsService.ListQueueTags(ctx, &sqs.ListQueueTagsInput{
		QueueUrl: aws.String(queueURL),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list queue tags: %w", err)
	}

	// Build external state
	externalState := s.buildExternalStateFromAttributes(attrsOutput.Attributes, arnStr)

	// Add tags if present (filtering out Bluelink provenance tags)
	// Uses shared utility which sorts tags by key for consistent ordering
	if tagsOutput != nil && len(tagsOutput.Tags) > 0 {
		tagsNode := utils.UserTagsToMappingNode(tagsOutput.Tags, input.ProviderContext)
		if len(tagsNode.Items) > 0 {
			externalState["tags"] = tagsNode
		}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: externalState,
		},
	}, nil
}

// buildExternalStateFromAttributes converts SQS queue attributes to mapping node fields.
func (s *sqsQueueResourceActions) buildExternalStateFromAttributes(
	attrs map[string]string,
	arn string,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{
		"arn": core.MappingNodeFromString(arn),
	}

	// Extract queue name from the ARN we already have
	fields["queueName"] = core.MappingNodeFromString(extractQueueNameFromARN(arn))

	// Integer attributes - must be converted from string to int to match schema
	integerAttrs := map[string]string{
		"DelaySeconds":                  "delaySeconds",
		"MaximumMessageSize":            "maximumMessageSize",
		"MessageRetentionPeriod":        "messageRetentionPeriod",
		"ReceiveMessageWaitTimeSeconds": "receiveMessageWaitTimeSeconds",
		"VisibilityTimeout":             "visibilityTimeout",
		"KmsDataKeyReusePeriodSeconds":  "kmsDataKeyReusePeriodSeconds",
	}
	for awsAttr, specField := range integerAttrs {
		if value, ok := attrs[awsAttr]; ok && value != "" {
			if intVal, err := strconv.Atoi(value); err == nil {
				fields[specField] = core.MappingNodeFromInt(intVal)
			}
		}
	}

	// Boolean attributes - must be converted from string to bool to match schema
	booleanAttrs := map[string]string{
		"SqsManagedSseEnabled":      "sqsManagedSseEnabled",
		"FifoQueue":                 "fifoQueue",
		"ContentBasedDeduplication": "contentBasedDeduplication",
	}
	for awsAttr, specField := range booleanAttrs {
		if value, ok := attrs[awsAttr]; ok && value != "" {
			fields[specField] = core.MappingNodeFromBool(value == "true")
		}
	}

	// String attributes
	stringAttrs := map[string]string{
		"KmsMasterKeyId":      "kmsMasterKeyId",
		"DeduplicationScope":  "deduplicationScope",
		"FifoThroughputLimit": "fifoThroughputLimit",
	}
	for awsAttr, specField := range stringAttrs {
		if value, ok := attrs[awsAttr]; ok && value != "" {
			fields[specField] = core.MappingNodeFromString(value)
		}
	}

	// JSON object attributes - must be parsed into structured objects to match schema
	jsonAttrs := map[string]string{
		"RedrivePolicy":      "redrivePolicy",
		"RedriveAllowPolicy": "redriveAllowPolicy",
		"Policy":             "policy",
	}
	for awsAttr, specField := range jsonAttrs {
		if value, ok := attrs[awsAttr]; ok && value != "" {
			var parsed interface{}
			if err := json.Unmarshal([]byte(value), &parsed); err == nil {
				if node, err := pluginutils.AnyToMappingNode(parsed); err == nil {
					fields[specField] = node
				}
			}
		}
	}

	return fields
}

// extractQueueNameFromARN extracts the queue name from an SQS queue ARN.
func extractQueueNameFromARN(arn string) string {
	// ARN format: arn:aws:sqs:region:account-id:queue-name
	for i := len(arn) - 1; i >= 0; i-- {
		if arn[i] == ':' {
			return arn[i+1:]
		}
	}
	return arn
}

// lookupQueueARNByTags attempts to find an SQS queue by its Bluelink provenance tags.
// This is used as a fallback when the ARN is not available (e.g., interrupted resource creation).
// Returns empty string if no matching resource is found.
func (s *sqsQueueResourceActions) lookupQueueARNByTags(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (string, error) {
	tagFilters := utils.BuildBluelinkTagFiltersForLookup(input)
	if tagFilters == nil {
		// Tagging is not enabled, cannot perform fallback lookup
		return "", nil
	}

	taggingService, err := s.getResourceGroupTaggingService(ctx, input.ProviderContext)
	if err != nil {
		return "", err
	}

	result, err := taggingService.GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters:          tagFilters,
		ResourceTypeFilters: []string{"sqs"},
	})
	if err != nil {
		return "", err
	}

	if len(result.ResourceTagMappingList) == 0 {
		return "", nil
	}

	return aws.ToString(result.ResourceTagMappingList[0].ResourceARN), nil
}
