package cloudcontrol

import (
	"encoding/json"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func mappingNodeToCFNJSON(node *core.MappingNode) (string, error) {
	if node == nil {
		return "{}", nil
	}

	data, err := json.Marshal(node)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Parses a Cloud Control properties JSON string into a mapping
// node. It deliberately avoids MappingNode.UnmarshalJSON (which parses ${..} as
// substitutions) by unmarshalling into plain Go values first, CFN property values
// can legitimately contain ${..} (e.g. IAM policy documents).
func cfnJSONToMappingNode(propertiesJSON string) (*core.MappingNode, error) {
	if propertiesJSON == "" {
		return &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}, nil
	}

	var raw any
	if err := json.Unmarshal([]byte(propertiesJSON), &raw); err != nil {
		return nil, err
	}

	return pluginutils.AnyToMappingNode(raw)
}
