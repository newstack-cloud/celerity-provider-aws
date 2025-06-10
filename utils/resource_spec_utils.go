package utils

import (
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

// GetCurrentResourceStateSpecData returns the spec data for the current
// resource state from the changes object.
func GetCurrentResourceStateSpecData(changes *provider.Changes) *core.MappingNode {
	if changes == nil {
		return &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}
	}

	appliedResourceInfo := changes.AppliedResourceInfo
	if appliedResourceInfo.CurrentResourceState == nil {
		return &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}
	}

	return appliedResourceInfo.CurrentResourceState.SpecData
}

// GetResolvedResourceSpecData returns the resolved spec data for the resource
// from changes.
func GetResolvedResourceSpecData(changes *provider.Changes) *core.MappingNode {
	if changes == nil || changes.AppliedResourceInfo.ResourceWithResolvedSubs == nil {
		return &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}
	}

	return changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
}

// SpecValueSetter is a helper struct that can be used to set a value in a
// from a resource spec to a target API-specific struct used to update
// or create the resource in the upstream provider.
type SpecValueSetter[Target any] struct {
	PathInSpec   string
	SetValueFunc func(
		value *core.MappingNode,
		target Target,
	)
	didSet bool
}

func (u *SpecValueSetter[Target]) Set(
	value *core.MappingNode,
	target Target,
) {
	value, hasValue := GetSpecValueByPath(u.PathInSpec, value)
	if !hasValue {
		return
	}

	u.SetValueFunc(value, target)
	u.didSet = true
}

func (u *SpecValueSetter[Target]) DidSet() bool {
	return u.didSet
}

// GetSpecValueByPath is a helper function to extract a value from a resource
// spec that is a light weight wrapper around core.GetPathValue.
// Unlike core.GetPathValue, this function will not return an error,
// instead it will return nil and false if the value is not found
// or the provided path is not valid.
func GetSpecValueByPath(
	fieldPath string,
	specData *core.MappingNode,
) (*core.MappingNode, bool) {
	value, err := core.GetPathValue(
		fieldPath,
		specData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, false
	}

	return value, true
}
