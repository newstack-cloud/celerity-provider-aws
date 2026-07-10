A parameter path representing a configuration namespace as a whole.

```blueprintlang
version "2025-11-02"

resource appConfig: aws/ssm/parameterPath {
    metadata {
        displayName = "Application configuration namespace"
    }
    spec {
        path = "/my-app/config"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    appConfig:
        type: aws/ssm/parameterPath
        metadata:
            displayName: Application configuration namespace
        spec:
            path: /my-app/config
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "appConfig": {
      "type": "aws/ssm/parameterPath",
      "metadata": {
        "displayName": "Application configuration namespace"
      },
      "spec": {
        "path": "/my-app/config"
      }
    }
  }
}
```
