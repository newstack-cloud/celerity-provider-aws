package cloudcontrol

import (
	"errors"

	"github.com/aws/smithy-go"
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
