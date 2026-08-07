//go:build integration

package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Deleting a VPC means emptying it first, in an order AWS enforces: endpoints hold
// network interfaces, security groups hold each other through their rules, and subnets
// hold anything still attached to them.
//
// The network interfaces Lambda attaches to a placed function are the slow part. They
// outlive the function by minutes, and while any remain the security groups and subnets
// referencing them cannot be deleted, so the wait comes before anything else.
func (c *sweepClients) deleteVPC(ctx context.Context, vpcID string) error {
	if err := c.deleteVPCEndpoints(ctx, vpcID); err != nil {
		return err
	}

	if err := c.awaitENIRelease(ctx, vpcID); err != nil {
		return err
	}

	// Rules are stripped from every group before any group is deleted, because a group
	// referenced by another group's rules cannot be deleted and the endpoint groups
	// reference the caller's group by design.
	if err := c.revokeVPCSecurityGroupRules(ctx, vpcID); err != nil {
		return err
	}

	if err := c.deleteVPCSecurityGroups(ctx, vpcID); err != nil {
		return err
	}

	if err := c.deleteVPCSubnets(ctx, vpcID); err != nil {
		return err
	}

	if err := c.deleteVPCRouteTables(ctx, vpcID); err != nil {
		return err
	}

	_, err := c.ec2.DeleteVpc(ctx, &ec2.DeleteVpcInput{VpcId: aws.String(vpcID)})

	return err
}

func (c *sweepClients) deleteVPCEndpoints(ctx context.Context, vpcID string) error {
	output, err := c.ec2.DescribeVpcEndpoints(ctx, &ec2.DescribeVpcEndpointsInput{
		Filters: vpcFilter(vpcID),
	})
	if err != nil {
		return err
	}
	if len(output.VpcEndpoints) == 0 {
		return nil
	}

	ids := make([]string, 0, len(output.VpcEndpoints))
	for _, endpoint := range output.VpcEndpoints {
		ids = append(ids, aws.ToString(endpoint.VpcEndpointId))
	}

	_, err = c.ec2.DeleteVpcEndpoints(ctx, &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: ids,
	})

	return err
}

func (c *sweepClients) awaitENIRelease(ctx context.Context, vpcID string) error {
	deadline := time.Now().Add(leakSweepENIDeadline)
	for {
		output, err := c.ec2.DescribeNetworkInterfaces(
			ctx,
			&ec2.DescribeNetworkInterfacesInput{Filters: vpcFilter(vpcID)},
		)
		if err != nil {
			return err
		}
		if len(output.NetworkInterfaces) == 0 {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf(
				"%d network interface(s) still attached after %s, so the VPC cannot be deleted",
				len(output.NetworkInterfaces),
				leakSweepENIDeadline,
			)
		}

		c.detachAvailableENIs(ctx, output.NetworkInterfaces)

		// The per-VPC deadline above does not bound the sweep on its own, since each
		// leaked VPC gets its own. The shared budget is what stops several of them
		// adding up to an unbounded wait.
		if err := sleepOrDone(ctx, leakSweepENIInterval); err != nil {
			return fmt.Errorf(
				"%d network interface(s) still attached when the sweep budget ran out",
				len(output.NetworkInterfaces),
			)
		}
	}
}

// An interface that has come free but has not been reclaimed is deleted directly. The
// ones still in use belong to Lambda and are left to it.
func (c *sweepClients) detachAvailableENIs(
	ctx context.Context,
	interfaces []ec2types.NetworkInterface,
) {
	for _, eni := range interfaces {
		if eni.Status != ec2types.NetworkInterfaceStatusAvailable {
			continue
		}
		_, err := c.ec2.DeleteNetworkInterface(ctx, &ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: eni.NetworkInterfaceId,
		})
		if err != nil {
			sweepWarn("delete network interface "+aws.ToString(eni.NetworkInterfaceId), err)
		}
	}
}

func (c *sweepClients) revokeVPCSecurityGroupRules(ctx context.Context, vpcID string) error {
	groups, err := c.nonDefaultSecurityGroups(ctx, vpcID)
	if err != nil {
		return err
	}

	for _, group := range groups {
		groupID := aws.ToString(group.GroupId)
		if len(group.IpPermissions) > 0 {
			_, err = c.ec2.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
				GroupId:       aws.String(groupID),
				IpPermissions: group.IpPermissions,
			})
			if err != nil {
				sweepWarn("revoke ingress on "+groupID, err)
			}
		}
		if len(group.IpPermissionsEgress) > 0 {
			_, err = c.ec2.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
				GroupId:       aws.String(groupID),
				IpPermissions: group.IpPermissionsEgress,
			})
			if err != nil {
				sweepWarn("revoke egress on "+groupID, err)
			}
		}
	}

	return nil
}

func (c *sweepClients) deleteVPCSecurityGroups(ctx context.Context, vpcID string) error {
	groups, err := c.nonDefaultSecurityGroups(ctx, vpcID)
	if err != nil {
		return err
	}

	for _, group := range groups {
		_, err = c.ec2.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: group.GroupId,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// The default group belongs to the VPC and is removed with it, so it is never deleted
// directly.
func (c *sweepClients) nonDefaultSecurityGroups(
	ctx context.Context,
	vpcID string,
) ([]ec2types.SecurityGroup, error) {
	output, err := c.ec2.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: vpcFilter(vpcID),
	})
	if err != nil {
		return nil, err
	}

	groups := []ec2types.SecurityGroup{}
	for _, group := range output.SecurityGroups {
		if aws.ToString(group.GroupName) != "default" {
			groups = append(groups, group)
		}
	}

	return groups, nil
}

func (c *sweepClients) deleteVPCSubnets(ctx context.Context, vpcID string) error {
	output, err := c.ec2.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: vpcFilter(vpcID),
	})
	if err != nil {
		return err
	}

	for _, subnet := range output.Subnets {
		_, err = c.ec2.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{SubnetId: subnet.SubnetId})
		if err != nil {
			return err
		}
	}

	return nil
}

// The main route table is removed with the VPC and cannot be deleted on its own.
func (c *sweepClients) deleteVPCRouteTables(ctx context.Context, vpcID string) error {
	output, err := c.ec2.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: vpcFilter(vpcID),
	})
	if err != nil {
		return err
	}

	for _, routeTable := range output.RouteTables {
		if isMainRouteTable(routeTable) {
			continue
		}
		_, err = c.ec2.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
			RouteTableId: routeTable.RouteTableId,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func isMainRouteTable(routeTable ec2types.RouteTable) bool {
	for _, association := range routeTable.Associations {
		if aws.ToBool(association.Main) {
			return true
		}
	}

	return false
}

func vpcFilter(vpcID string) []ec2types.Filter {
	return []ec2types.Filter{
		{
			Name:   aws.String("vpc-id"),
			Values: []string{vpcID},
		},
	}
}
