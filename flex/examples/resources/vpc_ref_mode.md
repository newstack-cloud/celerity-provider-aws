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
            mode: reference
```
