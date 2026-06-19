package lambda

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type eventSourceMappingCreate struct {
	input *lambda.CreateEventSourceMappingInput
}

func (e *eventSourceMappingCreate) Name() string {
	return "create event source mapping"
}

func (e *eventSourceMappingCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToCreateEventSourceMappingInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}
	e.input = input
	return hasValues, saveOpCtx, nil
}

func (e *eventSourceMappingCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createEventSourceMappingOutput, err := lambdaService.CreateEventSourceMapping(ctx, e.input)
	if err != nil {
		return saveOpCtx, err
	}

	id := aws.ToString(createEventSourceMappingOutput.UUID)
	eventSourceMappingArn := aws.ToString(createEventSourceMappingOutput.EventSourceMappingArn)

	// Set the ProviderUpstreamID to the ARN for tagging operations
	newSaveOpCtx.ProviderUpstreamID = eventSourceMappingArn
	newSaveOpCtx.Data["createEventSourceMappingOutput"] = createEventSourceMappingOutput
	newSaveOpCtx.Data["id"] = id
	newSaveOpCtx.Data["eventSourceMappingArn"] = eventSourceMappingArn

	return newSaveOpCtx, nil
}

func changesToCreateEventSourceMappingInput(
	specData *core.MappingNode,
) (*lambda.CreateEventSourceMappingInput, bool, error) {
	input := &lambda.CreateEventSourceMappingInput{}

	valueSetters := []*pluginutils.ValueSetter[*lambda.CreateEventSourceMappingInput]{
		pluginutils.NewValueSetter(
			"$.functionName",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.FunctionName = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.eventSourceArn",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.EventSourceArn = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.batchSize",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.BatchSize = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.enabled",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.Enabled = aws.Bool(core.BoolValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.startingPosition",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.StartingPosition = types.EventSourcePosition(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.startingPositionTimestamp",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				timestamp := core.FloatValue(value)
				input.StartingPositionTimestamp = aws.Time(time.Unix(int64(timestamp), 0))
			},
		),
		pluginutils.NewValueSetter(
			"$.maximumBatchingWindowInSeconds",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.MaximumBatchingWindowInSeconds = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.maximumRecordAgeInSeconds",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.MaximumRecordAgeInSeconds = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.maximumRetryAttempts",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.MaximumRetryAttempts = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.bisectBatchOnFunctionError",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.BisectBatchOnFunctionError = aws.Bool(core.BoolValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.parallelizationFactor",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.ParallelizationFactor = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.tumblingWindowInSeconds",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.TumblingWindowInSeconds = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.kmsKeyArn",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.KMSKeyArn = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.functionResponseTypes",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				for _, item := range value.Items {
					input.FunctionResponseTypes = append(input.FunctionResponseTypes, types.FunctionResponseType(core.StringValue(item)))
				}
			},
		),
		pluginutils.NewValueSetter(
			"$.topics",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				for _, item := range value.Items {
					input.Topics = append(input.Topics, core.StringValue(item))
				}
			},
		),
		pluginutils.NewValueSetter(
			"$.queues",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.Queues = core.StringSliceValue(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.filterCriteria",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.FilterCriteria = buildFilterCriteriaFromSpecNode(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.destinationConfig",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.DestinationConfig = buildDestinationConfigFromSpecNode(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.sourceAccessConfigurations",
			func(value *core.MappingNode, input *lambda.CreateEventSourceMappingInput) {
				input.SourceAccessConfigurations = buildSourceAccessConfigurationsFromSpecNode(value)
			},
		),
	}

	hasValuesToSave := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasValuesToSave = hasValuesToSave || valueSetter.DidSet()
	}

	// FunctionName is required
	if input.FunctionName == nil {
		return nil, false, fmt.Errorf("functionName is required")
	}

	return input, hasValuesToSave, nil
}

func buildFilterCriteriaFromSpecNode(node *core.MappingNode) *types.FilterCriteria {
	if node == nil || node.Fields["filters"] == nil {
		return nil
	}

	var filters []types.Filter
	for _, filterNode := range node.Fields["filters"].Items {
		filter := types.Filter{}
		if patternNode := filterNode.Fields["pattern"]; patternNode != nil {
			filter.Pattern = aws.String(core.StringValue(patternNode))
		}
		filters = append(filters, filter)
	}

	return &types.FilterCriteria{
		Filters: filters,
	}
}

func buildDestinationConfigFromSpecNode(node *core.MappingNode) *types.DestinationConfig {
	if node == nil {
		return nil
	}

	destConfig := &types.DestinationConfig{}

	if onFailureNode := node.Fields["onFailure"]; onFailureNode != nil {
		if destinationNode := onFailureNode.Fields["destination"]; destinationNode != nil {
			destConfig.OnFailure = &types.OnFailure{
				Destination: aws.String(core.StringValue(destinationNode)),
			}
		}
	}

	if onSuccessNode := node.Fields["onSuccess"]; onSuccessNode != nil {
		if destinationNode := onSuccessNode.Fields["destination"]; destinationNode != nil {
			destConfig.OnSuccess = &types.OnSuccess{
				Destination: aws.String(core.StringValue(destinationNode)),
			}
		}
	}

	return destConfig
}

func buildSourceAccessConfigurationsFromSpecNode(node *core.MappingNode) []types.SourceAccessConfiguration {
	if node == nil || node.Items == nil {
		return nil
	}

	var configs []types.SourceAccessConfiguration
	for _, configNode := range node.Items {
		config := types.SourceAccessConfiguration{}
		if typeNode := configNode.Fields["type"]; typeNode != nil {
			config.Type = types.SourceAccessType(core.StringValue(typeNode))
		}
		if uriNode := configNode.Fields["uri"]; uriNode != nil {
			config.URI = aws.String(core.StringValue(uriNode))
		}
		configs = append(configs, config)
	}

	return configs
}
