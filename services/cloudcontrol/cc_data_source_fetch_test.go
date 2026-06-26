//go:build unit

package cloudcontrol

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	awscctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	cloudcontrolmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/cloudcontrol_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type CCDataSourceFetchSuite struct {
	suite.Suite
}

const (
	dsTableName = "orders"
	dsTableARN  = "arn:aws:dynamodb:us-west-2:123456789012:table/orders"
	dsStreamARN = "arn:aws:dynamodb:us-west-2:123456789012:table/orders/stream/2024"
)

// The Cloud Control properties JSON (CFN PascalCase) for the
// orders table, including a write-only field and a Bluelink provenance tag.
const fullTableProperties = `{
  "TableName": "orders",
  "Arn": "arn:aws:dynamodb:us-west-2:123456789012:table/orders",
  "BillingMode": "PAY_PER_REQUEST",
  "StreamArn": "arn:aws:dynamodb:us-west-2:123456789012:table/orders/stream/2024",
  "ProvisionedThroughput": {"ReadCapacityUnits": 10, "WriteCapacityUnits": 5},
  "Tags": [{"Key": "team", "Value": "orders"}, {"Key": "bluelink:instance-id", "Value": "x"}],
  "Secret": "should-be-stripped"
}`

func tableDataSourceConfig() CCDataSourceConfig {
	stringSchema := func() *provider.ResourceDefinitionsSchema {
		return &provider.ResourceDefinitionsSchema{Type: provider.ResourceDefinitionsSchemaTypeString, Label: "s"}
	}
	return CCDataSourceConfig{
		BlueprintType:           "aws/dynamodb/table",
		CFNType:                 "AWS::DynamoDB::Table",
		Label:                   "AWS DynamoDB Table",
		DeriveIdentifierFromARN: true,
		Schema: &provider.ResourceDefinitionsSchema{
			Type:  provider.ResourceDefinitionsSchemaTypeObject,
			Label: "Table",
			Attributes: map[string]*provider.ResourceDefinitionsSchema{
				"tableName":   stringSchema(),
				"arn":         stringSchema(),
				"billingMode": stringSchema(),
				"streamArn":   stringSchema(),
				"secret":      stringSchema(),
				"provisionedThroughput": {
					Type:  provider.ResourceDefinitionsSchemaTypeObject,
					Label: "PT",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"readCapacityUnits":  {Type: provider.ResourceDefinitionsSchemaTypeInteger, Label: "r"},
						"writeCapacityUnits": {Type: provider.ResourceDefinitionsSchemaTypeInteger, Label: "w"},
					},
				},
				"tags": {
					Type:  provider.ResourceDefinitionsSchemaTypeArray,
					Label: "Tags",
					Items: &provider.ResourceDefinitionsSchema{
						Type:  provider.ResourceDefinitionsSchemaTypeObject,
						Label: "Tag",
						Attributes: map[string]*provider.ResourceDefinitionsSchema{
							"key":   stringSchema(),
							"value": stringSchema(),
						},
					},
				},
			},
		},
		Meta: CCResourceMeta{
			PrimaryIdentifierField:  "tableName",
			PrimaryIdentifierFields: []string{"tableName"},
			ComputedFields:          []string{"arn", "streamArn"},
			WriteOnlyFields:         []string{"secret"},
			TagPropertyName:         "tags",
			TagShape:                TagShapeKeyValueList,
		},
		FilterFields: map[string]*provider.DataSourceFilterSchema{
			"tableName":   {Type: provider.DataSourceFilterSearchValueTypeString},
			"arn":         {Type: provider.DataSourceFilterSearchValueTypeString},
			"billingMode": {Type: provider.DataSourceFilterSearchValueTypeString},
		},
		ExportFields: map[string]*provider.DataSourceSpecSchema{
			"tableName":   {Type: provider.DataSourceSpecTypeString},
			"arn":         {Type: provider.DataSourceSpecTypeString},
			"billingMode": {Type: provider.DataSourceSpecTypeString},
			"streamArn":   {Type: provider.DataSourceSpecTypeString},
			"provisionedThroughput.readCapacityUnits": {Type: provider.DataSourceSpecTypeInteger},
			"tags": {Type: provider.DataSourceSpecTypeArray},
		},
	}
}

func (s *CCDataSourceFetchSuite) actions(mock cloudcontrolservice.Service) *ccDataSourceActions {
	loader := &testutils.MockAWSConfigLoader{}
	return &ccDataSourceActions{
		config: tableDataSourceConfig(),
		cloudControlServiceFactory: func(*aws.Config, provider.Context) cloudcontrolservice.Service {
			return mock
		},
		awsConfigStore: utils.NewAWSConfigStore(
			[]string{}, utils.AWSConfigFromProviderContext, loader, utils.AWSConfigCacheKey,
		),
	}
}

func (s *CCDataSourceFetchSuite) fetchInput(filters *provider.ResolvedDataSourceFilters) *provider.DataSourceFetchInput {
	return &provider.DataSourceFetchInput{
		ProviderContext: plugintestutils.NewTestProviderContext(
			"aws",
			map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
			map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test")},
		),
		DataSourceWithResolvedSubs: &provider.ResolvedDataSource{Filter: filters},
	}
}

func (s *CCDataSourceFetchSuite) assertOrdersExports(out *provider.DataSourceFetchOutput) {
	s.Equal(dsTableName, core.StringValue(out.Data["tableName"]))
	s.Equal(dsTableARN, core.StringValue(out.Data["arn"]))
	s.Equal("PAY_PER_REQUEST", core.StringValue(out.Data["billingMode"]))
	s.Equal(dsStreamARN, core.StringValue(out.Data["streamArn"]))
	// Nested scalar flattened to dot-notation.
	s.Equal(10, core.IntValue(out.Data["provisionedThroughput.readCapacityUnits"]))
	// Array kept whole; Bluelink provenance tag filtered out, user tag retained.
	s.Require().NotNil(out.Data["tags"])
	s.Require().Len(out.Data["tags"].Items, 1)
	s.Equal("team", core.StringValue(out.Data["tags"].Items[0].Fields["key"]))
	// Write-only field is not exported.
	s.NotContains(out.Data, "secret")
}

func (s *CCDataSourceFetchSuite) Test_fast_path_by_primary_identifier() {
	mock := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &awscctypes.ResourceDescription{
				Properties: aws.String(fullTableProperties),
			},
		}),
	)
	out, err := s.actions(mock).Fetch(context.Background(), s.fetchInput(
		pluginutils.CreateStringEqualsFilter("tableName", dsTableName),
	))
	s.Require().NoError(err)
	s.assertOrdersExports(out)
	// No listing on the fast path.
	mock.AssertNotCalled(&s.Suite, "ListResources")
	mock.AssertCalled(&s.Suite, "GetResource")
}

func (s *CCDataSourceFetchSuite) Test_fast_path_by_arn_derives_identifier() {
	mock := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &awscctypes.ResourceDescription{Properties: aws.String(fullTableProperties)},
		}),
	)
	out, err := s.actions(mock).Fetch(context.Background(), s.fetchInput(
		pluginutils.CreateStringEqualsFilter("arn", dsTableARN),
	))
	s.Require().NoError(err)
	s.assertOrdersExports(out)
	mock.AssertNotCalled(&s.Suite, "ListResources")
	// GetResource identifier was derived from the ARN suffix.
	mock.AssertCalledWith(&s.Suite, "GetResource", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*awscc.GetResourceInput)
		return ok && aws.ToString(in.Identifier) == dsTableName
	})
}

func (s *CCDataSourceFetchSuite) Test_filter_path_advanced_operator() {
	mock := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithListResourcesOutput(&awscc.ListResourcesOutput{
			ResourceDescriptions: []awscctypes.ResourceDescription{
				{Identifier: aws.String("payments"), Properties: aws.String(`{"TableName":"payments","BillingMode":"PROVISIONED"}`)},
				{Identifier: aws.String(dsTableName), Properties: aws.String(`{"TableName":"orders","BillingMode":"PAY_PER_REQUEST"}`)},
			},
		}),
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &awscctypes.ResourceDescription{Properties: aws.String(fullTableProperties)},
		}),
	)
	out, err := s.actions(mock).Fetch(context.Background(), s.fetchInput(filters(
		filter("billingMode", schema.DataSourceFilterOperatorContains, core.MappingNodeFromString("PAY")),
	)))
	s.Require().NoError(err)
	s.assertOrdersExports(out)
	mock.AssertCalled(&s.Suite, "ListResources")
}

func (s *CCDataSourceFetchSuite) Test_filter_path_no_match_errors() {
	mock := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithListResourcesOutput(&awscc.ListResourcesOutput{
			ResourceDescriptions: []awscctypes.ResourceDescription{
				{Identifier: aws.String("payments"), Properties: aws.String(`{"TableName":"payments","BillingMode":"PROVISIONED"}`)},
			},
		}),
	)
	_, err := s.actions(mock).Fetch(context.Background(), s.fetchInput(filters(
		filter("billingMode", schema.DataSourceFilterOperatorEquals, core.MappingNodeFromString("PAY_PER_REQUEST")),
	)))
	s.Require().Error(err)
	s.Contains(err.Error(), "no \"aws/dynamodb/table\" resource matched")
}

func (s *CCDataSourceFetchSuite) Test_filter_path_ambiguous_errors() {
	mock := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithListResourcesOutput(&awscc.ListResourcesOutput{
			ResourceDescriptions: []awscctypes.ResourceDescription{
				{Identifier: aws.String("orders"), Properties: aws.String(`{"TableName":"orders","BillingMode":"PAY_PER_REQUEST"}`)},
				{Identifier: aws.String("orders-archive"), Properties: aws.String(`{"TableName":"orders-archive","BillingMode":"PAY_PER_REQUEST"}`)},
			},
		}),
	)
	_, err := s.actions(mock).Fetch(context.Background(), s.fetchInput(filters(
		filter("billingMode", schema.DataSourceFilterOperatorEquals, core.MappingNodeFromString("PAY_PER_REQUEST")),
	)))
	s.Require().Error(err)
	s.Contains(err.Error(), "matched more than one")
}

func TestCCDataSourceFetchSuite(t *testing.T) {
	suite.Run(t, new(CCDataSourceFetchSuite))
}
