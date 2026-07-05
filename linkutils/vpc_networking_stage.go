package linkutils

import (
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// NetworkAccessFieldName is the synthetic link-data field name used to signal, in staged
// changes, that a VPC-attached caller's network access to a link target will be configured at
// deploy time (a VPC endpoint or a security-group rule opened by ActivateLinkNetworking).
func NetworkAccessFieldName(callerResourceName string) string {
	return fmt.Sprintf("%sNetworkAccess", callerResourceName)
}

// StageNetworkAccessKnownOnDeploy appends a known-on-deploy network-access signal to the link
// changes when the caller (resource A) is detectably attached to a VPC.
//
// Networking activation (interface/gateway VPC endpoints, security-group rules) happens at
// deploy and is a no-op for callers not attached to a VPC, and its concrete values (endpoint
// id, rule) are only known at deploy. So this is a best-effort, value-less signal: it fires
// when the caller's current state or resolved spec already shows a VPC attachment (an explicit
// vpcConfig, or one drift-mapped from a placement link on a re-deploy), and stays silent
// otherwise.
func StageNetworkAccessKnownOnDeploy(
	callerChanges *provider.Changes,
	changes *provider.LinkChanges,
) {
	if callerChanges == nil || !callerHasVPCAttachment(callerChanges) {
		return
	}
	changes.FieldChangesKnownOnDeploy = append(
		changes.FieldChangesKnownOnDeploy,
		NetworkAccessFieldName(pluginutils.GetResourceName(&callerChanges.AppliedResourceInfo)),
	)
}

func callerHasVPCAttachment(callerChanges *provider.Changes) bool {
	if callerChanges.AppliedResourceInfo.CurrentResourceState != nil &&
		hasVPCSubnets(callerChanges.AppliedResourceInfo.CurrentResourceState.SpecData) {
		return true
	}
	if callerChanges.AppliedResourceInfo.ResourceWithResolvedSubs != nil &&
		hasVPCSubnets(callerChanges.AppliedResourceInfo.ResourceWithResolvedSubs.Spec) {
		return true
	}
	return false
}

func hasVPCSubnets(spec *core.MappingNode) bool {
	node, has := pluginutils.GetValueByPath("$.vpcConfig.subnetIds", spec)
	return has && node != nil && len(node.Items) > 0
}
