**JSON with Commas and Comments**

```javascript
{
    "resources": {
        "myVPC": {
            "type": "aws/flex/vpc",
            "linkSelector": {
                "byLabel": {
                    "network": "myVPC"
                }
            },
            "spec": {
                "name": "myVPC",
                "preset": "standard",
                // Pick a CIDR block that will comfortably fit the number of resources
                // that will be deployed to the VPC.
                "cidrBlock": "10.0.0.0/16",
                "region": "eu-west-1",
                "tags": {
                    "Environment": "Production",
                    "Project": "MyProject"
                }
            }
        }
    }
}
```
