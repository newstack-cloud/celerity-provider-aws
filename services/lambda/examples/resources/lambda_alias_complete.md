Create a complete Lambda setup with a function, a version, and a production alias.

```blueprintlang
version "2025-11-02"

resource myFunction: aws/lambda/function {
    metadata {
        displayName = "My Sample Function"
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
        description = "Version 1"
    }
}

resource prodAlias: aws/lambda/alias {
    metadata {
        displayName = "Production Alias"
    }
    spec {
        functionName = resources.myFunction.spec.functionName
        name = "PROD"
        functionVersion = resources.version1.spec.version
        description = "Production alias"
    }
}
```

```yaml
version: 2025-11-02

resources:
  myFunction:
    type: aws/lambda/function
    metadata:
      displayName: My Sample Function
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
      description: Version 1

  prodAlias:
    type: aws/lambda/alias
    metadata:
      displayName: Production Alias
    spec:
      functionName: ${resources.myFunction.spec.functionName}
      name: PROD
      functionVersion: ${resources.version1.spec.version}
      description: Production alias
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "My Sample Function"
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
        "description": "Version 1"
      }
    },
    "prodAlias": {
      "type": "aws/lambda/alias",
      "metadata": {
        "displayName": "Production Alias"
      },
      "spec": {
        "functionName": "${resources.myFunction.spec.functionName}",
        "name": "PROD",
        "functionVersion": "${resources.version1.spec.version}",
        "description": "Production alias"
      }
    }
  }
}
```
