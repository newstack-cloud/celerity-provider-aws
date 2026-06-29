//go:build unit

package kinesislambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type StreamFunctionLinkStageChangesSuite struct {
	suite.Suite
}

func streamFunctionStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, cloudcontrolservice.Service, *aws.Config, lambdaservice.Service],
) provider.Link {
	build := StreamFunctionLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service {
			return iammock.CreateIamServiceMock()
		},
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, cloudcontrolservice.Service, *aws.Config, lambdaservice.Service],
	) provider.Link {
		return build(StreamToFunctionLinkDeps(deps))
	}
}

const (
	esmStreamARN   = "arn:aws:kinesis:us-west-2:123456789012:stream/events-stream"
	esmFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:process-stream"
	// The link-owned event source mapping intermediary id (stream__function__suffix).
	streamFunctionESMID = "eventsStream__processStreamFunction__event-source-mapping"
)

func esmLeaf(leaf string) string {
	return "[\"intermediaries\"][\"" + streamFunctionESMID + "\"][\"" + leaf + "\"]"
}

func (s *StreamFunctionLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		stageStreamFunctionNewResourcesTestCase(),
		stageStreamFunctionArnChangesTestCase(),
		stageStreamFunctionNoChangesTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		streamFunctionStageLinkFactory(),
		&s.Suite,
	)
}

func stageStreamFunctionNewResourcesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "marks the event source mapping ARNs and execution-role grant as known on deploy when resources are new",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "eventsStream",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
					},
				},
				// A new stream: a user-set field appears in NewFields (making the
				// resource "new"), while the computed ARN is known only on deploy.
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.name", NewValue: core.MappingNodeFromString("events-stream")},
				},
				FieldChangesKnownOnDeploy: []string{"spec.arn"},
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
					{
						FieldPath: esmLeaf("resourceType"),
						NewValue:  core.MappingNodeFromString("aws/lambda/eventSourceMapping"),
					},
				},
				FieldChangesKnownOnDeploy: []string{
					esmLeaf("eventSourceArn"),
					esmLeaf("functionArn"),
					"processStreamFunctionExecutionRole",
				},
			},
		},
	}
}

func stageStreamFunctionArnChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "stages create of the event source mapping with resolved ARNs",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "eventsStream",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(esmStreamARN),
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
					{
						FieldPath: esmLeaf("resourceType"),
						NewValue:  core.MappingNodeFromString("aws/lambda/eventSourceMapping"),
					},
					{
						FieldPath: esmLeaf("eventSourceArn"),
						NewValue:  core.MappingNodeFromString(esmStreamARN),
					},
					{
						FieldPath: esmLeaf("functionArn"),
						NewValue:  core.MappingNodeFromString(esmFunctionARN),
					},
				},
			},
		},
	}
}

func stageStreamFunctionNoChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "stages no changes when ARNs already match the link state",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "eventsStream",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(esmStreamARN),
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
						streamFunctionESMID: {Fields: map[string]*core.MappingNode{
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

func TestStreamFunctionLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(StreamFunctionLinkStageChangesSuite))
}
