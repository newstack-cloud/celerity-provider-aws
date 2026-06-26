package overlays

import "sort"

// Holds the set of Bluelink resource types whose computed fields
// only become fully available once the resource has finished provisioning (e.g. an
// RDS DB instance endpoint, or a flex VPC's NAT gateways becoming available).
//
// Two things key off this set:
//   - Producers: such a resource's Create returns as soon as its identifier is known
//     (it does not block for the whole multi-minute provision); its computed fields are
//     finalised at stabilisation instead.
//   - Consumers: every generated Cloud Control resource declares this set as its
//     stabilised dependencies, so any resource that depends on one of these types waits
//     for it to stabilise before deploying. This makes sure that it resolves the late computed
//     fields. Hand-written resources can do the same.
var stabiliseRequired = map[string]bool{}

// RegisterStabiliseRequired marks a Bluelink resource type as slow to stabilise.
// This is intended to be called from per-type init functions.
func RegisterStabiliseRequired(blueprintType string) {
	stabiliseRequired[blueprintType] = true
}

// IsStabiliseRequired reports whether a resource type is registered as slow to
// stabilise.
func IsStabiliseRequired(blueprintType string) bool {
	return stabiliseRequired[blueprintType]
}

// StabiliseRequiredTypes returns the registered stabilise-required resource types,
// sorted for deterministic output. Used to populate a resource's stabilised
// dependencies so consumers wait for these types to stabilise.
func StabiliseRequiredTypes() []string {
	types := make([]string, 0, len(stabiliseRequired))
	for blueprintType := range stabiliseRequired {
		types = append(types, blueprintType)
	}
	sort.Strings(types)
	return types
}
