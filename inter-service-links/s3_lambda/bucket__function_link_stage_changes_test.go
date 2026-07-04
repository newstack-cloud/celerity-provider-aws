//go:build unit

package s3lambda

import (
	"context"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/suite"
)

type BucketFunctionLinkStageChangesSuite struct {
	suite.Suite
}

func permissionLeaf(leaf string) string {
	return "[\"intermediaries\"][\"" + bfResourceID + "\"][\"" + leaf + "\"]"
}

// When the resources are new, the link projects the Lambda permission intermediary and the
// bucket notification configuration as known-on-deploy changes (their values depend on the
// destination ARN, resolved at deploy time).
func (s *BucketFunctionLinkStageChangesSuite) Test_stage_changes_projects_permission_and_notification() {
	out, err := (&bucketFunctionLinkActions{}).StageChanges(
		context.Background(),
		&provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "ordersBucket",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
					},
				},
				// A new bucket: a user-set field makes the resource "new", while the
				// computed ARN is known only on deploy.
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.bucketName", NewValue: core.MappingNodeFromString("orders")},
				},
				FieldChangesKnownOnDeploy: []string{"spec.arn"},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "processOrderFunction",
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

	// The Lambda permission intermediary is staged as a new resource.
	var hasPermissionResourceType bool
	for _, fc := range out.Changes.NewFields {
		if fc.FieldPath == permissionLeaf("resourceType") &&
			core.StringValue(fc.NewValue) == "aws/lambda/permission" {
			hasPermissionResourceType = true
		}
	}
	s.True(hasPermissionResourceType, "expected the lambda/permission intermediary to be staged")

	// The permission ARNs and the bucket notification configuration are known on deploy.
	s.Contains(out.Changes.FieldChangesKnownOnDeploy, permissionLeaf("sourceArn"))
	s.Contains(out.Changes.FieldChangesKnownOnDeploy, permissionLeaf("functionArn"))
	s.Contains(
		out.Changes.FieldChangesKnownOnDeploy,
		"ordersBucket.notificationConfiguration.lambdaConfigurations",
	)
}

// When neither resource is new, the notification configuration is not surfaced as a
// known-on-deploy change (only an existing entry is reconciled, not added).
func (s *BucketFunctionLinkStageChangesSuite) Test_stage_changes_omits_notification_when_no_new_resources() {
	out, err := (&bucketFunctionLinkActions{}).StageChanges(
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
					ResourceName: "processOrderFunction",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:process-order"),
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
		"ordersBucket.notificationConfiguration.lambdaConfigurations",
	)
}

func TestBucketFunctionLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(BucketFunctionLinkStageChangesSuite))
}
