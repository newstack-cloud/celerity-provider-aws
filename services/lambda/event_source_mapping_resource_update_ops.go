package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type eventSourceMappingUpdate struct {
	input *lambda.UpdateEventSourceMappingInput
}

func (e *eventSourceMappingUpdate) Name() string {
	return "update event source mapping"
}

func (e *eventSourceMappingUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToUpdateEventSourceMappingInput(specData, changes)
	if err != nil {
		return false, saveOpCtx, err
	}
	e.input = input
	return hasValues, saveOpCtx, nil
}

func (e *eventSourceMappingUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	updateEventSourceMappingOutput, err := lambdaService.UpdateEventSourceMapping(ctx, e.input)
	if err != nil {
		return saveOpCtx, err
	}

	id := aws.ToString(updateEventSourceMappingOutput.UUID)
	eventSourceMappingArn := aws.ToString(updateEventSourceMappingOutput.EventSourceMappingArn)

	// Set the ProviderUpstreamID to the ARN for tagging operations
	newSaveOpCtx.ProviderUpstreamID = eventSourceMappingArn
	newSaveOpCtx.Data["updateEventSourceMappingOutput"] = updateEventSourceMappingOutput
	newSaveOpCtx.Data["id"] = id
	newSaveOpCtx.Data["eventSourceMappingArn"] = eventSourceMappingArn

	return newSaveOpCtx, nil
}

func changesToUpdateEventSourceMappingInput(
	specData *core.MappingNode,
	changes *provider.Changes,
) (*lambda.UpdateEventSourceMappingInput, bool, error) {
	input := &lambda.UpdateEventSourceMappingInput{}

	// ID is required for updates
	if idField := specData.Fields["id"]; idField != nil {
		input.UUID = aws.String(core.StringValue(idField))
	} else {
		return nil, false, fmt.Errorf("id is required for updates")
	}

	valueSetters := []*pluginutils.ValueSetter[*lambda.UpdateEventSourceMappingInput]{
		pluginutils.NewValueSetter(
			"$.functionName",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.FunctionName = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.batchSize",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.BatchSize = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.enabled",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.Enabled = aws.Bool(core.BoolValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.maximumBatchingWindowInSeconds",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.MaximumBatchingWindowInSeconds = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.maximumRecordAgeInSeconds",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.MaximumRecordAgeInSeconds = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.maximumRetryAttempts",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.MaximumRetryAttempts = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.bisectBatchOnFunctionError",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.BisectBatchOnFunctionError = aws.Bool(core.BoolValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.parallelizationFactor",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.ParallelizationFactor = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.tumblingWindowInSeconds",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.TumblingWindowInSeconds = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.kmsKeyArn",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.KMSKeyArn = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.functionResponseTypes",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				for _, item := range value.Items {
					input.FunctionResponseTypes = append(input.FunctionResponseTypes, types.FunctionResponseType(core.StringValue(item)))
				}
			},
		),
		pluginutils.NewValueSetter(
			"$.filterCriteria",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.FilterCriteria = buildFilterCriteriaForUpdate(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.destinationConfig",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.DestinationConfig = buildDestinationConfigFromSpecNode(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.sourceAccessConfigurations",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.SourceAccessConfigurations = buildSourceAccessConfigurationsFromSpecNode(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.metricsConfig",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.MetricsConfig = buildMetricsConfigForUpdate(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.scalingConfig",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.ScalingConfig = buildScalingConfigForUpdate(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.provisionedPollerConfig",
			func(value *core.MappingNode, input *lambda.UpdateEventSourceMappingInput) {
				input.ProvisionedPollerConfig = buildProvisionedPollerConfigForUpdate(value)
			},
		),
	}

	hasValuesToSave := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasValuesToSave = hasValuesToSave || valueSetter.DidSet()
	}

	return input, hasValuesToSave, nil
}

func buildFilterCriteriaForUpdate(node *core.MappingNode) *types.FilterCriteria {
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

func buildMetricsConfigForUpdate(node *core.MappingNode) *types.EventSourceMappingMetricsConfig {
	if node == nil || node.Fields["metrics"] == nil {
		return nil
	}

	var metrics []types.EventSourceMappingMetric
	for _, metricNode := range node.Fields["metrics"].Items {
		metrics = append(metrics, types.EventSourceMappingMetric(core.StringValue(metricNode)))
	}

	return &types.EventSourceMappingMetricsConfig{
		Metrics: metrics,
	}
}

func buildScalingConfigForUpdate(node *core.MappingNode) *types.ScalingConfig {
	if node == nil {
		return nil
	}

	scalingConfig := &types.ScalingConfig{}
	if maxConcurrencyNode := node.Fields["maximumConcurrency"]; maxConcurrencyNode != nil {
		scalingConfig.MaximumConcurrency = aws.Int32(int32(core.IntValue(maxConcurrencyNode)))
	}

	return scalingConfig
}

func buildProvisionedPollerConfigForUpdate(node *core.MappingNode) *types.ProvisionedPollerConfig {
	if node == nil {
		return nil
	}

	pollerConfig := &types.ProvisionedPollerConfig{}
	if minPollersNode := node.Fields["minimumPollers"]; minPollersNode != nil {
		pollerConfig.MinimumPollers = aws.Int32(int32(core.IntValue(minPollersNode)))
	}
	if maxPollersNode := node.Fields["maximumPollers"]; maxPollersNode != nil {
		pollerConfig.MaximumPollers = aws.Int32(int32(core.IntValue(maxPollersNode)))
	}

	return pollerConfig
}
