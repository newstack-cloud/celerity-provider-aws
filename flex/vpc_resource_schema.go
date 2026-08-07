package flex

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/validation"
)

func vpcResourceSchema() *provider.ResourceDefinitionsSchema {
	presetDescription, _ := descriptions.ReadFile("resources/vpc_preset.md")

	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "FlexVPCDefinition",
		Description: "The definition of an AWS Flex VPC.",
		Required:    []string{"name", "preset", "cidrBlock", "region"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"name": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name of the flex VPC that will be unique in the target AWS account. " +
					"This will be attached as a tag to the concrete AWS VPC and associated resources.",
				FormattedDescription: "The name of the flex VPC that will be unique in the target AWS account. " +
					"This will be attached as a tag to the concrete AWS VPC and associated resources. " +
					"The format for the tag is `bluelink:flex-vpc:name={name}`.",
				MustRecreate: true,
			},
			"mode": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The mode of the flex VPC resource. This can be one of \"create\" or \"reference\". " +
					"\"create\" is used for when the flex VPC will be created in the AWS account. " +
					"\"reference\" is used when the flex VPC definition will be a reference to an existing VPC (and related resources) in the AWS account. " +
					"The default value is \"create\".",
				FormattedDescription: "The mode of the flex VPC resource. " +
					"This can be one of the following values: " +
					"- `create`: The flex VPC will be created in the AWS account. \n" +
					"- `reference`: The flex VPC definition will be a reference to an existing VPC (and related resources) in the AWS account.\n" +
					"The default value is `create`.",
				MustRecreate: true,
				Default:      core.MappingNodeFromString("create"),
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("create"),
					core.MappingNodeFromString("reference"),
				},
			},
			"preset": {
				Type:                 provider.ResourceDefinitionsSchemaTypeString,
				Description:          vpcResourcePresetDescription,
				FormattedDescription: string(presetDescription),
				MustRecreate:         true,
				Default:              core.MappingNodeFromString("standard"),
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("standard"),
					core.MappingNodeFromString("public"),
					core.MappingNodeFromString("isolated"),
					core.MappingNodeFromString("light"),
					core.MappingNodeFromString("light-public"),
				},
				ValidateFunc: validation.ConflictsWithResourceDefinition(
					"$.mode",
					&validation.ConflictOptions{
						ConflictsOnValue: core.ScalarFromString("reference"),
					},
				),
			},
			"cidrBlock": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The CIDR block of the flex VPC. \n" +
					"This is the IPv4 address range for the VPC. \n" +
					"The CIDR block must be large enough to accommodate the number of subnets specified in the preset.",
				FormattedDescription: "The CIDR block of the flex VPC. \n" +
					"This is the IPv4 address range for the VPC. \n" +
					"The CIDR block must be large enough to accommodate the number of subnets specified in the preset.",
				MustRecreate: true,
				Default:      core.MappingNodeFromString("10.0.0.0/16"),
				ValidateFunc: validation.ConflictsWithResourceDefinition(
					"$.mode",
					&validation.ConflictOptions{
						ConflictsOnValue: core.ScalarFromString("reference"),
					},
				),
			},
			"enableDNSSupport": {
				Type: provider.ResourceDefinitionsSchemaTypeBoolean,
				Description: "Indicates whether DNS resolution is supported for the VPC. If enabled, queries to the " +
					"Amazon provided DNS server at the 169.254.169.253 IP address, or the reserved IP address at the base of the " +
					"VPC network range \"plus two\" succeed. If disabled, the Amazon provided DNS service in the VPC that resolves " +
					"public DNS hostnames to IP addresses is not enabled. Enabled by default.",
				FormattedDescription: "Indicates whether DNS resolution is supported for the VPC. If enabled, queries to the " +
					"Amazon provided DNS server at the 169.254.169.253 IP address, or the reserved IP address at the base of the " +
					"VPC network range \"plus two\" succeed. If disabled, the Amazon provided DNS service in the VPC that resolves " +
					"public DNS hostnames to IP addresses is not enabled. Enabled by default. For more information, see [DNS attributes in your VPC.](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-dns.html#vpc-dns-support)",
				MustRecreate: false,
				Default:      core.MappingNodeFromBool(true),
				ValidateFunc: validation.ConflictsWithResourceDefinition(
					"$.mode",
					&validation.ConflictOptions{
						ConflictsOnValue: core.ScalarFromString("reference"),
					},
				),
			},
			"enableDNSHostnames": {
				Type: provider.ResourceDefinitionsSchemaTypeBoolean,
				Description: "Indicates whether the instances launched in the VPC get DNS hostnames. " +
					"If enabled, instances in the VPC get DNS hostnames; otherwise, they do not. " +
					"Disabled by default for nondefault VPCs. You will want to enable DNS hostnames if you intend to use VPC endpoints " +
					"to connect to AWS services using the public host names that SDKs use to connect to the services by default.",
				FormattedDescription: "Indicates whether the instances launched in the VPC get DNS hostnames. " +
					"If enabled, instances in the VPC get DNS hostnames; otherwise, they do not. " +
					"Disabled by default for nondefault VPCs. For more information, see [DNS attributes in your VPC.](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-dns.html#vpc-dns-hostnames)\n" +
					"You will want to enable DNS hostnames if you intend to use VPC endpoints " +
					"to connect to AWS services using the public host names that SDKs use to connect to the services by default.",
				MustRecreate: false,
				Default:      core.MappingNodeFromBool(false),
				ValidateFunc: validation.ConflictsWithResourceDefinition(
					"$.mode",
					&validation.ConflictOptions{
						ConflictsOnValue: core.ScalarFromString("reference"),
					},
				),
			},
			"region": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The region of the flex VPC. \n" +
					"This is the region where the flex VPC will be created. \n" +
					"The fallback value is the region of the current deployment.",
				FormattedDescription: "The region of the flex VPC. \n" +
					"This is the region where the flex VPC will be created. \n" +
					"The fallback value is the region of the current deployment.",
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("us-east-1"),
				},
				MustRecreate: true,
				ValidateFunc: validation.ConflictsWithResourceDefinition(
					"$.mode",
					&validation.ConflictOptions{
						ConflictsOnValue: core.ScalarFromString("reference"),
					},
				),
			},
			"tags": {
				Type: provider.ResourceDefinitionsSchemaTypeMap,
				Description: "The tags of the flex VPC. \n" +
					"This is the tags that will be attached to the flex VPC and associated resources.",
				FormattedDescription: "The tags of the flex VPC. \n" +
					"This is the tags that will be attached to the flex VPC and associated resources.",
				Examples: []*core.MappingNode{
					core.MappingNodeFields(
						"bluelink:flex-vpc:name", core.ScalarFromString("my-flex-vpc"),
					),
				},
				MustRecreate: false,
				ValidateFunc: validation.ConflictsWithResourceDefinition(
					"$.mode",
					&validation.ConflictOptions{
						ConflictsOnValue: core.ScalarFromString("reference"),
					},
				),
			},

			// Computed fields.
			// Apart from the VPC ID, all the other computed fields are tracked for drift detection
			// to allow for detecting changes to the core network infrastructure resources created
			// for a flex VPC (subnets, route tables, security groups, network ACLs, gateways).
			"vpcId": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The ID of the concrete AWS VPC that will be created or referenced.",
				Computed:    true,
			},
			"subnets": {
				Type:        provider.ResourceDefinitionsSchemaTypeMap,
				Description: "A map of the subnets that will be created or referenced.",
				Computed:    true,
				TrackDrift:  true,
				MapValues: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeObject,
					Label:       "SubnetInfo",
					Description: "Basic information about the subnet.",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"id": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The ID of the subnet.",
						},
						"availabilityZone": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The availability zone of the subnet.",
						},
						"ipv6CidrBlock": {
							Type: provider.ResourceDefinitionsSchemaTypeString,
							Description: "The IPv6 CIDR block associated with the subnet. Only present " +
								"for dual-stack subnets. Links that place resources in the VPC use this " +
								"to decide whether the placed resource can be given outbound IPv6.",
						},
						"subnetType": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The tier of the subnet, either \"public\" or \"private\". Links that place resources in the VPC use this to select subnets by tier.",
							AllowedValues: []*core.MappingNode{
								core.MappingNodeFromString("public"),
								core.MappingNodeFromString("private"),
							},
						},
					},
				},
			},
			"privateSubnetIds": {
				Type: provider.ResourceDefinitionsSchemaTypeArray,
				Description: "The IDs of the private-tier subnets, ordered by availability zone. " +
					"Directly referenceable for placing resources in the private tier, such as the " +
					"subnetIds of an ElastiCache or RDS subnet group. " +
					"The \"standard\" and \"isolated\" presets provide 3 private subnets across 3 distinct " +
					"availability zones, satisfying services that require subnets in at least 2 zones " +
					"(such as RDS subnet groups and RDS Proxy). " +
					"The \"light\" preset provides a single private subnet in one availability zone, which does " +
					"not satisfy those services. " +
					"For presets without private subnets (\"public\", \"light-public\"), this is an empty list. " +
					"In \"reference\" mode, this reflects the referenced VPC's private-tagged subnets with no guaranteed minimum.",
				Computed:   true,
				TrackDrift: true,
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The ID of a private-tier subnet.",
				},
			},
			"publicSubnetIds": {
				Type: provider.ResourceDefinitionsSchemaTypeArray,
				Description: "The IDs of the public-tier subnets, ordered by availability zone. " +
					"Directly referenceable for placing resources in the public tier. " +
					"The \"standard\" and \"public\" presets provide 3 public subnets across 3 distinct " +
					"availability zones; the \"light\" and \"light-public\" presets provide a single public subnet. " +
					"For the \"isolated\" preset, this is an empty list. " +
					"In \"reference\" mode, this reflects the referenced VPC's public-tagged subnets with no guaranteed minimum.",
				Computed:   true,
				TrackDrift: true,
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The ID of a public-tier subnet.",
				},
			},
			"routeTables": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of the route tables that will be created or referenced.",
				Computed:    true,
				TrackDrift:  true,
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeObject,
					Label:       "RouteTableInfo",
					Description: "Basic information about the route table.",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"id": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The ID of the route table.",
						},
						"subnetIds": {
							Type:        provider.ResourceDefinitionsSchemaTypeArray,
							Description: "The subnet IDs associated with this route table.",
							Items: &provider.ResourceDefinitionsSchema{
								Type:        provider.ResourceDefinitionsSchemaTypeString,
								Description: "The ID of a subnet associated with this route table.",
							},
						},
					},
				},
			},
			"securityGroups": {
				Type:                 provider.ResourceDefinitionsSchemaTypeArray,
				Description:          vpcResourceSecurityGroupsDescription,
				FormattedDescription: vpcResourceSecurityGroupsDescription,
				MustRecreate:         false,
				Nullable:             true,
				Items: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeString,
					Description: "A name for the group, unique within this VPC. It is the key " +
						"the group is exposed under in securityGroupIdsByName.",
				},
			},
			"securityGroupIdsByName": {
				Type: provider.ResourceDefinitionsSchemaTypeMap,
				Description: "The ID of the security group created for each name in securityGroups, " +
					"keyed by that name. A resource in the VPC references one of these from its own " +
					"security group field, and the link between a workload and that resource pairs " +
					"the two groups.",
				Computed:   true,
				TrackDrift: true,
				MapValues: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The ID of the named security group.",
				},
			},
			"securityGroupIds": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of the security groups that will be created or referenced.",
				Computed:    true,
				TrackDrift:  true,
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The ID of the security group.",
				},
			},
			"networkAcls": {
				Type:        provider.ResourceDefinitionsSchemaTypeMap,
				Description: "A map of the network ACLs that will be created or referenced.",
				Computed:    true,
				TrackDrift:  true,
				MapValues: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeObject,
					Label:       "NetworkAclInfo",
					Description: "Basic information about the network ACL.",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"id": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The ID of the network ACL.",
						},
						"subnetIds": {
							Type:        provider.ResourceDefinitionsSchemaTypeArray,
							Description: "The subnet IDs associated with this network ACL.",
							Items: &provider.ResourceDefinitionsSchema{
								Type:        provider.ResourceDefinitionsSchemaTypeString,
								Description: "The ID of a subnet associated with this network ACL.",
							},
						},
					},
				},
			},
			"gateways": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "A structure containing internet and NAT gateways that will be created or referenced.",
				Computed:    true,
				TrackDrift:  true,
				Required:    []string{"natGateways"},
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"internetGatewayId": {
						Type: provider.ResourceDefinitionsSchemaTypeString,
						Description: "The ID of the internet gateway. Absent for presets that " +
							"provision no public subnets, such as the isolated preset.",
					},
					"egressOnlyInternetGatewayId": {
						Type: provider.ResourceDefinitionsSchemaTypeString,
						Description: "The ID of the egress-only internet gateway that carries outbound " +
							"IPv6 traffic from private subnets. Only present for presets whose private " +
							"subnets route to the internet.",
					},
					"natGateways": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "A list of NAT gateways.",
						Items: &provider.ResourceDefinitionsSchema{
							Type:        provider.ResourceDefinitionsSchemaTypeObject,
							Label:       "NATGatewayInfo",
							Description: "Basic information about the NAT gateway.",
							Attributes: map[string]*provider.ResourceDefinitionsSchema{
								"id": {
									Type:        provider.ResourceDefinitionsSchemaTypeString,
									Description: "The ID of the NAT gateway.",
								},
								"elasticIpId": {
									Type:        provider.ResourceDefinitionsSchemaTypeString,
									Description: "The ID of the elastic IP associated with this NAT gateway.",
								},
								"inPublicSubnetId": {
									Type:        provider.ResourceDefinitionsSchemaTypeString,
									Description: "The ID of the subnet where this NAT gateway is located.",
								},
								"forPrivateSubnetId": {
									Type:        provider.ResourceDefinitionsSchemaTypeString,
									Description: "The ID of the private subnet that the NAT gateway is has been created for.",
								},
							},
						},
					},
				},
			},
		},
	}
}

const vpcResourceSecurityGroupsDescription = `
Names of empty security groups to create in this VPC, for resources that need an identity of their
own within it.

A database or cache in the VPC references one of these groups from its own spec, through
` + "`securityGroupIdsByName`" + `. The link between a workload and that resource then opens a path
between the workload's group and this one, on the port the resource listens on.

Each name gets a group of its own so that reach is granted per resource rather than to everything in
the VPC. Two databases sharing a group means a workload linked to one of them can reach the other on
the same port, whether or not it is linked to it.

The groups are created empty and stay that way unless a link opens something on them. The VPC never
writes rules of its own.
`

const vpcResourcePresetDescription = `
The preset of the flex VPC resource.
This can be one of the following values:
- standard: A multi-AZ VPC with public and private subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 public subnets and 3 private subnets in total, each with a unique, non-overlapping CIDR block.
- public: A multi-AZ VPC with only public subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 public subnets in total, each with a unique, non-overlapping CIDR block.
- isolated: A multi-AZ VPC with only private subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 private subnets in total, each with a unique, non-overlapping CIDR block.
- light: A VPC with a public and a private subnet in a single availability zone, with a NAT gateway for the private subnet. This is a single-AZ equivalent of the "standard" preset, for workloads that do not need multi-AZ redundancy. Single-AZ private subnets do not satisfy services that require subnets in at least two availability zones, such as RDS subnet groups and RDS Proxy.
- light-public: A VPC with only a public subnet in a single availability zone. This is the most cost-effective option for a VPC that does not require private subnets for resources. The cost savings are primarily due to the fact that a NAT gateway is not required.
`
