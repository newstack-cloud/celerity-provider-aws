//go:build integration

package e2e

import (
	"path/filepath"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/changes"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/require"
)

// Stage loads the named blueprintlang file and runs change staging (no deploy),
// returning the staged change set. Data sources are resolved against real AWS
// during change staging, so a blueprint that declares a data source (plus the
// minimum one placeholder resource the framework requires) yields the resolved
// data source values without creating any infrastructure. The placeholder is
// never deployed.
func (h *Harness) Stage(
	t *testing.T,
	blueprintFile string,
	vars map[string]*core.ScalarValue,
) *changes.BlueprintChanges {
	t.Helper()
	params := h.params(vars)
	bp, err := h.Loader.Load(h.Ctx, filepath.Join("testdata", blueprintFile), params)
	require.NoErrorf(t, err, "load blueprint %s", blueprintFile)
	return h.stage(t, bp, "it-stage-"+uniqueSuffix(), params, false)
}

// ResolvedExport returns the resolved value of a blueprint export from a staged
// change set (used to read data source outputs surfaced via exports during
// change staging).
func ResolvedExport(t *testing.T, changeSet *changes.BlueprintChanges, name string) *core.MappingNode {
	t.Helper()
	fieldChange, ok := changeSet.NewExports[name]
	require.Truef(t, ok, "export %q not found in staged changes", name)
	return fieldChange.NewValue
}
