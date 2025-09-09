package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type queueCreate struct {
	input *sqs.CreateQueueInput
}

func (q *queueCreate) Name() string {
	return "create queue"
}

func (q *queueCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToCreateQueueInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}

	// Generate a unique queue name if not provided.
	if input.QueueName == nil || aws.ToString(input.QueueName) == "" {
		generateName := utils.SQSQueueNameGenerator(
			input.Attributes["FifoQueue"] == "true",
		)

		inputData, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
		if !ok || inputData == nil {
			return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in SaveOperationContext.Data")
		}

		generatedName, err := generateName(inputData)
		if err != nil {
			return false, saveOpCtx, err
		}

		input.QueueName = aws.String(generatedName)
		hasValues = true
	}

	q.input = input
	return hasValues, saveOpCtx, nil
}

func (q *queueCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	sqsService sqsservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createQueueOutput, err := sqsService.CreateQueue(ctx, q.input)
	if err != nil {
		return saveOpCtx, err
	}

	attributesOutput, err := sqsService.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: createQueueOutput.QueueUrl,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameQueueArn,
		},
	})
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = attributesOutput.Attributes["QueueArn"]
	newSaveOpCtx.Data["createQueueOutput"] = createQueueOutput
	return newSaveOpCtx, nil
}

func changesToCreateQueueInput(specData *core.MappingNode) (*sqs.CreateQueueInput, bool, error) {
	input := &sqs.CreateQueueInput{
		Attributes: map[string]string{},
	}

	// Handle serialisation of nested attribute objects
	// outside of value setters to handle serialisation errors.
	redrivePolicy, exists := pluginutils.GetValueByPath("$.redrivePolicy", specData)
	if exists {
		redrivePolicyBytes, err := json.Marshal(redrivePolicy)
		if err != nil {
			return nil, false, err
		}
		input.Attributes["RedrivePolicy"] = string(redrivePolicyBytes)
	}

	redriveAllowPolicy, exists := pluginutils.GetValueByPath("$.redriveAllowPolicy", specData)
	if exists {
		redriveAllowPolicyBytes, err := json.Marshal(redriveAllowPolicy)
		if err != nil {
			return nil, false, err
		}
		input.Attributes["RedriveAllowPolicy"] = string(redriveAllowPolicyBytes)
	}

	policy, exists := pluginutils.GetValueByPath("$.policy", specData)
	if exists {
		policyBytes, err := json.Marshal(policy)
		if err != nil {
			return nil, false, err
		}
		input.Attributes["Policy"] = string(policyBytes)
	}

	valueSetters := []*pluginutils.ValueSetter[*sqs.CreateQueueInput]{
		pluginutils.NewValueSetter(
			"$.queueName",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.QueueName = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.delaySeconds",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["DelaySeconds"] = strconv.Itoa(core.IntValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.messageRetentionPeriod",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["MessageRetentionPeriod"] = strconv.Itoa(core.IntValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.visibilityTimeoutSeconds",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["VisibilityTimeout"] = strconv.Itoa(core.IntValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.fifoQueue",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["FifoQueue"] = strconv.FormatBool(core.BoolValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.fifoThroughputLimit",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["FifoThroughputLimit"] = core.StringValue(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.contentBasedDeduplication",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["ContentBasedDeduplication"] = strconv.FormatBool(core.BoolValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.deduplicationScope",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["DeduplicationScope"] = core.StringValue(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.maximumMessageSize",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["MaximumMessageSize"] = strconv.Itoa(core.IntValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.kmsMasterKeyId",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["KmsMasterKeyId"] = core.StringValue(value)
			},
		),
		pluginutils.NewValueSetter(
			"$.kmsDataKeyReusePeriodSeconds",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["KmsDataKeyReusePeriodSeconds"] = strconv.Itoa(core.IntValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.sqsManagedSseEnabled",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Attributes["SqsManagedSseEnabled"] = strconv.FormatBool(core.BoolValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.tags",
			func(value *core.MappingNode, input *sqs.CreateQueueInput) {
				input.Tags = utils.ToTagsMap(value)
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
