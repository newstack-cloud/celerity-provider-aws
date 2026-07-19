//go:build unit

package linkutils

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/suite"
)

type PhysicalNamesSuite struct {
	suite.Suite
}

func resourceInfoWithSpec(fields map[string]*core.MappingNode) *provider.ResourceInfo {
	return &provider.ResourceInfo{
		CurrentResourceState: &state.ResourceState{
			SpecData: &core.MappingNode{Fields: fields},
		},
	}
}

func (s *PhysicalNamesSuite) Test_returns_name_field_when_present() {
	info := resourceInfoWithSpec(map[string]*core.MappingNode{
		"tableName": core.MappingNodeFromString("orders"),
		"arn":       core.MappingNodeFromString("arn:aws:dynamodb:eu-west-1:123456789012:table/other"),
	})
	name, ok := PhysicalResourceName(info, "tableName")
	s.Require().True(ok)
	s.Equal("orders", name, "an explicit name should win over the ARN")
}

func (s *PhysicalNamesSuite) Test_derives_table_name_from_arn_when_name_missing() {
	info := resourceInfoWithSpec(map[string]*core.MappingNode{
		"arn": core.MappingNodeFromString("arn:aws:dynamodb:eu-west-1:123456789012:table/orders-8f2k1"),
	})
	name, ok := PhysicalResourceName(info, "tableName")
	s.Require().True(ok)
	s.Equal("orders-8f2k1", name)
}

func (s *PhysicalNamesSuite) Test_derives_bucket_name_from_pathless_arn() {
	info := resourceInfoWithSpec(map[string]*core.MappingNode{
		"arn": core.MappingNodeFromString("arn:aws:s3:::assets-bucket-8f2k1"),
	})
	name, ok := PhysicalResourceName(info, "bucketName")
	s.Require().True(ok)
	s.Equal("assets-bucket-8f2k1", name)
}

func (s *PhysicalNamesSuite) Test_empty_name_field_falls_back_to_arn() {
	info := resourceInfoWithSpec(map[string]*core.MappingNode{
		"bucketName": core.MappingNodeFromString(""),
		"arn":        core.MappingNodeFromString("arn:aws:s3:::assets-bucket"),
	})
	name, ok := PhysicalResourceName(info, "bucketName")
	s.Require().True(ok)
	s.Equal("assets-bucket", name)
}

func (s *PhysicalNamesSuite) Test_not_found_without_name_or_arn() {
	info := resourceInfoWithSpec(map[string]*core.MappingNode{
		"region": core.MappingNodeFromString("eu-west-1"),
	})
	_, ok := PhysicalResourceName(info, "tableName")
	s.False(ok)
}

func (s *PhysicalNamesSuite) Test_not_found_for_malformed_arn() {
	info := resourceInfoWithSpec(map[string]*core.MappingNode{
		"arn": core.MappingNodeFromString("not-an-arn"),
	})
	_, ok := PhysicalResourceName(info, "tableName")
	s.False(ok)
}

func (s *PhysicalNamesSuite) Test_not_found_for_nil_state() {
	_, ok := PhysicalResourceName(&provider.ResourceInfo{}, "tableName")
	s.False(ok)
}

func TestPhysicalNamesSuite(t *testing.T) {
	suite.Run(t, new(PhysicalNamesSuite))
}
