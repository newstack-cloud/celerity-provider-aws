package s3sns

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *bucketTopicLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	// This link grants a topic policy as a link-owned intermediary resource (not a tracked
	// spec field of either resource) and merges an entry into the bucket's notification
	// configuration. The policy is projected into the link's linkData so its
	// create/update/destroy is surfaced as link changes; the source bucket and target topic
	// ARNs are derived from the linked resources.
	changes := &provider.LinkChanges{}
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	identity := topicPolicyIntermediaryIdentity(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	err := linkutils.CollectIntermediaryChanges(currentLinkData, changes, linkutils.StageIntermediary{
		Identity: identity,
		DerivedLeaves: []linkutils.DerivedLeaf{
			{
				Leaf:             "sourceArn",
				ResourceChanges:  input.ResourceAChanges,
				ResourceSpecPath: "$.spec.arn",
			},
			{
				Leaf:             "topicArn",
				ResourceChanges:  input.ResourceBChanges,
				ResourceSpecPath: "$.spec.topicArn",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	linkutils.CollectS3NotificationChanges(
		changes,
		input.ResourceAChanges.AppliedResourceInfo.ResourceName,
		linkutils.S3TopicTarget,
		input.ResourceAChanges,
		input.ResourceBChanges,
	)

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
