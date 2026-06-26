//go:build unit

package cloudcontrol

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/stretchr/testify/suite"
)

type CCDataSourceFilterSuite struct {
	suite.Suite
}

func intNode(value int) *core.MappingNode {
	return &core.MappingNode{Scalar: &core.ScalarValue{IntValue: &value}}
}

func filter(field string, op schema.DataSourceFilterOperator, values ...*core.MappingNode) *provider.ResolvedDataSourceFilter {
	return &provider.ResolvedDataSourceFilter{
		Field:    core.ScalarFromString(field),
		Operator: &schema.DataSourceFilterOperatorWrapper{Value: op},
		Search:   &provider.ResolvedDataSourceFilterSearch{Values: values},
	}
}

func filters(items ...*provider.ResolvedDataSourceFilter) *provider.ResolvedDataSourceFilters {
	return &provider.ResolvedDataSourceFilters{Filters: items}
}

func (s *CCDataSourceFilterSuite) fields() map[string]*core.MappingNode {
	return map[string]*core.MappingNode{
		"tableName":   core.MappingNodeFromString("orders-table"),
		"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
		"readUnits":   intNode(10),
		"tags": {Items: []*core.MappingNode{
			core.MappingNodeFromString("prod"),
			core.MappingNodeFromString("orders"),
		}},
		"streamSpec": {Fields: map[string]*core.MappingNode{
			"streamViewType": core.MappingNodeFromString("NEW_IMAGE"),
		}},
	}
}

func (s *CCDataSourceFilterSuite) assertMatch(f *provider.ResolvedDataSourceFilters, want bool) {
	s.Equal(want, MatchesFilters(f, s.fields()))
}

func (s *CCDataSourceFilterSuite) Test_equals_and_not_equals() {
	s.assertMatch(filters(filter(
		"tableName",
		schema.DataSourceFilterOperatorEquals,
		core.MappingNodeFromString("orders-table"),
	)), true)
	s.assertMatch(filters(filter(
		"tableName",
		schema.DataSourceFilterOperatorEquals,
		core.MappingNodeFromString("other"))), false)
	s.assertMatch(filters(filter(
		"tableName",
		schema.DataSourceFilterOperatorNotEquals,
		core.MappingNodeFromString("other"),
	)), true)
	s.assertMatch(filters(filter(
		"tableName",
		schema.DataSourceFilterOperatorNotEquals,
		core.MappingNodeFromString("orders-table"),
	)), false)
}

func (s *CCDataSourceFilterSuite) Test_in_and_not_in() {
	s.assertMatch(filters(filter("billingMode", schema.DataSourceFilterOperatorIn,
		core.MappingNodeFromString("PROVISIONED"), core.MappingNodeFromString("PAY_PER_REQUEST"))), true)
	s.assertMatch(filters(filter("billingMode", schema.DataSourceFilterOperatorNotIn,
		core.MappingNodeFromString("PROVISIONED"))), true)
	s.assertMatch(filters(filter("billingMode", schema.DataSourceFilterOperatorNotIn,
		core.MappingNodeFromString("PAY_PER_REQUEST"))), false)
}

func (s *CCDataSourceFilterSuite) Test_contains_string_and_array() {
	// Substring on a scalar string.
	s.assertMatch(filters(filter(
		"tableName",
		schema.DataSourceFilterOperatorContains,
		core.MappingNodeFromString("orders"),
	)), true)
	// Membership in an array.
	s.assertMatch(filters(filter(
		"tags",
		schema.DataSourceFilterOperatorContains,
		core.MappingNodeFromString("prod"),
	)), true)
	s.assertMatch(filters(filter(
		"tags",
		schema.DataSourceFilterOperatorContains,
		core.MappingNodeFromString("absent"),
	)), false)
	s.assertMatch(filters(filter(
		"tags",
		schema.DataSourceFilterOperatorNotContains,
		core.MappingNodeFromString("absent"),
	)), true)
}

func (s *CCDataSourceFilterSuite) Test_starts_with_ends_with() {
	s.assertMatch(filters(filter(
		"tableName",
		schema.DataSourceFilterOperatorStartsWith,
		core.MappingNodeFromString("orders"),
	)), true)
	s.assertMatch(filters(filter(
		"tableName",
		schema.DataSourceFilterOperatorEndsWith,
		core.MappingNodeFromString("table"),
	)), true)
	s.assertMatch(filters(filter(
		"tableName",
		schema.DataSourceFilterOperatorNotStartsWith,
		core.MappingNodeFromString("xyz"),
	)), true)
}

func (s *CCDataSourceFilterSuite) Test_numeric_comparisons() {
	s.assertMatch(filters(filter(
		"readUnits",
		schema.DataSourceFilterOperatorGreaterThan,
		intNode(5),
	)), true)
	s.assertMatch(filters(filter(
		"readUnits",
		schema.DataSourceFilterOperatorGreaterThan,
		intNode(10),
	)), false)
	s.assertMatch(filters(filter(
		"readUnits",
		schema.DataSourceFilterOperatorGreaterThanOrEqual,
		intNode(10),
	)), true)
	s.assertMatch(filters(filter(
		"readUnits",
		schema.DataSourceFilterOperatorLessThan,
		intNode(20),
	)), true)
	s.assertMatch(filters(filter(
		"readUnits",
		schema.DataSourceFilterOperatorLessThanOrEqual,
		intNode(10),
	)), true)
}

func (s *CCDataSourceFilterSuite) Test_has_key() {
	s.assertMatch(filters(filter(
		"streamSpec",
		schema.DataSourceFilterOperatorHasKey,
		core.MappingNodeFromString("streamViewType"),
	)), true)
	s.assertMatch(filters(filter(
		"streamSpec",
		schema.DataSourceFilterOperatorHasKey,
		core.MappingNodeFromString("absent"),
	)), false)
	s.assertMatch(filters(filter(
		"streamSpec",
		schema.DataSourceFilterOperatorNotHasKey,
		core.MappingNodeFromString("absent"),
	)), true)
}

func (s *CCDataSourceFilterSuite) Test_filters_are_anded() {
	s.assertMatch(filters(
		filter(
			"tableName",
			schema.DataSourceFilterOperatorEquals,
			core.MappingNodeFromString("orders-table"),
		),
		filter(
			"readUnits",
			schema.DataSourceFilterOperatorGreaterThan,
			intNode(5),
		),
	), true)
	s.assertMatch(filters(
		filter(
			"tableName",
			schema.DataSourceFilterOperatorEquals,
			core.MappingNodeFromString("orders-table"),
		),
		filter(
			"readUnits",
			schema.DataSourceFilterOperatorGreaterThan,
			intNode(50),
		),
	), false)
}

func (s *CCDataSourceFilterSuite) Test_region_pseudo_filter_is_ignored() {
	s.assertMatch(filters(filter(
		"region",
		schema.DataSourceFilterOperatorEquals,
		core.MappingNodeFromString("us-east-1"),
	)), true)
}

func (s *CCDataSourceFilterSuite) Test_absent_field_positive_false_negative_true() {
	s.assertMatch(filters(filter(
		"missing",
		schema.DataSourceFilterOperatorEquals,
		core.MappingNodeFromString("x"),
	)), false)
	s.assertMatch(filters(filter(
		"missing",
		schema.DataSourceFilterOperatorNotEquals,
		core.MappingNodeFromString("x"),
	)), true)
}

func TestCCDataSourceFilterSuite(t *testing.T) {
	suite.Run(t, new(CCDataSourceFilterSuite))
}
