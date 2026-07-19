package linkutils

import (
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// PhysicalResourceName reads a physical name field (e.g. "bucketName" or
// "tableName") from a linked resource's current state spec data, falling back
// to deriving the name from the state's "arn" field when no name is present.
//
// The fallback covers auto-named resources: a computed-when-omitted name field
// may not be visible in the linked resource's state at link-update time, but
// the ARN — whose final path segment is the physical name — is a schema-level
// computed field that is always captured at deploy time.
func PhysicalResourceName(resourceInfo *provider.ResourceInfo, nameField string) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(resourceInfo)
	if nameNode, has := pluginutils.GetValueByPath("$."+nameField, spec); has {
		if name := core.StringValue(nameNode); name != "" {
			return name, true
		}
	}

	arnNode, hasARN := pluginutils.GetValueByPath("$.arn", spec)
	if !hasARN {
		return "", false
	}
	if name := nameFromARN(core.StringValue(arnNode)); name != "" {
		return name, true
	}
	return "", false
}

// The name is the final path segment of the ARN's resource portion
// ("arn:aws:dynamodb:eu-west-1:123456789012:table/orders" -> "orders"), or the
// whole resource portion when it has no path ("arn:aws:s3:::my-bucket" ->
// "my-bucket").
func nameFromARN(arn string) string {
	parts := strings.SplitN(arn, ":", 6)
	if len(parts) < 6 {
		return ""
	}
	resource := parts[5]
	if idx := strings.LastIndex(resource, "/"); idx >= 0 {
		return resource[idx+1:]
	}
	return resource
}
