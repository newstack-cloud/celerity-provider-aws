package flex

import "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/overlays"

// The flex VPC is slow to stabilise. It creates NAT gateways that take minutes to
// become available (its Stabilised polls for this). Registering it as
// stabilise-required means any resource placed in the VPC waits for it to stabilise
// before deploying. The registry is shared with the generated Cloud Control engine so
// both honour the same set of slow-to-stabilise types.
func init() {
	overlays.RegisterStabiliseRequired("aws/flex/vpc")
}
