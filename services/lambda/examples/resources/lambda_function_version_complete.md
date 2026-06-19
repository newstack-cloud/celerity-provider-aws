Create a Lambda function with inline code and publish an initial version of it.

```blueprintlang
version "2025-11-02"

resource myFunction: aws/lambda/function {
    metadata {
        displayName = "My Function"
    }
    spec {
        functionName = "my-sample-function"
        runtime = "nodejs18.x"
        handler = "index.handler"
        role = "arn:aws:iam::123456789012:role/lambda-role"
        code = {
            zipFile = """
exports.handler = async (event) => {
  return {
    statusCode: 200,
    body: JSON.stringify('Hello from Lambda!')
  };
};
"""
        }
    }
}

resource version1: aws/lambda/functionVersion {
    metadata {
        displayName = "Function Version 1"
    }
    spec {
        functionName = resources.myFunction.spec.functionName
        description = "Initial release with basic functionality"
    }
}
```

```yaml
version: 2025-11-02

resources:
  myFunction:
    type: aws/lambda/function
    metadata:
      displayName: My Function
    spec:
      functionName: my-sample-function
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            return {
              statusCode: 200,
              body: JSON.stringify('Hello from Lambda!')
            };
          };

  version1:
    type: aws/lambda/functionVersion
    metadata:
      displayName: Function Version 1
    spec:
      functionName: ${resources.myFunction.spec.functionName}
      description: "Initial release with basic functionality"
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "My Function"
      },
      "spec": {
        "functionName": "my-sample-function",
        "runtime": "nodejs18.x",
        "handler": "index.handler",
        "role": "arn:aws:iam::123456789012:role/lambda-role",
        "code": {
          "zipFile": "exports.handler = async (event) => {\n  return {\n    statusCode: 200,\n    body: JSON.stringify('Hello from Lambda!')\n  };\n};\n"
        }
      }
    },
    "version1": {
      "type": "aws/lambda/functionVersion",
      "metadata": {
        "displayName": "Function Version 1"
      },
      "spec": {
        "functionName": "${resources.myFunction.spec.functionName}",
        "description": "Initial release with basic functionality"
      }
    }
  }
}
```
