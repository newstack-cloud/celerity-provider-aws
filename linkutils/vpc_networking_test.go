//go:build unit

package linkutils

import (
	"context"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/assert"
)

func TestCallerNetworking_attached(t *testing.T) {
	cases := []struct {
		name     string
		caller   CallerNetworking
		attached bool
	}{
		{
			name:     "fully attached",
			caller:   CallerNetworking{VPCID: "vpc-1", SubnetIDs: []string{"subnet-1"}, SecurityGroupIDs: []string{"sg-1"}},
			attached: true,
		},
		{name: "no VPC id", caller: CallerNetworking{SubnetIDs: []string{"subnet-1"}, SecurityGroupIDs: []string{"sg-1"}}},
		{name: "no subnets", caller: CallerNetworking{VPCID: "vpc-1", SecurityGroupIDs: []string{"sg-1"}}},
		{name: "no security groups", caller: CallerNetworking{VPCID: "vpc-1", SubnetIDs: []string{"subnet-1"}}},
		{name: "empty", caller: CallerNetworking{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.attached, tc.caller.attached())
		})
	}
}

// A caller that is not VPC-attached reaches AWS services over the public API, so
// ReconcileLinkNetworking returns the output unchanged without resolving the flex VPC or
// touching the EC2 service (a nil service here would panic if it were used).
func TestActivateLinkNetworking_noOpWhenNotAttached(t *testing.T) {
	output := &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData: core.MappingNodeFields("marker", core.MappingNodeFromString("unchanged")),
	}

	result, err := ReconcileLinkNetworking(
		context.Background(),
		nil, // EC2 service must not be used on the no-op path.
		&provider.LinkUpdateIntermediaryResourcesInput{},
		NetworkingActivation{Caller: CallerNetworking{}, Region: "us-west-2", AWSService: "sns"},
		output,
	)

	assert.NoError(t, err)
	assert.Same(t, output, result)
}
