package utils

import (
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// ExtractARNFromCurrentState extracts the ARN from the current state spec data,
// this only works when the "arn" field is present as a a top-level field in provided
// spec data.
func ExtractARNFromCurrentState(
	currentStateSpecData *core.MappingNode,
	context string,
) (string, error) {
	if currentStateSpecData == nil {
		return "", fmt.Errorf("current state spec data is required for %s", context)
	}
	arn, hasArn := pluginutils.GetValueByPath("$.arn", currentStateSpecData)
	if !hasArn {
		return "", fmt.Errorf("ARN is required for %s", context)
	}
	return core.StringValue(arn), nil
}

// ExtractARNFromResourceInfo extracts the ARN from the resource info,
// this only works when the "arn" field is present as a a top-level field in provided
// spec data.
func ExtractARNFromResourceInfo(resourceInfo *provider.ResourceInfo) (string, bool) {
	specData := pluginutils.GetCurrentStateSpecDataFromResourceInfo(
		resourceInfo,
	)
	arn, hasARN := pluginutils.GetValueByPath(
		"$.arn",
		specData,
	)
	return core.StringValue(arn), hasARN
}
