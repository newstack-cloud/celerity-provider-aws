A parameter path namespace with plaintext and encrypted parameters beneath it, read by a Lambda function through a single link.

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

resource appConfig: aws/ssm/parameterPath {
    metadata {
        displayName = "Application configuration namespace"
        labels = {
            configStore = "app-config"
        }
    }
    spec {
        path = "/my-app/config"
    }
}

resource dbHost: aws/ssm/parameter {
    spec {
        name = "/my-app/config/db-host"
        type = "String"
        value = "db.internal.example.com"
    }
}

resource dbPassword: aws/ssm/parameter {
    spec {
        name = "/my-app/config/db-password"
        type = "SecureString"
        secureValue = "${variables.dbPassword}"
    }
}

resource apiFunctionRole: aws/iam/role {
    spec {
        name = "api-role"
        # ... role configuration
    }
}

exports {
    configStorePath: string {
        field = "resources.appConfig.spec.path"
    }
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
        type: aws/ssm/parameterPath
        metadata:
            displayName: Application configuration namespace
            labels:
                configStore: app-config
        spec:
            path: /my-app/config
    dbHost:
        type: aws/ssm/parameter
        spec:
            name: /my-app/config/db-host
            type: String
            value: db.internal.example.com
    dbPassword:
        type: aws/ssm/parameter
        spec:
            name: /my-app/config/db-password
            type: SecureString
            secureValue: "${variables.dbPassword}"
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
      "type": "aws/ssm/parameterPath",
      "metadata": {
        "displayName": "Application configuration namespace",
        "labels": {
          "configStore": "app-config"
        }
      },
      "spec": {
        "path": "/my-app/config"
      }
    },
    "dbHost": {
      "type": "aws/ssm/parameter",
      "spec": {
        "name": "/my-app/config/db-host",
        "type": "String",
        "value": "db.internal.example.com"
      }
    },
    "dbPassword": {
      "type": "aws/ssm/parameter",
      "spec": {
        "name": "/my-app/config/db-password",
        "type": "SecureString",
        "secureValue": "${variables.dbPassword}"
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
