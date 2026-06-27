package lambdadynamodb

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *dynamoDBTableLambdaFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	// This link creates a Lambda event source mapping as a link-owned intermediary
	// resource (the DynamoDB table's stream cannot be modelled as a spec field of either
	// resource). It is projected into the link's linkData under the "intermediaries" map so
	// its create/update is surfaced as link changes; the stream/function ARNs are derived
	// from the linked resources and are known on deploy when either resource is new.
	changes := &provider.LinkChanges{}
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	identity := tableFunctionESMIdentity(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	err := linkutils.CollectIntermediaryChanges(currentLinkData, changes, linkutils.StageIntermediary{
		Identity: identity,
		DerivedLeaves: []linkutils.DerivedLeaf{
			{Leaf: "eventSourceArn", ResourceChanges: input.ResourceAChanges, ResourceSpecPath: "$.spec.streamArn"},
			{Leaf: "functionArn", ResourceChanges: input.ResourceBChanges, ResourceSpecPath: "$.spec.arn"},
		},
	})
	if err != nil {
		return nil, err
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
