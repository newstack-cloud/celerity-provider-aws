## API Gateway v2 Authorizer to Lambda Function

Grants API Gateway permission to invoke a Lambda (REQUEST) authorizer's backing function.

A Lambda authorizer is an `aws/apigatewayv2/authorizer` resource (`authorizerType: REQUEST`) whose `authorizerUri` points at a Lambda function. Before API Gateway can call that function to authorize a request, the function needs a resource-based permission allowing `apigateway.amazonaws.com` to invoke it. This link manages exactly that one permission:

- **A Lambda permission** allowing `apigateway.amazonaws.com` to invoke the authorizer's backing function, scoped to the authorizer's `execute-api` source ARN (`arn:aws:execute-api:<region>:<account>:<apiId>/authorizers/<authorizerId>`).

The authorizer is the authored resource, so it already exists (with its `apiId` and generated `authorizerId`) by the time this link runs, the source ARN is constructed from those plus the region and the account (from the function ARN). Nothing else is created or modified: the authorizer, its `authorizerUri` and the routes it protects are all authored/referenced elsewhere.

### Role across authorization scenarios

- **Lambda (REQUEST) authorizer** — this link applies: it grants the invoke permission for the authorizer's function.
- **JWT authorizer** — this link does **not** apply: a JWT authorizer has no backing Lambda, so there is no invoke permission to grant. Routes reference the JWT authorizer's id directly (see the `aws/apigatewayv2/api` → `aws/lambda/function` link).
- **External authorizer** — if the authorizer (and its permission) live outside the blueprint, this link is not used; routes reference the external authorizer id directly.

The permission is available as an authorable `aws/lambda/permission` resource for manual control; this link is the low-boilerplate way to wire it.

### Example

```blueprintlang
version "2025-11-02"

resource ordersApi: aws/apigatewayv2/api {
    spec {
        name = "orders-api"
        protocolType = "HTTP"
    }
}

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

    # selects the authorizer's backing function → this link grants the invoke permission
    select by label { guard = "orders" }
}
```
