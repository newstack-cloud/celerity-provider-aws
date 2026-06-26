//go:build integration

package e2e

import (
	"fmt"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/require"
)

// TestResources deploys each resource blueprint end-to-end through the framework
// container against real AWS and asserts the computed fields surfaced via blueprint
// exports and the deployed resource state. Teardown (and thus the destroy path) is
// handled by the harness. Each subtest runs in parallel with its own harness (a unique
// name prefix and isolated state container); concurrency against AWS is bounded by the
// harness operation gate.
func TestResources(t *testing.T) {
	t.Parallel()

	t.Run("sqs queue", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		inst := h.Deploy(t, "queue_resource.blueprint", nil)
		expectedARN := fmt.Sprintf("arn:aws:sqs:%s:%s:%s-queue", h.Region, h.Account.AccountID, h.NamePrefix)
		require.Equal(t, expectedARN, core.StringValue(inst.Export(t, "queueArn")))
		require.Contains(t, core.StringValue(inst.Export(t, "queueUrl")), h.Account.AccountID)
		require.Equal(t, expectedARN, core.StringValue(inst.ResourceSpec(t, "queue").Fields["arn"]))
	})

	t.Run("dynamodb table", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		inst := h.Deploy(t, "table_resource.blueprint", nil)
		expectedARN := fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/%s-table", h.Region, h.Account.AccountID, h.NamePrefix)
		require.Equal(t, expectedARN, core.StringValue(inst.Export(t, "tableArn")))
		require.NotEmpty(t, core.StringValue(inst.Export(t, "streamArn")), "expected a stream ARN")
	})

	t.Run("iam role", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		inst := h.Deploy(t, "role_resource.blueprint", nil)
		expectedARN := fmt.Sprintf("arn:aws:iam::%s:role/%s-role", h.Account.AccountID, h.NamePrefix)
		require.Equal(t, expectedARN, core.StringValue(inst.Export(t, "roleArn")))
	})

	t.Run("lambda function", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		inst := h.Deploy(t, "function_resource.blueprint", nil)
		expectedARN := fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s-fn", h.Region, h.Account.AccountID, h.NamePrefix)
		require.Equal(t, expectedARN, core.StringValue(inst.Export(t, "functionArn")))
	})
}
