package apigatewayv2lambda

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *authorizerFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	// The invoke permission is a link-owned intermediary; project it so its create/update/destroy
	// surfaces as a link change. Its function ARN derives from the linked function.
	changes := &provider.LinkChanges{}
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	identity := authorizerPermissionIdentity(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	err := linkutils.CollectIntermediaryChanges(currentLinkData, changes, linkutils.StageIntermediary{
		Identity: identity,
		DerivedLeaves: []linkutils.DerivedLeaf{
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

	return &provider.LinkStageChangesOutput{Changes: changes}, nil
}
