A parameter tree combining plaintext and encrypted entries (with hierarchical keys, a customer managed KMS key, tags, and a target region), read by a Lambda function through a single link.

```blueprintlang
version "2025-11-02"

variables {
    dbPassword: string {
        secret = true
    }
}

resource apiFunction: aws/lambda/function {
    metadata {
        labels = {
            configStore = "app-config"
        }
        annotations = {
            "aws.lambda.ssm.appConfig.envVarName" = "APP_CONFIG_STORE_PATH"
        }
    }

    select by label {
        configStore = "app-config"
    }

    spec {
        functionName = "api"
        role = apiFunctionRole.spec.arn
        # ... other function configuration
    }
}

resource appConfig: aws/ssm/parameterTree {
    metadata {
        displayName = "Application configuration store"
        labels = {
            configStore = "app-config"
        }
    }
    spec {
        path = "/my-app/config"
        values = {
            logLevel = "info"
            "db/host" = "db.internal.example.com"
            "db/port" = "5432"
        }
        secureValues = {
            "db/password" = "${variables.dbPassword}"
        }
        keyId = "alias/my-app-config"
        tier = "Standard"
        region = "us-east-1"
        tags = {
            Environment = "production"
        }
    }
}

resource apiFunctionRole: aws/iam/role {
    spec {
        name = "api-role"
        # ... role configuration
    }
}

export configStorePath: string {
    field = resources.appConfig.spec.path
}
```

```yaml
version: "2025-11-02"
variables:
    dbPassword:
        type: string
        secret: true
resources:
    apiFunction:
        type: aws/lambda/function
        metadata:
            labels:
                configStore: app-config
            annotations:
                aws.lambda.ssm.appConfig.envVarName: APP_CONFIG_STORE_PATH
        linkSelector:
            byLabel:
                configStore: app-config
        spec:
            functionName: api
            role: "${apiFunctionRole.spec.arn}"
    appConfig:
        type: aws/ssm/parameterTree
        metadata:
            displayName: Application configuration store
            labels:
                configStore: app-config
        spec:
            path: /my-app/config
            values:
                logLevel: info
                db/host: db.internal.example.com
                db/port: "5432"
            secureValues:
                db/password: "${variables.dbPassword}"
            keyId: alias/my-app-config
            tier: Standard
            region: us-east-1
            tags:
                Environment: production
    apiFunctionRole:
        type: aws/iam/role
        spec:
            name: api-role
exports:
    configStorePath:
        type: string
        field: resources.appConfig.spec.path
```

```javascript
{
  "version": "2025-11-02",
  "variables": {
    "dbPassword": {
      "type": "string",
      "secret": true
    }
  },
  "resources": {
    "apiFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "labels": {
          "configStore": "app-config"
        },
        "annotations": {
          "aws.lambda.ssm.appConfig.envVarName": "APP_CONFIG_STORE_PATH"
        }
      },
      "linkSelector": {
        "byLabel": {
          "configStore": "app-config"
        }
      },
      "spec": {
        "functionName": "api",
        "role": "${apiFunctionRole.spec.arn}"
      }
    },
    "appConfig": {
      "type": "aws/ssm/parameterTree",
      "metadata": {
        "displayName": "Application configuration store",
        "labels": {
          "configStore": "app-config"
        }
      },
      "spec": {
        "path": "/my-app/config",
        "values": {
          "logLevel": "info",
          "db/host": "db.internal.example.com",
          "db/port": "5432"
        },
        "secureValues": {
          "db/password": "${variables.dbPassword}"
        },
        "keyId": "alias/my-app-config",
        "tier": "Standard",
        "region": "us-east-1",
        "tags": {
          "Environment": "production"
        }
      }
    },
    "apiFunctionRole": {
      "type": "aws/iam/role",
      "spec": {
        "name": "api-role"
      }
    }
  },
  "exports": {
    "configStorePath": {
      "type": "string",
      "field": "resources.appConfig.spec.path"
    }
  }
}
```
