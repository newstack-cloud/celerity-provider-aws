package s3sqs

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *bucketQueueLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	// This link grants an SQS queue policy as a link-owned intermediary resource (not a
	// tracked spec field of either resource) and merges an entry into the bucket's
	// notification configuration. The queue policy is projected into the link's linkData so
	// its create/update/destroy is surfaced as link changes; the source bucket and target
	// queue ARNs are derived from the linked resources.
	changes := &provider.LinkChanges{}
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	identity := queuePolicyIntermediaryIdentity(
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
				Leaf:             "queueArn",
				ResourceChanges:  input.ResourceBChanges,
				ResourceSpecPath: "$.spec.arn",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	linkutils.CollectS3NotificationChanges(
		changes,
		input.ResourceAChanges.AppliedResourceInfo.ResourceName,
		linkutils.S3QueueTarget,
		input.ResourceAChanges,
		input.ResourceBChanges,
	)

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
