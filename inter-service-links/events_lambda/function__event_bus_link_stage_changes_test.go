//go:build unit

package eventslambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FunctionEventBusLinkStageChangesSuite struct {
	suite.Suite
}

func functionEventBusLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service],
) provider.Link {
	build := FunctionEventBusLink(iammock.CreateIamServiceMockFactory())
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service],
	) provider.Link {
		return build(FunctionToEventBusLinkDeps(deps))
	}
}

func (s *FunctionEventBusLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		stageEventBusEnvVarChangeTestCase(),
		stageEventBusNoChangesTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionEventBusLinkFactory(),
		&s.Suite,
	)
}

func stageEventBusEnvVarChangeTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	eventsservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		Name: "stages env var change for the linked event bus",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "publishOrderFunction",
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "orderEventBus",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"name": core.MappingNodeFromString("order-events"),
							},
						},
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.name",
						NewValue:  core.MappingNodeFromString("order-events"),
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "publishOrderFunction.environmentVariables[\"EVENT_BUS_orderEventBus\"]",
						NewValue:  core.MappingNodeFromString("order-events"),
					},
				},
				// The event bus is new (it has NewFields), so the execution-role
				// permission data for this link is only known at deploy time.
				FieldChangesKnownOnDeploy: []string{"publishOrderFunctionExecutionRole"},
			},
		},
	}
}

func stageEventBusNoChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	eventsservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		Name: "stages no changes when the env var already matches",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "publishOrderFunction",
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "orderEventBus",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"name": core.MappingNodeFromString("order-events"),
							},
						},
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"publishOrderFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									"EVENT_BUS_orderEventBus": core.MappingNodeFromString("order-events"),
								},
							},
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				UnchangedFields: []string{
					"publishOrderFunction.environmentVariables[\"EVENT_BUS_orderEventBus\"]",
				},
			},
		},
	}
}

func TestFunctionEventBusLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionEventBusLinkStageChangesSuite))
}
