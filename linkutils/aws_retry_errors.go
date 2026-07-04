package linkutils

import (
	"errors"
	"strings"

	"github.com/aws/smithy-go"
)

// IsRoleNotYetPropagatedError reports whether an error is the transient validation failure
// AWS returns when a Lambda event source mapping is created before the execution role's
// freshly-granted stream/queue read permissions have propagated (IAM eventual consistency).
//
// It anchors on the smithy API error code (InvalidParameterValueException) as well as a
// message substring so it is decoupled from unrelated errors sharing the same wording. The
// broader "Cannot access" substring covers the stream ("Cannot access stream ...") and queue
// variants of the message.
//
// This is an eventual-consistency retry the AWS SDK's default retryer will NOT perform
// (InvalidParameterValueException is a 400-level validation error, outside the SDK's
// retryable set). Callers wrap CreateEventSourceMapping with pluginutils.Retryable so the
// blueprint engine re-runs the idempotent intermediary update, re-asserting the role grant
// before retrying.
func IsRoleNotYetPropagatedError(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == "InvalidParameterValueException" {
		msg := apiErr.ErrorMessage()
		return strings.Contains(msg, "Cannot access") ||
			strings.Contains(msg, "ensure the role can perform")
	}
	return false
}
