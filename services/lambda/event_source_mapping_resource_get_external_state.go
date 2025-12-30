package lambda

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaEventSourceMappingResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	// Try to get UUID from current spec first
	uuid := ""
	idValue, hasID := pluginutils.GetValueByPath("$.id", input.CurrentResourceSpec)
	if hasID {
		uuid = core.StringValue(idValue)
	}

	// If no UUID, attempt fallback lookup by Bluelink tags
	if uuid == "" {
		fallbackUUID, err := l.lookupEventSourceMappingUUIDByTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup event source mapping by tags: %w", err)
		}
		if fallbackUUID == "" {
			// Resource doesn't exist yet
			return &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
			}, nil
		}
		uuid = fallbackUUID
	}

	getEventSourceMappingInput := &lambda.GetEventSourceMappingInput{
		UUID: aws.String(uuid),
	}

	result, err := lambdaService.GetEventSourceMapping(ctx, getEventSourceMappingInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get event source mapping: %w", err)
	}

	// Build resource spec state from AWS response
	resourceSpecState := l.buildBaseResourceSpecState(result)

	// Add optional fields if they exist
	err = l.addOptionalConfigurationsToSpec(result, resourceSpecState.Fields)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (l *lambdaEventSourceMappingResourceActions) buildBaseResourceSpecState(
	output *lambda.GetEventSourceMappingOutput,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"id":                    core.MappingNodeFromString(aws.ToString(output.UUID)),
			"eventSourceMappingArn": core.MappingNodeFromString(aws.ToString(output.EventSourceMappingArn)),
			"functionArn":           core.MappingNodeFromString(aws.ToString(output.FunctionArn)),
			"enabled":               core.MappingNodeFromBool(aws.ToString(output.State) == "Enabled"),
		},
	}
}

func (l *lambdaEventSourceMappingResourceActions) addOptionalConfigurationsToSpec(
	output *lambda.GetEventSourceMappingOutput,
	specFields map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetEventSourceMappingOutput]{
		{
			Name: "eventSourceArn",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.EventSourceArn != nil
			},
			Fields: []string{"eventSourceArn"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.EventSourceArn)),
				}, nil
			},
		},
		{
			Name: "batchSize",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.BatchSize != nil
			},
			Fields: []string{"batchSize"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.BatchSize))),
				}, nil
			},
		},
		{
			Name: "state",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.State != nil
			},
			Fields: []string{"state"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.State)),
				}, nil
			},
		},
		{
			Name: "startingPosition",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.StartingPosition != ""
			},
			Fields: []string{"startingPosition"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(string(output.StartingPosition)),
				}, nil
			},
		},
		{
			Name: "maximumBatchingWindowInSeconds",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.MaximumBatchingWindowInSeconds != nil
			},
			Fields: []string{"maximumBatchingWindowInSeconds"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.MaximumBatchingWindowInSeconds))),
				}, nil
			},
		},
		{
			Name: "maximumRecordAgeInSeconds",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.MaximumRecordAgeInSeconds != nil
			},
			Fields: []string{"maximumRecordAgeInSeconds"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.MaximumRecordAgeInSeconds))),
				}, nil
			},
		},
		{
			Name: "maximumRetryAttempts",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.MaximumRetryAttempts != nil
			},
			Fields: []string{"maximumRetryAttempts"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.MaximumRetryAttempts))),
				}, nil
			},
		},
		{
			Name: "bisectBatchOnFunctionError",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.BisectBatchOnFunctionError != nil
			},
			Fields: []string{"bisectBatchOnFunctionError"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromBool(aws.ToBool(output.BisectBatchOnFunctionError)),
				}, nil
			},
		},
		{
			Name: "parallelizationFactor",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.ParallelizationFactor != nil
			},
			Fields: []string{"parallelizationFactor"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.ParallelizationFactor))),
				}, nil
			},
		},
		{
			Name: "tumblingWindowInSeconds",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.TumblingWindowInSeconds != nil
			},
			Fields: []string{"tumblingWindowInSeconds"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.TumblingWindowInSeconds))),
				}, nil
			},
		},
		{
			Name: "kmsKeyArn",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.KMSKeyArn != nil
			},
			Fields: []string{"kmsKeyArn"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.KMSKeyArn)),
				}, nil
			},
		},
		{
			Name: "functionResponseTypes",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return len(output.FunctionResponseTypes) > 0
			},
			Fields: []string{"functionResponseTypes"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				items := make([]*core.MappingNode, len(output.FunctionResponseTypes))
				for i, responseType := range output.FunctionResponseTypes {
					items[i] = core.MappingNodeFromString(string(responseType))
				}
				return []*core.MappingNode{
					{Items: items},
				}, nil
			},
		},
		{
			Name: "topics",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return len(output.Topics) > 0
			},
			Fields: []string{"topics"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				items := make([]*core.MappingNode, len(output.Topics))
				for i, topic := range output.Topics {
					items[i] = core.MappingNodeFromString(topic)
				}
				return []*core.MappingNode{
					{Items: items},
				}, nil
			},
		},
		{
			Name: "queues",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return len(output.Queues) > 0
			},
			Fields: []string{"queues"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				items := make([]*core.MappingNode, len(output.Queues))
				for i, queue := range output.Queues {
					items[i] = core.MappingNodeFromString(queue)
				}
				return []*core.MappingNode{
					{Items: items},
				}, nil
			},
		},
		{
			Name: "filterCriteria",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.FilterCriteria != nil && len(output.FilterCriteria.Filters) > 0
			},
			Fields: []string{"filterCriteria"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					eventSourceMappingFilterCriteriaToMappingNode(output.FilterCriteria),
				}, nil
			},
		},
		{
			Name: "destinationConfig",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return output.DestinationConfig != nil
			},
			Fields: []string{"destinationConfig"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					eventSourceMappingDestinationConfigToMappingNode(output.DestinationConfig),
				}, nil
			},
		},
		{
			Name: "sourceAccessConfigurations",
			Condition: func(output *lambda.GetEventSourceMappingOutput) bool {
				return len(output.SourceAccessConfigurations) > 0
			},
			Fields: []string{"sourceAccessConfigurations"},
			Values: func(output *lambda.GetEventSourceMappingOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					eventSourceMappingSourceAccessConfigurationsToMappingNode(output.SourceAccessConfigurations),
				}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(
		output,
		specFields,
		extractors,
	)
}

func eventSourceMappingFilterCriteriaToMappingNode(
	filterCriteria *types.FilterCriteria,
) *core.MappingNode {
	if filterCriteria == nil || len(filterCriteria.Filters) == 0 {
		return &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	}

	filters := make([]*core.MappingNode, len(filterCriteria.Filters))
	for i, filter := range filterCriteria.Filters {
		filters[i] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"pattern": core.MappingNodeFromString(aws.ToString(filter.Pattern)),
			},
		}
	}

	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"filters": {Items: filters},
		},
	}
}

func eventSourceMappingDestinationConfigToMappingNode(
	destinationConfig *types.DestinationConfig,
) *core.MappingNode {
	if destinationConfig == nil {
		return &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	}

	fields := map[string]*core.MappingNode{}

	if destinationConfig.OnFailure != nil {
		fields["onFailure"] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"destination": core.MappingNodeFromString(aws.ToString(destinationConfig.OnFailure.Destination)),
			},
		}
	}

	if destinationConfig.OnSuccess != nil {
		fields["onSuccess"] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"destination": core.MappingNodeFromString(aws.ToString(destinationConfig.OnSuccess.Destination)),
			},
		}
	}

	return &core.MappingNode{Fields: fields}
}

func eventSourceMappingSourceAccessConfigurationsToMappingNode(
	configurations []types.SourceAccessConfiguration,
) *core.MappingNode {
	if len(configurations) == 0 {
		return &core.MappingNode{Items: []*core.MappingNode{}}
	}

	items := make([]*core.MappingNode, len(configurations))
	for i, config := range configurations {
		items[i] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"type": core.MappingNodeFromString(string(config.Type)),
				"uri":  core.MappingNodeFromString(aws.ToString(config.URI)),
			},
		}
	}

	return &core.MappingNode{Items: items}
}

// lookupEventSourceMappingUUIDByTags attempts to find a Lambda event source mapping by its Bluelink provenance tags.
// This is used as a fallback when the UUID is not available (e.g., interrupted resource creation).
// Returns empty string if no matching resource is found.
func (l *lambdaEventSourceMappingResourceActions) lookupEventSourceMappingUUIDByTags(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (string, error) {
	tagFilters := utils.BuildBluelinkTagFiltersForLookup(input)
	if tagFilters == nil {
		// Tagging is not enabled, cannot perform fallback lookup
		return "", nil
	}

	taggingService, err := l.getResourceGroupTaggingService(ctx, input.ProviderContext)
	if err != nil {
		return "", err
	}

	result, err := taggingService.GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters:          tagFilters,
		ResourceTypeFilters: []string{"lambda:event-source-mapping"},
	})
	if err != nil {
		return "", err
	}

	if len(result.ResourceTagMappingList) == 0 {
		return "", nil
	}

	// The ARN format is: arn:aws:lambda:region:account:event-source-mapping:uuid
	// Extract the UUID from the ARN
	arn := aws.ToString(result.ResourceTagMappingList[0].ResourceARN)
	parts := strings.Split(arn, ":")
	if len(parts) < 7 {
		return "", fmt.Errorf("invalid event source mapping ARN format: %s", arn)
	}
	return parts[len(parts)-1], nil
}
