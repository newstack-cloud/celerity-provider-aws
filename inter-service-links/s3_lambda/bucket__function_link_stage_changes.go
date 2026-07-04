package s3lambda

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *bucketFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	// This link grants a Lambda permission as a link-owned intermediary resource (not a
	// tracked spec field of either resource) and merges an entry into the bucket's
	// notification configuration. The permission is projected into the link's linkData so
	// its create/update/destroy is surfaced as link changes; the source bucket and target
	// function ARNs are derived from the linked resources.
	changes := &provider.LinkChanges{}
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	identity := permissionIntermediaryIdentity(
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
				Leaf:             "functionArn",
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
		linkutils.S3LambdaTarget,
		input.ResourceAChanges,
		input.ResourceBChanges,
	)

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
