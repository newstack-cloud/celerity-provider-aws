//go:build integration

package integration

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/stretchr/testify/require"
)

// EqualsFilter builds a single-Equals filter set (the data-source fast path when
// the field is the primary identifier or a derivable ARN).
func EqualsFilter(field, value string) *provider.ResolvedDataSourceFilters {
	return Filters(Filter(field, schema.DataSourceFilterOperatorEquals, core.MappingNodeFromString(value)))
}

// Filter builds a single resolved data-source filter.
func Filter(
	field string,
	op schema.DataSourceFilterOperator,
	values ...*core.MappingNode,
) *provider.ResolvedDataSourceFilter {
	return &provider.ResolvedDataSourceFilter{
		Field:    core.ScalarFromString(field),
		Operator: &schema.DataSourceFilterOperatorWrapper{Value: op},
		Search:   &provider.ResolvedDataSourceFilterSearch{Values: values},
	}
}

// Filters combines filters into a resolved filter set (AND-combined by the engine).
func Filters(items ...*provider.ResolvedDataSourceFilter) *provider.ResolvedDataSourceFilters {
	return &provider.ResolvedDataSourceFilters{Filters: items}
}

// EqFilter builds a single string Equals filter (for combining into a multi-field
// filter set with Filters).
func EqFilter(field, value string) *provider.ResolvedDataSourceFilter {
	return Filter(field, schema.DataSourceFilterOperatorEquals, core.MappingNodeFromString(value))
}

// Fetch looks up a data source by type from the provider and fetches it with the
// given filters, returning the exported field map.
func (h *Harness) Fetch(
	t *testing.T,
	dataSourceType string,
	filters *provider.ResolvedDataSourceFilters,
) map[string]*core.MappingNode {
	t.Helper()

	dataSource, err := h.Provider.DataSource(h.Ctx, dataSourceType)
	require.NoError(t, err)

	output, err := dataSource.Fetch(h.Ctx, &provider.DataSourceFetchInput{
		ProviderContext:            h.ProviderCtx,
		DataSourceWithResolvedSubs: &provider.ResolvedDataSource{Filter: filters},
	})
	require.NoErrorf(t, err, "failed to fetch %s", dataSourceType)
	require.NotNil(t, output)
	return output.Data
}
