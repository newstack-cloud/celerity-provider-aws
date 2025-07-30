package ec2service

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the Amazon EC2 service
// used by the EC2 resource implementations.
type Service interface {
	// Creates a VPC with the specified CIDR blocks. For more information, see [IP addressing for your VPCs and subnets] in the
	// Amazon VPC User Guide.
	//
	// You can optionally request an IPv6 CIDR block for the VPC. You can request an
	// Amazon-provided IPv6 CIDR block from Amazon's pool of IPv6 addresses or an IPv6
	// CIDR block from an IPv6 address pool that you provisioned through bring your own
	// IP addresses ([BYOIP] ).
	//
	// By default, each instance that you launch in the VPC has the default DHCP
	// options, which include only a default DNS server that we provide
	// (AmazonProvidedDNS). For more information, see [DHCP option sets]in the Amazon VPC User Guide.
	//
	// You can specify the instance tenancy value for the VPC when you create it. You
	// can't change this value for the VPC after you create it. For more information,
	// see [Dedicated Instances]in the Amazon EC2 User Guide.
	//
	// [BYOIP]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-byoip.html
	// [Dedicated Instances]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/dedicated-instance.html
	// [DHCP option sets]: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_DHCP_Options.html
	// [IP addressing for your VPCs and subnets]: https://docs.aws.amazon.com/vpc/latest/userguide/vpc-ip-addressing.html
	CreateVpc(ctx context.Context, params *ec2.CreateVpcInput, optFns ...func(*ec2.Options)) (*ec2.CreateVpcOutput, error)
	// Describes your VPCs. The default is to describe all your VPCs. Alternatively,
	// you can specify specific VPC IDs or filter the results to include only the VPCs
	// that match specific criteria.
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	// Modifies the specified attribute of the specified VPC.
	ModifyVpcAttribute(ctx context.Context, params *ec2.ModifyVpcAttributeInput, optFns ...func(*ec2.Options)) (*ec2.ModifyVpcAttributeOutput, error)
	// Creates a subnet in the specified VPC. For an IPv4 only subnet, specify an IPv4
	// CIDR block. If the VPC has an IPv6 CIDR block, you can create an IPv6 only
	// subnet or a dual stack subnet instead. For an IPv6 only subnet, specify an IPv6
	// CIDR block. For a dual stack subnet, specify both an IPv4 CIDR block and an IPv6
	// CIDR block.
	//
	// A subnet CIDR block must not overlap the CIDR block of an existing subnet in
	// the VPC. After you create a subnet, you can't change its CIDR block.
	//
	// The allowed size for an IPv4 subnet is between a /28 netmask (16 IP addresses)
	// and a /16 netmask (65,536 IP addresses). Amazon Web Services reserves both the
	// first four and the last IPv4 address in each subnet's CIDR block. They're not
	// available for your use.
	//
	// If you've associated an IPv6 CIDR block with your VPC, you can associate an
	// IPv6 CIDR block with a subnet when you create it.
	//
	// If you add more than one subnet to a VPC, they're set up in a star topology
	// with a logical router in the middle.
	//
	// When you stop an instance in a subnet, it retains its private IPv4 address.
	// It's therefore possible to have a subnet with no running instances (they're all
	// stopped), but no remaining IP addresses available.
	//
	// For more information, see [Subnets] in the Amazon VPC User Guide.
	//
	// [Subnets]: https://docs.aws.amazon.com/vpc/latest/userguide/configure-subnets.html
	CreateSubnet(ctx context.Context, params *ec2.CreateSubnetInput, optFns ...func(*ec2.Options)) (*ec2.CreateSubnetOutput, error)
	// Describes the Availability Zones, Local Zones, and Wavelength Zones that are
	// available to you.
	//
	// For more information about Availability Zones, Local Zones, and Wavelength
	// Zones, see [Regions and zones]in the Amazon EC2 User Guide.
	//
	// The order of the elements in the response, including those within nested
	// structures, might vary. Applications should not assume the elements appear in a
	// particular order.
	//
	// [Regions and zones]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
	DescribeAvailabilityZones(ctx context.Context, params *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAvailabilityZonesOutput, error)
	// Modifies a subnet attribute. You can only modify one attribute at a time.
	//
	// Use this action to modify subnets on Amazon Web Services Outposts.
	//
	//   - To modify a subnet on an Outpost rack, set both MapCustomerOwnedIpOnLaunch
	//     and CustomerOwnedIpv4Pool . These two parameters act as a single attribute.
	//
	//   - To modify a subnet on an Outpost server, set either EnableLniAtDeviceIndex
	//     or DisableLniAtDeviceIndex .
	//
	// For more information about Amazon Web Services Outposts, see the following:
	//
	// [Outpost servers]
	//
	// [Outpost racks]
	//
	// [Outpost servers]: https://docs.aws.amazon.com/outposts/latest/userguide/how-servers-work.html
	// [Outpost racks]: https://docs.aws.amazon.com/outposts/latest/userguide/how-racks-work.html
	ModifySubnetAttribute(ctx context.Context, params *ec2.ModifySubnetAttributeInput, optFns ...func(*ec2.Options)) (*ec2.ModifySubnetAttributeOutput, error)
	// Creates a route table for the specified VPC. After you create a route table,
	// you can add routes and associate the table with a subnet.
	//
	// For more information, see [Route tables] in the Amazon VPC User Guide.
	//
	// [Route tables]: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Route_Tables.html
	CreateRouteTable(ctx context.Context, params *ec2.CreateRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.CreateRouteTableOutput, error)
	// Associates a subnet in your VPC or an internet gateway or virtual private
	// gateway attached to your VPC with a route table in your VPC. This association
	// causes traffic from the subnet or gateway to be routed according to the routes
	// in the route table. The action returns an association ID, which you need in
	// order to disassociate the route table later. A route table can be associated
	// with multiple subnets.
	//
	// For more information, see [Route tables] in the Amazon VPC User Guide.
	//
	// [Route tables]: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Route_Tables.html
	AssociateRouteTable(ctx context.Context, params *ec2.AssociateRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.AssociateRouteTableOutput, error)
	// Creates an internet gateway for use with a VPC. After creating the internet
	// gateway, you attach it to a VPC using AttachInternetGateway.
	//
	// For more information, see [Internet gateways] in the Amazon VPC User Guide.
	//
	// [Internet gateways]: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Internet_Gateway.html
	CreateInternetGateway(ctx context.Context, params *ec2.CreateInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.CreateInternetGatewayOutput, error)
	// Attaches an internet gateway or a virtual private gateway to a VPC, enabling
	// connectivity between the internet and the VPC. For more information, see [Internet gateways]in the
	// Amazon VPC User Guide.
	//
	// [Internet gateways]: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Internet_Gateway.html
	AttachInternetGateway(ctx context.Context, params *ec2.AttachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.AttachInternetGatewayOutput, error)
	// Creates a route in a route table within a VPC.
	//
	// You must specify either a destination CIDR block or a prefix list ID. You must
	// also specify exactly one of the resources from the parameter list.
	//
	// When determining how to route traffic, we use the route with the most specific
	// match. For example, traffic is destined for the IPv4 address 192.0.2.3 , and the
	// route table includes the following two IPv4 routes:
	//
	//   - 192.0.2.0/24 (goes to some target A)
	//
	//   - 192.0.2.0/28 (goes to some target B)
	//
	// Both routes apply to the traffic destined for 192.0.2.3 . However, the second
	// route in the list covers a smaller number of IP addresses and is therefore more
	// specific, so we use that route to determine where to target the traffic.
	//
	// For more information about route tables, see [Route tables] in the Amazon VPC User Guide.
	//
	// [Route tables]: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Route_Tables.html
	CreateRoute(ctx context.Context, params *ec2.CreateRouteInput, optFns ...func(*ec2.Options)) (*ec2.CreateRouteOutput, error)
	// Creates a NAT gateway in the specified subnet. This action creates a network
	// interface in the specified subnet with a private IP address from the IP address
	// range of the subnet. You can create either a public NAT gateway or a private NAT
	// gateway.
	//
	// With a public NAT gateway, internet-bound traffic from a private subnet can be
	// routed to the NAT gateway, so that instances in a private subnet can connect to
	// the internet.
	//
	// With a private NAT gateway, private communication is routed across VPCs and
	// on-premises networks through a transit gateway or virtual private gateway.
	// Common use cases include running large workloads behind a small pool of
	// allowlisted IPv4 addresses, preserving private IPv4 addresses, and communicating
	// between overlapping networks.
	//
	// For more information, see [NAT gateways] in the Amazon VPC User Guide.
	//
	// When you create a public NAT gateway and assign it an EIP or secondary EIPs,
	// the network border group of the EIPs must match the network border group of the
	// Availability Zone (AZ) that the public NAT gateway is in. If it's not the same,
	// the NAT gateway will fail to launch. You can see the network border group for
	// the subnet's AZ by viewing the details of the subnet. Similarly, you can view
	// the network border group of an EIP by viewing the details of the EIP address.
	// For more information about network border groups and EIPs, see [Allocate an Elastic IP address]in the Amazon
	// VPC User Guide.
	//
	// [NAT gateways]: https://docs.aws.amazon.com/vpc/latest/userguide/vpc-nat-gateway.html
	// [Allocate an Elastic IP address]: https://docs.aws.amazon.com/vpc/latest/userguide/WorkWithEIPs.html
	CreateNatGateway(ctx context.Context, params *ec2.CreateNatGatewayInput, optFns ...func(*ec2.Options)) (*ec2.CreateNatGatewayOutput, error)
	// Describes your NAT gateways. The default is to describe all your NAT gateways.
	// Alternatively, you can specify specific NAT gateway IDs or filter the results to
	// include only the NAT gateways that match specific criteria.
	DescribeNatGateways(ctx context.Context, params *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error)
	// Allocates an Elastic IP address to your Amazon Web Services account. After you
	// allocate the Elastic IP address you can associate it with an instance or network
	// interface. After you release an Elastic IP address, it is released to the IP
	// address pool and can be allocated to a different Amazon Web Services account.
	//
	// You can allocate an Elastic IP address from an address pool owned by Amazon Web
	// Services or from an address pool created from a public IPv4 address range that
	// you have brought to Amazon Web Services for use with your Amazon Web Services
	// resources using bring your own IP addresses (BYOIP). For more information, see [Bring Your Own IP Addresses (BYOIP)]
	// in the Amazon EC2 User Guide.
	//
	// If you release an Elastic IP address, you might be able to recover it. You
	// cannot recover an Elastic IP address that you released after it is allocated to
	// another Amazon Web Services account. To attempt to recover an Elastic IP address
	// that you released, specify it in this operation.
	//
	// For more information, see [Elastic IP Addresses] in the Amazon EC2 User Guide.
	//
	// You can allocate a carrier IP address which is a public IP address from a
	// telecommunication carrier, to a network interface which resides in a subnet in a
	// Wavelength Zone (for example an EC2 instance).
	//
	// [Bring Your Own IP Addresses (BYOIP)]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-byoip.html
	// [Elastic IP Addresses]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/elastic-ip-addresses-eip.html
	AllocateAddress(ctx context.Context, params *ec2.AllocateAddressInput, optFns ...func(*ec2.Options)) (*ec2.AllocateAddressOutput, error)
	// Creates a security group.
	//
	// A security group acts as a virtual firewall for your instance to control
	// inbound and outbound traffic. For more information, see [Amazon EC2 security groups]in the Amazon EC2 User
	// Guide and [Security groups for your VPC]in the Amazon VPC User Guide.
	//
	// When you create a security group, you specify a friendly name of your choice.
	// You can't have two security groups for the same VPC with the same name.
	//
	// You have a default security group for use in your VPC. If you don't specify a
	// security group when you launch an instance, the instance is launched into the
	// appropriate default security group. A default security group includes a default
	// rule that grants instances unrestricted network access to each other.
	//
	// You can add or remove rules from your security groups using AuthorizeSecurityGroupIngress, AuthorizeSecurityGroupEgress, RevokeSecurityGroupIngress, and RevokeSecurityGroupEgress.
	//
	// For more information about VPC security group limits, see [Amazon VPC Limits].
	//
	// [Amazon VPC Limits]: https://docs.aws.amazon.com/vpc/latest/userguide/amazon-vpc-limits.html
	// [Amazon EC2 security groups]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-network-security.html
	// [Security groups for your VPC]: https://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_SecurityGroups.html
	CreateSecurityGroup(ctx context.Context, params *ec2.CreateSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error)
	// Removes the specified outbound (egress) rules from the specified security group.
	//
	// You can specify rules using either rule IDs or security group rule properties.
	// If you use rule properties, the values that you specify (for example, ports)
	// must match the existing rule's values exactly. Each rule has a protocol, from
	// and to ports, and destination (CIDR range, security group, or prefix list). For
	// the TCP and UDP protocols, you must also specify the destination port or range
	// of ports. For the ICMP protocol, you must also specify the ICMP type and code.
	// If the security group rule has a description, you do not need to specify the
	// description to revoke the rule.
	//
	// For a default VPC, if the values you specify do not match the existing rule's
	// values, no error is returned, and the output describes the security group rules
	// that were not revoked.
	//
	// Amazon Web Services recommends that you describe the security group to verify
	// that the rules were removed.
	//
	// Rule changes are propagated to instances within the security group as quickly
	// as possible. However, a small delay might occur.
	RevokeSecurityGroupEgress(ctx context.Context, params *ec2.RevokeSecurityGroupEgressInput, optFns ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupEgressOutput, error)
	// Creates a network ACL in a VPC. Network ACLs provide an optional layer of
	// security (in addition to security groups) for the instances in your VPC.
	//
	// For more information, see [Network ACLs] in the Amazon VPC User Guide.
	//
	// [Network ACLs]: https://docs.aws.amazon.com/vpc/latest/userguide/vpc-network-acls.html
	CreateNetworkAcl(ctx context.Context, params *ec2.CreateNetworkAclInput, optFns ...func(*ec2.Options)) (*ec2.CreateNetworkAclOutput, error)
	// Creates an entry (a rule) in a network ACL with the specified rule number. Each
	// network ACL has a set of numbered ingress rules and a separate set of numbered
	// egress rules. When determining whether a packet should be allowed in or out of a
	// subnet associated with the ACL, we process the entries in the ACL according to
	// the rule numbers, in ascending order. Each network ACL has a set of ingress
	// rules and a separate set of egress rules.
	//
	// We recommend that you leave room between the rule numbers (for example, 100,
	// 110, 120, ...), and not number them one right after the other (for example, 101,
	// 102, 103, ...). This makes it easier to add a rule between existing ones without
	// having to renumber the rules.
	//
	// After you add an entry, you can't modify it; you must either replace it, or
	// create an entry and delete the old one.
	//
	// For more information about network ACLs, see [Network ACLs] in the Amazon VPC User Guide.
	//
	// [Network ACLs]: https://docs.aws.amazon.com/vpc/latest/userguide/vpc-network-acls.html
	CreateNetworkAclEntry(ctx context.Context, params *ec2.CreateNetworkAclEntryInput, optFns ...func(*ec2.Options)) (*ec2.CreateNetworkAclEntryOutput, error)
	// Changes which network ACL a subnet is associated with. By default when you
	// create a subnet, it's automatically associated with the default network ACL. For
	// more information, see [Network ACLs]in the Amazon VPC User Guide.
	//
	// This is an idempotent operation.
	//
	// [Network ACLs]: https://docs.aws.amazon.com/vpc/latest/userguide/vpc-network-acls.html
	ReplaceNetworkAclAssociation(ctx context.Context, params *ec2.ReplaceNetworkAclAssociationInput, optFns ...func(*ec2.Options)) (*ec2.ReplaceNetworkAclAssociationOutput, error)
	// Describes your network ACLs. The default is to describe all your network ACLs.
	// Alternatively, you can specify specific network ACL IDs or filter the results to
	// include only the network ACLs that match specific criteria.
	//
	// For more information, see [Network ACLs] in the Amazon VPC User Guide.
	//
	// [Network ACLs]: https://docs.aws.amazon.com/vpc/latest/userguide/vpc-network-acls.html
	DescribeNetworkAcls(ctx context.Context, params *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error)
	// Adds or overwrites only the specified tags for the specified Amazon EC2
	// resource or resources. When you specify an existing tag key, the value is
	// overwritten with the new value. Each resource can have a maximum of 50 tags.
	// Each tag consists of a key and optional value. Tag keys must be unique per
	// resource.
	//
	// For more information about tags, see [Tag your Amazon EC2 resources] in the Amazon Elastic Compute Cloud User
	// Guide. For more information about creating IAM policies that control users'
	// access to resources based on tags, see [Supported resource-level permissions for Amazon EC2 API actions]in the Amazon Elastic Compute Cloud User
	// Guide.
	//
	// [Supported resource-level permissions for Amazon EC2 API actions]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-supported-iam-actions-resources.html
	// [Tag your Amazon EC2 resources]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html
	CreateTags(ctx context.Context, params *ec2.CreateTagsInput, optFns ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	// Deletes the specified set of tags from the specified set of resources.
	//
	// To list the current tags, use DescribeTags. For more information about tags, see [Tag your Amazon EC2 resources] in the
	// Amazon Elastic Compute Cloud User Guide.
	//
	// [Tag your Amazon EC2 resources]: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html
	DeleteTags(ctx context.Context, params *ec2.DeleteTagsInput, optFns ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
	// Deletes the specified network ACL. You can't delete the ACL if it's associated
	// with any subnets. You can't delete the default network ACL.
	DeleteNetworkAcl(ctx context.Context, params *ec2.DeleteNetworkAclInput, optFns ...func(*ec2.Options)) (*ec2.DeleteNetworkAclOutput, error)
	// Deletes a security group.
	//
	// If you attempt to delete a security group that is associated with an instance
	// or network interface, is referenced by another security group in the same VPC,
	// or has a VPC association, the operation fails with DependencyViolation .
	DeleteSecurityGroup(ctx context.Context, params *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)
	// Describes your route tables. The default is to describe all your route tables.
	// Alternatively, you can specify specific route table IDs or filter the results to
	// include only the route tables that match specific criteria.
	//
	// Each subnet in your VPC must be associated with a route table. If a subnet is
	// not explicitly associated with any route table, it is implicitly associated with
	// the main route table. This command does not return the subnet ID for implicit
	// associations.
	//
	// For more information, see [Route tables] in the Amazon VPC User Guide.
	//
	// [Route tables]: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Route_Tables.html
	DescribeRouteTables(ctx context.Context, params *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error)
	// Disassociates a subnet or gateway from a route table.
	//
	// After you perform this action, the subnet no longer uses the routes in the
	// route table. Instead, it uses the routes in the VPC's main route table. For more
	// information about route tables, see [Route tables]in the Amazon VPC User Guide.
	//
	// [Route tables]: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Route_Tables.html
	DisassociateRouteTable(ctx context.Context, params *ec2.DisassociateRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.DisassociateRouteTableOutput, error)
	// Deletes the specified route table. You must disassociate the route table from
	// any subnets before you can delete it. You can't delete the main route table.
	DeleteRouteTable(ctx context.Context, params *ec2.DeleteRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error)
	// Deletes the specified NAT gateway. Deleting a public NAT gateway disassociates
	// its Elastic IP address, but does not release the address from your account.
	// Deleting a NAT gateway does not delete any NAT gateway routes in your route
	// tables.
	DeleteNatGateway(ctx context.Context, params *ec2.DeleteNatGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteNatGatewayOutput, error)
	// Releases the specified Elastic IP address.
	//
	// [Default VPC] Releasing an Elastic IP address automatically disassociates it
	// from any instance that it's associated with. To disassociate an Elastic IP
	// address without releasing it, use DisassociateAddress.
	//
	// [Nondefault VPC] You must use DisassociateAddress to disassociate the Elastic IP address before
	// you can release it. Otherwise, Amazon EC2 returns an error (
	// InvalidIPAddress.InUse ).
	//
	// After releasing an Elastic IP address, it is released to the IP address pool.
	// Be sure to update your DNS records and any servers or devices that communicate
	// with the address. If you attempt to release an Elastic IP address that you
	// already released, you'll get an AuthFailure error if the address is already
	// allocated to another Amazon Web Services account.
	//
	// After you release an Elastic IP address, you might be able to recover it. For
	// more information, see AllocateAddress.
	ReleaseAddress(ctx context.Context, params *ec2.ReleaseAddressInput, optFns ...func(*ec2.Options)) (*ec2.ReleaseAddressOutput, error)
	// Detaches an internet gateway from a VPC, disabling connectivity between the
	// internet and the VPC. The VPC must not contain any running instances with
	// Elastic IP addresses or public IPv4 addresses.
	DetachInternetGateway(ctx context.Context, params *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error)
	// Deletes the specified internet gateway. You must detach the internet gateway
	// from the VPC before you can delete it.
	DeleteInternetGateway(ctx context.Context, params *ec2.DeleteInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error)
	// Deletes the specified subnet. You must terminate all running instances in the
	// subnet before you can delete the subnet.
	DeleteSubnet(ctx context.Context, params *ec2.DeleteSubnetInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error)
	// Deletes the specified VPC. You must detach or delete all gateways and resources
	// that are associated with the VPC before you can delete it. For example, you must
	// terminate all instances running in the VPC, delete all security groups
	// associated with the VPC (except the default one), delete all route tables
	// associated with the VPC (except the default one), and so on. When you delete the
	// VPC, it deletes the default security group, network ACL, and route table for the
	// VPC.
	//
	// If you created a flow log for the VPC that you are deleting, note that flow
	// logs for deleted VPCs are eventually automatically removed.
	DeleteVpc(ctx context.Context, params *ec2.DeleteVpcInput, optFns ...func(*ec2.Options)) (*ec2.DeleteVpcOutput, error)
	// Describes the specified attribute of the specified VPC. You can specify only
	// one attribute at a time.
	DescribeVpcAttribute(ctx context.Context, params *ec2.DescribeVpcAttributeInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcAttributeOutput, error)
	// Describes your subnets. The default is to describe all your subnets.
	// Alternatively, you can specify specific subnet IDs or filter the results to
	// include only the subnets that match specific criteria.
	//
	// For more information, see [Subnets] in the Amazon VPC User Guide.
	//
	// [Subnets]: https://docs.aws.amazon.com/vpc/latest/userguide/configure-subnets.html
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	// Describes the specified security groups or all of your security groups.
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	// Describes your internet gateways. The default is to describe all your internet
	// gateways. Alternatively, you can specify specific internet gateway IDs or filter
	// the results to include only the internet gateways that match specific criteria.
	DescribeInternetGateways(ctx context.Context, params *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error)
}

// NewService creates a new instance of the Amazon EC2 service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return ec2.NewFromConfig(
		*awsConfig,
		ec2.WithEndpointResolverV2(
			&ec2EndpointResolverV2{
				providerContext,
			},
		),
	)
}

type ec2EndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *ec2EndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params ec2.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	ec2Aliases := utils.Services["ec2"]
	ec2Endpoint, hasEC2Endpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"ec2",
		ec2Aliases,
	)
	if hasEC2Endpoint && !core.IsScalarNil(ec2Endpoint) {
		u, err := url.Parse(core.StringValueFromScalar(ec2Endpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return ec2.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
