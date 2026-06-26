package overlays

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

func init() {
	// An EventBridge connection's AuthParameters is a polymorphic, secret-bearing
	// CloudFormation object (a oneOf of mutually exclusive API-key / basic / OAuth auth
	// shapes, with nested write-only credentials). The generator cannot model it as a
	// clean typed schema and falls back to a plain string, but Cloud Control validates
	// desired state against the CloudFormation schema where AuthParameters is a JSON
	// object, so a string is rejected. Re-typing it as an open map lets it be authored as
	// a native object (better than a serialised JSON string) whose keys and values are
	// passed through to Cloud Control as they are.
	Register("aws/events/connection", func(schema *provider.ResourceDefinitionsSchema) *provider.ResourceDefinitionsSchema {
		retypeAtPath(schema, "authParameters", func(existing *provider.ResourceDefinitionsSchema) *provider.ResourceDefinitionsSchema {
			return &provider.ResourceDefinitionsSchema{
				Type:        provider.ResourceDefinitionsSchemaTypeMap,
				Label:       "Auth Parameters",
				Description: existing.Description,
				Nullable:    true,
			}
		})
		return schema
	})
}
