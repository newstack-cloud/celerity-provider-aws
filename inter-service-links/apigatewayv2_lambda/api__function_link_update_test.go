//go:build unit

package apigatewayv2lambda

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

const (
	afRoleARN              = "arn:aws:iam::123456789012:role/list-orders-role"
	afRoleName             = "list-orders-role"
	afRoleResource         = "listOrdersRole"
	afManageConnectionsSID = "APIGatewayManageConnectionslistOrders"
	afManageConnectionsARN = "arn:aws:execute-api:us-west-2:123456789012:api-abc123/$default/POST/@connections/*"
)

func afRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: afRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(afRoleName),
			"arn", core.MappingNodeFromString(afRoleARN),
		),
	}
}

// A WebSocket route handler (manageConnections defaults true)
// gets execute-api:ManageConnections on its execution role.
func (s *APIFunctionLinkUpdateSuite) Test_create_websocket_grants_manage_connections() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(afFunctionARN),
				Role:        aws.String(afRoleARN),
			},
		}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithDeployOutput(afDeployOutput()),
		resourceservicemock.WithLookupResourceInState(afRoleState()),
	)

	actions := &apiFunctionLinkActions{
		awsConfigStore:       testConfigStore(loader),
		lambdaServiceFactory: func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		iamServiceFactory:    func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
	}

	input := s.updateInput("WEBSOCKET")
	input.ResourceBInfo = afFunctionInfo() // manageConnections defaults true
	input.ResourceService = rs

	out, err := actions.UpdateIntermediaryResources(context.Background(), input)
	s.Require().NoError(err)

	iamSvc.AssertCalledWith(&s.Suite, "PutRolePolicy", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*iam.PutRolePolicyInput)
		return ok && aws.ToString(in.RoleName) == afRoleName &&
			afManageConnectionsStatementPresent(aws.ToString(in.PolicyDocument))
	})

	// The grant is projected into the link data under the execution-role key.
	s.NotNil(out.LinkData.Fields["listOrdersExecutionRole"])
}

func afManageConnectionsStatementPresent(policyDocument string) bool {
	var doc struct {
		Statement []struct {
			Sid      string
			Action   string
			Resource string
		}
	}
	if err := json.Unmarshal([]byte(policyDocument), &doc); err != nil {
		return false
	}
	for _, statement := range doc.Statement {
		if statement.Sid == afManageConnectionsSID {
			return statement.Action == "execute-api:ManageConnections" &&
				statement.Resource == afManageConnectionsARN
		}
	}
	return false
}

type APIFunctionLinkUpdateSuite struct {
	suite.Suite
}

func (s *APIFunctionLinkUpdateSuite) updateInput(protocolType string) *provider.LinkUpdateIntermediaryResourcesInput {
	return &provider.LinkUpdateIntermediaryResourcesInput{
		LinkID:         "link-1",
		LinkUpdateType: provider.LinkUpdateTypeCreate,
		ResourceAInfo:  afAPIInfo(protocolType),
		ResourceBInfo:  afFunctionInfo(),
		LinkContext:    testLinkContext(),
		InstanceName:   "instance",
	}
}

func (s *APIFunctionLinkUpdateSuite) Test_create_deploys_integration_route_and_permission() {
	loader := &testutils.MockAWSConfigLoader{}
	rs := resourceservicemock.Create(resourceservicemock.WithDeployOutput(afDeployOutput()))

	input := s.updateInput("HTTP")
	input.ResourceService = rs

	out, err := testActions(loader).UpdateIntermediaryResources(context.Background(), input)
	s.Require().NoError(err)

	// Three intermediaries, deployed in order: integration, route, permission.
	s.Require().Len(rs.DeployCalls, 3)

	integration := rs.DeployCalls[0]
	s.Equal("aws/apigatewayv2/integration", integration.ResourceType)
	s.Equal(afIntegrationResourceID, integration.Input.DeployInput.ResourceID)
	iSpec := integration.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(afAPIID, core.StringValue(iSpec.Fields["apiId"]))
	s.Equal("AWS_PROXY", core.StringValue(iSpec.Fields["integrationType"]))
	s.Equal(afFunctionARN, core.StringValue(iSpec.Fields["integrationUri"]))
	s.Equal("2.0", core.StringValue(iSpec.Fields["payloadFormatVersion"]))

	route := rs.DeployCalls[1]
	s.Equal("aws/apigatewayv2/route", route.ResourceType)
	s.Equal(afRouteResourceID, route.Input.DeployInput.ResourceID)
	rSpec := route.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(afAPIID, core.StringValue(rSpec.Fields["apiId"]))
	s.Equal(afRouteKey, core.StringValue(rSpec.Fields["routeKey"]))
	s.Equal("integrations/"+afIntegrationID, core.StringValue(rSpec.Fields["target"]))

	permission := rs.DeployCalls[2]
	s.Equal("aws/lambda/permission", permission.ResourceType)
	s.Equal(afPermissionResourceID, permission.Input.DeployInput.ResourceID)
	pSpec := permission.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(afFunctionARN, core.StringValue(pSpec.Fields["functionName"]))
	s.Equal("lambda:InvokeFunction", core.StringValue(pSpec.Fields["action"]))
	s.Equal("apigateway.amazonaws.com", core.StringValue(pSpec.Fields["principal"]))
	s.Equal(afExecuteAPIARN, core.StringValue(pSpec.Fields["sourceArn"]))

	// Output records all three intermediary states and projects them into link data.
	s.Require().Len(out.IntermediaryResourceStates, 3)
	intermediaries := out.LinkData.Fields["intermediaries"].Fields
	s.Contains(intermediaries, afIntegrationResourceID)
	s.Contains(intermediaries, afRouteResourceID)
	s.Contains(intermediaries, afPermissionResourceID)
}

func (s *APIFunctionLinkUpdateSuite) Test_create_guarded_route_sets_authorizer() {
	loader := &testutils.MockAWSConfigLoader{}
	rs := resourceservicemock.Create(resourceservicemock.WithDeployOutput(afDeployOutput()))

	fn := afFunctionInfo()
	// Add the guard annotations (resolved authorizer id + type) to the route function.
	fn.ResourceWithResolvedSubs.Metadata.Annotations.Fields["aws.apigatewayv2.lambda.authorizerId"] =
		core.MappingNodeFromString("auth-xyz")
	fn.ResourceWithResolvedSubs.Metadata.Annotations.Fields["aws.apigatewayv2.lambda.authorizationType"] =
		core.MappingNodeFromString("JWT")

	input := s.updateInput("HTTP")
	input.ResourceBInfo = fn
	input.ResourceService = rs

	_, err := testActions(loader).UpdateIntermediaryResources(context.Background(), input)
	s.Require().NoError(err)

	rSpec := rs.DeployCalls[1].Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal("JWT", core.StringValue(rSpec.Fields["authorizationType"]))
	s.Equal("auth-xyz", core.StringValue(rSpec.Fields["authorizerId"]))
}

func (s *APIFunctionLinkUpdateSuite) Test_create_unguarded_route_omits_authorizer() {
	loader := &testutils.MockAWSConfigLoader{}
	rs := resourceservicemock.Create(resourceservicemock.WithDeployOutput(afDeployOutput()))

	input := s.updateInput("HTTP")
	input.ResourceService = rs

	_, err := testActions(loader).UpdateIntermediaryResources(context.Background(), input)
	s.Require().NoError(err)

	rSpec := rs.DeployCalls[1].Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	_, hasType := rSpec.Fields["authorizationType"]
	_, hasID := rSpec.Fields["authorizerId"]
	s.False(hasType)
	s.False(hasID)
}

func (s *APIFunctionLinkUpdateSuite) Test_create_websocket_two_way_deploys_responses() {
	loader := &testutils.MockAWSConfigLoader{}
	rs := resourceservicemock.Create(resourceservicemock.WithDeployOutput(afDeployOutput()))

	fn := afFunctionInfo()
	fn.ResourceWithResolvedSubs.Metadata.Annotations.Fields["aws.apigatewayv2.lambda.routeKey"] =
		core.MappingNodeFromString("$default")
	fn.ResourceWithResolvedSubs.Metadata.Annotations.Fields["aws.apigatewayv2.lambda.websocketTwoWay"] =
		core.MappingNodeFromBool(true)
	// Focus this test on the responses; the ManageConnections grant has its own test.
	fn.ResourceWithResolvedSubs.Metadata.Annotations.Fields["aws.apigatewayv2.lambda.manageConnections"] =
		core.MappingNodeFromBool(false)

	input := s.updateInput("WEBSOCKET")
	input.ResourceBInfo = fn
	input.ResourceService = rs

	out, err := testActions(loader).UpdateIntermediaryResources(context.Background(), input)
	s.Require().NoError(err)

	// integration, route, permission, integrationResponse, routeResponse.
	s.Require().Len(rs.DeployCalls, 5)
	s.Equal("aws/apigatewayv2/integrationResponse", rs.DeployCalls[3].ResourceType)
	irSpec := rs.DeployCalls[3].Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(afIntegrationID, core.StringValue(irSpec.Fields["integrationId"]))
	s.Equal("$default", core.StringValue(irSpec.Fields["integrationResponseKey"]))

	s.Equal("aws/apigatewayv2/routeResponse", rs.DeployCalls[4].ResourceType)
	rrSpec := rs.DeployCalls[4].Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(afRouteID, core.StringValue(rrSpec.Fields["routeId"]))
	s.Equal("$default", core.StringValue(rrSpec.Fields["routeResponseKey"]))

	s.Require().Len(out.IntermediaryResourceStates, 5)
}

func (s *APIFunctionLinkUpdateSuite) Test_create_websocket_omits_payload_format_version() {
	loader := &testutils.MockAWSConfigLoader{}
	rs := resourceservicemock.Create(resourceservicemock.WithDeployOutput(afDeployOutput()))

	fn := afFunctionInfo()
	fn.ResourceWithResolvedSubs.Metadata.Annotations.Fields["aws.apigatewayv2.lambda.manageConnections"] =
		core.MappingNodeFromBool(false)

	input := s.updateInput("WEBSOCKET")
	input.ResourceBInfo = fn
	input.ResourceService = rs

	_, err := testActions(loader).UpdateIntermediaryResources(context.Background(), input)
	s.Require().NoError(err)

	s.Require().Len(rs.DeployCalls, 3)
	iSpec := rs.DeployCalls[0].Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	_, hasPayloadFormat := iSpec.Fields["payloadFormatVersion"]
	s.False(hasPayloadFormat, "WebSocket integrations must not set payloadFormatVersion")
}

func (s *APIFunctionLinkUpdateSuite) Test_destroy_removes_all_intermediaries() {
	loader := &testutils.MockAWSConfigLoader{}
	rs := resourceservicemock.Create()

	input := s.updateInput("HTTP")
	input.LinkUpdateType = provider.LinkUpdateTypeDestroy
	input.ResourceService = rs
	input.CurrentLinkState = &state.LinkState{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
			afIntermediaryState(afIntegrationResourceID, "aws/apigatewayv2/integration"),
			afIntermediaryState(afRouteResourceID, "aws/apigatewayv2/route"),
			afIntermediaryState(afPermissionResourceID, "aws/lambda/permission"),
		},
	}

	_, err := testActions(loader).UpdateIntermediaryResources(context.Background(), input)
	s.Require().NoError(err)

	// The route is destroyed before the integration it targets; the permission is independent.
	s.Require().Len(rs.DestroyCalls, 3)
	s.Equal(afRouteResourceID, rs.DestroyCalls[0].Input.ResourceID)
	s.Equal(afIntegrationResourceID, rs.DestroyCalls[1].Input.ResourceID)
	s.Equal(afPermissionResourceID, rs.DestroyCalls[2].Input.ResourceID)
}

func TestAPIFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(APIFunctionLinkUpdateSuite))
}
