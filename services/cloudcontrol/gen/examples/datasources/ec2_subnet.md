Look up an existing AWS EC2 Subnet by subnetId and export its assignIpv6AddressOnCreation.

```blueprintlang
version "2025-11-02"

data exampleSubnet: aws/ec2/subnet {
    filter "subnetId" == "example-subnetid"

    export assignIpv6AddressOnCreation: boolean
    export availabilityZone: string
    export availabilityZoneId: string
}

export exampleSubnetAssignIpv6AddressOnCreation: boolean {
    field = datasources.exampleSubnet.assignIpv6AddressOnCreation
}
```

```yaml
version: 2025-11-02

datasources:
  exampleSubnet:
    type: aws/ec2/subnet
    filter:
      field: subnetId
      operator: "="
      search: example-subnetid
    exports:
      assignIpv6AddressOnCreation:
        type: boolean
      availabilityZone:
        type: string
      availabilityZoneId:
        type: string

exports:
  exampleSubnetAssignIpv6AddressOnCreation:
    type: boolean
    field: datasources.exampleSubnet.assignIpv6AddressOnCreation
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleSubnet": {
      "type": "aws/ec2/subnet",
      "filter": { "field": "subnetId", "operator": "=", "search": "example-subnetid" },
      "exports": {
        "assignIpv6AddressOnCreation": { "type": "boolean" },
        "availabilityZone": { "type": "string" },
        "availabilityZoneId": { "type": "string" }
      }
    }
  }
}
```
