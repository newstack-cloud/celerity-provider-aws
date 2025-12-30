package lambda

import (
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func lambdaSchemaTags(resourceType string) *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:             provider.ResourceDefinitionsSchemaTypeArray,
		SortArrayByField: "key",
		Description:      fmt.Sprintf("A list of tags to apply to the %s.", resourceType),
		FormattedDescription: fmt.Sprintf(
			"A list of [tags](https://docs.aws.amazon.com/lambda/latest/dg/tagging.html) "+
				"to apply to the %s.",
			resourceType,
		),
		Items: &provider.ResourceDefinitionsSchema{
			Type:                 provider.ResourceDefinitionsSchemaTypeObject,
			Label:                "Tag",
			Description:          fmt.Sprintf("A tag to apply to the %s.", resourceType),
			FormattedDescription: fmt.Sprintf("A [tag](https://docs.aws.amazon.com/lambda/latest/dg/tagging.html) to apply to the %s.", resourceType),
			Required:             []string{"key", "value"},
			Attributes: map[string]*provider.ResourceDefinitionsSchema{
				"key": {
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The key of the tag.",
					MinLength:   1,
					MaxLength:   128,
				},
				"value": {
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The value of the tag.",
					MinLength:   0,
					MaxLength:   256,
				},
			},
		},
	}
}
