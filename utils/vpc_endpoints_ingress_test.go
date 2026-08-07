//go:build unit

package utils

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/require"
)

// EC2 returns a group reference in a non-default VPC with GroupId populated and
// GroupName empty, so matching on the name reports every rule as absent. That made the
// endpoint idempotency check permanently false: it would keep trying to add an ingress
// rule it had already added.
func TestHasIngressFromSecurityGroupIDMatchesByIDNotName(t *testing.T) {
	securityGroup := &ec2types.SecurityGroup{
		IpPermissions: []ec2types.IpPermission{
			{
				UserIdGroupPairs: []ec2types.UserIdGroupPair{
					// As returned for a VPC group reference: an ID and no name.
					{GroupId: aws.String("sg-caller")},
				},
			},
		},
	}

	require.True(t, HasIngressFromSecurityGroupID(securityGroup, "sg-caller"))
	require.False(t, HasIngressFromSecurityGroupID(securityGroup, "sg-other"))
}

func TestHasIngressFromSecurityGroupIDIsFalseWithNoRules(t *testing.T) {
	require.False(
		t,
		HasIngressFromSecurityGroupID(&ec2types.SecurityGroup{}, "sg-caller"),
	)
}

// EC2 matches a tag with a "tag:" prefix on the filter name. Passing the bare key makes
// it reject the entire call with InvalidParameterValue.
//
// Both call sites did exactly that, and neither was reachable until the links stopped
// returning early on destroy, at which point every teardown failed on a malformed
// describe rather than on anything to do with networking.
func TestCreateTagFilterBluelinkServicePrefixesTheTagKey(t *testing.T) {
	filter := CreateTagFilterBluelinkService("com.amazonaws.eu-west-2.sqs")

	require.Equal(t, "tag:"+TagBluelinkService, aws.ToString(filter.Name))
	require.Equal(t, []string{"com.amazonaws.eu-west-2.sqs"}, filter.Values)
	require.NotEqual(
		t,
		TagBluelinkService,
		aws.ToString(filter.Name),
		"a bare tag key is not a valid EC2 filter name",
	)
}
