package snslambda

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *subscriptionFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	// This link grants a Lambda permission as a link-owned intermediary resource, which is
	// not a tracked spec field of either resource. It is projected into the link's linkData
	// so its create/update/destroy is surfaced as link changes.
	changes := &provider.LinkChanges{}
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	identity := subscriptionFunctionIntermediaryIdentity(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	err := linkutils.CollectIntermediaryChanges(currentLinkData, changes, linkutils.StageIntermediary{
		Identity: identity,
		DerivedLeaves: []linkutils.DerivedLeaf{
			{
				Leaf:             "sourceArn",
				ResourceChanges:  input.ResourceAChanges,
				ResourceSpecPath: "$.spec.topicArn",
			},
			{
				Leaf:             "functionName",
				ResourceChanges:  input.ResourceBChanges,
				ResourceSpecPath: "$.spec.arn",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
