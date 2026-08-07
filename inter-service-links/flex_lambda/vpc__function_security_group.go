package flexlambda

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink-provider-aws/flex"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Identifies the workload a placed function's security group belongs to.
//
// A group is scoped to one function within one VPC within one blueprint instance.
// The instance is part of the identity because a VPC in reference mode is shared
// between applications, and two of them may each place a function of the same name.
type workloadGroupIdentity struct {
	vpcID        string
	vpcName      string
	functionName string
	instanceID   string
}

func (i *workloadGroupIdentity) groupName() string {
	return fmt.Sprintf("bluelink-flexvpc-%s-%s-%s", i.vpcName, i.functionName, i.instanceID)
}

// EC2 restricts security group descriptions to a-zA-Z0-9 and a fixed punctuation set
// that does not include the double quote, so names are interpolated with %s rather
// than %q. A quoted name is rejected with InvalidParameterValue.
func (i *workloadGroupIdentity) description() string {
	return fmt.Sprintf(
		"Managed by Bluelink for the %s function placed in the %s flex VPC.",
		i.functionName,
		i.vpcName,
	)
}

func (i *workloadGroupIdentity) tags() []types.Tag {
	return []types.Tag{
		{Key: aws.String(flex.TagFlexVPCName), Value: aws.String(i.vpcName)},
		{Key: aws.String(flex.TagFlexVPCResource), Value: aws.String("true")},
		{Key: aws.String(flex.TagFlexVPCWorkloadSecurityGroup), Value: aws.String(i.functionName)},
		{Key: aws.String(flex.TagFlexVPCWorkloadOwner), Value: aws.String(i.instanceID)},
	}
}

// Matches exactly the group this workload owns, so a lookup never returns the VPC's
// base group, a policy group, or a peer application's workload group.
func (i *workloadGroupIdentity) filters() []types.Filter {
	return []types.Filter{
		{Name: aws.String("vpc-id"), Values: []string{i.vpcID}},
		{
			Name:   aws.String(fmt.Sprintf("tag:%s", flex.TagFlexVPCWorkloadSecurityGroup)),
			Values: []string{i.functionName},
		},
		{
			Name:   aws.String(fmt.Sprintf("tag:%s", flex.TagFlexVPCWorkloadOwner)),
			Values: []string{i.instanceID},
		},
		{
			Name:   aws.String(fmt.Sprintf("tag:%s", flex.TagFlexVPCName)),
			Values: []string{i.vpcName},
		},
	}
}

func workloadIdentity(
	flexVPCSpecData *core.MappingNode,
	functionName string,
	instanceID string,
) (*workloadGroupIdentity, error) {
	vpcIDNode, _ := pluginutils.GetValueByPath("$.vpcId", flexVPCSpecData)
	vpcID := core.StringValue(vpcIDNode)
	if vpcID == "" {
		return nil, errors.New("the linked flex VPC exposes no VPC ID")
	}

	vpcNameNode, _ := pluginutils.GetValueByPath("$.name", flexVPCSpecData)
	vpcName := core.StringValue(vpcNameNode)
	if vpcName == "" {
		return nil, errors.New("the linked flex VPC exposes no name")
	}

	return &workloadGroupIdentity{
		vpcID:        vpcID,
		vpcName:      vpcName,
		functionName: functionName,
		instanceID:   instanceID,
	}, nil
}

// Returns the ID of the security group for this placed function, creating it if it
// does not exist yet.
//
// Link updates run repeatedly against the same workload, so this is a find-or-create
// rather than a create: a second deploy of an unchanged blueprint must reuse the group
// rather than fail on a duplicate name or leak a second one.
func resolveWorkloadSecurityGroup(
	ctx context.Context,
	service ec2service.Service,
	identity *workloadGroupIdentity,
) (string, error) {
	existing, err := findWorkloadSecurityGroup(ctx, service, identity)
	if err != nil {
		return "", err
	}
	if existing != "" {
		return existing, nil
	}

	return createWorkloadSecurityGroup(ctx, service, identity)
}

func findWorkloadSecurityGroup(
	ctx context.Context,
	service ec2service.Service,
	identity *workloadGroupIdentity,
) (string, error) {
	output, err := service.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: identity.filters(),
	})
	if err != nil {
		return "", err
	}

	if output == nil || len(output.SecurityGroups) == 0 {
		return "", nil
	}

	return aws.ToString(output.SecurityGroups[0].GroupId), nil
}

func createWorkloadSecurityGroup(
	ctx context.Context,
	service ec2service.Service,
	identity *workloadGroupIdentity,
) (string, error) {
	output, err := service.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(identity.groupName()),
		Description: aws.String(identity.description()),
		VpcId:       aws.String(identity.vpcID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSecurityGroup,
				Tags:         identity.tags(),
			},
		},
	})
	if err != nil {
		return "", err
	}

	if output == nil || output.GroupId == nil {
		return "", errors.New("no security group was returned when creating one for the placed function")
	}

	groupID := aws.ToString(output.GroupId)

	// EC2 attaches an allow-all egress rule to every new group. Leaving it would make
	// the per-workload group pointless: the function could reach anything in the VPC
	// and the internet regardless of what it is linked to. Reach is granted per link
	// from here on.
	_, err = service.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
		GroupId: aws.String(groupID),
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				IpRanges:   []types.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
				Ipv6Ranges: []types.Ipv6Range{{CidrIpv6: aws.String("::/0")}},
			},
		},
	})
	if err != nil {
		return "", err
	}

	return groupID, nil
}

// Opens outbound access to destinations outside the VPC on the workload's group.
//
// Reach *within* the VPC is granted per link by the access links, so nothing here
// touches in-VPC destinations. This covers only what the link graph cannot express:
// the public internet, or a set of CIDR ranges the author named.
//
// Rules are authorised rather than reconciled, and duplicates are tolerated, because
// a link update runs repeatedly against the same group.
func authorizeWorkloadEgress(
	ctx context.Context,
	service ec2service.Service,
	groupID string,
	plan *egressPlan,
) error {
	permissions := egressPermissions(plan)
	if len(permissions) == 0 {
		return nil
	}

	_, err := service.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId:       aws.String(groupID),
		IpPermissions: permissions,
	})
	return ignoreDuplicateEgressRule(err)
}

func egressPermissions(plan *egressPlan) []types.IpPermission {
	if plan == nil || plan.reach == egressNone {
		return nil
	}

	// All protocols to the destination: narrowing by port belongs to the access
	// links, which know the port because they know the target.
	permission := types.IpPermission{IpProtocol: aws.String("-1")}

	if len(plan.cidrs) > 0 {
		for _, cidr := range plan.cidrs {
			if strings.Contains(cidr, ":") {
				permission.Ipv6Ranges = append(
					permission.Ipv6Ranges,
					types.Ipv6Range{CidrIpv6: aws.String(cidr)},
				)
				continue
			}
			// An IPv4 range is unreachable from a public subnet, where the function
			// has no public IPv4 address. Authorising it anyway would be a rule that
			// silently does nothing, so it is left out and only IPv6 is opened.
			if plan.reach == egressFull {
				permission.IpRanges = append(
					permission.IpRanges,
					types.IpRange{CidrIp: aws.String(cidr)},
				)
			}
		}

		if len(permission.IpRanges) == 0 && len(permission.Ipv6Ranges) == 0 {
			return nil
		}

		return []types.IpPermission{permission}
	}

	permission.Ipv6Ranges = []types.Ipv6Range{{CidrIpv6: aws.String("::/0")}}
	if plan.reach == egressFull {
		permission.IpRanges = []types.IpRange{{CidrIp: aws.String("0.0.0.0/0")}}
	}

	return []types.IpPermission{permission}
}

// A rule that is already present is the expected outcome of re-running an unchanged
// link, not a failure.
func ignoreDuplicateEgressRule(err error) error {
	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidPermission.Duplicate" {
		return nil
	}

	return err
}
