package cloudcontrol

import (
	"errors"
	"fmt"
	"strings"

	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// The Cloud Control / CloudFormation API error codes that
// represent transient throttling or concurrency contention and are safe to retry.
var retryableCCErrorCodes = map[string]bool{
	"ThrottlingException":             true,
	"RequestLimitExceeded":            true,
	"TooManyRequestsException":        true,
	"ConcurrentOperationException":    true,
	"ConcurrentModificationException": true,
	"ServiceInternalError":            true,
	"ServiceInternalErrorException":   true,
}

// Fragments of the status messages Cloud Control reports when an operation failed only
// because IAM has not finished propagating a role the blueprint created moments earlier.
//
// These arrive on a failed progress event rather than as an API error, so the code-based
// classification above never sees them. A blueprint that creates a role and a resource
// that assumes it is the ordinary case, and whether it succeeds comes down to whether
// anything slow happened to be deployed in between.
//
// Matching on message text is unpleasant and deliberate: the failure carries the
// downstream service's own error code (InvalidRequest for Lambda), which is far too broad
// to retry on, and the propagation delay is reported nowhere else.
var retryableCCStatusMessageFragments = []string{
	"cannot be assumed by",
	"is not authorized to perform: sts:AssumeRole",
	"Cannot access",
	"does not have permissions to call",
	"ensure the role can perform",
}

// IsCCPropagationFailure reports whether a failed Cloud Control operation failed for a
// reason that clears on its own, so the deployment can be retried rather than abandoned.
func IsCCPropagationFailure(statusMessage string) bool {
	for _, fragment := range retryableCCStatusMessageFragments {
		if strings.Contains(statusMessage, fragment) {
			return true
		}
	}

	return false
}

// The error for a failed Cloud Control operation, marked retryable when the failure is
// one that clears on its own.
//
// A failed operation arrives as a progress event rather than an API error, so returning a
// plain error here abandons the whole deployment for a condition that would have passed
// on the next attempt. Wrapping it as a provider.RetryableError hands it to the
// provider's retry policy, which is what already carries the deployment through the same
// class of failure elsewhere.
func ccOperationFailedError(
	cfnType string,
	errorCode cctypes.HandlerErrorCode,
	statusMessage string,
) error {
	err := fmt.Errorf(
		"cloud control operation for %s failed (%s): %s",
		cfnType,
		errorCode,
		statusMessage,
	)

	if IsCCPropagationFailure(statusMessage) {
		return &provider.RetryableError{ChildError: err}
	}

	return err
}

// Classifies a Cloud Control API error as retryable, matching on the smithy API error
// code so it is decoupled from concrete generated type names.
func isCCErrorRetryable(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		return retryableCCErrorCodes[apiErr.ErrorCode()]
	}
	return false
}
