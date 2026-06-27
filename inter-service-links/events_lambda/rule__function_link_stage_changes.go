package eventslambda

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *ruleFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	identity := ruleFunctionIntermediaryIdentity(
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
