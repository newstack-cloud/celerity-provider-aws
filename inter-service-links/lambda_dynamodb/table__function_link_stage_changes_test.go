//go:build unit

package lambdadynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type TableFunctionLinkStageChangesSuite struct {
	suite.Suite
}

func tableFunctionStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, dynamodbservice.Service, *aws.Config, lambdaservice.Service],
) provider.Link {
	build := DynamoDBTableLambdaFunctionLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service {
			return iammock.CreateIamServiceMock()
		},
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, dynamodbservice.Service, *aws.Config, lambdaservice.Service],
	) provider.Link {
		return build(DynamoDBTableToLambdaFunctionLinkDeps(deps))
	}
}

const (
	esmStreamARN   = "arn:aws:dynamodb:us-west-2:123456789012:table/orders/stream/2024"
	esmFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:process-stream"
	// The link-owned event source mapping intermediary id (table__function__suffix).
	tableFunctionESMID = "ordersTable__processStreamFunction__event-source-mapping"
)

func esmLeaf(leaf string) string {
	return "[\"intermediaries\"][\"" + tableFunctionESMID + "\"][\"" + leaf + "\"]"
}

func (s *TableFunctionLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		stageTableFunctionNewResourcesTestCase(),
		stageTableFunctionArnChangesTestCase(),
		stageTableFunctionNoChangesTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		tableFunctionStageLinkFactory(),
		&s.Suite,
	)
}

func stageTableFunctionNewResourcesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "marks the event source mapping ARNs as known on deploy when resources are new",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "ordersTable",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
					},
				},
				FieldChangesKnownOnDeploy: []string{"spec.streamArn"},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "processStreamFunction",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
					},
				},
				FieldChangesKnownOnDeploy: []string{"spec.arn"},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{FieldPath: esmLeaf("resourceType"), NewValue: core.MappingNodeFromString("aws/lambda/eventSourceMapping")},
				},
				FieldChangesKnownOnDeploy: []string{
					esmLeaf("eventSourceArn"),
					esmLeaf("functionArn"),
				},
			},
		},
	}
}

func stageTableFunctionArnChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "stages create of the event source mapping with resolved ARNs",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "ordersTable",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
							"streamArn": core.MappingNodeFromString(esmStreamARN),
						}},
					},
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "processStreamFunction",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(esmFunctionARN),
						}},
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
					{FieldPath: esmLeaf("resourceType"), NewValue: core.MappingNodeFromString("aws/lambda/eventSourceMapping")},
					{FieldPath: esmLeaf("eventSourceArn"), NewValue: core.MappingNodeFromString(esmStreamARN)},
					{FieldPath: esmLeaf("functionArn"), NewValue: core.MappingNodeFromString(esmFunctionARN)},
				},
			},
		},
	}
}

func stageTableFunctionNoChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "stages no changes when ARNs already match the link state",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "ordersTable",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
							"streamArn": core.MappingNodeFromString(esmStreamARN),
						}},
					},
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "processStreamFunction",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(esmFunctionARN),
						}},
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"intermediaries": {Fields: map[string]*core.MappingNode{
						tableFunctionESMID: {Fields: map[string]*core.MappingNode{
							"resourceType":   core.MappingNodeFromString("aws/lambda/eventSourceMapping"),
							"eventSourceArn": core.MappingNodeFromString(esmStreamARN),
							"functionArn":    core.MappingNodeFromString(esmFunctionARN),
						}},
					}},
				},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				UnchangedFields: []string{
					esmLeaf("resourceType"),
					esmLeaf("eventSourceArn"),
					esmLeaf("functionArn"),
				},
			},
		},
	}
}

func TestTableFunctionLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(TableFunctionLinkStageChangesSuite))
}
