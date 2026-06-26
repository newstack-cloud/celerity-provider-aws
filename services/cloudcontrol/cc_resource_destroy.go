package cloudcontrol

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/aws/smithy-go"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Destroy deletion polling bounds.
var (
	destroyPollInterval = 5 * time.Second
	destroyPollTimeout  = 30 * time.Minute
)

func (a *ccResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	service, err := a.getCloudControlService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	identifier := ""
	if input.ResourceState != nil {
		identifier = primaryIdentifier(
			input.ResourceState.SpecData,
			a.config.Meta,
		)
	}
	if identifier == "" {
		// Nothing recorded to delete so treat as already destroyed.
		return nil
	}

	deleteResource := pluginutils.RetryableReturnValue(
		func(
			ctx context.Context,
			in *awscc.DeleteResourceInput,
		) (*awscc.DeleteResourceOutput, error) {
			return service.DeleteResource(ctx, in)
		},
		isCCErrorRetryable,
	)

	output, err := deleteResource(ctx, &awscc.DeleteResourceInput{
		TypeName:    aws.String(a.config.CFNType),
		Identifier:  aws.String(identifier),
		ClientToken: aws.String(clientToken("delete", input.InstanceID, input.ResourceID)),
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return err
	}

	if output.ProgressEvent == nil {
		return nil
	}
	return a.waitForDeletion(ctx, service, aws.ToString(output.ProgressEvent.RequestToken))
}

func (a *ccResourceActions) waitForDeletion(
	ctx context.Context,
	service cloudcontrolservice.Service,
	token string,
) error {
	if token == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, destroyPollTimeout)
	defer cancel()

	getStatus := pluginutils.RetryableReturnValue(
		func(ctx context.Context, in *awscc.GetResourceRequestStatusInput) (*awscc.GetResourceRequestStatusOutput, error) {
			return service.GetResourceRequestStatus(ctx, in)
		},
		isCCErrorRetryable,
	)

	for {
		output, err := getStatus(ctx, &awscc.GetResourceRequestStatusInput{
			RequestToken: aws.String(token),
		})
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return err
		}

		if progress := output.ProgressEvent; progress != nil {
			done, err := a.deletionStatusFromProgress(progress)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf(
				"timed out waiting for %s deletion to complete",
				a.config.CFNType,
			)
		case <-time.After(destroyPollInterval):
		}
	}
}

func (a *ccResourceActions) deletionStatusFromProgress(
	progress *cctypes.ProgressEvent,
) (done bool, err error) {
	switch progress.OperationStatus {
	case cctypes.OperationStatusSuccess:
		return true, nil
	case cctypes.OperationStatusFailed,
		cctypes.OperationStatusCancelInProgress,
		cctypes.OperationStatusCancelComplete:
		if progress.ErrorCode == cctypes.HandlerErrorCodeNotFound {
			return true, nil
		}
		return false, fmt.Errorf(
			"cloud control delete for %s failed (%s): %s",
			a.config.CFNType,
			progress.ErrorCode,
			aws.ToString(progress.StatusMessage),
		)
	}
	return false, nil
}

func isNotFoundError(err error) bool {
	if _, ok := errors.AsType[*cctypes.ResourceNotFoundException](err); ok {
		return true
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		return apiErr.ErrorCode() == "ResourceNotFoundException"
	}
	return false
}
