//go:build unit

package apigatewayv2lambda

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

const (
	afInstanceID    = "instance-1"
	afAPIID         = "api-abc123"
	afFunctionARN   = "arn:aws:lambda:us-west-2:123456789012:function:list-orders"
	afIntegrationID = "intg-1"
	afRouteID       = "route-1"
	afRouteKey      = "GET /orders"
	afExecuteAPIARN = "arn:aws:execute-api:us-west-2:123456789012:api-abc123/*/*"

	afIntegrationResourceID = "ordersApi__listOrders__integration"
	afRouteResourceID       = "ordersApi__listOrders__route"
	afPermissionResourceID  = "ordersApi__listOrders__apigw-invoke-permission"
)

func testLinkContext() provider.LinkContext {
	return plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {"region": core.ScalarFromString("us-west-2")},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func testConfigStore(loader *testutils.MockAWSConfigLoader) pluginutils.ServiceConfigStore[*aws.Config] {
	return utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
}

func testActions(loader *testutils.MockAWSConfigLoader) *apiFunctionLinkActions {
	return &apiFunctionLinkActions{
		awsConfigStore: testConfigStore(loader),
	}
}

func afAPIInfo(protocolType string) *provider.ResourceInfo {
	spec := core.MappingNodeFields(
		"apiId", core.MappingNodeFromString(afAPIID),
		"protocolType", core.MappingNodeFromString(protocolType),
	)
	return &provider.ResourceInfo{
		ResourceName:         "ordersApi",
		InstanceID:           afInstanceID,
		CurrentResourceState: &state.ResourceState{SpecData: spec},
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Spec: spec,
		},
	}
}

func afFunctionInfo() *provider.ResourceInfo {
	spec := core.MappingNodeFields("arn", core.MappingNodeFromString(afFunctionARN))
	return &provider.ResourceInfo{
		ResourceName: "listOrders",
		InstanceID:   afInstanceID,
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: core.MappingNodeFields(
					"aws.apigatewayv2.lambda.routeKey", core.MappingNodeFromString(afRouteKey),
				),
			},
			Spec: spec,
		},
		CurrentResourceState: &state.ResourceState{SpecData: spec},
	}
}

func afDeployOutput() *provider.ResourceDeployOutput {
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.integrationId": core.MappingNodeFromString(afIntegrationID),
			"spec.routeId":       core.MappingNodeFromString(afRouteID),
		},
	}
}

func afIntermediaryState(resourceID, resourceType string) *state.LinkIntermediaryResourceState {
	return &state.LinkIntermediaryResourceState{
		ResourceID:   resourceID,
		ResourceType: resourceType,
		InstanceID:   afInstanceID,
	}
}
