**YAML**

```yaml
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
```
