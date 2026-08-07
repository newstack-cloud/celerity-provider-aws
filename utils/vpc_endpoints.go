package utils

import (
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// VPCEndpointInSubnets checks if the VPC endpoint is in the given subnets.
func VPCEndpointInSubnets(
	endpoint *ec2types.VpcEndpoint,
	subnets []string,
) bool {
	if len(endpoint.SubnetIds) == 0 {
		return false
	}

	return slices.ContainsFunc(
		endpoint.SubnetIds,
		func(subnetID string) bool {
			return slices.Contains(subnets, subnetID)
		},
	)
}

// HasVPCEndpointTagForLink checks if the VPC endpoint has a tag
// that associates it with the given link ID.
func HasVPCEndpointTagForLink(
	endpoint *ec2types.VpcEndpoint,
	linkID string,
) bool {
	if len(endpoint.Tags) == 0 {
		return false
	}

	linkIDTag := CreateTagBlueprintLinkID(linkID)
	return slices.ContainsFunc(
		endpoint.Tags,
		func(tag ec2types.Tag) bool {
			return tag.Key != nil &&
				tag.Value != nil &&
				aws.ToString(tag.Key) == aws.ToString(linkIDTag.Key) &&
				aws.ToString(tag.Value) == "true"
		},
	)
}

// HasIngressFromSecurityGroupID checks if the security group has an ingress rule
// admitting the given source security group.
//
// Matched on GroupId rather than GroupName: a group in a non-default VPC is always
// referenced by ID, and EC2 does not populate GroupName on such a reference, so a
// name-based check can never match and would report every rule as absent.
func HasIngressFromSecurityGroupID(
	securityGroup *ec2types.SecurityGroup,
	sourceSecurityGroupID string,
) bool {
	return slices.ContainsFunc(
		securityGroup.IpPermissions,
		func(permission ec2types.IpPermission) bool {
			return slices.ContainsFunc(
				permission.UserIdGroupPairs,
				checkSourceSecurityGroupID(sourceSecurityGroupID),
			)
		},
	)
}

func checkSourceSecurityGroupID(
	sourceSecurityGroupID string,
) func(pair ec2types.UserIdGroupPair) bool {
	return func(pair ec2types.UserIdGroupPair) bool {
		return pair.GroupId != nil &&
			aws.ToString(pair.GroupId) == sourceSecurityGroupID
	}
}

// HasSecurityGroupTagForLink checks if the security group has a tag
// that associates it with the given link ID.
func HasSecurityGroupTagForLink(
	securityGroup *ec2types.SecurityGroup,
	linkID string,
) bool {
	if len(securityGroup.Tags) == 0 {
		return false
	}

	linkIDTag := CreateTagBlueprintLinkID(linkID)
	return slices.ContainsFunc(
		securityGroup.Tags,
		func(tag ec2types.Tag) bool {
			return tag.Key != nil &&
				tag.Value != nil &&
				aws.ToString(tag.Key) == aws.ToString(linkIDTag.Key) &&
				aws.ToString(tag.Value) == "true"
		},
	)
}

// GetOtherLinkTagsFromVPCEndpoint gets all link tags from a VPC endpoint
// except for the specified link ID.
func GetOtherLinkTagsFromVPCEndpoint(
	endpoint *ec2types.VpcEndpoint,
	excludeLinkID string,
) []ec2types.Tag {
	if len(endpoint.Tags) == 0 {
		return []ec2types.Tag{}
	}

	var otherLinkTags []ec2types.Tag
	excludeLinkIDTag := CreateTagBlueprintLinkID(excludeLinkID)

	for _, tag := range endpoint.Tags {
		if tag.Key != nil &&
			aws.ToString(tag.Key) == aws.ToString(excludeLinkIDTag.Key) &&
			aws.ToString(tag.Value) == "true" {
			continue
		}

		if tag.Key != nil &&
			strings.HasPrefix(aws.ToString(tag.Key), TagBlueprintLinkIDPrefix) &&
			aws.ToString(tag.Value) == "true" {
			otherLinkTags = append(otherLinkTags, tag)
		}
	}

	return otherLinkTags
}

// GetOtherLinkTagsFromSecurityGroup gets all link tags from a security group
// except for the specified link ID.
func GetOtherLinkTagsFromSecurityGroup(
	securityGroup *ec2types.SecurityGroup,
	excludeLinkID string,
) []ec2types.Tag {
	if len(securityGroup.Tags) == 0 {
		return []ec2types.Tag{}
	}

	var otherLinkTags []ec2types.Tag
	excludeLinkIDTag := CreateTagBlueprintLinkID(excludeLinkID)

	for _, tag := range securityGroup.Tags {
		if tag.Key != nil &&
			aws.ToString(tag.Key) == aws.ToString(excludeLinkIDTag.Key) &&
			aws.ToString(tag.Value) == "true" {
			continue
		}

		if tag.Key != nil &&
			strings.HasPrefix(aws.ToString(tag.Key), TagBlueprintLinkIDPrefix) &&
			aws.ToString(tag.Value) == "true" {
			otherLinkTags = append(otherLinkTags, tag)
		}
	}

	return otherLinkTags
}
