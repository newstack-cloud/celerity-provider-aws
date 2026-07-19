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

// Auto-named identifier fields (optional, non-read-only primary identifier
// components) must be computed-when-omitted so references to them resolve as
// known-after-deploy when a blueprint leaves the name to AWS.
func TestAutoNamedIdentifierFieldsAreComputedWhenOmitted(t *testing.T) {
	resources := GeneratedResources(
		func(*aws.Config, provider.Context) cloudcontrolservice.Service { return nil },
		func(*aws.Config, provider.Context) resgrouptagservice.Service { return nil },
		nil,
	)

	expectations := map[string]string{
		"aws/s3/bucket":       "bucketName",
		"aws/dynamodb/table":  "tableName",
		"aws/lambda/function": "functionName",
		"aws/iam/role":        "roleName",
		"aws/logs/logGroup":   "logGroupName",
	}

	ctx := context.Background()
	for resourceType, field := range expectations {
		resource, ok := resources[resourceType]
		require.Truef(t, ok, "%s should be registered", resourceType)

		specOutput, err := resource.GetSpecDefinition(ctx, &provider.ResourceGetSpecDefinitionInput{})
		require.NoError(t, err)

		attribute := specOutput.SpecDefinition.Schema.Attributes[field]
		require.NotNilf(t, attribute, "%s should have a %s attribute", resourceType, field)
		require.Truef(
			t,
			attribute.ComputedWhenOmitted,
			"%s %s should be computed-when-omitted",
			resourceType,
			field,
		)
		require.Falsef(t, attribute.Computed, "%s %s must stay user-settable", resourceType, field)
	}
}
