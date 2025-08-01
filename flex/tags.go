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
	// TagFlexVPCNATGateway is the tag name that indicates that the resource
	// is a NAT gateway created for a flex VPC.
	TagFlexVPCNATGateway = "bluelink:flex-vpc:nat-gateway"
	// TagFlexVPCElasticIP is the tag name that indicates that the resource
	// is an Elastic IP address created for a flex VPC.
	TagFlexVPCElasticIP = "bluelink:flex-vpc:elastic-ip"
	// TagFlexVPCSecurityGroup is the tag name that indicates that the resource
	// is a security group created for a flex VPC.
	TagFlexVPCSecurityGroup = "bluelink:flex-vpc:security-group"
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
