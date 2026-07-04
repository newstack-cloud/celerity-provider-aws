//go:build unit

package linkutils

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type RetryErrorsSuite struct {
	suite.Suite
}

func apiError(code, message string) error {
	return &smithy.GenericAPIError{Code: code, Message: message}
}

func (s *RetryErrorsSuite) Test_IsRoleNotYetPropagatedError() {
	// Matching: the InvalidParameterValueException code AND a known message substring.
	s.True(IsRoleNotYetPropagatedError(apiError(
		"InvalidParameterValueException",
		"Cannot access stream arn:aws:kinesis:...; ensure the role can perform GetRecords",
	)))
	s.True(IsRoleNotYetPropagatedError(apiError(
		"InvalidParameterValueException",
		"Cannot access the resource",
	)))

	// Right message, wrong code → not this transient case.
	s.False(IsRoleNotYetPropagatedError(apiError("AccessDeniedException", "Cannot access stream")))
	// Right code, unrelated message → a genuinely bad parameter, not eventual consistency.
	s.False(IsRoleNotYetPropagatedError(apiError("InvalidParameterValueException", "BatchSize is invalid")))
	// Non-API error and nil.
	s.False(IsRoleNotYetPropagatedError(errors.New("boom")))
	s.False(IsRoleNotYetPropagatedError(nil))
}

func (s *RetryErrorsSuite) Test_isS3DestinationNotReadyError() {
	// Matching: the InvalidArgument code AND a known message substring.
	s.True(isS3DestinationNotReadyError(apiError(
		"InvalidArgument",
		"Unable to validate the following destination configurations",
	)))
	s.True(isS3DestinationNotReadyError(apiError(
		"InvalidArgument",
		"Permissions on the destination queue do not allow S3 to publish notifications",
	)))

	// Right message, wrong code.
	s.False(isS3DestinationNotReadyError(apiError("MalformedXML", "Unable to validate the following destination configurations")))
	// Right code, unrelated message.
	s.False(isS3DestinationNotReadyError(apiError("InvalidArgument", "Filter rule name must be prefix or suffix")))
	// Non-API error and nil.
	s.False(isS3DestinationNotReadyError(errors.New("boom")))
	s.False(isS3DestinationNotReadyError(nil))
}

// The retry wrapper turns a matching error into a provider.RetryableError (so the engine
// retries the idempotent operation) and leaves a non-matching error untouched (terminal).
func (s *RetryErrorsSuite) Test_wrapper_marks_only_matching_errors_retryable() {
	matching := apiError("InvalidParameterValueException", "Cannot access stream orders")
	retryable := pluginutils.Retryable(
		func(_ context.Context, _ any) error { return matching },
		IsRoleNotYetPropagatedError,
	)
	err := retryable(context.Background(), nil)
	var retryErr *provider.RetryableError
	s.Require().True(errors.As(err, &retryErr), "matching error should be wrapped as retryable")

	nonMatching := apiError("AccessDeniedException", "nope")
	terminal := pluginutils.Retryable(
		func(_ context.Context, _ any) error { return nonMatching },
		IsRoleNotYetPropagatedError,
	)
	err = terminal(context.Background(), nil)
	var notRetry *provider.RetryableError
	s.False(errors.As(err, &notRetry), "non-matching error should not be wrapped")
	s.ErrorIs(err, nonMatching, "the raw error should be returned unchanged")
}

func TestRetryErrorsSuite(t *testing.T) {
	suite.Run(t, new(RetryErrorsSuite))
}
