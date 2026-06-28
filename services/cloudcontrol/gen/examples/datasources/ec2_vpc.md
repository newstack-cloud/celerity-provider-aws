Look up an existing AWS EC2 VPC by vpcId and export its cidrBlock.

```blueprintlang
version "2025-11-02"

data exampleVpc: aws/ec2/vpc {
    filter "vpcId" == "example-vpcid"

    export cidrBlock: string
    export defaultNetworkAcl: string
    export defaultSecurityGroup: string
}

export exampleVpcCidrBlock: string {
    field = datasources.exampleVpc.cidrBlock
}
```

```yaml
version: 2025-11-02

datasources:
  exampleVpc:
    type: aws/ec2/vpc
    filter:
      field: vpcId
      operator: "="
      search: example-vpcid
    exports:
      cidrBlock:
        type: string
      defaultNetworkAcl:
        type: string
      defaultSecurityGroup:
        type: string

exports:
  exampleVpcCidrBlock:
    type: string
    field: datasources.exampleVpc.cidrBlock
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleVpc": {
      "type": "aws/ec2/vpc",
      "filter": { "field": "vpcId", "operator": "=", "search": "example-vpcid" },
      "exports": {
        "cidrBlock": { "type": "string" },
        "defaultNetworkAcl": { "type": "string" },
        "defaultSecurityGroup": { "type": "string" }
      }
    }
  }
}
```
