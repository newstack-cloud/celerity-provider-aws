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
					},
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
				Required:    []string{"internetGatewayId", "natGateways"},
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"internetGatewayId": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The ID of the internet gateway.",
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

const vpcResourcePresetDescription = `
The preset of the flex VPC resource.
This can be one of the following values:
- standard: A multi-AZ VPC with public and private subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 public subnets and 3 private subnets in total, each with a unique, non-overlapping CIDR block.
- public: A multi-AZ VPC with only public subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 public subnets in total, each with a unique, non-overlapping CIDR block.
- isolated: A multi-AZ VPC with only private subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 private subnets in total, each with a unique, non-overlapping CIDR block.
- light: A VPC with only a public subnet in a single availability zone. This is the most cost-effective option for a VPC that does not require private subnets for resources. The cost savings are primarily due to the fact that a NAT gateway is not required.
- light-public: A VPC with only a public subnet in a single availability zone. This is the most cost-effective option for a VPC that does not require private subnets for resources. The cost savings are primarily due to the fact that a NAT gateway is not required.
`
