## VPC to Lambda Function

Runs a Lambda function inside a VPC.

When a VPC links to a function, the function is placed in the VPC's subnets so it runs on the VPC's private network. Use the `aws.flexvpc.lambda.subnetType` annotation to choose which subnet tier it runs in.

Once a function runs inside the VPC, links from the function to other resources in that VPC (such as a database or cache) set up the network access it needs to reach them.

### Annotations

- `aws.flexvpc.lambda.subnetType` is the subnet tier to place the function in, `public` or `private` (default `private`).
  - `private` lets the function reach private resources in the VPC and, where the VPC provides a NAT gateway, the internet.
  - `public` lets the function reach private resources in the VPC but has no outbound internet access, which avoids NAT gateway cost for VPCs provisioned without one.

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
