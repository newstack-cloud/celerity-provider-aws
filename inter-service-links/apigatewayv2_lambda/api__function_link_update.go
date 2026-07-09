package apigatewayv2lambda

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

const lambdaAnnotationPrefix = "aws.apigatewayv2.lambda"

func (l *apiFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The API is not modified by this link; the integration, route and permission are managed
	// as intermediaries in UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *apiFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The function is not modified by this link; the invoke permission is a separate,
	// link-owned aws/lambda/permission intermediary handled in UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

// UpdateIntermediaryResources deploys the AWS_PROXY integration, the route targeting it, and the
// Lambda permission that lets API Gateway invoke the function. The integration is deployed first
// because the route targets it by its generated id.
func (l *apiFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	instanceID := pluginutils.GetInstanceID(input.ResourceAInfo)
	intermediaries := newAPIFunctionIntermediaries(input)
	annotations := getAPILinkAnnotations(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		// Revoke any execute-api:ManageConnections grant before tearing down the intermediaries
		// (region/account/apiId are unused on the revoke path).
		if _, _, err := l.reconcileManageConnectionsGrant(ctx, providerCtx, input, "", "", "", annotations); err != nil {
			return nil, err
		}
		return destroyAPIFunctionIntermediaries(ctx, input.ResourceService, instanceID, providerCtx, intermediaries)
	}

	deployment, err := l.newAPIFunctionDeployment(ctx, input, providerCtx, instanceID, intermediaries)
	if err != nil {
		return nil, err
	}

	integrationID, err := deployment.deployIntegration(ctx)
	if err != nil {
		return nil, err
	}
	routeState, err := deployment.deployRoute(ctx, integrationID)
	if err != nil {
		return nil, err
	}
	if err := deployment.deployInvokePermission(ctx); err != nil {
		return nil, err
	}
	if err := deployment.syncWebSocketResponses(ctx, integrationID, routeState); err != nil {
		return nil, err
	}

	// WebSocket ManageConnections grant on the handler's execution role (opt-in, WebSocket only).
	grantLinkData, grantMappings, err := l.reconcileManageConnectionsGrant(
		ctx, providerCtx, input, deployment.apiID, deployment.region, deployment.account, annotations,
	)
	if err != nil {
		return nil, err
	}

	return deployment.output(grantLinkData, grantMappings), nil
}

type intermediaryRef struct {
	identity linkutils.IntermediaryIdentity
	prior    *state.LinkIntermediaryResourceState
}

type apiFunctionIntermediaries struct {
	integration         intermediaryRef
	route               intermediaryRef
	permission          intermediaryRef
	integrationResponse intermediaryRef
	routeResponse       intermediaryRef
}

func newAPIFunctionIntermediaries(
	input *provider.LinkUpdateIntermediaryResourcesInput,
) apiFunctionIntermediaries {
	ref := func(identity linkutils.IntermediaryIdentity) intermediaryRef {
		return intermediaryRef{
			identity: identity,
			prior:    linkutils.FindIntermediaryState(input.CurrentLinkState, identity.ResourceID),
		}
	}
	return apiFunctionIntermediaries{
		integration:         ref(integrationIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)),
		route:               ref(routeIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)),
		permission:          ref(permissionIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)),
		integrationResponse: ref(integrationResponseIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)),
		routeResponse:       ref(routeResponseIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)),
	}
}

func destroyAPIFunctionIntermediaries(
	ctx context.Context,
	resourceService provider.ResourceService,
	instanceID string,
	providerCtx provider.Context,
	intermediaries apiFunctionIntermediaries,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Destroy children before parents: responses before the route/integration they belong to,
	// the route before the integration it targets; the permission is independent.
	priors := []*state.LinkIntermediaryResourceState{
		intermediaries.routeResponse.prior,
		intermediaries.integrationResponse.prior,
		intermediaries.route.prior,
		intermediaries.integration.prior,
		intermediaries.permission.prior,
	}
	for _, prior := range priors {
		if err := linkutils.DestroyManagedIntermediary(
			ctx, resourceService, instanceID, providerCtx, prior,
		); err != nil {
			return nil, err
		}
	}
	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

type apiFunctionDeployment struct {
	input          *provider.LinkUpdateIntermediaryResourcesInput
	providerCtx    provider.Context
	instanceID     string
	intermediaries apiFunctionIntermediaries

	apiID       string
	functionARN string
	region      string
	account     string
	sourceARN   string
	annotations *apiLinkAnnotations
	httpAPI     bool

	deployed []linkutils.DeployedIntermediary
	states   []*state.LinkIntermediaryResourceState
}

func (l *apiFunctionLinkActions) newAPIFunctionDeployment(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	providerCtx provider.Context,
	instanceID string,
	intermediaries apiFunctionIntermediaries,
) (*apiFunctionDeployment, error) {
	apiID, hasAPIID := apiID(input.ResourceAInfo)
	if !hasAPIID {
		return nil, fmt.Errorf(
			"API id could not be retrieved from the API Gateway v2 API %q",
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}
	functionARN, hasFunctionARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the Lambda function %q",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	region, err := l.getRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}
	account, err := accountFromARN(functionARN)
	if err != nil {
		return nil, err
	}

	return &apiFunctionDeployment{
		input:          input,
		providerCtx:    providerCtx,
		instanceID:     instanceID,
		intermediaries: intermediaries,
		apiID:          apiID,
		functionARN:    functionARN,
		region:         region,
		account:        account,
		sourceARN:      executeAPISourceARN(region, account, apiID),
		annotations:    getAPILinkAnnotations(input.ResourceBInfo),
		httpAPI:        isHTTPProtocol(input.ResourceAInfo),
	}, nil
}

func (d *apiFunctionDeployment) deploy(
	ctx context.Context,
	ref intermediaryRef,
	spec *core.MappingNode,
	linkDataLeaves map[string]*core.MappingNode,
) (*state.LinkIntermediaryResourceState, error) {
	intermediaryState, err := linkutils.DeployManagedIntermediary(
		ctx, d.input.ResourceService, d.instanceID, d.input.InstanceName, d.providerCtx, ref.prior,
		linkutils.ManagedIntermediary{
			ResourceType: ref.identity.ResourceType,
			ResourceID:   ref.identity.ResourceID,
			ResourceName: ref.identity.ResourceName,
			Spec:         spec,
		},
	)
	if err != nil {
		return nil, err
	}
	d.deployed = append(d.deployed, linkutils.DeployedIntermediary{
		Identity: ref.identity,
		Leaves:   linkDataLeaves,
	})
	d.states = append(d.states, intermediaryState)
	return intermediaryState, nil
}

func (d *apiFunctionDeployment) deployIntegration(ctx context.Context) (string, error) {
	integrationState, err := d.deploy(
		ctx,
		d.intermediaries.integration,
		integrationSpec(d.apiID, d.functionARN, d.annotations, d.httpAPI),
		map[string]*core.MappingNode{
			"apiId":          core.MappingNodeFromString(d.apiID),
			"integrationUri": core.MappingNodeFromString(d.functionARN),
		},
	)
	if err != nil {
		return "", err
	}

	integrationID, hasIntegrationID := integrationIDFromState(integrationState)
	if !hasIntegrationID {
		return "", fmt.Errorf(
			"integration id could not be determined for the API Gateway v2 integration created for function %q",
			pluginutils.GetResourceName(d.input.ResourceBInfo),
		)
	}
	return integrationID, nil
}

func (d *apiFunctionDeployment) deployRoute(
	ctx context.Context,
	integrationID string,
) (*state.LinkIntermediaryResourceState, error) {
	return d.deploy(
		ctx,
		d.intermediaries.route,
		routeSpec(d.apiID, integrationID, d.annotations),
		map[string]*core.MappingNode{"apiId": core.MappingNodeFromString(d.apiID)},
	)
}

func (d *apiFunctionDeployment) deployInvokePermission(ctx context.Context) error {
	_, err := d.deploy(
		ctx,
		d.intermediaries.permission,
		core.MappingNodeFields(
			"functionName", core.MappingNodeFromString(d.functionARN),
			"action", core.MappingNodeFromString("lambda:InvokeFunction"),
			"principal", core.MappingNodeFromString("apigateway.amazonaws.com"),
			"sourceArn", core.MappingNodeFromString(d.sourceARN),
		),
		map[string]*core.MappingNode{"functionArn": core.MappingNodeFromString(d.functionARN)},
	)
	return err
}

// Deploys the $default integration/route responses that let a two-way
// WebSocket route reply synchronously to the sender. When the route is not (or no longer)
// two-way, any responses this link previously created are removed instead.
func (d *apiFunctionDeployment) syncWebSocketResponses(
	ctx context.Context,
	integrationID string,
	routeState *state.LinkIntermediaryResourceState,
) error {
	if !d.annotations.websocketTwoWay || d.httpAPI {
		return d.destroyResponses(ctx)
	}

	routeID, hasRouteID := routeIDFromState(routeState)
	if !hasRouteID {
		return fmt.Errorf(
			"route id could not be determined for the API Gateway v2 route created for function %q",
			pluginutils.GetResourceName(d.input.ResourceBInfo),
		)
	}

	apiIDLeaves := map[string]*core.MappingNode{"apiId": core.MappingNodeFromString(d.apiID)}
	_, err := d.deploy(
		ctx,
		d.intermediaries.integrationResponse,
		core.MappingNodeFields(
			"apiId", core.MappingNodeFromString(d.apiID),
			"integrationId", core.MappingNodeFromString(integrationID),
			"integrationResponseKey", core.MappingNodeFromString("$default"),
		),
		apiIDLeaves,
	)
	if err != nil {
		return err
	}

	_, err = d.deploy(
		ctx,
		d.intermediaries.routeResponse,
		core.MappingNodeFields(
			"apiId", core.MappingNodeFromString(d.apiID),
			"routeId", core.MappingNodeFromString(routeID),
			"routeResponseKey", core.MappingNodeFromString("$default"),
		),
		apiIDLeaves,
	)
	return err
}

func (d *apiFunctionDeployment) destroyResponses(ctx context.Context) error {
	priors := []*state.LinkIntermediaryResourceState{
		d.intermediaries.routeResponse.prior,
		d.intermediaries.integrationResponse.prior,
	}
	for _, prior := range priors {
		if err := linkutils.DestroyManagedIntermediary(
			ctx, d.input.ResourceService, d.instanceID, d.providerCtx, prior,
		); err != nil {
			return err
		}
	}
	return nil
}

func (d *apiFunctionDeployment) output(
	grantLinkData *core.MappingNode,
	grantMappings map[string]string,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkData := linkutils.IntermediaryLinkData(d.deployed...)
	if grantLinkData != nil && len(grantLinkData.Fields) > 0 {
		linkData = core.MergeMaps(linkData, grantLinkData)
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:                   linkData,
		ResourceDataMappings:       grantMappings,
		IntermediaryResourceStates: d.states,
	}
}

// Builds the route spec, attaching the authorizer (authorizationType + authorizerId)
// when the route function names one. Without an authorizerId the route is left open
// (authorizationType NONE, the API Gateway default).
func routeSpec(apiID, integrationID string, annotations *apiLinkAnnotations) *core.MappingNode {
	spec := core.MappingNodeFields(
		"apiId", core.MappingNodeFromString(apiID),
		"routeKey", core.MappingNodeFromString(annotations.routeKey),
		"target", core.MappingNodeFromString("integrations/"+integrationID),
	)
	if annotations.authorizerID != "" {
		spec.Fields["authorizationType"] = core.MappingNodeFromString(annotations.authorizationType)
		spec.Fields["authorizerId"] = core.MappingNodeFromString(annotations.authorizerID)
	}
	return spec
}

func integrationSpec(
	apiID, functionARN string,
	annotations *apiLinkAnnotations,
	http bool,
) *core.MappingNode {
	spec := core.MappingNodeFields(
		"apiId", core.MappingNodeFromString(apiID),
		"integrationType", core.MappingNodeFromString("AWS_PROXY"),
		"integrationUri", core.MappingNodeFromString(functionARN),
	)
	// WebSocket AWS_PROXY integrations do not take a payload format version.
	if http {
		spec.Fields["payloadFormatVersion"] = core.MappingNodeFromString(annotations.payloadFormatVersion)
	}
	return spec
}
