# Flex VPC Design

This document describes the design of the Flex VPC resource used to allow practitioners to define a VPC and its associated resources based on presets that follow industry best practices.

## Overview

A flex VPC is typically defined in each blueprint that requires a VPC and is assigned a globally unique name (global to a single AWS account). The flex VPC resource will only be created if one with the same name does not already exist in the AWS account. Each resource that should be deployed in the VPC should have a link from the flex VPC where the labels match the `linkSelector.byLabel` field of the flex VPC resource. This will then enable links that connect resources in a blueprint together to create networking elements on the fly that "activate" the link such as security group rules, VPC endpoints, etc. This is only possible when a flex VPC is defined in the same blueprint and linked to the resources that should be deployed in the VPC. If this is not the case, all the links can do when it comes to networking is evaluate the current networking setup and return requested changes that would be required to enable the link.

An example of how this can be used is shown below:
```yaml
version: 2025-05-12
resources:
    myVPC:
        type: aws/flex/vpc
        linkSelector:
            byLabel:
                network: myVPC
        spec:
            name: myVPC
            preset: standard
            cidrBlock: 10.0.0.0/16
            region: eu-west-1
            tags:
                Environment: Production
                Project: MyProject
    myWebServer:
        type: aws/ec2/instance
        metadata:
            labels:
                network: myVPC
        linkSelector:
            byLabel:
                system: mySystem
        spec:
            name: myWebServer
            image: ami-01234567890123456
    myDatabase:
        type: aws/rds/dbCluster
        metadata:
            labels:
                network: myVPC
                system: mySystem
        spec:
            name: myDatabase
            engine: postgres
```

## Fields

### `name`

The name of the flex VPC.

This will be used to uniquely identify the flex VPC in the AWS account.

The name must be unique within the AWS account.

### `mode`

The mode of the flex VPC resource.

This can be one of the following values:

- `create`: The flex VPC will be created in the AWS account.
- `reference`: The flex VPC definition will be a reference to an existing VPC (and related resources) in the AWS account.

**Default:** `create`

### `preset`

A preset is a pre-defined configuration for the Flex VPC resource. It is used to specify the desired VPC configuration and the associated resources.

See the [Presets](#presets) section for more details on the available presets.

**Default:** `standard`

### `cidrBlock`

The CIDR block for the VPC.

The CIDR block must be a valid IPv4 CIDR block.

The CIDR block must be large enough to accommodate the number of subnets specified in the preset.

**Default:** `10.0.0.0/16`

If not specified, the VPC will be created with the default CIDR bluck.

**Note:** If you plan to peer this VPC with others or connect to on-premises networks, ensure the CIDR block does not overlap with any other network.

### `region`

The region for the VPC.

This will also be used to determine the availability zones for the subnets.

If not provided, the region will be inferred from the current deployment.

### `enableDNSSupport`

Whether DNS resolution is enabled for the VPC.

**Default:** `true`

### `enableDNSHostnames`

Whether DNS hostnames are enabled for the VPC.
This is useful for internal host names for EC2 instances.

**Default:** `false`

### `tags`

A map of custom tags to be applied to the VPC and its associated resources.

## Computed Fields

### `vpcId`

The ID of the concrete AWS VPC that will be created or referenced.

This will be computed after the VPC is created.

### `subnets`

A map of the subnets that will be created or referenced.
The keys are logical names of the subnets that are derived from the flex VPC preset.

An example of this would be:

```json
{
    "subnets": {
        "public-az-1": {
            "id": "subnet-01234567890123456",
            "availabilityZone": "us-east-1a",
        }
    }
}
```

### `routeTables`

A list of the route tables that will be created or referenced.
The route table structure will contain the ID of the route table and the subnet IDs that it is associated with.

An example of this would be:

```json
{
    "routeTables": [
        {
            "id": "rtb-01234567890123456",
            "subnetIds": [
                "subnet-01234567890123456",
                "subnet-01234567890123457",
                "subnet-01234567890123458"
            ]
        }
    ]
}
```

### `securityGroups`

A list of the security group IDs that will be created or referenced.

An example of this would be:

```json
{
    "securityGroups": [
        "sg-01234567890123456",
        "sg-01234567890123457",
        "sg-01234567890123458"
    ]
}
```

### `networkAcls`

A mapping of network ACLs to subnet IDs that they are associated with.
The key is a logical name of the network ACL as defined by the flex VPC resource implementation.

An example of this would be:

```json
{
    "networkAcls": {
        "default": {
            "id": "acl-01234567890123456",
            "subnetIds": [
                "subnet-01234567890123456",
                "subnet-01234567890123457",
                "subnet-01234567890123458"
            ]
        }
    }
}
```

### `gateways`

A structure containing of internet and NAT gateways that will be created or referenced.

An example of this would be:

```json
{
    "gateways": {
        "internetGatewayId": "igw-01234567890123456",
        "natGateways": [
            {
                "id": "nat-01234567890123456",
                "elasticIpId": "eipalloc-01234567890123456",
                "inPublicSubnetId": "subnet-01234567890123456",
                "forPrivateSubnetId": "subnet-01234567890123457"
            },
            {
                "id": "nat-01234567890123457",
                "elasticIpId": "eipalloc-01234567890123457",
                "inPublicSubnetId": "subnet-01234567890123457",
                "forPrivateSubnetId": "subnet-01234567890123458"
            }
        ]
    }
}
```

## Presets

The Flex VPC resource supports the following presets:

### `standard`

A multi-AZ VPC with public and private subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 public subnets and 3 private subnets in total, each with a unique, non-overlapping CIDR block.

- **Subnets**: 3 public and 3 private subnets, each with a unique, non-overlapping CIDR block. Subnet CIDR blocks are automatically and deterministically derived from the VPC's CIDR block defined in the `cidrBlock` field. This process makes sure there are no overlaps and that consistent sizing is used. For example, a `/16` VPC with 6 subnets will result in each subnet having a `/19` block.
- **Availiability Zones**: The availability zones are selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **NAT Gateways**: 3 NAT Gateways, one in each public subnet to cover each availability zone, providing zone-local egress for private subnets.
- **Internet Gateway**: An internet gateway is attached to the VPC, enabling inbound/outbound internet access for public subnets. IPv6 will be enabled by default in addition to IPv4.
- **Route Tables**: Public subnets route 0.0.0.0/0 to the internet gateway; private subnets route 0.0.0.0/0 to the NAT gateway in their AZ.
- **Security**: An initial security group is provisioned that denies all traffic by default, the usage of links between resources in a blueprint will enable rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable for production workloads that require high availability, scalability and secure internet access for private resources. Examples of use cases would be web applications, data processing workloads and databases.
- **Tagging**: All resources in the VPC will be tagged with Bluelink's default tags to identify the resources as a part of a Flex VPC along with user-defined tags for the flex VPC resource.

### `public`

A multi-AZ VPC with only public subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 public subnets in total, each with a unique, non-overlapping CIDR block.

- **Subnets**: 3 public subnets, each with a unique, non-overlapping CIDR block. Subnet CIDR blocks are automatically and deterministically derived from the VPC's CIDR block defined in the `cidrBlock` field. This process makes sure there are no overlaps and that consistent sizing is used. For example, a `/16` VPC with N subnets will result in each subnet having a `/18` block.
- **Availiability Zones**: The availability zones are selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **Internet Gateway**: An internet gateway is attached to the VPC, enabling inbound/outbound internet access for public subnets. IPv6 will be enabled by default in addition to IPv4.
- **Route Tables**: Public subnets route 0.0.0.0/0 to the internet gateway.
- **Security**: An initial security group is provisioned to deny all traffic by default. A deny-all NACL will also be created as an additional layer of security for public-only VPCs as a starting point. The usage of links between resources in a blueprint will enable security group and NACL rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable when all resources need direct internet access and there is no requirement for resources to be isolated from the public internet. This is not typical for production workloads, but can be very valuable for development, testing and public-facing services with lenient security requirements.
- **Tagging**: All resources in the VPC will be tagged with Bluelink's default tags to identify the resources as a part of a Flex VPC along with user-defined tags for the flex VPC resource.

### `isolated`

A multi-AZ VPC with only private subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 private subnets in total, each with a unique, non-overlapping CIDR block.

- **Subnets**: 3 private subnets, each with a unique, non-overlapping CIDR block. Subnet CIDR blocks are automatically and deterministically derived from the VPC's CIDR block defined in the `cidrBlock` field. This process makes sure there are no overlaps and that consistent sizing is used. For example, a `/16` VPC with N subnets will result in each subnet having a `/18` block.
- **Availiability Zones**: The availability zones are selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **Internet Gateway**: No internet gateway is attached to the VPC, no inbound/outbound internet access is allowed for subnets.
- **NAT Gateways**: No NAT Gateways are deployed.
- **Route Tables**: There are no routes to 0.0.0.0/0 (the internet) or to any NAT or internet gateway. Route entries will be added dynamically by links for specific VPC endpoints and VPC peering connections.
- **Security**: An initial security group is provisioned to deny all traffic by default, the usage of links between resources in a blueprint will enable rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable for workloads that need to be isolated from the public internet. This is useful for internal databases and workloads with strict security requirements.
- **Tagging**: All resources in the VPC will be tagged with Bluelink's default tags to identify the resources as a part of a Flex VPC along with user-defined tags for the flex VPC resource.

### `light`

A VPC with one public and one private subnet in a single availability zone. The public and private subnets have a unique, non-overlapping CIDR block.

This preset is the most cost-effective option for a VPC that still requires private subnets for some resources.

- **Subnets**: 1 public and 1 private subnet, each with a unique, non-overlapping CIDR block. Subnet CIDR blocks are automatically and deterministically derived from the VPC's CIDR block defined in the `cidrBlock` field. This process makes sure there are no overlaps and that consistent sizing is used. For example, a `/16` VPC with 2 subnets will result in each subnet having a `/17` block.
- **Availiability Zones**: The availability zone is selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **Internet Gateway**: An internet gateway is attached to the VPC, enabling inbound/outbound internet access for the public subnet. IPv6 will be enabled by default in addition to IPv4.
- **NAT Gateways**: a NAT gateway is deployed to allow the private subnet access to the public internet.
- **Route Tables**: The public subnet route 0.0.0.0/0 to the internet gateway; the private subnet route 0.0.0.0/0 to the NAT gateway in their AZ.
- **Security**: An initial security group is provisioned to deny all traffic by default, the usage of links between resources in a blueprint will enable rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable for small workloads that do not require high availability. This is useful for development, testing and internal services that aren't business critical.

### `light-public`

A VPC with only a public subnet in a single availability zone.

This preset is the most cost-effective option for a VPC that does not require private subnets for resources. The cost savings are primarily due to the fact that a NAT gateway is not required.

- **Subnets**: 1 public subnet, the CIDR block will be derived from the VPC's CIDR block defined in the `cidrBlock` field.
- **Availiability Zones**: The availability zone is selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **Internet Gateway**: An internet gateway is attached to the VPC, enabling inbound/outbound internet access for the public subnet. IPv6 will be enabled by default in addition to IPv4.
- **Route Tables**: The public subnet route 0.0.0.0/0 to the internet gateway.
- **Security**: An initial security group is provisioned to deny all traffic by default. A deny-all NACL will also be created as an additional layer of security for public-only VPCs as a starting point. The usage of links between resources in a blueprint will enable security group and NACL rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable for small workloads that do not require high availability. This is useful for development, testing and internal services that aren't business critical.

## Validation

Validation should be performed to ensure that the selected preset is compatible with the provided `cidrBlock` and `region`.

## Links and Subnet selection

The link implementations for `aws/flex/vpc` -> `<targetResourceType>`, by default, will determine the subnets to place the target resources in based on the target resource type and best practices (e.g., EC2 → private, LB → public, RDS → private).

This can, however, be overridden by the user by specifying the `aws.flex.vpc.subnetType` annotation on the target resource that will determine whether the target resource should be placed in a public or private subnet. When an invalid subnet type is specified, at the load stage of a deployment/change staging process, validation will fail.

For resources that can be configured with multi-AZ support, the link implementation will configure the resource to be deployed to all the availability zones of the selected flex VPC preset. The user can override this by specifying the specific availability zones to deploy the resource to using the `aws.flex.vpc.availabilityZones` annotation (This will be a comma-separated list of availability zones). If an invalid availability zone is specified, a failure will occur at the deployment stage as the details of the deployed infrastructure for a flex VPC is not tracked in state, only the high-level definition.

_Only one `aws/flex/vpc` resource should be linked to any given resource in the same blueprint._

## State management and Flex VPCs

Flex VPCs are designed to be flexible resources that allow links to modify/augment with rules and networking elements that activate connections between resources. This means that the state of all the resources that make up a flex VPC is not tracked in state, only the high-level definition, otherwise, issues such as constant drift detection will require reconciliation of changes made to the networking infrastructure by links across multiple application and infrastructure blueprints will be required.

## Note on Manual Changes

Manual changes to a flex VPC may not be detected by the system and could lead to unexpected behaviour. If you need more control over the networking infrastructure, it is recommended to use the `aws/vpc` and related networking resources instead, like you would with other IaC tools.

## Sharing Flex VPCs between blueprints

The same Flex VPC can be used in multiple blueprints for multiple application and infrastructure blueprints. To be able to accurately track which resources and rules are associated with which links for resources, a tagging approach is used to ensure that the correct resources are associated with the correct links. Often, the same security group rules, route table entries and VPC endpoints will be used to activate links between resources across multiple blueprints. To make sure that rules and resources are not duplicated, orphaned or incorrectly removed, tags will be used to track usage of resources and rules in links. Links should always check for the presence of tags from other blueprint links to determine whether or not a resource should be removed.

Links should also use detection of standardised flex VPC tags to determine whether a resource can be modified by a link. If a resource is not tagged with the standard flex VPC tags, the link should instead report the changes that would be required to enable the link.

The link tag format is as follows: `bluelink:links:{blueprintInstanceId}={linkName1},{linkName2}` where `{blueprintInstanceId}` is the instance ID of the blueprint that the link is defined in and `{linkName1..N}` is the logical name of the link relative to the blueprint.

For links within the same blueprint, a local blueprint lock will be acquired to ensure only one link in the same blueprint can modify the networking infrastructure for a flex VPC at a time.

### Link cleanup behaviour

When a link is removed, its name is removed from the tags associated with flex VPC resources and the resource is only deleted if there are no other links associated with the resource.

### Definitions across multiple blueprints

Only one flex VPC definition across all blueprints should have configuration for the same name. All other flex VPC resource definitions must be set with the `mode` field set to `reference`, where validation will fail if a flex VPC in reference mode contains additional configuration.

If you don't set the `mode` field to `reference`, the deployment will fail if the flex VPC has already been created in the target AWS account.
