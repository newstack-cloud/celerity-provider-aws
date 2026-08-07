//go:build unit

package provider

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/require"
)

// The provider must declare its own retry policy rather than inherit the framework
// default, which is sized for an API that reaches consistency quickly.
//
// AWS does not. The measurement that set this budget: the same teardown released the
// network interfaces holding a security group after 18 seconds on one run and 977 on
// the next. The default's five retries cover roughly four minutes end to end, so a
// destroy that would have succeeded gets reported as failed.
func TestProviderDeclaresARetryBudgetThatCoversAWSConsistencyDelays(t *testing.T) {
	require.NotNil(t, awsRetryPolicy, "the provider must not fall back to the framework default")

	// Backoff budget, excluding the work each attempt does. Enough to outlast the
	// slowest release actually observed.
	const slowestObservedReleaseSeconds = 977.0
	require.Greater(
		t,
		backoffBudgetSeconds(awsRetryPolicy),
		slowestObservedReleaseSeconds,
		"the retry budget must outlast the slowest AWS release measured against a real account",
	)

	// A short first delay keeps the common case fast: most retryable failures here
	// clear in seconds, and reaching the long tail sooner would penalise all of them.
	require.LessOrEqual(
		t,
		awsRetryPolicy.FirstRetryDelay,
		5.0,
		"a long first delay would slow down the short-lived failures that dominate",
	)
	require.Greater(t, awsRetryPolicy.BackoffFactor, 1.0)
	require.True(t, awsRetryPolicy.Jitter, "concurrent retries should not converge on the same instants")
}

// Total seconds spent waiting between attempts for a policy that exhausts its retries.
func backoffBudgetSeconds(policy *provider.RetryPolicy) float64 {
	total := 0.0
	delay := policy.FirstRetryDelay
	for i := 0; i < policy.MaxRetries; i++ {
		if policy.MaxDelay > 0 && delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
		total += delay
		delay *= policy.BackoffFactor
	}

	return total
}
