package flex

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/ec2util"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
)

// Scoped to the owning blueprint instance as well as the VPC: a referenced VPC is
// shared, so the workload groups belonging to peer applications must not be swept up
// with this application's own.
// Removes the security groups links created for VPC endpoints in this VPC.
//
// Like the workload groups above, these are created by a link and deleted here, for a
// reason the link cannot work around. A link that yields while its group is still held
// by the endpoint's network interfaces does not get a second chance: by the time the
// destroy is retried, the resources at each end of the link are gone, and the framework
// skips a link whose linked resource has no persisted state. The cleanup then never
// runs, the group survives, and its ingress rule keeps the caller's group, and with it
// the VPC, undeletable.
//
// Scoped to this blueprint instance, since a referenced VPC is shared and a peer
// application's endpoint groups are not this blueprint's to remove.
func linkEndpointSecurityGroupIDs(
	ctx context.Context,
	service ec2service.Service,
	vpcID string,
	instanceName string,
) ([]string, error) {
	if instanceName == "" {
		return nil, nil
	}

	output, err := service.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", utils.TagLinkSecurityGroup)),
				Values: []string{"true"},
			},
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", utils.TagBlueprintInstanceName)),
				Values: []string{instanceName},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, nil
	}

	groupIDs := make([]string, 0, len(output.SecurityGroups))
	for _, group := range output.SecurityGroups {
		groupIDs = append(groupIDs, aws.ToString(group.GroupId))
	}

	return groupIDs, nil
}

func workloadSecurityGroupIDs(
	ctx context.Context,
	service ec2service.Service,
	vpcID string,
	instanceID string,
) ([]string, error) {
	// Scoped by VPC ID rather than the flex VPC name: the ID is always present in
	// destroy state, whereas the name is only guaranteed when the ID is missing.
	output, err := service.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCWorkloadOwner)),
				Values: []string{instanceID},
			},
			{
				Name:   aws.String("tag-key"),
				Values: []string{TagFlexVPCWorkloadSecurityGroup},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, nil
	}

	groupIDs := make([]string, 0, len(output.SecurityGroups))
	for _, group := range output.SecurityGroups {
		groupIDs = append(groupIDs, aws.ToString(group.GroupId))
	}

	return groupIDs, nil
}

// Removes every security group a link created in this VPC for this blueprint instance.
//
// One function so the create-mode and reference-mode teardowns cannot drift as both
// require the same cleanup, and the only difference between them is whether the VPC itself is
// removed afterwards.
//
// Every rule is revoked across the whole set before anything is deleted, because these
// groups reference each other: a workload's group has egress to the endpoint's group and
// the endpoint's group has ingress from the workload's. Neither can be deleted while the
// other's rule points at it, so clearing one group's rules and deleting it immediately
// fails on a rule the next step was about to remove. Breaking every reference first
// makes deletion order irrelevant.
func (l *vpcResourceActions) deleteLinkOwnedSecurityGroups(
	ctx context.Context,
	service ec2service.Service,
	vpcID string,
	instanceID string,
	instanceName string,
) error {
	endpointGroupIDs, err := linkEndpointSecurityGroupIDs(ctx, service, vpcID, instanceName)
	if err != nil {
		return err
	}

	workloadGroupIDs, err := workloadSecurityGroupIDs(ctx, service, vpcID, instanceID)
	if err != nil {
		return err
	}

	groupIDs := append(endpointGroupIDs, workloadGroupIDs...)
	if len(groupIDs) == 0 {
		return nil
	}

	// Fixed order so a partial failure leaves the same groups behind every time.
	sort.Strings(groupIDs)

	for _, groupID := range groupIDs {
		if err := revokeAllRules(ctx, service, groupID); err != nil {
			return err
		}
	}

	for _, groupID := range groupIDs {
		if err := ec2util.DeleteSecurityGroupWhenUnused(ctx, service, groupID); err != nil {
			return err
		}
	}

	return nil
}
