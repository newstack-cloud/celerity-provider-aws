package linkutils

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// IsRoleNotYetPropagatedError reports whether an error is the transient validation failure
// AWS returns when a Lambda event source mapping is created before the execution role's
// freshly-granted stream/queue read permissions have propagated (IAM eventual consistency).
//
// It anchors on the smithy API error code (InvalidParameterValueException) as well as a
// message substring so it is decoupled from unrelated errors sharing the same wording. The
// "Cannot access" substring covers the stream ("Cannot access stream ...") variants,
// "does not have permissions to call" the SQS variant ("The [provided] execution role does
// not have permissions to call ReceiveMessage on SQS"), and "ensure the role can perform"
// the guidance-styled variants.
//
// This is an eventual-consistency retry the AWS SDK's default retryer will NOT perform
// (InvalidParameterValueException is a 400-level validation error, outside the SDK's
// retryable set). Callers wrap CreateEventSourceMapping with RetryOnIAMPropagation, which
// retries in-call for the propagation window and then defers to the blueprint engine's
// retry policy for the idempotent intermediary update.
func IsRoleNotYetPropagatedError(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == "InvalidParameterValueException" {
		msg := apiErr.ErrorMessage()
		return strings.Contains(msg, "Cannot access") ||
			strings.Contains(msg, "does not have permissions to call") ||
			strings.Contains(msg, "ensure the role can perform")
	}
	return false
}

// IsESMInUseError reports whether an error is the transient failure AWS returns when a
// Lambda event source mapping is deleted (or updated) while it is still processing or
// mid-update ("Cannot delete the event source mapping because it is in use"). It is the
// canonical retryable on the ESM delete side, exactly as the role-not-yet-propagated
// validation failure is on the create side.
func IsESMInUseError(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == "ResourceInUseException" {
		msg := apiErr.ErrorMessage()
		return strings.Contains(msg, "in use") ||
			strings.Contains(msg, "event source mapping") ||
			strings.Contains(msg, "currently being updated")
	}
	return false
}

// IsLambdaFunctionConflictError reports whether an error is the transient failure AWS
// returns when a Lambda function configuration update is attempted while another update
// to the same function is still applying ("The operation cannot be performed at this
// time. An update is in progress ...") or while the function is still initialising right
// after creation ("The function is currently in the following state: Pending").
//
// Concurrent link updates that target the same function (e.g. two config-store links
// each injecting environment variables into one handler) serialize on this error:
// Lambda accepts one writer and rejects the rest with ResourceConflictException until
// the in-flight update completes.
func IsLambdaFunctionConflictError(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == "ResourceConflictException" {
		msg := apiErr.ErrorMessage()
		return strings.Contains(msg, "update is in progress") ||
			strings.Contains(msg, "cannot be performed at this time") ||
			strings.Contains(msg, "Pending")
	}
	return false
}

var transientRetryBackoff = []time.Duration{
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
	16 * time.Second,
	30 * time.Second,
	30 * time.Second,
	30 * time.Second,
}

// RetryOnIAMPropagation wraps a call that AWS validates against freshly-granted IAM
// permissions (e.g. CreateEventSourceMapping), retrying it in-call with backoff for
// roughly two minutes while it fails with the role-not-yet-propagated validation error.
// If the window is exhausted, the final error is surfaced as a provider.RetryableError
// so the blueprint engine's retry policy re-runs the idempotent link update as a last
// resort. Non-matching errors are returned immediately and unchanged.
func RetryOnIAMPropagation[Arg any, Value any](
	fn pluginutils.ContextFuncReturnValue[Arg, Value],
) pluginutils.ContextFuncReturnValue[Arg, Value] {
	return retryTransient(fn, IsRoleNotYetPropagatedError)
}

// RetryOnESMInUse wraps DeleteEventSourceMapping (or any call that can transiently
// fail while an event source mapping is in use), retrying it in-call with backoff for
// roughly two minutes. If the window is exhausted, the final error is surfaced as a
// provider.RetryableError so the blueprint engine's retry policy re-runs the
// idempotent link update as a last resort. Non-matching errors are returned
// immediately and unchanged.
func RetryOnESMInUse[Arg any, Value any](
	fn pluginutils.ContextFuncReturnValue[Arg, Value],
) pluginutils.ContextFuncReturnValue[Arg, Value] {
	return retryTransient(fn, IsESMInUseError)
}

func retryTransient[Arg any, Value any](
	fn pluginutils.ContextFuncReturnValue[Arg, Value],
	isTransient func(error) bool,
) pluginutils.ContextFuncReturnValue[Arg, Value] {
	return func(ctx context.Context, arg Arg) (Value, error) {
		value, err := fn(ctx, arg)
		if err == nil || !isTransient(err) {
			return value, err
		}

		for _, delay := range transientRetryBackoff {
			select {
			case <-ctx.Done():
				return value, ctx.Err()
			case <-time.After(delay):
			}

			value, err = fn(ctx, arg)
			if err == nil || !isTransient(err) {
				return value, err
			}
		}

		return value, &provider.RetryableError{ChildError: err}
	}
}
