//go:build unit

package s3sqs

import (
	"context"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/suite"
)

type BucketQueueLinkStageChangesSuite struct {
	suite.Suite
}

func policyLeaf(leaf string) string {
	return "[\"intermediaries\"][\"" + bqResourceID + "\"][\"" + leaf + "\"]"
}

// When the resources are new, the link projects the queue policy intermediary and the
// bucket notification configuration as known-on-deploy changes.
func (s *BucketQueueLinkStageChangesSuite) Test_stage_changes_projects_policy_and_notification() {
	out, err := (&bucketQueueLinkActions{}).StageChanges(
		context.Background(),
		&provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "ordersBucket",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.bucketName", NewValue: core.MappingNodeFromString("orders")},
				},
				FieldChangesKnownOnDeploy: []string{"spec.arn"},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "orderEventsQueue",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
					},
				},
				FieldChangesKnownOnDeploy: []string{"spec.arn"},
			},
			CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
		},
	)
	s.Require().NoError(err)

	var hasPolicyResourceType bool
	for _, fc := range out.Changes.NewFields {
		if fc.FieldPath == policyLeaf("resourceType") &&
			core.StringValue(fc.NewValue) == "aws/sqs/queueInlinePolicy" {
			hasPolicyResourceType = true
		}
	}
	s.True(hasPolicyResourceType, "expected the sqs/queueInlinePolicy intermediary to be staged")

	s.Contains(out.Changes.FieldChangesKnownOnDeploy, policyLeaf("sourceArn"))
	s.Contains(out.Changes.FieldChangesKnownOnDeploy, policyLeaf("queueArn"))
	s.Contains(
		out.Changes.FieldChangesKnownOnDeploy,
		"ordersBucket.notificationConfiguration.queueConfigurations",
	)
}

func (s *BucketQueueLinkStageChangesSuite) Test_stage_changes_omits_notification_when_no_new_resources() {
	out, err := (&bucketQueueLinkActions{}).StageChanges(
		context.Background(),
		&provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "ordersBucket",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString("arn:aws:s3:::orders"),
						}},
					},
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "orderEventsQueue",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:order-events"),
						}},
					},
				},
			},
			CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
		},
	)
	s.Require().NoError(err)
	s.NotContains(
		out.Changes.FieldChangesKnownOnDeploy,
		"ordersBucket.notificationConfiguration.queueConfigurations",
	)
}

func TestBucketQueueLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(BucketQueueLinkStageChangesSuite))
}
