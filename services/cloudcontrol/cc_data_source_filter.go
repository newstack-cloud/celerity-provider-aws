package cloudcontrol

import (
	"slices"
	"strconv"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// regionFilterField is the pseudo-filter used to select the AWS region; it is not a
// resource property and is excluded from client-side matching.
const regionFilterField = "region"

// MatchesFilters reports whether a resource (its flattened export fields) satisfies every
// resolved filter. Cloud Control offers no server-side filtering, so the engine lists resources
// and evaluates the framework's filter operators here. All filters are combined with logical
// AND.
func MatchesFilters(
	filters *provider.ResolvedDataSourceFilters,
	fields map[string]*core.MappingNode,
) bool {
	if filters == nil {
		return true
	}
	for _, filter := range filters.Filters {
		if filter == nil {
			continue
		}

		field := core.StringValueFromScalar(filter.Field)
		if field == "" || field == regionFilterField {
			continue
		}

		if !matchesFilter(filter, fields[field]) {
			return false
		}
	}
	return true
}

func matchesFilter(
	filter *provider.ResolvedDataSourceFilter,
	value *core.MappingNode,
) bool {
	operator := pluginutils.GetDataSourceFilterOperator(filter)
	searchValues := pluginutils.GetDataSourceFilterSearchValues(filter)

	present := value != nil
	switch operator {
	case schema.DataSourceFilterOperatorEquals:
		return present && anyMatch(value, searchValues, scalarEqual)
	case schema.DataSourceFilterOperatorNotEquals:
		return !present || !anyMatch(value, searchValues, scalarEqual)
	case schema.DataSourceFilterOperatorIn:
		return present && anyMatch(value, searchValues, scalarEqual)
	case schema.DataSourceFilterOperatorNotIn:
		return !present || !anyMatch(value, searchValues, scalarEqual)
	case schema.DataSourceFilterOperatorContains:
		return present && anyMatch(value, searchValues, nodeContains)
	case schema.DataSourceFilterOperatorNotContains:
		return !present || !anyMatch(value, searchValues, nodeContains)
	case schema.DataSourceFilterOperatorStartsWith:
		return present && anyMatch(value, searchValues, nodeStartsWith)
	case schema.DataSourceFilterOperatorNotStartsWith:
		return !present || !anyMatch(value, searchValues, nodeStartsWith)
	case schema.DataSourceFilterOperatorEndsWith:
		return present && anyMatch(value, searchValues, nodeEndsWith)
	case schema.DataSourceFilterOperatorNotEndsWith:
		return !present || !anyMatch(value, searchValues, nodeEndsWith)
	case schema.DataSourceFilterOperatorHasKey:
		return nodeHasKey(value, searchValues)
	case schema.DataSourceFilterOperatorNotHasKey:
		return !nodeHasKey(value, searchValues)
	case schema.DataSourceFilterOperatorGreaterThan:
		return present && compareNumeric(value, searchValues, numGreater)
	case schema.DataSourceFilterOperatorGreaterThanOrEqual:
		return present && compareNumeric(value, searchValues, numGreaterOrEqual)
	case schema.DataSourceFilterOperatorLessThan:
		return present && compareNumeric(value, searchValues, numLess)
	case schema.DataSourceFilterOperatorLessThanOrEqual:
		return present && compareNumeric(value, searchValues, numLessOrEqual)
	default:
		return false
	}
}

func anySearch(
	searchValues []*core.MappingNode,
	predicate func(*core.MappingNode) bool,
) bool {
	return slices.ContainsFunc(searchValues, predicate)
}

// anyMatch reports whether value matches any of the search values under the given
// binary predicate (value held fixed, each search value tried in turn).
func anyMatch(
	value *core.MappingNode,
	searchValues []*core.MappingNode,
	match func(value, search *core.MappingNode) bool,
) bool {
	return anySearch(searchValues, func(s *core.MappingNode) bool {
		return match(value, s)
	})
}

func scalarEqual(value, search *core.MappingNode) bool {
	return nodeString(value) == nodeString(search)
}

func nodeStartsWith(value, search *core.MappingNode) bool {
	return strings.HasPrefix(nodeString(value), nodeString(search))
}

func nodeEndsWith(value, search *core.MappingNode) bool {
	return strings.HasSuffix(nodeString(value), nodeString(search))
}

func numGreater(a, b float64) bool        { return a > b }
func numGreaterOrEqual(a, b float64) bool { return a >= b }
func numLess(a, b float64) bool           { return a < b }
func numLessOrEqual(a, b float64) bool    { return a <= b }

// Reports whether value contains search.
// Will test for a substring for scalar strings, or membership for
// an array of scalars.
func nodeContains(value, search *core.MappingNode) bool {
	if value != nil && value.Items != nil {
		for _, item := range value.Items {
			if nodeString(item) == nodeString(search) {
				return true
			}
		}
		return false
	}
	return strings.Contains(nodeString(value), nodeString(search))
}

func nodeHasKey(value *core.MappingNode, searchValues []*core.MappingNode) bool {
	if value == nil || value.Fields == nil {
		return false
	}
	return anySearch(searchValues, func(s *core.MappingNode) bool {
		_, ok := value.Fields[nodeString(s)]
		return ok
	})
}

func compareNumeric(value *core.MappingNode, searchValues []*core.MappingNode, cmp func(a, b float64) bool) bool {
	left, ok := nodeFloat(value)
	if !ok {
		return false
	}
	return anySearch(searchValues, func(s *core.MappingNode) bool {
		right, ok := nodeFloat(s)
		return ok && cmp(left, right)
	})
}

// Renders a scalar mapping node to its canonical string form.
func nodeString(node *core.MappingNode) string {
	if node == nil || node.Scalar == nil {
		return ""
	}

	switch {
	case node.Scalar.StringValue != nil:
		return *node.Scalar.StringValue
	case node.Scalar.IntValue != nil:
		return strconv.Itoa(*node.Scalar.IntValue)
	case node.Scalar.FloatValue != nil:
		return strconv.FormatFloat(*node.Scalar.FloatValue, 'f', -1, 64)
	case node.Scalar.BoolValue != nil:
		return strconv.FormatBool(*node.Scalar.BoolValue)
	default:
		return ""
	}
}

// Extracts a numeric value from a scalar mapping node (int, float, or a numeric string).
func nodeFloat(node *core.MappingNode) (float64, bool) {
	if node == nil || node.Scalar == nil {
		return 0, false
	}
	switch {
	case node.Scalar.IntValue != nil:
		return float64(*node.Scalar.IntValue), true
	case node.Scalar.FloatValue != nil:
		return *node.Scalar.FloatValue, true
	case node.Scalar.StringValue != nil:
		parsed, err := strconv.ParseFloat(*node.Scalar.StringValue, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}
