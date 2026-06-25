//go:build unit

package gen

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/require"
)

// TestSQSQueueGeneratedSchema validates the generated AWS::SQS::Queue schema against
// the native implementation as an oracle: the expected property surface is present,
// and the overlay has refined the free-form CloudFormation constructs.
func TestSQSQueueGeneratedSchema(t *testing.T) {
	schema := sqsQueueSpecSchema(t)

	// Property coverage: every field the native resource exposes should be present.
	expectedFields := []string{
		"queueName", "fifoQueue", "visibilityTimeout", "delaySeconds",
		"maximumMessageSize", "messageRetentionPeriod", "receiveMessageWaitTimeSeconds",
		"redrivePolicy", "redriveAllowPolicy", "kmsMasterKeyId",
		"kmsDataKeyReusePeriodSeconds", "sqsManagedSseEnabled", "contentBasedDeduplication",
		"deduplicationScope", "fifoThroughputLimit", "tags", "arn", "queueUrl",
	}
	for _, field := range expectedFields {
		require.Contains(t, schema.Attributes, field, "generated schema should expose %s", field)
	}

	// RedrivePolicy is a free-form object, modelled as an open map (string keys,
	// arbitrary values) rather than a JSON string, so it can be authored natively.
	redrive := schema.Attributes["redrivePolicy"]
	require.Equal(t, provider.ResourceDefinitionsSchemaTypeMap, redrive.Type)
	require.Nil(t, redrive.MapValues, "an open map has no declared value schema")

	// Overlay: numeric bounds documented only in prose are restored.
	visibility := schema.Attributes["visibilityTimeout"]
	require.NotNil(t, visibility.Minimum, "visibilityTimeout should have a minimum bound from the overlay")
	require.NotNil(t, visibility.Maximum, "visibilityTimeout should have a maximum bound from the overlay")

	// Stamping carried through from the CloudFormation pointer lists.
	require.True(t, schema.Attributes["queueName"].MustRecreate, "createOnly queueName should be MustRecreate")
	require.True(t, schema.Attributes["arn"].Computed, "readOnly arn should be Computed")
}

func sqsQueueSpecSchema(t *testing.T) *provider.ResourceDefinitionsSchema {
	t.Helper()
	resource := sqsQueueResource(
		func(*aws.Config, provider.Context) cloudcontrolservice.Service { return nil },
		func(*aws.Config, provider.Context) resgrouptagservice.Service { return nil },
		nil,
	)
	out, err := resource.GetSpecDefinition(context.Background(), &provider.ResourceGetSpecDefinitionInput{})
	require.NoError(t, err)
	return out.SpecDefinition.Schema
}
