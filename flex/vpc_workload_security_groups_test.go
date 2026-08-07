//go:build unit

package flex

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

// A link places a workload in a flex VPC by preparing it a security group of its own, and
// nothing else knows those groups exist. If a destroy leaves them behind, they hold the
// VPC open and the teardown cannot finish.
//
// Reference mode is the shortest path through Destroy that still has to sweep: the
// shared topology survives, but this application's own link-owned groups do not.
// Driving the resource's own Destroy is what proves the sweep is wired in rather than
// merely present.
type FlexVPCWorkloadSecurityGroupsSuite struct {
	suite.Suite
}

// Runs the resource's exported Destroy for a reference-mode flex VPC against the given
// EC2 mock.
//
// The table harness is not used here because these cases assert on the relative order
// of calls and the shape of filters, which its action expectations cannot express.
// Destroy is the resource's own method either way.
func (s *FlexVPCWorkloadSecurityGroupsSuite) runReferenceModeDestroy(
	service ec2service.Service,
) error {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	resource := VPCResource(
		func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		func(config *aws.Config, ctx provider.Context) resgrouptagservice.Service {
			return &mockResourceGroupTaggingService{}
		},
		utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
	)

	return resource.Destroy(context.Background(), &provider.ResourceDestroyInput{
		InstanceID:      "instance-1",
		InstanceName:    "it-instance-1",
		ProviderContext: providerCtx,
		ResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"vpcId", core.MappingNodeFromString("vpc-1"),
				"name", core.MappingNodeFromString("TestVPC"),
				"mode", core.MappingNodeFromString("reference"),
			),
		},
	})
}

// Deleted in a fixed order so a partial failure leaves the same groups behind on every
// attempt, which is what makes a retried destroy converge.
func (s *FlexVPCWorkloadSecurityGroupsSuite) Test_link_owned_groups_are_deleted_in_a_fixed_order() {
	service := ec2mock.CreateEc2ServiceMock(
		// The first lookup answers the endpoint groups, the second the workload
		// groups, deliberately returned out of order.
		ec2mock.WithDescribeSecurityGroupsOutputs([]*ec2.DescribeSecurityGroupsOutput{
			{},
			{
				SecurityGroups: []types.SecurityGroup{
					{GroupId: aws.String("sg-workload-b")},
					{GroupId: aws.String("sg-workload-a")},
				},
			},
		}),
		ec2mock.WithRevokeSecurityGroupIngressOutput(&ec2.RevokeSecurityGroupIngressOutput{}),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
		ec2mock.WithDescribeNetworkInterfacesOutputs(
			[]*ec2.DescribeNetworkInterfacesOutput{{}},
		),
		ec2mock.WithDeleteSecurityGroupOutput(&ec2.DeleteSecurityGroupOutput{}),
	)

	s.Require().NoError(s.runReferenceModeDestroy(service))
	s.Equal([]string{"sg-workload-a", "sg-workload-b"}, service.DeletedSecurityGroupIDs())
}

// Scoped to this VPC and this blueprint instance. A referenced VPC is shared, so the
// workload groups belonging to peer applications must not be swept up with these.
func (s *FlexVPCWorkloadSecurityGroupsSuite) Test_the_sweep_is_scoped_to_the_owning_instance() {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{}),
	)

	s.Require().NoError(s.runReferenceModeDestroy(service))

	// The endpoint group lookup runs first, so the workload lookup is the second
	// call. Its filters are what keep a peer application's groups out of the sweep.
	service.AssertCalledWith(
		&s.Suite,
		"DescribeSecurityGroups",
		1,
		plugintestutils.Any,
		func(arg any) bool {
			in, ok := arg.(*ec2.DescribeSecurityGroupsInput)
			if !ok {
				return false
			}
			byName := map[string][]string{}
			for _, filter := range in.Filters {
				byName[aws.ToString(filter.Name)] = filter.Values
			}
			return len(byName["vpc-id"]) == 1 &&
				byName["vpc-id"][0] == "vpc-1" &&
				len(byName["tag:"+TagFlexVPCWorkloadOwner]) == 1 &&
				byName["tag:"+TagFlexVPCWorkloadOwner][0] == "instance-1" &&
				len(byName["tag-key"]) == 1 &&
				byName["tag-key"][0] == TagFlexVPCWorkloadSecurityGroup
		},
	)

	s.Empty(
		service.DeletedSecurityGroupIDs(),
		"nothing matched, so nothing should have been deleted",
	)
}

// Rules are cleared across every link-owned group before any of them is deleted.
//
// The groups reference each other, and AWS will not delete a group that another group's
// rules point at. Revoking and deleting one group at a time deadlocks on the first pair.
func (s *FlexVPCWorkloadSecurityGroupsSuite) Test_every_rule_is_revoked_before_any_group_is_deleted() {
	service := ec2mock.CreateEc2ServiceMock(
		// First lookup answers the endpoint groups, second the workload groups. Every
		// later describe is the rule revocation inspecting a group, so those must carry
		// rules or nothing is revoked and the ordering assertion proves nothing.
		ec2mock.WithDescribeSecurityGroupsOutputs([]*ec2.DescribeSecurityGroupsOutput{
			{SecurityGroups: []types.SecurityGroup{{GroupId: aws.String("sg-endpoint")}}},
			{SecurityGroups: []types.SecurityGroup{{GroupId: aws.String("sg-workload")}}},
			{
				SecurityGroups: []types.SecurityGroup{
					{
						GroupId:             aws.String("sg-referencing"),
						IpPermissions:       []types.IpPermission{{IpProtocol: aws.String("tcp")}},
						IpPermissionsEgress: []types.IpPermission{{IpProtocol: aws.String("tcp")}},
					},
				},
			},
		}),
		ec2mock.WithRevokeSecurityGroupIngressOutput(&ec2.RevokeSecurityGroupIngressOutput{}),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
		ec2mock.WithDescribeNetworkInterfacesOutputs(
			[]*ec2.DescribeNetworkInterfacesOutput{{}},
		),
		ec2mock.WithDeleteSecurityGroupOutput(&ec2.DeleteSecurityGroupOutput{}),
	)

	s.Require().NoError(s.runReferenceModeDestroy(service))

	s.ElementsMatch(
		[]string{"sg-endpoint", "sg-workload"},
		service.DeletedSecurityGroupIDs(),
		"both link-owned groups should be deleted in one pass",
	)
	s.Less(
		service.LastRevokeCallOrder(),
		service.FirstDeleteSecurityGroupCallOrder(),
		"every rule must be revoked before the first group is deleted, or a group is "+
			"blocked by a rule the next step would have removed",
	)
}

func TestFlexVPCWorkloadSecurityGroupsSuite(t *testing.T) {
	suite.Run(t, new(FlexVPCWorkloadSecurityGroupsSuite))
}
