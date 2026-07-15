package ssm

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *parameterResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	// The parameter must be deleted from the region it was created in, so the region
	// recorded in the resource state (when set) re-targets the client.
	service, _, err := a.getSSMServiceWithRegion(
		ctx,
		input.ProviderContext,
		parameterRegionMeta(input.ResourceState.SpecData),
	)
	if err != nil {
		return err
	}

	nameValue, hasName := pluginutils.GetValueByPath("$.name", input.ResourceState.SpecData)
	if !hasName {
		return errors.New("name is required to destroy an SSM parameter")
	}

	_, err = service.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(core.StringValue(nameValue)),
	})
	if err != nil {
		return err
	}

	return nil
}
