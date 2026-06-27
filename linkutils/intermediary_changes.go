package linkutils

import (
	"fmt"
	"maps"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// This file projects a link's managed intermediary resources into the link's linkData
// so their create/update/destroy is surfaced as ordinary link field changes at change
// staging (plan/preview) time. Intermediaries are not spec fields of either linked
// resource, so without this projection an intermediary-only link reports no changes yet
// silently deploys a resource (e.g. an aws/lambda/permission). The projection is modelled
// as an "intermediaries" map keyed by the intermediary's deterministic ResourceID, so the
// same identity is staged, persisted and reconciled.

// IntermediaryLinkDataKey is the top-level linkData key under which a link's managed
// intermediaries are projected.
const IntermediaryLinkDataKey = "intermediaries"

// IntermediaryIdentity is the deploy-stable identity of a single link-owned intermediary.
// It is computed identically at stage and update time so the surfaced change matches what
// is deployed.
type IntermediaryIdentity struct {
	ResourceType string
	ResourceID   string
	ResourceName string
}

// DerivedLeaf is an intermediary linkData leaf whose value comes from a linked resource's
// spec field (so it may be resolvable now or known on deploy).
type DerivedLeaf struct {
	// Leaf is the linkData leaf name under the intermediary entry, e.g. "sourceArn".
	Leaf string
	// ResourceChanges is the linked resource (ResourceAChanges or ResourceBChanges) the
	// value is read from.
	ResourceChanges *provider.Changes
	// ResourceSpecPath is the path of the source value in the resource spec, e.g.
	// "$.spec.arn".
	ResourceSpecPath string
}

// StageIntermediary describes one link-owned intermediary to project into link changes
// at stage time.
type StageIntermediary struct {
	Identity      IntermediaryIdentity
	DerivedLeaves []DerivedLeaf
}

// CollectIntermediaryChanges projects a managed intermediary into the given LinkChanges by
// diffing its linkData entry against the prior linkData. The constant "resourceType" leaf
// is collected directly; each derived leaf is collected from its source resource spec so
// the existing helper routes unresolved/computed values to FieldChangesKnownOnDeploy.
func CollectIntermediaryChanges(
	currentLinkData *core.MappingNode,
	out *provider.LinkChanges,
	intermediary StageIntermediary,
) error {
	err := linkhelpers.CollectLinkDataChanges(
		IntermediaryLeafPath(intermediary.Identity.ResourceID, "resourceType"),
		currentLinkData,
		out,
		core.MappingNodeFromString(intermediary.Identity.ResourceType),
	)
	if err != nil {
		return err
	}

	for _, leaf := range intermediary.DerivedLeaves {
		err := linkhelpers.CollectChanges(
			leaf.ResourceSpecPath,
			IntermediaryLeafPath(intermediary.Identity.ResourceID, leaf.Leaf),
			currentLinkData,
			leaf.ResourceChanges,
			out,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeployedIntermediary is a resolved intermediary to persist into link data at update time.
// Leaves excludes "resourceType", which is derived from Identity.
type DeployedIntermediary struct {
	Identity IntermediaryIdentity
	Leaves   map[string]*core.MappingNode
}

// IntermediaryLinkData builds the linkData node ({intermediaries: {<id>: {resourceType,
// ...leaves}}}) a link should return from UpdateIntermediaryResources, so that what is
// persisted matches the projection diffed at stage time.
func IntermediaryLinkData(intermediaries ...DeployedIntermediary) *core.MappingNode {
	entries := map[string]*core.MappingNode{}
	for _, intermediary := range intermediaries {
		fields := map[string]*core.MappingNode{
			"resourceType": core.MappingNodeFromString(intermediary.Identity.ResourceType),
		}
		maps.Copy(fields, intermediary.Leaves)
		entries[intermediary.Identity.ResourceID] = &core.MappingNode{Fields: fields}
	}

	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			IntermediaryLinkDataKey: {
				Fields: entries,
			},
		},
	}
}

// IntermediaryLeafPath returns the linkData path of a single intermediary leaf, e.g.
// $["intermediaries"]["<resourceID>"]["sourceArn"].
func IntermediaryLeafPath(resourceID, leaf string) string {
	return fmt.Sprintf("$[%q][%q][%q]", IntermediaryLinkDataKey, resourceID, leaf)
}
