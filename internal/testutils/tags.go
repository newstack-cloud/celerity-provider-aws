package testutils

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/suite"
)

// AssertTags is for asserting that tags in the spec data of a resource
// in the form of a list of objects with "key" and "value" fields match the
// tags in the output of the upstream service.
func AssertTags(
	suite *suite.Suite,
	expectedSpecTags []*core.MappingNode,
	actualUpstreamTags map[string]string,
) {
	for _, tag := range expectedSpecTags {
		key := core.StringValue(tag.Fields["key"])
		value := core.StringValue(tag.Fields["value"])
		suite.Assert().Equal(actualUpstreamTags[key], value)
	}
}
