package testutils

import (
	"slices"

	"github.com/newstack-cloud/celerity/libs/blueprint/core"
)

// ShallowCopy creates a shallow copy of a map of MappingNodes, excluding
// the keys in the ignoreKeys slice.
func ShallowCopy(
	fields map[string]*core.MappingNode,
	ignoreKeys ...string,
) map[string]*core.MappingNode {
	copy := make(map[string]*core.MappingNode, len(fields))
	for k, v := range fields {
		if !slices.Contains(ignoreKeys, k) {
			copy[k] = v
		}
	}
	return copy
}
