//go:build unit

package apigatewayv2lambda

import (
	"context"
	"testing"

	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/suite"
)

type AuthorizerFunctionLinkUpdateSuite struct {
	suite.Suite
}

const (
	afAuthorizerID   = "authz-1"
	afAuthzSourceARN = "arn:aws:execute-api:us-west-2:123456789012:api-abc123/authorizers/authz-1"
	afAuthzPermID    = "ordersAuth__authFn__apigw-authorizer-permission"
)

func afAuthorizerInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersAuth",
		InstanceID:   afInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"apiId", core.MappingNodeFromString(afAPIID),
				"authorizerId", core.MappingNodeFromString(afAuthorizerID),
			),
		},
	}
}

func afGuardFunctionInfo() *provider.ResourceInfo {
	spec := core.MappingNodeFields("arn", core.MappingNodeFromString(afFunctionARN))
	return &provider.ResourceInfo{
		ResourceName:             "authFn",
		InstanceID:               afInstanceID,
		CurrentResourceState:     &state.ResourceState{SpecData: spec},
		ResourceWithResolvedSubs: &provider.ResolvedResource{Spec: spec},
	}
}

func (s *AuthorizerFunctionLinkUpdateSuite) Test_create_grants_invoke_permission() {
	loader := &testutils.MockAWSConfigLoader{}
	rs := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{}))

	out, err := (&authorizerFunctionLinkActions{awsConfigStore: testConfigStore(loader)}).
		UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
			LinkID:          "link-1",
			LinkUpdateType:  provider.LinkUpdateTypeCreate,
			ResourceAInfo:   afAuthorizerInfo(),
			ResourceBInfo:   afGuardFunctionInfo(),
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
			ResourceService: rs,
		})
	s.Require().NoError(err)

	s.Require().Len(rs.DeployCalls, 1)
	call := rs.DeployCalls[0]
	s.Equal("aws/lambda/permission", call.ResourceType)
	s.Equal(afAuthzPermID, call.Input.DeployInput.ResourceID)
	spec := call.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(afFunctionARN, core.StringValue(spec.Fields["functionName"]))
	s.Equal("apigateway.amazonaws.com", core.StringValue(spec.Fields["principal"]))
	s.Equal(afAuthzSourceARN, core.StringValue(spec.Fields["sourceArn"]))
	s.Require().Len(out.IntermediaryResourceStates, 1)
}

func (s *AuthorizerFunctionLinkUpdateSuite) Test_destroy_removes_permission() {
	loader := &testutils.MockAWSConfigLoader{}
	rs := resourceservicemock.Create()

	_, err := (&authorizerFunctionLinkActions{awsConfigStore: testConfigStore(loader)}).
		UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
			LinkID:          "link-1",
			LinkUpdateType:  provider.LinkUpdateTypeDestroy,
			ResourceAInfo:   afAuthorizerInfo(),
			ResourceBInfo:   afGuardFunctionInfo(),
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
			ResourceService: rs,
			CurrentLinkState: &state.LinkState{
				IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
					afIntermediaryState(afAuthzPermID, "aws/lambda/permission"),
				},
			},
		})
	s.Require().NoError(err)

	s.Require().Len(rs.DestroyCalls, 1)
	s.Equal(afAuthzPermID, rs.DestroyCalls[0].Input.ResourceID)
}

func TestAuthorizerFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(AuthorizerFunctionLinkUpdateSuite))
}
