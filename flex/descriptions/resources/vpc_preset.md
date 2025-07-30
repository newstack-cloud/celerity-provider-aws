
The Flex VPC resource supports the following presets:

**`standard`**

A multi-AZ VPC with public and private subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 public subnets and 3 private subnets in total, each with a unique, non-overlapping CIDR block.

- **Subnets**: 3 public and 3 private subnets, each with a unique, non-overlapping CIDR block. Subnet CIDR blocks are automatically and deterministically derived from the VPC's CIDR block defined in the `cidrBlock` field. This process makes sure there are no overlaps and that consistent sizing is used. For example, a `/16` VPC with 6 subnets will result in each subnet having a `/19` block.
- **Availiability Zones**: The availability zones are selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **NAT Gateways**: 3 NAT Gateways, one in each public subnet to cover each availability zone, providing zone-local egress for private subnets.
- **Internet Gateway**: An internet gateway is attached to the VPC, enabling inbound/outbound internet access for public subnets. IPv6 will be enabled by default in addition to IPv4.
- **Route Tables**: Public subnets route 0.0.0.0/0 to the internet gateway; private subnets route 0.0.0.0/0 to the NAT gateway in their AZ.
- **Security**: Default security groups and NACLs deny all traffic by default, the usage of links between resources in a blueprint will enable rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable for production workloads that require high availability, scalability and secure internet access for private resources. Examples of use cases would be web applications, data processing workloads and databases.
- **Tagging**: All resources in the VPC will be tagged with Bluelink's default tags to identify the resources as a part of a Flex VPC along with user-defined tags for the flex VPC resource.

**`public`**

A multi-AZ VPC with only public subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 public subnets in total, each with a unique, non-overlapping CIDR block.

- **Subnets**: 3 public subnets, each with a unique, non-overlapping CIDR block. Subnet CIDR blocks are automatically and deterministically derived from the VPC's CIDR block defined in the `cidrBlock` field. This process makes sure there are no overlaps and that consistent sizing is used. For example, a `/16` VPC with N subnets will result in each subnet having a `/18` block.
- **Availiability Zones**: The availability zones are selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **Internet Gateway**: An internet gateway is attached to the VPC, enabling inbound/outbound internet access for public subnets. IPv6 will be enabled by default in addition to IPv4.
- **Route Tables**: Public subnets route 0.0.0.0/0 to the internet gateway.
- **Security**: Default security groups and NACLs deny all traffic by default, the usage of links between resources in a blueprint will enable rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable when all resources need direct internet access and there is no requirement for resources to be isolated from the public internet. This is not typical for production workloads, but can be very valuable for development, testing and public-facing services with lenient security requirements.
- **Tagging**: All resources in the VPC will be tagged with Bluelink's default tags to identify the resources as a part of a Flex VPC along with user-defined tags for the flex VPC resource.

**`isolated`**

A multi-AZ VPC with only private subnets in 3 availability zones (AZs) for high availability and fault tolerance. That is 3 private subnets in total, each with a unique, non-overlapping CIDR block.

- **Subnets**: 3 private subnets, each with a unique, non-overlapping CIDR block. Subnet CIDR blocks are automatically and deterministically derived from the VPC's CIDR block defined in the `cidrBlock` field. This process makes sure there are no overlaps and that consistent sizing is used. For example, a `/16` VPC with N subnets will result in each subnet having a `/18` block.
- **Availiability Zones**: The availability zones are selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **Internet Gateway**: No internet gateway is attached to the VPC, no inbound/outbound internet access is allowed for subnets.
- **NAT Gateways**: No NAT Gateways are deployed.
- **Route Tables**: There are no routes to 0.0.0.0/0 (the internet) or to any NAT or internet gateway. Route entries will be added dynamically by links for specific VPC endpoints and VPC peering connections.
- **Security**: Default security groups and NACLs deny all traffic by default, the usage of links between resources in a blueprint will enable rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable for workloads that need to be isolated from the public internet. This is useful for internal databases and workloads with strict security requirements.
- **Tagging**: All resources in the VPC will be tagged with Bluelink's default tags to identify the resources as a part of a Flex VPC along with user-defined tags for the flex VPC resource.

**`light`**

A VPC with one public and one private subnet in a single availability zone. The public and private subnets have a unique, non-overlapping CIDR block.

This preset is the most cost-effective option for a VPC that still requires private subnets for some resources.

- **Subnets**: 1 public and 1 private subnet, each with a unique, non-overlapping CIDR block. Subnet CIDR blocks are automatically and deterministically derived from the VPC's CIDR block defined in the `cidrBlock` field. This process makes sure there are no overlaps and that consistent sizing is used. For example, a `/16` VPC with 2 subnets will result in each subnet having a `/17` block.
- **Availiability Zones**: The availability zone is selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **Internet Gateway**: An internet gateway is attached to the VPC, enabling inbound/outbound internet access for the public subnet. IPv6 will be enabled by default in addition to IPv4.
- **NAT Gateways**: a NAT gateway is deployed to allow the private subnet access to the public internet.
- **Route Tables**: The public subnet route 0.0.0.0/0 to the internet gateway; the private subnet route 0.0.0.0/0 to the NAT gateway in their AZ.
- **Security**: Default security groups and NACLs deny all traffic by default, the usage of links between resources in a blueprint will enable rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable for small workloads that do not require high availability. This is useful for development, testing and internal services that aren't business critical.

**`light-public`**

A VPC with only a public subnet in a single availability zone.

This preset is the most cost-effective option for a VPC that does not require private subnets for resources. The cost savings are primarily due to the fact that a NAT gateway is not required.

- **Subnets**: 1 public subnet, the CIDR block will be derived from the VPC's CIDR block defined in the `cidrBlock` field.
- **Availiability Zones**: The availability zone is selected based on the current region of the deployment or a custom region that can be specified for the flex VPC resource.
- **Internet Gateway**: An internet gateway is attached to the VPC, enabling inbound/outbound internet access for the public subnet. IPv6 will be enabled by default in addition to IPv4.
- **Route Tables**: The public subnet route 0.0.0.0/0 to the internet gateway.
- **Security**: Default security groups and NACLs deny all traffic by default, the usage of links between resources in a blueprint will enable rules that allow specific traffic based on the linked resources.
- **Use Case**: This preset is suitable for small workloads that do not require high availability. This is useful for development, testing and internal services that aren't business critical.
