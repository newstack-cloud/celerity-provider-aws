//go:build unit

package cloudcontrol

import (
	"testing"

	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/require"
)

// A blueprint that creates an IAM role and a resource that assumes it is the ordinary
// case, and IAM takes a moment to propagate. Whether the deployment survived that used to
// come down to whether something slow happened to be deployed in between: a VPC gave IAM
// several minutes, a queue gave it none.
//
// The failure arrives as a progress event rather than an API error, so the error-code
// classification never saw it and the whole deployment was abandoned for a condition that
// passes on the next attempt.
func TestCCOperationFailedErrorIsRetryableForRolePropagation(t *testing.T) {
	err := ccOperationFailedError(
		"AWS::Lambda::Function",
		cctypes.HandlerErrorCodeInvalidRequest,
		"The role defined for the function cannot be assumed by Lambda. "+
			"(Service: Lambda, Status Code: 400)",
	)

	var retryable *provider.RetryableError
	require.ErrorAs(t, err, &retryable)
	require.Contains(t, retryable.ChildError.Error(), "AWS::Lambda::Function")
}

func TestCCOperationFailedErrorIsRetryableForAssumeRoleDenial(t *testing.T) {
	err := ccOperationFailedError(
		"AWS::Lambda::Function",
		cctypes.HandlerErrorCodeAccessDenied,
		"User is not authorized to perform: sts:AssumeRole on resource",
	)

	var retryable *provider.RetryableError
	require.ErrorAs(t, err, &retryable)
}

// The downstream service's error code is far too broad to retry on by itself. A genuinely
// invalid request carries the same InvalidRequest code as the propagation failure, and
// retrying it would turn an immediate, accurate error into a slow one.
func TestCCOperationFailedErrorIsNotRetryableForAGenuineFailure(t *testing.T) {
	err := ccOperationFailedError(
		"AWS::Lambda::Function",
		cctypes.HandlerErrorCodeInvalidRequest,
		"The runtime parameter of nodejs10.x is no longer supported",
	)

	require.Error(t, err)
	var retryable *provider.RetryableError
	require.False(
		t,
		asRetryable(err, &retryable),
		"a permanently invalid request must fail immediately rather than be retried",
	)
}

func asRetryable(err error, target **provider.RetryableError) bool {
	retryable, ok := err.(*provider.RetryableError)
	if ok {
		*target = retryable
	}

	return ok
}
