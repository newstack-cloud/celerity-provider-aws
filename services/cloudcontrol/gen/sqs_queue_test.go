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

func TestSQSQueueResourceRegistration(t *testing.T) {
	resources := GeneratedResources(
		func(*aws.Config, provider.Context) cloudcontrolservice.Service { return nil },
		func(*aws.Config, provider.Context) resgrouptagservice.Service { return nil },
		nil,
	)

	resource, ok := resources["aws/sqs/queue"]
	require.True(t, ok, "aws/sqs/queue should be registered")

	ctx := context.Background()

	typeOutput, err := resource.GetType(ctx, &provider.ResourceGetTypeInput{})
	require.NoError(t, err)
	require.Equal(t, "aws/sqs/queue", typeOutput.Type)

	specOutput, err := resource.GetSpecDefinition(ctx, &provider.ResourceGetSpecDefinitionInput{})
	require.NoError(t, err)
	require.NotNil(t, specOutput.SpecDefinition)

	schema := specOutput.SpecDefinition.Schema
	require.Equal(t, provider.ResourceDefinitionsSchemaTypeObject, schema.Type)
	require.Equal(t, "queueUrl", specOutput.SpecDefinition.IDField)
	require.Equal(t, provider.TaggingSupportFull, specOutput.SpecDefinition.TaggingSupport)

	// The engine must have injected its internal bookkeeping fields as computed,
	// drift-ignored strings.
	for _, internal := range []string{"__ccRequestToken", "__ccPrimaryIdentifier"} {
		field, exists := schema.Attributes[internal]
		require.Truef(t, exists, "internal field %s should be present", internal)
		require.True(t, field.Computed, "internal field should be computed")
		require.True(t, field.IgnoreDrift, "internal field should ignore drift")
	}

	// Every object-type schema node must carry a Label (repo-wide requirement).
	assertObjectLabels(t, schema)
}

func assertObjectLabels(t *testing.T, schema *provider.ResourceDefinitionsSchema) {
	t.Helper()
	if schema == nil {
		return
	}
	if schema.Type == provider.ResourceDefinitionsSchemaTypeObject {
		require.NotEmpty(t, schema.Label, "object schema node must have a label")
		for _, attr := range schema.Attributes {
			assertObjectLabels(t, attr)
		}
	}
	assertObjectLabels(t, schema.Items)
	assertObjectLabels(t, schema.MapValues)
	for _, branch := range schema.OneOf {
		assertObjectLabels(t, branch)
	}
}
