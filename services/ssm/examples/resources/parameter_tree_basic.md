A parameter tree managing a small configuration store of plaintext values beneath a path prefix.

```blueprintlang
version "2025-11-02"

resource appConfig: aws/ssm/parameterTree {
    spec {
        path = "/my-app/config"
        values = {
            logLevel = "info"
            featureFlags = "search,recommendations"
        }
    }
}

export configStorePath: string {
    field = resources.appConfig.spec.path
}
```

```yaml
version: "2025-11-02"
resources:
    appConfig:
        type: aws/ssm/parameterTree
        spec:
            path: /my-app/config
            values:
                logLevel: info
                featureFlags: search,recommendations
exports:
    configStorePath:
        type: string
        field: resources.appConfig.spec.path
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "appConfig": {
      "type": "aws/ssm/parameterTree",
      "spec": {
        "path": "/my-app/config",
        "values": {
          "logLevel": "info",
          "featureFlags": "search,recommendations"
        }
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
