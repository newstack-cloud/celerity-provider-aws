//go:build integration

package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/require"
)

// TestLinks deploys each link blueprint end-to-end. The framework establishes the
// link (label selector) and runs the link's resource/intermediary updates during
// deploy, so the real AWS side effects are asserted via the raw AWS APIs after a
// successful deploy. Teardown (including the link destroy path) is handled by the
// harness. Each subtest runs in parallel with its own harness (a unique name prefix
// and isolated state container); concurrency against AWS is bounded by the harness
// operation gate.
func TestLinks(t *testing.T) {
	t.Parallel()

	// 1. Direct resource attribute: SQS queue -> queue (dead-letter redrive).
	t.Run("queue to queue dlq", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "link_queue_dlq.blueprint", nil)
		dlqARN := fmt.Sprintf("arn:aws:sqs:%s:%s:%s-dlq-target", h.Region, h.Account.AccountID, h.NamePrefix)
		attrs := h.QueueAttributes(t, h.NamePrefix+"-dlq-src")
		redrive := map[string]any{}
		require.NoError(t, json.Unmarshal([]byte(attrs["RedrivePolicy"]), &redrive))
		require.Equal(t, dlqARN, redrive["deadLetterTargetArn"])
		require.EqualValues(t, 5, redrive["maxReceiveCount"])
	})

	// 2. Managed intermediary via the engine's ResourceService: rule -> SQS queue.
	t.Run("rule to queue", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "link_rule_queue.blueprint", nil)
		ruleARN := fmt.Sprintf("arn:aws:events:%s:%s:rule/%s-rq-rule", h.Region, h.Account.AccountID, h.NamePrefix)
		policy := CompactJSON(t, h.QueueAttributes(t, h.NamePrefix+"-rq-queue")["Policy"])
		require.Contains(t, policy, "events.amazonaws.com")
		require.Contains(t, policy, "sqs:SendMessage")
		require.Contains(t, policy, ruleARN)
	})

	// 3. Managed intermediary (Lambda permission): rule -> Lambda function.
	t.Run("rule to function", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "link_rule_function.blueprint", nil)
		ruleARN := fmt.Sprintf("arn:aws:events:%s:%s:rule/%s-rf-rule", h.Region, h.Account.AccountID, h.NamePrefix)
		policy, ok := h.FunctionResourcePolicy(t, h.NamePrefix+"-rf-fn")
		require.True(t, ok, "expected a resource policy granting EventBridge invoke")
		require.Contains(t, policy, "events.amazonaws.com")
		require.Contains(t, policy, ruleARN)
	})

	// 4. Existing-role allocator + env vars: Lambda function -> DynamoDB table.
	t.Run("function to table", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "link_function_table.blueprint", nil)
		tableARN := fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/%s-ft-table", h.Region, h.Account.AccountID, h.NamePrefix)
		policy, ok := h.RoleInlinePolicy(t, h.NamePrefix+"-ft-role")
		require.True(t, ok, "expected the shared inline access policy")
		require.Contains(t, policy, "dynamodb:")
		require.Contains(t, policy, tableARN)
		require.Contains(t, h.FunctionEnvValues(t, h.NamePrefix+"-ft-fn"), h.NamePrefix+"-ft-table")
	})

	// 5. Stream trigger (managed event source mapping + role grant): table -> function.
	t.Run("table to function", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		inst := h.Deploy(t, "link_table_function.blueprint", nil)
		streamARN := core.StringValue(inst.Export(t, "streamArn"))
		require.NotEmpty(t, streamARN)
		policy, ok := h.RoleInlinePolicy(t, h.NamePrefix+"-tf-role")
		require.True(t, ok, "expected the shared inline access policy")
		require.Contains(t, policy, "dynamodb:GetRecords")
		require.Contains(t, policy, streamARN)
		mappings := h.FunctionEventSourceMappings(t, h.NamePrefix+"-tf-fn")
		found := false
		for _, m := range mappings {
			if aws.ToString(m.EventSourceArn) == streamARN {
				found = true
			}
		}
		require.True(t, found, "expected an event source mapping for the table stream")
	})

	// 6. Env vars + existing-role grant: Lambda function -> EventBridge event bus.
	t.Run("function to event bus", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "link_function_eventbus.blueprint", nil)
		busARN := fmt.Sprintf("arn:aws:events:%s:%s:event-bus/%s-feb-bus", h.Region, h.Account.AccountID, h.NamePrefix)
		policy, ok := h.RoleInlinePolicy(t, h.NamePrefix+"-feb-role")
		require.True(t, ok, "expected the shared inline access policy")
		require.Contains(t, policy, "events:PutEvents")
		require.Contains(t, policy, busARN)
		require.Contains(t, h.FunctionEnvValues(t, h.NamePrefix+"-feb-fn"), h.NamePrefix+"-feb-bus")
	})

	// 7. Lambda invoke (env vars + existing-role grant): function -> function.
	t.Run("function to function", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "link_function_function.blueprint", nil)
		calleeARN := fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s-ff-callee", h.Region, h.Account.AccountID, h.NamePrefix)
		policy, ok := h.RoleInlinePolicy(t, h.NamePrefix+"-ff-role")
		require.True(t, ok, "expected the shared inline access policy")
		require.Contains(t, policy, "lambda:InvokeFunction")
		require.Contains(t, policy, calleeARN)
		require.Contains(t, h.FunctionEnvValues(t, h.NamePrefix+"-ff-caller"), calleeARN)
	})

	// 8. Existing-role allocator: EventBridge rule -> API destination.
	t.Run("rule to api destination", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		inst := h.Deploy(t, "link_rule_apidestination.blueprint", nil)
		destARN := core.StringValue(inst.Export(t, "destArn"))
		policy, ok := h.RoleInlinePolicy(t, h.NamePrefix+"-apidest-role")
		require.True(t, ok, "expected the shared inline access policy")
		require.Contains(t, policy, "events:InvokeApiDestination")
		require.Contains(t, policy, destARN)
	})

	// 9. Code signing: Lambda function -> code signing config (the CSC needs a real
	// Signer profile version ARN, created via SDK fixture and passed as a variable).
	t.Run("function to code signing config", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		profileVersionARN := h.SigningProfileVersionARN(t, "csc-link-profile")
		inst := h.Deploy(t, "link_function_codesigningconfig.blueprint", map[string]*core.ScalarValue{
			"signingProfileVersionArn": core.ScalarFromString(profileVersionARN),
		})
		cscARN := core.StringValue(inst.Export(t, "cscArn"))
		require.NotEmpty(t, cscARN)
		require.Equal(t, cscARN, h.FunctionCodeSigningConfigArn(t, h.NamePrefix+"-csc-fn"))
	})
}
