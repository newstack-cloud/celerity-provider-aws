package cloudcontrol

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *ccResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	token := readRequestToken(input.ResourceSpec)
	if token == "" {
		// No in-flight operation (e.g. a no-op update) so is already stable.
		return &provider.ResourceHasStabilisedOutput{Stabilised: true}, nil
	}

	service, err := a.getCloudControlService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	getStatus := pluginutils.RetryableReturnValue(
		func(
			ctx context.Context,
			in *awscc.GetResourceRequestStatusInput,
		) (*awscc.GetResourceRequestStatusOutput, error) {
			return service.GetResourceRequestStatus(ctx, in)
		},
		isCCErrorRetryable,
	)

	output, err := getStatus(ctx, &awscc.GetResourceRequestStatusInput{
		RequestToken: aws.String(token),
	})
	if err != nil {
		return nil, err
	}

	if output.ProgressEvent == nil {
		return nil, fmt.Errorf(
			"cloud control request status for %s returned no progress event",
			a.config.CFNType,
		)
	}

	switch output.ProgressEvent.OperationStatus {
	case cctypes.OperationStatusSuccess:
		return a.stabilisedSuccessOutput(ctx, service, output.ProgressEvent, input)
	case cctypes.OperationStatusFailed,
		cctypes.OperationStatusCancelInProgress,
		cctypes.OperationStatusCancelComplete:
		return nil, fmt.Errorf(
			"cloud control operation for %s failed (%s): %s",
			a.config.CFNType,
			output.ProgressEvent.ErrorCode,
			aws.ToString(output.ProgressEvent.StatusMessage),
		)
	default:
		return &provider.ResourceHasStabilisedOutput{Stabilised: false}, nil
	}
}

func (a *ccResourceActions) stabilisedSuccessOutput(
	ctx context.Context,
	service cloudcontrolservice.Service,
	progress *cctypes.ProgressEvent,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// On success the resource model carries every computed field, including any (e.g. an
	// endpoint address) that were not yet available when first captured at config-complete.
	// Returning them lets the engine finalise them in the resource's persisted state.
	identifier := aws.ToString(progress.Identifier)
	if identifier == "" {
		identifier = primaryIdentifier(input.ResourceSpec, a.config.Meta)
	}
	specState, err := a.completedSpecState(
		ctx, service, progress, identifier, input.ProviderContext,
	)
	if err != nil {
		return nil, err
	}
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
		ComputedFieldValues: a.computedFieldValues(
			specState,
			identifier,
			"",
			a.omittedAutoNamedFields(input.ResourceSpec),
		),
	}, nil
}

func readRequestToken(spec *core.MappingNode) string {
	value, ok := pluginutils.GetValueByPath("$."+fieldRequestToken, spec)
	if !ok {
		return ""
	}
	return core.StringValue(value)
}
