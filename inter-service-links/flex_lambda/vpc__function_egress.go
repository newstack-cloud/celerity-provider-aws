package flexlambda

import (
	"fmt"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// What a placed function is allowed to reach outside the VPC.
//
// Three states rather than two, because a public subnet reaches the internet over
// IPv6 and not over IPv4: a VPC-attached function is never assigned a public IPv4
// address, so the internet gateway is unusable for IPv4 no matter which tier the
// function sits in. Collapsing that into "internet" would promise IPv4 that never
// works, and into "none" would withhold IPv6 that does.
type egressReach int

const (
	egressNone egressReach = iota
	egressIPv6Only
	egressFull
)

// The resolved outbound configuration for a placed function.
type egressPlan struct {
	reach egressReach
	// Set only when the author named specific destinations, in which case they are
	// authorised instead of the whole internet.
	cidrs []string
}

const (
	egressAnnotationKey = "aws.flexvpc.lambda.egress"
	egressValueInternet = "internet"
	egressValueNone     = "none"
)

// Resolves what the function may reach outside the VPC from the author's annotation
// and the VPC's own topology.
//
// Unset is not the same as "none": left unset the function keeps whatever the
// topology can deliver, so placing a workload in a VPC never silently takes away
// outbound access it had before. Setting the annotation only ever narrows.
func resolveEgressPlan(
	resourceInfo *provider.ResourceInfo,
	flexVPCSpecData *core.MappingNode,
	subnetType string,
) (*egressPlan, error) {
	available := availableEgressReach(flexVPCSpecData, subnetType)

	declared, hasDeclared := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{Key: egressAnnotationKey},
	)
	if !hasDeclared || strings.TrimSpace(declared) == "" {
		return &egressPlan{reach: available}, nil
	}

	declared = strings.TrimSpace(declared)
	if declared == egressValueNone {
		return &egressPlan{reach: egressNone}, nil
	}

	if declared == egressValueInternet {
		if available == egressNone {
			return nil, fmt.Errorf(
				"the %s annotation requests internet egress, but the linked flex VPC provides no "+
					"outbound path for a function in a %q subnet",
				egressAnnotationKey,
				subnetType,
			)
		}
		return &egressPlan{reach: available}, nil
	}

	cidrs := parseEgressCIDRs(declared)
	if len(cidrs) == 0 {
		return nil, fmt.Errorf(
			"the %s annotation must be %q, %q, or a comma-separated list of CIDR ranges, got %q",
			egressAnnotationKey,
			egressValueInternet,
			egressValueNone,
			declared,
		)
	}
	if available == egressNone {
		return nil, fmt.Errorf(
			"the %s annotation requests egress to %s, but the linked flex VPC provides no "+
				"outbound path for a function in a %q subnet",
			egressAnnotationKey,
			strings.Join(cidrs, ", "),
			subnetType,
		)
	}

	return &egressPlan{reach: available, cidrs: cidrs}, nil
}

// What the VPC can actually deliver for a function in the given subnet tier.
//
// A private subnet reaches IPv4 only through a NAT gateway, so its presence in the
// VPC's computed gateways is what separates full internet from none. A public subnet
// has no NAT gateway by design and reaches IPv6 through the internet gateway.
func availableEgressReach(flexVPCSpecData *core.MappingNode, subnetType string) egressReach {
	if subnetType == "public" {
		if hasInternetGateway(flexVPCSpecData) {
			return egressIPv6Only
		}
		return egressNone
	}

	if hasNATGateway(flexVPCSpecData) {
		return egressFull
	}

	// A private subnet with no NAT gateway is the isolated preset, which routes
	// neither address family to the internet.
	return egressNone
}

func hasNATGateway(flexVPCSpecData *core.MappingNode) bool {
	natGateways, ok := pluginutils.GetValueByPath("$.gateways.natGateways", flexVPCSpecData)
	return ok && natGateways != nil && len(natGateways.Items) > 0
}

func hasInternetGateway(flexVPCSpecData *core.MappingNode) bool {
	igw, ok := pluginutils.GetValueByPath("$.gateways.internetGatewayId", flexVPCSpecData)
	return ok && core.StringValue(igw) != ""
}

func parseEgressCIDRs(declared string) []string {
	cidrs := []string{}
	for part := range strings.SplitSeq(declared, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Anything that is not a CIDR is a typo rather than a destination; the caller
		// turns an empty result into an error naming the accepted forms.
		if !strings.Contains(part, "/") {
			return nil
		}
		cidrs = append(cidrs, part)
	}

	return cidrs
}
