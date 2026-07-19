//go:build unit

package linkutils

import (
	"context"
	"errors"
	"testing"
	"time"

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
	// The SQS variants observed at CreateEventSourceMapping time.
	s.True(IsRoleNotYetPropagatedError(apiError(
		"InvalidParameterValueException",
		"The provided execution role does not have permissions to call ReceiveMessage on SQS",
	)))
	s.True(IsRoleNotYetPropagatedError(apiError(
		"InvalidParameterValueException",
		"The function execution role does not have permissions to call ReceiveMessage on SQS",
	)))

	// Right message, wrong code → not this transient case.
	s.False(IsRoleNotYetPropagatedError(apiError("AccessDeniedException", "Cannot access stream")))
	// Right code, unrelated message → a genuinely bad parameter, not eventual consistency.
	s.False(IsRoleNotYetPropagatedError(apiError("InvalidParameterValueException", "BatchSize is invalid")))
	// Non-API error and nil.
	s.False(IsRoleNotYetPropagatedError(errors.New("boom")))
	s.False(IsRoleNotYetPropagatedError(nil))
}

func (s *RetryErrorsSuite) Test_IsESMInUseError() {
	// Matching: the ResourceInUseException code AND a known message substring.
	s.True(IsESMInUseError(apiError(
		"ResourceInUseException",
		"Cannot delete the event source mapping because it is in use",
	)))
	s.True(IsESMInUseError(apiError(
		"ResourceInUseException",
		"The operation cannot be performed at this time. The event source mapping is currently being updated.",
	)))

	// Right message, wrong code → not this transient case.
	s.False(IsESMInUseError(apiError("InvalidParameterValueException", "it is in use")))
	// Right code, unrelated message.
	s.False(IsESMInUseError(apiError("ResourceInUseException", "Function is being deleted")))
	// Non-API error and nil.
	s.False(IsESMInUseError(errors.New("boom")))
	s.False(IsESMInUseError(nil))
}

func (s *RetryErrorsSuite) Test_IsLambdaFunctionConflictError() {
	// Matching: the ResourceConflictException code AND a known message substring.
	// The mid-update variant observed when two links update the same function.
	s.True(IsLambdaFunctionConflictError(apiError(
		"ResourceConflictException",
		"The operation cannot be performed at this time. An update is in progress for resource: arn:aws:lambda:eu-west-2:123456789012:function:orders-api",
	)))
	// The just-created variant, before the function reaches the Active state.
	s.True(IsLambdaFunctionConflictError(apiError(
		"ResourceConflictException",
		"The function is currently in the following state: Pending",
	)))

	// Right message, wrong code → not this transient case.
	s.False(IsLambdaFunctionConflictError(apiError("InvalidParameterValueException", "An update is in progress")))
	// Right code, unrelated message → e.g. a concurrent delete, not a transient update conflict.
	s.False(IsLambdaFunctionConflictError(apiError("ResourceConflictException", "Function is being deleted")))
	// Non-API error and nil.
	s.False(IsLambdaFunctionConflictError(errors.New("boom")))
	s.False(IsLambdaFunctionConflictError(nil))
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

func (s *RetryErrorsSuite) withFastBackoff(run func()) {
	original := transientRetryBackoff
	transientRetryBackoff = []time.Duration{time.Millisecond, time.Millisecond}
	defer func() { transientRetryBackoff = original }()
	run()
}

func (s *RetryErrorsSuite) Test_RetryOnIAMPropagation_retries_until_propagated() {
	s.withFastBackoff(func() {
		propagationErr := apiError(
			"InvalidParameterValueException",
			"The provided execution role does not have permissions to call ReceiveMessage on SQS",
		)
		attempts := 0
		call := RetryOnIAMPropagation(func(_ context.Context, _ any) (string, error) {
			attempts++
			if attempts < 3 {
				return "", propagationErr
			}
			return "esm-uuid", nil
		})

		value, err := call(context.Background(), nil)
		s.Require().NoError(err)
		s.Equal("esm-uuid", value)
		s.Equal(3, attempts)
	})
}

func (s *RetryErrorsSuite) Test_RetryOnIAMPropagation_returns_non_matching_errors_immediately() {
	s.withFastBackoff(func() {
		terminal := apiError("InvalidParameterValueException", "BatchSize is invalid")
		attempts := 0
		call := RetryOnIAMPropagation(func(_ context.Context, _ any) (string, error) {
			attempts++
			return "", terminal
		})

		_, err := call(context.Background(), nil)
		s.ErrorIs(err, terminal)
		s.Equal(1, attempts, "a genuinely bad parameter should not be retried")
	})
}

func (s *RetryErrorsSuite) Test_RetryOnIAMPropagation_defers_to_engine_when_window_exhausted() {
	s.withFastBackoff(func() {
		propagationErr := apiError(
			"InvalidParameterValueException",
			"Cannot access stream arn:aws:kinesis:...; ensure the role can perform GetRecords",
		)
		attempts := 0
		call := RetryOnIAMPropagation(func(_ context.Context, _ any) (string, error) {
			attempts++
			return "", propagationErr
		})

		_, err := call(context.Background(), nil)
		var retryErr *provider.RetryableError
		s.Require().True(
			errors.As(err, &retryErr),
			"an exhausted in-call window should defer to the engine's retry policy",
		)
		s.ErrorIs(retryErr.ChildError, propagationErr)
		s.Equal(3, attempts, "initial attempt plus one per backoff step")
	})
}

func (s *RetryErrorsSuite) Test_RetryOnESMInUse_retries_until_mapping_released() {
	s.withFastBackoff(func() {
		inUseErr := apiError(
			"ResourceInUseException",
			"Cannot delete the event source mapping because it is in use",
		)
		attempts := 0
		call := RetryOnESMInUse(func(_ context.Context, _ any) (string, error) {
			attempts++
			if attempts < 3 {
				return "", inUseErr
			}
			return "deleted", nil
		})

		value, err := call(context.Background(), nil)
		s.Require().NoError(err)
		s.Equal("deleted", value)
		s.Equal(3, attempts)
	})
}

func (s *RetryErrorsSuite) Test_RetryOnESMInUse_defers_to_engine_when_window_exhausted() {
	s.withFastBackoff(func() {
		inUseErr := apiError(
			"ResourceInUseException",
			"Cannot delete the event source mapping because it is in use",
		)
		call := RetryOnESMInUse(func(_ context.Context, _ any) (string, error) {
			return "", inUseErr
		})

		_, err := call(context.Background(), nil)
		var retryErr *provider.RetryableError
		s.Require().True(
			errors.As(err, &retryErr),
			"an exhausted in-call window should defer to the engine's retry policy",
		)
		s.ErrorIs(retryErr.ChildError, inUseErr)
	})
}

func (s *RetryErrorsSuite) Test_RetryOnIAMPropagation_stops_waiting_on_context_cancellation() {
	propagationErr := apiError(
		"InvalidParameterValueException",
		"Cannot access stream arn:aws:kinesis:...",
	)
	call := RetryOnIAMPropagation(func(_ context.Context, _ any) (string, error) {
		return "", propagationErr
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := call(ctx, nil)
	s.ErrorIs(err, context.Canceled)
}

func TestRetryErrorsSuite(t *testing.T) {
	suite.Run(t, new(RetryErrorsSuite))
}
