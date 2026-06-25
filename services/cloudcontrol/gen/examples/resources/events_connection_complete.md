An AWS Events Connection using OAuth client-credentials authentication, with the auth parameters authored as a native object.

```blueprintlang
version "2025-11-02"

resource connection: aws/events/connection {
    metadata {
        displayName = "Partner API connection"
    }
    spec {
        name = "partner-api-connection"
        description = "Connection to the partner API"
        authorizationType = "OAUTH_CLIENT_CREDENTIALS"
        authParameters = {
            OAuthParameters = {
                AuthorizationEndpoint = "https://partner.example.com/oauth/token"
                HttpMethod = "POST"
                ClientParameters = {
                    ClientID = "example-client-id"
                    ClientSecret = "example-client-secret"
                }
            }
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    connection:
        type: aws/events/connection
        metadata:
            displayName: Partner API connection
        spec:
            name: partner-api-connection
            description: Connection to the partner API
            authorizationType: OAUTH_CLIENT_CREDENTIALS
            authParameters:
                OAuthParameters:
                    AuthorizationEndpoint: https://partner.example.com/oauth/token
                    HttpMethod: POST
                    ClientParameters:
                        ClientID: example-client-id
                        ClientSecret: example-client-secret
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "connection": {
      "type": "aws/events/connection",
      "metadata": {
        "displayName": "Partner API connection"
      },
      "spec": {
        "name": "partner-api-connection",
        "description": "Connection to the partner API",
        "authorizationType": "OAUTH_CLIENT_CREDENTIALS",
        "authParameters": {
          "OAuthParameters": {
            "AuthorizationEndpoint": "https://partner.example.com/oauth/token",
            "HttpMethod": "POST",
            "ClientParameters": {
              "ClientID": "example-client-id",
              "ClientSecret": "example-client-secret"
            }
          }
        }
      }
    }
  }
}
```
