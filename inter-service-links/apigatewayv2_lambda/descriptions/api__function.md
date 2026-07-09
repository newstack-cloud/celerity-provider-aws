## API Gateway v2 API to Lambda Function

Routes API Gateway v2 requests to a Lambda function.

The API selects handler functions by label; each function declares its own route key. When a function links to an API, the link manages three intermediary resources:

1. **An `AWS_PROXY` integration** whose `IntegrationUri` is the function's invoke ARN. For HTTP APIs the integration's `PayloadFormatVersion` comes from `aws.apigatewayv2.lambda.payloadFormatVersion` (default `2.0`); WebSocket API integrations do not take a payload format version, so it is omitted for them.

2. **A route** whose `RouteKey` comes from `aws.apigatewayv2.lambda.routeKey` (default `$default`) and whose `Target` is `integrations/<integrationId>` of the integration above. The integration is deployed first so its generated id is known.

3. **A Lambda permission** allowing `apigateway.amazonaws.com` to invoke the function, scoped to the API's `execute-api` source ARN.

The API Gateway v2 API resource exposes no ARN attribute, so the `execute-api` source ARN (`arn:aws:execute-api:<region>:<account>:<apiId>/*/*`) is constructed from the API id (API state), the region and the account (from the function ARN).

The integration and route remain available as authorable resources for manual configuration; this link is the low-boilerplate way to expose a function through an API.

### WebSocket two-way responses

For a WebSocket API, set `aws.apigatewayv2.lambda.websocketTwoWay = true` on a route function to have the link also manage a `$default` **integration response** and **route response** on the route/integration it creates, so the route can reply synchronously to the sender on the same connection. This is ignored for HTTP APIs (which do not use route/integration responses), and the responses are torn down again if the flag is later removed.

### WebSocket outbound push (`@connections`)

The other way a WebSocket handler sends messages is the **`@connections` management API** - replying to the sender out-of-band or broadcasting to other connected clients. That requires `execute-api:ManageConnections` on the handler's execution role. For a WebSocket route handler this link grants it **by default** (`aws.apigatewayv2.lambda.manageConnections`, default `true`), based on the anticipation that most WebSocket handlers push, including `$connect`/`$disconnect` handlers that broadcast join/leave to others, this is scoped to `arn:aws:execute-api:<region>:<account>:<apiId>/<stage>/POST/@connections/*` (stage from `aws.apigatewayv2.lambda.manageConnectionsStage`, default `$default`; set `*` for all stages). Set `manageConnections = false` for a handler that only receives. This is ignored for HTTP APIs.

### Authorization

To protect a route, set two annotations on the route function:

- `aws.apigatewayv2.lambda.authorizerId` — the id of the authorizer, usually a reference to an authored authorizer's id (`${apiGuard.spec.authorizerId}`), or a literal id for an authorizer defined outside the blueprint.
- `aws.apigatewayv2.lambda.authorizationType` — `CUSTOM` for a Lambda (REQUEST) authorizer or `JWT` for a JWT authorizer (default `CUSTOM`).

When `authorizerId` is set the link creates the route with that authorizer attached; when it is absent the route is open (`authorizationType` `NONE`). A "default guard" that protects every route is simply these two annotations set on every route function.

**This link's role in authorization is deliberately narrow: it only stamps `authorizationType` + `authorizerId` onto the routes it creates. It never creates, owns, or looks up authorizers.** Authorizers are `aws/apigatewayv2/authorizer` resources you author (or reference by id). That keeps the model uniform across every case:

| Scenario | Authorizer | This link's role | Also involved |
| --- | --- | --- | --- |
| **JWT** | Authored `aws/apigatewayv2/authorizer` (`type JWT`) | Stamps `authorizationType = JWT` + the authorizer id onto routes | — (no Lambda) |
| **Lambda (REQUEST)** | Authored authorizer (`authorizerUri` = the guard fn's wrapped invoke ARN) | Stamps `authorizationType = CUSTOM` + the authorizer id onto routes | The `aws/apigatewayv2/authorizer` → `aws/lambda/function` link grants the authorizer's invoke permission |
| **External authorizer** | Defined outside the blueprint | Stamps the literal authorizer id onto routes | — |
| **Public / mixed** | None on those routes | Omits the auth fields (route open) | — |

### Examples

**Route to a Lambda, no auth:**

```blueprintlang
version "2025-11-02"

resource ordersApi: aws/apigatewayv2/api {
    metadata { labels = { api = "orders" } }
    spec {
        name = "orders-api"
        protocolType = "HTTP"
    }
    select by label { api = "orders" }
}

resource ordersApiStage: aws/apigatewayv2/stage {
    spec {
        apiId = ordersApi.spec.apiId
        stageName = "$default"
        autoDeploy = true
    }
}

resource listOrders: aws/lambda/function {
    metadata {
        labels = { api = "orders" }
        annotations = { "aws.apigatewayv2.lambda.routeKey" = "GET /orders" }
    }
    spec {
        functionName = "list-orders"
        # ... other function configuration
    }
}
```

**JWT authorizer protecting a route** (the authorizer is authored; the route references its id, this link only stamps the auth onto the route):

```blueprintlang
resource ordersJwtAuth: aws/apigatewayv2/authorizer {
    spec {
        apiId = ordersApi.spec.apiId
        name = "orders-jwt"
        authorizerType = "JWT"
        identitySource = ["$request.header.Authorization"]
        jwtConfiguration = {
            issuer = "https://identity.example.com/oauth2/v1/"
            audience = ["https://api.example.com"]
        }
    }
}

resource listOrders: aws/lambda/function {
    metadata {
        labels = { api = "orders" }
        annotations = {
            "aws.apigatewayv2.lambda.routeKey" = "GET /orders"
            "aws.apigatewayv2.lambda.authorizerId" = "${ordersJwtAuth.spec.authorizerId}"
            "aws.apigatewayv2.lambda.authorizationType" = "JWT"
        }
    }
    spec { functionName = "list-orders" }
}
```

**Lambda authorizer** (authored authorizer + the `authorizer → function` link for the invoke permission; the route references the authorizer id):

```blueprintlang
resource authFn: aws/lambda/function {
    metadata { labels = { guard = "orders" } }
    spec { functionName = "orders-authorizer" }
}

resource ordersLambdaAuth: aws/apigatewayv2/authorizer {
    spec {
        apiId = ordersApi.spec.apiId
        name = "orders-lambda-auth"
        authorizerType = "REQUEST"
        identitySource = ["$request.header.Authorization"]
        authorizerPayloadFormatVersion = "2.0"
        enableSimpleResponses = true
        authorizerUri = "arn:aws:apigateway:${aws.region}:lambda:path/2015-03-31/functions/${authFn.spec.arn}/invocations"
    }
    select by label { guard = "orders" }
}

resource listOrders: aws/lambda/function {
    metadata {
        labels = { api = "orders" }
        annotations = {
            "aws.apigatewayv2.lambda.routeKey" = "GET /orders"
            "aws.apigatewayv2.lambda.authorizerId" = "${ordersLambdaAuth.spec.authorizerId}"
            "aws.apigatewayv2.lambda.authorizationType" = "CUSTOM"
        }
    }
    spec { functionName = "list-orders" }
}
```

**External authorizer / mixed public routes** (no authorizer resource in the blueprint; open routes just omit the auth annotations):

```blueprintlang
resource health: aws/lambda/function {
    metadata {
        labels = { api = "orders" }
        annotations = { "aws.apigatewayv2.lambda.routeKey" = "GET /health" }
    }
    spec { functionName = "health" }
}

resource listOrders: aws/lambda/function {
    metadata {
        labels = { api = "orders" }
        annotations = {
            "aws.apigatewayv2.lambda.routeKey" = "GET /orders"
            "aws.apigatewayv2.lambda.authorizerId" = "abcd1234"
            "aws.apigatewayv2.lambda.authorizationType" = "CUSTOM"
        }
    }
    spec { functionName = "list-orders" }
}
```
