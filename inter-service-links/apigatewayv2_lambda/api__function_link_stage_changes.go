package apigatewayv2lambda

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *apiFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	// This link manages an integration, route and Lambda permission as link-owned intermediary
	// resources (not tracked spec fields of either resource). Each is projected into the link's
	// linkData so its create/update/destroy surfaces as link changes; derived leaves come from
	// the linked resources, and computed values (the integration id, the route target) resolve on
	// deploy.
	changes := &provider.LinkChanges{}
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	apiInfo := &input.ResourceAChanges.AppliedResourceInfo
	functionInfo := &input.ResourceBChanges.AppliedResourceInfo

	stagings := []linkutils.StageIntermediary{
		{
			Identity: integrationIntermediaryIdentity(apiInfo, functionInfo),
			DerivedLeaves: []linkutils.DerivedLeaf{
				{Leaf: "apiId", ResourceChanges: input.ResourceAChanges, ResourceSpecPath: "$.spec.apiId"},
				{Leaf: "integrationUri", ResourceChanges: input.ResourceBChanges, ResourceSpecPath: "$.spec.arn"},
			},
		},
		{
			Identity: routeIntermediaryIdentity(apiInfo, functionInfo),
			DerivedLeaves: []linkutils.DerivedLeaf{
				{Leaf: "apiId", ResourceChanges: input.ResourceAChanges, ResourceSpecPath: "$.spec.apiId"},
			},
		},
		{
			Identity: permissionIntermediaryIdentity(apiInfo, functionInfo),
			DerivedLeaves: []linkutils.DerivedLeaf{
				{Leaf: "functionArn", ResourceChanges: input.ResourceBChanges, ResourceSpecPath: "$.spec.arn"},
			},
		},
	}

	annotations := getAPILinkAnnotations(functionInfo)
	webSocket := !isHTTPProtocol(apiInfo)

	// WebSocket two-way responses are managed as additional intermediaries when the route opts in.
	if annotations.websocketTwoWay && webSocket {
		stagings = append(stagings,
			linkutils.StageIntermediary{
				Identity: integrationResponseIntermediaryIdentity(apiInfo, functionInfo),
				DerivedLeaves: []linkutils.DerivedLeaf{
					{Leaf: "apiId", ResourceChanges: input.ResourceAChanges, ResourceSpecPath: "$.spec.apiId"},
				},
			},
			linkutils.StageIntermediary{
				Identity: routeResponseIntermediaryIdentity(apiInfo, functionInfo),
				DerivedLeaves: []linkutils.DerivedLeaf{
					{Leaf: "apiId", ResourceChanges: input.ResourceAChanges, ResourceSpecPath: "$.spec.apiId"},
				},
			},
		)
	}

	for _, staging := range stagings {
		if err := linkutils.CollectIntermediaryChanges(currentLinkData, changes, staging); err != nil {
			return nil, err
		}
	}

	// The execute-api:ManageConnections grant on the handler's execution role references the API's
	// execute-api ARN, known on deploy. Surface it for a new WebSocket route that keeps the grant.
	if annotations.manageConnections && webSocket &&
		(pluginutils.IsResourceNew(input.ResourceAChanges) || pluginutils.IsResourceNew(input.ResourceBChanges)) {
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			executionRoleLinkDataName(functionInfo),
		)
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
