//go:build unit

package flexlambda

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/flex"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func sgTestSuite(t *testing.T) *suite.Suite {
	s := new(suite.Suite)
	s.SetT(t)
	return s
}

func testWorkloadIdentity() *workloadGroupIdentity {
	return &workloadGroupIdentity{
		vpcID:        "vpc-1",
		vpcName:      "orders-vpc",
		functionName: "getOrderFunction",
		instanceID:   "instance-1",
	}
}

// A new group is created in the VPC and carries the tags every later lookup filters
// on. Without the workload and owner tags the group cannot be found again, so the
// next deploy would create a second one and the first would leak.
func TestWorkloadSecurityGroupCreatedWithIdentifyingTags(t *testing.T) {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{}),
		ec2mock.WithCreateSecurityGroupOutput(&ec2.CreateSecurityGroupOutput{
			GroupId: aws.String("sg-workload"),
		}),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
	)

	identity := testWorkloadIdentity()
	groupID, err := resolveWorkloadSecurityGroup(context.Background(), service, identity)
	require.NoError(t, err)
	require.Equal(t, "sg-workload", groupID)

	service.AssertCalledWith(
		sgTestSuite(t),
		"CreateSecurityGroup",
		0,
		plugintestutils.Any,
		&ec2.CreateSecurityGroupInput{
			GroupName:   aws.String("bluelink-flexvpc-orders-vpc-getOrderFunction-instance-1"),
			Description: aws.String(identity.description()),
			VpcId:       aws.String("vpc-1"),
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeSecurityGroup,
					Tags:         identity.tags(),
				},
			},
		},
	)
}

// The group must not carry the tag the VPC resource filters its own securityGroups
// output on, or the VPC would adopt every workload group and hand them out to
// unrelated callers as though they were the base group.
func TestWorkloadSecurityGroupIsNotTaggedAsTheVPCBaseGroup(t *testing.T) {
	tags := testWorkloadIdentity().tags()

	keys := map[string]string{}
	for _, tag := range tags {
		keys[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	require.NotContains(t, keys, flex.TagFlexVPCSecurityGroup)
	require.Equal(t, "getOrderFunction", keys[flex.TagFlexVPCWorkloadSecurityGroup])
	require.Equal(t, "instance-1", keys[flex.TagFlexVPCWorkloadOwner])
	require.Equal(t, "orders-vpc", keys[flex.TagFlexVPCName])
}

// EC2 attaches an allow-all egress rule to every new group. Leaving it would defeat
// the point of a per-workload group: the function could reach anything in the VPC and
// the internet regardless of what it links to.
func TestWorkloadSecurityGroupHasDefaultEgressRevoked(t *testing.T) {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{}),
		ec2mock.WithCreateSecurityGroupOutput(&ec2.CreateSecurityGroupOutput{
			GroupId: aws.String("sg-workload"),
		}),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
	)

	_, err := resolveWorkloadSecurityGroup(context.Background(), service, testWorkloadIdentity())
	require.NoError(t, err)

	service.AssertCalledWith(
		sgTestSuite(t),
		"RevokeSecurityGroupEgress",
		0,
		plugintestutils.Any,
		&ec2.RevokeSecurityGroupEgressInput{
			GroupId: aws.String("sg-workload"),
			IpPermissions: []ec2types.IpPermission{
				{
					IpProtocol: aws.String("-1"),
					IpRanges:   []ec2types.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
					Ipv6Ranges: []ec2types.Ipv6Range{{CidrIpv6: aws.String("::/0")}},
				},
			},
		},
	)
}

// Link updates run repeatedly against the same workload, so a redeploy must reuse the
// existing group rather than fail on a duplicate group name or leak a second one.
func TestWorkloadSecurityGroupReusedWhenItAlreadyExists(t *testing.T) {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []ec2types.SecurityGroup{
				{GroupId: aws.String("sg-existing")},
			},
		}),
	)

	groupID, err := resolveWorkloadSecurityGroup(context.Background(), service, testWorkloadIdentity())
	require.NoError(t, err)
	require.Equal(t, "sg-existing", groupID)

	service.AssertNotCalled(sgTestSuite(t), "CreateSecurityGroup")
}

// Lookup is scoped to this VPC, this workload and this blueprint instance. A shared
// VPC holds groups belonging to peer applications and to the VPC's owner, and none of
// them may be adopted or deleted here.
func TestWorkloadSecurityGroupLookupIsFullyScoped(t *testing.T) {
	filters := testWorkloadIdentity().filters()

	byName := map[string][]string{}
	for _, filter := range filters {
		byName[aws.ToString(filter.Name)] = filter.Values
	}

	require.Equal(t, []string{"vpc-1"}, byName["vpc-id"])
	require.Equal(t, []string{"getOrderFunction"}, byName["tag:"+flex.TagFlexVPCWorkloadSecurityGroup])
	require.Equal(t, []string{"instance-1"}, byName["tag:"+flex.TagFlexVPCWorkloadOwner])
	require.Equal(t, []string{"orders-vpc"}, byName["tag:"+flex.TagFlexVPCName])
}

// Teardown of these groups moved to the flex VPC resource; see
// TestWorkloadSecurityGroupsSweptByVPCDestroy in the flex package. The placement link
// only detaches the function, because the group is referenced by rules the access
// links own and no link can make a peer revoke them first.

// EC2 accepts only a-zA-Z0-9 and a fixed punctuation set in a security group
// description, and rejects anything else with InvalidParameterValue.
//
// This is not a style check. Interpolating the workload names with %q wrapped them in
// double quotes, which are not in the set, so every placement failed at CreateSecurityGroup.
// No mock could catch it: the character set is enforced by the API alone.
func TestWorkloadSecurityGroupDescriptionUsesOnlyCharactersEC2Accepts(t *testing.T) {
	// The set from the API's own error message.
	const allowedPunctuation = " ._-:/()#,@[]+=&;{}!$*"

	allowed := func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return true
		default:
			return strings.ContainsRune(allowedPunctuation, r)
		}
	}

	description := testWorkloadIdentity().description()
	require.NotEmpty(t, description)
	require.Less(t, len(description), 256, "EC2 caps descriptions at 255 characters")

	for _, r := range description {
		require.Truef(t, allowed(r), "character %q is not accepted by EC2 in a security group description", r)
	}
}
