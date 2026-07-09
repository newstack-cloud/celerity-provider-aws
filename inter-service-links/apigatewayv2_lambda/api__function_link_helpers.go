package apigatewayv2lambda

import (
	"fmt"
	"strings"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func integrationIntermediaryIdentity(apiInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/apigatewayv2/integration",
		ResourceID:   intermediaryID(apiInfo, functionInfo, "integration"),
		ResourceName: intermediaryName(apiInfo, functionInfo, "Integration"),
	}
}

func routeIntermediaryIdentity(apiInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/apigatewayv2/route",
		ResourceID:   intermediaryID(apiInfo, functionInfo, "route"),
		ResourceName: intermediaryName(apiInfo, functionInfo, "Route"),
	}
}

func permissionIntermediaryIdentity(apiInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/lambda/permission",
		ResourceID:   intermediaryID(apiInfo, functionInfo, "apigw-invoke-permission"),
		ResourceName: intermediaryName(apiInfo, functionInfo, "InvokePermission"),
	}
}

func integrationResponseIntermediaryIdentity(apiInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/apigatewayv2/integrationResponse",
		ResourceID:   intermediaryID(apiInfo, functionInfo, "integration-response"),
		ResourceName: intermediaryName(apiInfo, functionInfo, "IntegrationResponse"),
	}
}

func routeResponseIntermediaryIdentity(apiInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/apigatewayv2/routeResponse",
		ResourceID:   intermediaryID(apiInfo, functionInfo, "route-response"),
		ResourceName: intermediaryName(apiInfo, functionInfo, "RouteResponse"),
	}
}

func intermediaryID(apiInfo, functionInfo *provider.ResourceInfo, suffix string) string {
	return fmt.Sprintf(
		"%s__%s__%s",
		pluginutils.GetResourceName(apiInfo),
		pluginutils.GetResourceName(functionInfo),
		suffix,
	)
}

func intermediaryName(apiInfo, functionInfo *provider.ResourceInfo, suffix string) string {
	return fmt.Sprintf(
		"%s%s%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(apiInfo)),
		suffix,
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(functionInfo)),
	)
}

func apiID(apiInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(apiInfo)
	node, has := pluginutils.GetValueByPath("$.apiId", spec)
	if !has || node == nil || core.StringValue(node) == "" {
		return "", false
	}
	return core.StringValue(node), true
}

// Reports whether the linked API is an HTTP API. WebSocket APIs (protocolType
// WEBSOCKET) do not take a payload format version on their integrations.
func isHTTPProtocol(apiInfo *provider.ResourceInfo) bool {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(apiInfo)
	node, has := pluginutils.GetValueByPath("$.protocolType", spec)
	if !has || node == nil {
		// Default to HTTP when the protocol is unknown (HTTP is the common case).
		return true
	}
	return strings.EqualFold(core.StringValue(node), "HTTP")
}

func integrationIDFromState(intermediaryState *state.LinkIntermediaryResourceState) (string, bool) {
	return computedIDFromState(intermediaryState, "$.integrationId")
}

func routeIDFromState(intermediaryState *state.LinkIntermediaryResourceState) (string, bool) {
	return computedIDFromState(intermediaryState, "$.routeId")
}

func computedIDFromState(intermediaryState *state.LinkIntermediaryResourceState, path string) (string, bool) {
	if intermediaryState == nil {
		return "", false
	}
	node, has := pluginutils.GetValueByPath(path, intermediaryState.ResourceSpecData)
	if !has || node == nil || core.StringValue(node) == "" {
		return "", false
	}
	return core.StringValue(node), true
}

func executeAPISourceARN(region, account, apiID string) string {
	return fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*", region, account, apiID)
}

func authorizerPermissionIdentity(authorizerInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/lambda/permission",
		ResourceID:   intermediaryID(authorizerInfo, functionInfo, "apigw-authorizer-permission"),
		ResourceName: intermediaryName(authorizerInfo, functionInfo, "AuthorizerPermission"),
	}
}

func authorizerAPIID(authorizerInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(authorizerInfo)
	node, has := pluginutils.GetValueByPath("$.apiId", spec)
	if !has || node == nil || core.StringValue(node) == "" {
		return "", false
	}
	return core.StringValue(node), true
}

func authorizerID(authorizerInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(authorizerInfo)
	node, has := pluginutils.GetValueByPath("$.authorizerId", spec)
	if !has || node == nil || core.StringValue(node) == "" {
		return "", false
	}
	return core.StringValue(node), true
}

func authorizerSourceARN(region, account, apiID, authorizerID string) string {
	return fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/authorizers/%s", region, account, apiID, authorizerID)
}

// Extracts the account id (the 5th colon-separated segment) from an ARN such as
// arn:aws:lambda:<region>:<account>:function:<name>.
func accountFromARN(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 5 || parts[4] == "" {
		return "", fmt.Errorf("account id could not be determined from ARN: %q", arn)
	}
	return parts[4], nil
}

type apiLinkAnnotations struct {
	routeKey               string
	payloadFormatVersion   string
	authorizerID           string
	authorizationType      string
	websocketTwoWay        bool
	manageConnections      bool
	manageConnectionsStage string
}

func getAPILinkAnnotations(functionInfo *provider.ResourceInfo) *apiLinkAnnotations {
	routeKey, _ := pluginutils.GetStringAnnotation(
		functionInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("%s.routeKey", lambdaAnnotationPrefix),
			Default: "$default",
		},
	)
	payloadFormatVersion, _ := pluginutils.GetStringAnnotation(
		functionInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("%s.payloadFormatVersion", lambdaAnnotationPrefix),
			Default: "2.0",
		},
	)
	authorizerID, _ := pluginutils.GetStringAnnotation(
		functionInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("%s.authorizerId", lambdaAnnotationPrefix),
		},
	)
	authorizationType, _ := pluginutils.GetStringAnnotation(
		functionInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("%s.authorizationType", lambdaAnnotationPrefix),
			Default: "CUSTOM",
		},
	)
	websocketTwoWay, _ := pluginutils.GetBoolAnnotation(
		functionInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     fmt.Sprintf("%s.websocketTwoWay", lambdaAnnotationPrefix),
			Default: false,
		},
	)
	manageConnections, _ := pluginutils.GetBoolAnnotation(
		functionInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     fmt.Sprintf("%s.manageConnections", lambdaAnnotationPrefix),
			Default: true,
		},
	)
	manageConnectionsStage, _ := pluginutils.GetStringAnnotation(
		functionInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("%s.manageConnectionsStage", lambdaAnnotationPrefix),
			Default: "$default",
		},
	)
	return &apiLinkAnnotations{
		routeKey:               routeKey,
		payloadFormatVersion:   payloadFormatVersion,
		authorizerID:           authorizerID,
		authorizationType:      authorizationType,
		websocketTwoWay:        websocketTwoWay,
		manageConnections:      manageConnections,
		manageConnectionsStage: manageConnectionsStage,
	}
}

func manageConnectionsSourceARN(region, account, apiID, stage string) string {
	return fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/POST/@connections/*", region, account, apiID, stage)
}

func manageConnectionsSID(functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("APIGatewayManageConnections%s", pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(functionInfo)))
}

func executionRoleLinkDataName(functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", pluginutils.GetResourceName(functionInfo))
}
