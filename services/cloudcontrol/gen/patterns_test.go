//go:build unit

package gen

import (
	"context"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/require"
)

// CloudFormation patterns are ECMA-flavoured while the blueprint framework
// validates with Go's RE2; a pattern RE2 cannot compile hard-fails spec
// validation for the whole resource. Every emitted pattern must compile.
func TestGeneratedSchemaPatternsCompile(t *testing.T) {
	resources := GeneratedResources(
		func(*aws.Config, provider.Context) cloudcontrolservice.Service { return nil },
		func(*aws.Config, provider.Context) resgrouptagservice.Service { return nil },
		nil,
	)
	require.NotEmpty(t, resources)

	ctx := context.Background()
	for resourceType, resource := range resources {
		specOutput, err := resource.GetSpecDefinition(ctx, &provider.ResourceGetSpecDefinitionInput{})
		require.NoError(t, err)
		require.NotNil(t, specOutput.SpecDefinition)

		assertPatternsCompile(t, resourceType, specOutput.SpecDefinition.Schema)
	}
}

func assertPatternsCompile(t *testing.T, resourceType string, schema *provider.ResourceDefinitionsSchema) {
	t.Helper()
	if schema == nil {
		return
	}
	if schema.Pattern != "" {
		_, err := regexp.Compile(schema.Pattern)
		require.NoErrorf(
			t,
			err,
			"%s: pattern %q does not compile under Go's RE2 regexp",
			resourceType,
			schema.Pattern,
		)
	}
	for _, attr := range schema.Attributes {
		assertPatternsCompile(t, resourceType, attr)
	}
	assertPatternsCompile(t, resourceType, schema.Items)
	assertPatternsCompile(t, resourceType, schema.MapValues)
	for _, branch := range schema.OneOf {
		assertPatternsCompile(t, resourceType, branch)
	}
}
