## Flex VPC to Lambda Function Link

This link places a Lambda function within a flex VPC by:

1. **Subnet Placement**: Sets the function's `VpcConfig` with the subnet IDs from the VPC's computed `subnets`, selected by the tier (`public` or `private`) chosen with the `aws.flexvpc.lambda.subnetType` annotation.

2. **Security Group**: Attaches the VPC's Bluelink-managed security group to the function, so its network interfaces sit behind the VPC's deny-all-egress group that access links open as needed.

### Requirements

The flex VPC must be defined as a resource in the same blueprint. This works for a VPC in `create` mode and one in `reference` mode (an existing VPC brought into the blueprint), since both expose the same computed `subnets` (with per-subnet tier) and `securityGroups` outputs that the link reads.

Placement is a prerequisite for VPC networking activation: once the function is VPC-attached, the access links between it and its targets (for example `aws/lambda/function` → `aws/rds/dbInstance`) open the specific security-group rules or provision the VPC endpoints the function needs to reach those targets.

### Annotations

- `aws.flexvpc.lambda.subnetType` — which subnet tier to place the function in, `public` or `private` (default `private`). `private` lets the function reach private in-VPC resources and, where a NAT gateway exists, the internet; `public` reaches in-VPC resources only which avoids NAT gateway cost.

### Example

```yaml
resources:
  appVpc:
    type: aws/flex/vpc
    metadata:
      labels:
        app: orders
    linkSelector:
      byLabel:
        app: orders
    spec:
      name: orders-vpc
      preset: standard
      cidrBlock: 10.0.0.0/16
      # ... other VPC configuration

  getOrderFunction:
    type: aws/lambda/function
    metadata:
      labels:
        app: orders
      annotations:
        aws.flexvpc.lambda.subnetType: private
    spec:
      functionName: get-order
      # ... other function configuration
```

