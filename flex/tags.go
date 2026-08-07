package flex

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	// TagFlexVPCName is the tag name the holds the name of the flex VPC.
	TagFlexVPCName = "bluelink:flex-vpc:name"
	// TagFlexVPCResource is the tag name that indicates that the resource is a part of a flex VPC.
	TagFlexVPCResource = "bluelink:flex-vpc:resource"
	// TagFlexVPCSubnetType is the tag name that indicates the type of the subnet.
	TagFlexVPCSubnetType = "bluelink:flex-vpc:subnet-type"
	// TagFlexVPCSubnetName is the tag name that indicates the name of the subnet,
	// this will be the static name of the subnet determined by the preset config.
	TagFlexVPCSubnetName = "bluelink:flex-vpc:subnet-name"
	// TagFlexVPCRouteTable is the tag name that indicates that the resource is a route table
	// created for a flex VPC.
	TagFlexVPCRouteTable = "bluelink:flex-vpc:route-table"
	// TagFlexVPCInternetGateway is the tag name that indicates that the resource
	// is an internet gateway created for a flex VPC.
	TagFlexVPCInternetGateway = "bluelink:flex-vpc:internet-gateway"
	// TagFlexVPCEgressOnlyInternetGateway is the tag name that indicates that the
	// resource is an egress-only internet gateway created for a flex VPC.
	TagFlexVPCEgressOnlyInternetGateway = "bluelink:flex-vpc:egress-only-internet-gateway"
	// TagFlexVPCNATGateway is the tag name that indicates that the resource
	// is a NAT gateway created for a flex VPC.
	TagFlexVPCNATGateway = "bluelink:flex-vpc:nat-gateway"
	// TagFlexVPCElasticIP is the tag name that indicates that the resource
	// is an Elastic IP address created for a flex VPC.
	TagFlexVPCElasticIP = "bluelink:flex-vpc:elastic-ip"
	// TagFlexVPCSecurityGroup is the tag name that indicates that the resource
	// is a security group created for a flex VPC.
	TagFlexVPCSecurityGroup = "bluelink:flex-vpc:security-group"
	// TagFlexVPCSecurityGroupName holds the declared name of a group created from the
	// VPC's securityGroups list, so the group can be found again and matched to the
	// name a resource references it by.
	//
	// Kept distinct from TagFlexVPCSecurityGroup for the same reason as the workload
	// tag below: that tag marks the VPC's own base group and drives the
	// securityGroupIds output, which a named group must not be adopted into.
	TagFlexVPCSecurityGroupName = "bluelink:flex-vpc:security-group-name"
	// TagFlexVPCSecurityGroupNameOwner holds the blueprint instance that owns a named
	// security group, so a shared VPC's named groups stay attributable to the
	// application that declared them.
	TagFlexVPCSecurityGroupNameOwner = "bluelink:flex-vpc:security-group-name-owner"
	// TagFlexVPCWorkloadSecurityGroup holds the name of the workload a security
	// group was created for by a placement link.
	//
	// This is deliberately not TagFlexVPCSecurityGroup as that tag marks the VPC's own base
	// group and is what the VPC resource filters its securityGroups output on, so a
	// workload group carrying it would be adopted into the VPC's state and handed
	// out to unrelated callers.
	TagFlexVPCWorkloadSecurityGroup = "bluelink:flex-vpc:workload-security-group"
	// TagFlexVPCWorkloadOwner holds the blueprint instance that owns a workload
	// security group, so a shared VPC's workload groups stay attributable to the
	// application that placed them, in the same way as
	// TagFlexVPCSecurityGroupNameOwner.
	TagFlexVPCWorkloadOwner = "bluelink:flex-vpc:workload-owner"
	// TagFlexVPCNetworkACL is the tag name that indicates that the resource
	// is a network ACL created for a flex VPC.
	TagFlexVPCNetworkACL = "bluelink:flex-vpc:network-acl"
)

// TagFilterFlexVPCName returns a filter for the flex VPC name
// for an EC2 networking resource.
func TagFilterFlexVPCName(name string) types.Filter {
	return types.Filter{
		Name: aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
		Values: []string{
			name,
		},
	}
}
