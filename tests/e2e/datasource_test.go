//go:build integration

package e2e

import (
	"fmt"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/require"
)

// TestDataSources deploys a real resource, then stages a blueprint containing a
// data source that reads it back. Data sources are resolved during change staging,
// so the resolved values surface as the staged values of exports that reference
// them. A second deployment is not needed and the staging placeholder resource is
// never created. Each subtest runs in parallel with its own isolated harness;
// concurrency against AWS is bounded by the harness operation gate.
func TestDataSources(t *testing.T) {
	t.Parallel()

	t.Run("sqs queue", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "queue_resource.blueprint", nil)
		changeSet := h.Stage(t, "queue_datasource.blueprint", nil)

		expectedARN := fmt.Sprintf("arn:aws:sqs:%s:%s:%s-queue", h.Region, h.Account.AccountID, h.NamePrefix)
		require.Equal(t, expectedARN, core.StringValue(ResolvedExport(t, changeSet, "queueArn")))
		require.Contains(t, core.StringValue(ResolvedExport(t, changeSet, "queueUrl")), h.Account.AccountID)
	})

	t.Run("dynamodb table", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "table_resource.blueprint", nil)
		changeSet := h.Stage(t, "table_datasource.blueprint", nil)

		expectedARN := fmt.Sprintf("arn:aws:dynamodb:%s:%s:table/%s-table", h.Region, h.Account.AccountID, h.NamePrefix)
		require.Equal(t, expectedARN, core.StringValue(ResolvedExport(t, changeSet, "tableArn")))
	})
}
