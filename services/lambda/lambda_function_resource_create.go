package lambda

import (
	"context"

	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaFunctionResourceActions) CreateFunc(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	return nil, nil
}
