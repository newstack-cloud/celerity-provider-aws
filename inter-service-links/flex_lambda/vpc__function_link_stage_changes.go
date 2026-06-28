package flexlambda

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// StageChanges reports that the function's VpcConfig will be set by this link. The
// concrete subnet IDs and security group are derived from the flex VPC's computed
// outputs, so they are only known at deploy time.
func (l *vpcFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	functionName := pluginutils.GetResourceName(&input.ResourceBChanges.AppliedResourceInfo)
	changes.FieldChangesKnownOnDeploy = append(
		changes.FieldChangesKnownOnDeploy,
		fmt.Sprintf("%s.vpcConfig.subnetIds", functionName),
		fmt.Sprintf("%s.vpcConfig.securityGroupIds", functionName),
	)

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
