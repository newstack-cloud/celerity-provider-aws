//go:build unit

package sqslinks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type QueueQueueLinkStageChangesSuite struct {
	suite.Suite
}

func (s *QueueQueueLinkStageChangesSuite) Test_stage_changes() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		createQueueQueueLinkChangesTestCase(providerCtx, loader),
		createQueueQueueLinkChangesKnownOnDeployTestCase(providerCtx, loader),
		createQueueQueueLinkChangesNoChangesTestCase(providerCtx, loader),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		QueueQueueLink,
		&s.Suite,
	)
}

func createQueueQueueLinkChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	dlqARN := "arn:aws:sqs:us-west-2:123456789012:dead-letter-queue"

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name: "has changes for adding dead-letter queue ARN for known value",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "source-queue",
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "dead-letter-queue",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"arn": core.MappingNodeFromString(dlqARN),
							},
						},
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.arn",
						NewValue:  core.MappingNodeFromString(dlqARN),
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
						FieldPath: "[\"source-queue\"].redrivePolicy.deadLetterTargetArn",
						NewValue:  core.MappingNodeFromString(dlqARN),
					},
					{
						FieldPath: "[\"source-queue\"].redrivePolicy.maxReceiveCount",
						NewValue:  core.MappingNodeFromInt(10),
					},
				},
			},
		},
	}
}

func createQueueQueueLinkChangesKnownOnDeployTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name: "has changes for adding DLQ ARN not known until deployment",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "source-queue",
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "dead-letter-queue",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{},
						},
					},
				},
				FieldChangesKnownOnDeploy: []string{
					"spec.arn",
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
						FieldPath: "[\"source-queue\"].redrivePolicy.maxReceiveCount",
						NewValue:  core.MappingNodeFromInt(10),
					},
				},
				FieldChangesKnownOnDeploy: []string{
					"[\"source-queue\"].redrivePolicy.deadLetterTargetArn",
				},
			},
		},
	}
}

func createQueueQueueLinkChangesNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	dlqARN := "arn:aws:sqs:us-west-2:123456789012:dead-letter-queue"

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		Name: "has no changes when DLQ ARN and maxReceiveCount are unchanged",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "source-queue",
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "dead-letter-queue",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"arn": core.MappingNodeFromString(dlqARN),
							},
						},
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"source-queue": {
						Fields: map[string]*core.MappingNode{
							"redrivePolicy": {
								Fields: map[string]*core.MappingNode{
									"deadLetterTargetArn": core.MappingNodeFromString(dlqARN),
									"maxReceiveCount":     core.MappingNodeFromInt(10),
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
					"[\"source-queue\"].redrivePolicy.deadLetterTargetArn",
					"[\"source-queue\"].redrivePolicy.maxReceiveCount",
				},
			},
		},
	}
}

func TestQueueQueueLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(QueueQueueLinkStageChangesSuite))
}
