package apigatewayv2lambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func apiFunctionLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.apigatewayv2.lambda.routeKey": {
			Name:  "aws.apigatewayv2.lambda.routeKey",
			Label: "Route Key",
			Type:  core.ScalarTypeString,
			Description: "The route key that maps to this function, e.g. 'GET /users' for an HTTP API or " +
				"'$connect' / '$default' for a WebSocket API. Defaults to '$default' (a catch-all route that " +
				"proxies all requests to the function).",
			DefaultValue: core.ScalarFromString("$default"),
			Examples: []*core.ScalarValue{
				core.ScalarFromString("GET /users"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.apigatewayv2.lambda.payloadFormatVersion": {
			Name:  "aws.apigatewayv2.lambda.payloadFormatVersion",
			Label: "Payload Format Version",
			Type:  core.ScalarTypeString,
			Description: "The format of the event payload API Gateway sends to the function for an HTTP API. " +
				"'2.0' is the HTTP API default; '1.0' matches the REST API format. Ignored for WebSocket APIs, " +
				"whose integrations do not take a payload format version. Applies to both route integrations " +
				"and (as the authorizer payload format version) Lambda authorizers.",
			DefaultValue: core.ScalarFromString("2.0"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("1.0"),
				core.ScalarFromString("2.0"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.apigatewayv2.lambda.authorizerId": {
			Name:  "aws.apigatewayv2.lambda.authorizerId",
			Label: "Authorizer Id",
			Type:  core.ScalarTypeString,
			Description: "The id of the authorizer that protects this route. Usually a reference to an authored " +
				"authorizer's id (e.g. ${apiGuard.spec.authorizerId}) for a JWT or Lambda authorizer, but may be a " +
				"literal id for an authorizer defined outside the blueprint. When set, the link creates the route " +
				"with this authorizer and authorizationType. A 'default guard' that protects every route is " +
				"expressed by setting this (and authorizationType) on every route function.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("${apiGuard.spec.authorizerId}"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.apigatewayv2.lambda.authorizationType": {
			Name:  "aws.apigatewayv2.lambda.authorizationType",
			Label: "Authorization Type",
			Type:  core.ScalarTypeString,
			Description: "The authorization type for the route when an authorizerId is set: 'CUSTOM' for a Lambda " +
				"(REQUEST) authorizer or 'JWT' for a JWT authorizer. Defaults to 'CUSTOM'. Ignored when no " +
				"authorizerId is set (the route is left open with authorizationType NONE).",
			DefaultValue: core.ScalarFromString("CUSTOM"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("CUSTOM"),
				core.ScalarFromString("JWT"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.apigatewayv2.lambda.websocketTwoWay": {
			Name:  "aws.apigatewayv2.lambda.websocketTwoWay",
			Label: "WebSocket Two-Way Responses",
			Type:  core.ScalarTypeBool,
			Description: "For a WebSocket API route, when true the link also manages a $default route response and " +
				"integration response so the route can reply synchronously to the sender on the same connection. " +
				"Ignored for HTTP APIs, which do not use route/integration responses.",
			DefaultValue: core.ScalarFromBool(false),
			Required:     false,
		},
		"aws/lambda/function::aws.apigatewayv2.lambda.manageConnections": {
			Name:  "aws.apigatewayv2.lambda.manageConnections",
			Label: "Manage WebSocket Connections",
			Type:  core.ScalarTypeBool,
			Description: "For a WebSocket API route handler, whether to grant the function's execution role " +
				"execute-api:ManageConnections so it can push messages to connected clients via the @connections " +
				"management API. Defaults to true (most WebSocket handlers reply to or broadcast to clients, including " +
				"$connect/$disconnect handlers that broadcast to others); set to false for a handler that never pushes. " +
				"Ignored for HTTP APIs.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.apigatewayv2.lambda.manageConnectionsStage": {
			Name:  "aws.apigatewayv2.lambda.manageConnectionsStage",
			Label: "Manage Connections Stage",
			Type:  core.ScalarTypeString,
			Description: "The stage scoped in the execute-api:ManageConnections grant. Defaults to '$default'; " +
				"set to '*' to allow pushing across all stages of the API.",
			DefaultValue: core.ScalarFromString("$default"),
			Required:     false,
		},
	}
}
