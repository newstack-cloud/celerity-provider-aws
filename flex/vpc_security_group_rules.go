package flex

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
)

// The permission set matching every protocol from anywhere, which is what AWS attaches
// as egress to a newly created security group.
func allTrafficPermissions() []types.IpPermission {
	return []types.IpPermission{
		{
			IpProtocol: aws.String("-1"),
			IpRanges: []types.IpRange{
				{CidrIp: aws.String("0.0.0.0/0")},
			},
			Ipv6Ranges: []types.Ipv6Range{
				{CidrIpv6: aws.String("::/0")},
			},
		},
	}
}

// A group with rules still attached cannot be deleted, and while a teardown is
// in progress a stripped group also stops granting anything.
func revokeAllRules(
	ctx context.Context,
	service ec2service.Service,
	groupID string,
) error {
	ingress, found, err := describeGroupIngress(ctx, service, groupID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	if len(ingress) > 0 {
		_, err = service.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       aws.String(groupID),
			IpPermissions: ingress,
		})
		if err != nil {
			return err
		}
	}

	egress, _, err := describeGroupEgress(ctx, service, groupID)
	if err != nil {
		return err
	}
	if len(egress) > 0 {
		_, err = service.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       aws.String(groupID),
			IpPermissions: egress,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func describeGroupEgress(
	ctx context.Context,
	service ec2service.Service,
	groupID string,
) ([]types.IpPermission, bool, error) {
	group, found, err := describeGroup(ctx, service, groupID)
	if err != nil || !found {
		return nil, found, err
	}

	return group.IpPermissionsEgress, true, nil
}

func describeGroupIngress(
	ctx context.Context,
	service ec2service.Service,
	groupID string,
) ([]types.IpPermission, bool, error) {
	group, found, err := describeGroup(ctx, service, groupID)
	if err != nil || !found {
		return nil, found, err
	}

	return group.IpPermissions, true, nil
}

// EC2 reports an unknown group ID as an InvalidGroup.NotFound error rather than
// an empty result, so absence has to be read off the error code.
func describeGroup(
	ctx context.Context,
	service ec2service.Service,
	groupID string,
) (*types.SecurityGroup, bool, error) {
	output, err := service.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{groupID},
	})
	if err != nil {
		if isGroupNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if output == nil || len(output.SecurityGroups) == 0 {
		return nil, false, nil
	}

	return &output.SecurityGroups[0], true, nil
}

func isGroupNotFoundError(err error) bool {
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		return apiErr.ErrorCode() == "InvalidGroup.NotFound"
	}

	return false
}
