Look up an existing AWS EC2 SecurityGroup by id and export its groupDescription.

```blueprintlang
version "2025-11-02"

data exampleSecurityGroup: aws/ec2/securityGroup {
    filter "id" == "example-id"

    export groupDescription: string
    export groupId: string
    export groupName: string
}

export exampleSecurityGroupGroupDescription: string {
    field = datasources.exampleSecurityGroup.groupDescription
}
```

```yaml
version: 2025-11-02

datasources:
  exampleSecurityGroup:
    type: aws/ec2/securityGroup
    filter:
      field: id
      operator: "="
      search: example-id
    exports:
      groupDescription:
        type: string
      groupId:
        type: string
      groupName:
        type: string

exports:
  exampleSecurityGroupGroupDescription:
    type: string
    field: datasources.exampleSecurityGroup.groupDescription
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleSecurityGroup": {
      "type": "aws/ec2/securityGroup",
      "filter": { "field": "id", "operator": "=", "search": "example-id" },
      "exports": {
        "groupDescription": { "type": "string" },
        "groupId": { "type": "string" },
        "groupName": { "type": "string" }
      }
    }
  }
}
```
