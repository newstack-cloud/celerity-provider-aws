package linkutils

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

// CapabilityNetworkAttached is the guarantee that a caller's VPC attachment is in place
// and readable from its live state.
//
// The flex VPC placement link establishes it; every link that calls
// ReconcileLinkNetworking depends on it, because that function reads the caller's live
// attachment to decide whether a VPC endpoint or a security group rule is needed. An
// access link that runs first sees an unattached caller, opens nothing, and reports
// success, leaving a deployment that completes but does not work at runtime.
//
// Both sides reference this constant so that a mismatch is a compile error rather than a
// link that silently deploys unordered.
const CapabilityNetworkAttached = "aws.flexvpc/network-attached"

// NetworkAttachedProvided returns the capability declaration for a link that attaches a
// caller to a VPC, where callerSide names the side of the relationship the caller sits
// on.
func NetworkAttachedProvided(
	callerSide provider.LinkPriorityResource,
) []provider.LinkCapability {
	return []provider.LinkCapability{
		{
			Name:     CapabilityNetworkAttached,
			Resource: callerSide,
		},
	}
}

// NetworkAttachedRequired returns the capability declaration for a link that reads a
// caller's live VPC attachment, where callerSide names the side of the relationship the
// caller sits on.
//
// MustExist is deliberately left false. A caller that was never placed in a VPC has no
// placement link, and its access links must still deploy: ReconcileLinkNetworking is a
// no-op for an unattached caller, which reaches AWS services over the public internet.
func NetworkAttachedRequired(
	callerSide provider.LinkPriorityResource,
) []provider.LinkCapability {
	return []provider.LinkCapability{
		{
			Name:     CapabilityNetworkAttached,
			Resource: callerSide,
		},
	}
}
