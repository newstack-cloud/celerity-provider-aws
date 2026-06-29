## VPC to Lambda Function

Runs a Lambda function inside a VPC.

When a VPC links to a function, the function is placed in the VPC's subnets so it runs on the VPC's private network. Use the `aws.flexvpc.lambda.subnetType` annotation to choose which subnet tier it runs in.

Once a function runs inside the VPC, links from the function to other resources in that VPC (such as a database or cache) set up the network access it needs to reach them.

### Example

```blueprintlang
version "2025-11-02"

resource appVpc: aws/flex/vpc {
    metadata {
        labels = {
            app = "orders"
        }
    }

    select by label {
        app = "orders"
    }

    spec {
        name = "orders-vpc"
        preset = "standard"
        cidrBlock = "10.0.0.0/16"
        # ... other VPC configuration
    }
}

resource getOrderFunction: aws/lambda/function {
    metadata {
        labels = {
            app = "orders"
        }
        annotations = {
            "aws.flexvpc.lambda.subnetType" = "private"
        }
    }

    spec {
        functionName = "get-order"
        # ... other function configuration
    }
}
```
