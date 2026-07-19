//go:build unit

package lambdadynamodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	dynamodbmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/dynamodb_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

const (
	tflTableName = "orders-table"
	tflStreamARN = "arn:aws:dynamodb:us-west-2:123456789012:table/orders/stream/2024-01-01T00:00:00.000"
	tflFuncARN   = "arn:aws:lambda:us-west-2:123456789012:function:process-stream"
	tflRoleARN   = "arn:aws:iam::123456789012:role/process-stream-role"
	tflExecRole  = "processStreamFunctionExecutionRole"
	tflStreamSID = "DynamoDBStreamReadordersTable"
	tflESMUUID   = "esm-uuid-1234"
)

type TableFunctionLinkUpdateSuite struct {
	suite.Suite
}

// tableFunctionLinkFactory adapts the curried link constructor to the
// underlying LinkServiceDeps type expected by the test harness.
func tableFunctionLinkFactory(
	iamSvc iamservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, dynamodbservice.Service, *aws.Config, lambdaservice.Service],
) provider.Link {
	build := DynamoDBTableLambdaFunctionLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, dynamodbservice.Service, *aws.Config, lambdaservice.Service],
	) provider.Link {
		return build(DynamoDBTableToLambdaFunctionLinkDeps(deps))
	}
}

// tableResourceInfoStreamsDisabled has no computed streamArn, so the link treats
// the table as having no active stream and enables one.
func tableResourceInfoStreamsDisabled() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersTable",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"tableName", core.MappingNodeFromString(tflTableName),
				"arn", core.MappingNodeFromString(
					"arn:aws:dynamodb:us-west-2:123456789012:table/orders",
				),
			),
		},
	}
}

// tableResourceInfoStreamsEnabled has the computed streamArn populated, the
// CFN-canonical signal that a stream is active.
func tableResourceInfoStreamsEnabled() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersTable",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"tableName", core.MappingNodeFromString(tflTableName),
				"streamSpecification", core.MappingNodeFields(
					"streamViewType", core.MappingNodeFromString("NEW_AND_OLD_IMAGES"),
				),
				"streamArn", core.MappingNodeFromString(tflStreamARN),
			),
		},
	}
}

func streamFunctionResourceInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "processStreamFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(tflFuncARN),
			),
		},
	}
}

// Carries two indexed filter-pattern
// annotations (filter.0 and filter.1) so the link builds a two-entry filter
// criteria on the event source mapping.
func streamFunctionResourceInfoWithFilters() *provider.ResourceInfo {
	info := streamFunctionResourceInfo()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.dynamodb.lambda.stream.filter.0",
				core.MappingNodeFromString(`{"eventName":["INSERT"]}`),
				"aws.dynamodb.lambda.stream.filter.1",
				core.MappingNodeFromString(`{"eventName":["MODIFY"]}`),
			),
		},
	}
	return info
}

// Carries the reportBatchItemFailures annotation with the given value, controlling
// whether the link sets the ReportBatchItemFailures function response type on the
// event source mapping.
func streamFunctionResourceInfoWithReportBatchItemFailures(enabled bool) *provider.ResourceInfo {
	info := streamFunctionResourceInfo()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.dynamodb.lambda.stream.reportBatchItemFailures",
				core.MappingNodeFromBool(enabled),
			),
		},
	}
	return info
}

// hasReportBatchItemFailures reports whether the create event source mapping input
// requests the ReportBatchItemFailures function response type.
func hasReportBatchItemFailures(input *lambda.CreateEventSourceMappingInput) bool {
	for _, responseType := range input.FunctionResponseTypes {
		if responseType == lambdatypes.FunctionResponseTypeReportBatchItemFailures {
			return true
		}
	}
	return false
}

func (s *TableFunctionLinkUpdateSuite) Test_report_batch_item_failures_annotation_sets_function_response_type() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(tflFuncARN),
				Role:        aws.String(tflRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID),
		}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "reportBatchItemFailures annotation sets the ReportBatchItemFailures function response type",
		ServiceFactoryA:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    tableResourceInfoStreamsEnabled(),
			ResourceBInfo:    streamFunctionResourceInfoWithReportBatchItemFailures(true),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(tflRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchStreamGrantOutput,
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			dynamodbservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]{testCase},
		tableFunctionLinkFactory(iamSvc),
		&s.Suite,
	)

	lambdaSvc.AssertCalledWith(
		&s.Suite,
		"CreateEventSourceMapping",
		0,
		plugintestutils.Any,
		func(arg any) bool {
			input, ok := arg.(*lambda.CreateEventSourceMappingInput)
			return ok && hasReportBatchItemFailures(input)
		},
	)
}

// When the stream was enabled by this link's own UpdateResourceA (table deployed
// without streams), the table's state snapshot has no streamArn, the
// intermediary update must resolve it live via DescribeTable instead of failing.
func (s *TableFunctionLinkUpdateSuite) Test_intermediary_update_resolves_stream_arn_via_describe_table_when_state_has_none() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(tflFuncARN),
				Role:        aws.String(tflRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID),
		}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "stream ARN missing from table state is resolved live via DescribeTable",
		ServiceFactoryA: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
				Table: &dynamodbtypes.TableDescription{
					TableName:       aws.String(tflTableName),
					LatestStreamArn: aws.String(tflStreamARN),
				},
			}),
		),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			InstanceName:   "test-instance",
			// No streamArn in the table's state — the pre-fix code errored here
			// with "ensure the table has DynamoDB Streams enabled".
			ResourceAInfo:    tableResourceInfoStreamsDisabled(),
			ResourceBInfo:    streamFunctionResourceInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(tflRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchStreamGrantOutput,
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			dynamodbservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]{testCase},
		tableFunctionLinkFactory(iamSvc),
		&s.Suite,
	)

	lambdaSvc.AssertCalledWith(
		&s.Suite,
		"CreateEventSourceMapping",
		0,
		plugintestutils.Any,
		func(arg any) bool {
			input, ok := arg.(*lambda.CreateEventSourceMappingInput)
			return ok && aws.ToString(input.EventSourceArn) == tflStreamARN
		},
	)
}

func (s *TableFunctionLinkUpdateSuite) Test_absent_report_batch_item_failures_annotation_leaves_response_types_unset() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(tflFuncARN),
				Role:        aws.String(tflRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID),
		}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "absent reportBatchItemFailures annotation leaves function response types unset",
		ServiceFactoryA:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    tableResourceInfoStreamsEnabled(),
			ResourceBInfo:    streamFunctionResourceInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(tflRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchStreamGrantOutput,
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			dynamodbservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]{testCase},
		tableFunctionLinkFactory(iamSvc),
		&s.Suite,
	)

	lambdaSvc.AssertCalledWith(
		&s.Suite,
		"CreateEventSourceMapping",
		0,
		plugintestutils.Any,
		func(arg any) bool {
			input, ok := arg.(*lambda.CreateEventSourceMappingInput)
			return ok && len(input.FunctionResponseTypes) == 0
		},
	)
}

// Regression: when reportBatchItemFailures flips from true to false, the update must
// send an empty (non-nil) FunctionResponseTypes so the previously enabled
// ReportBatchItemFailures setting is cleared; a nil field would leave it in place.
func (s *TableFunctionLinkUpdateSuite) Test_update_clears_function_response_types_when_report_batch_item_failures_disabled() {
	loader := &testutils.MockAWSConfigLoader{}

	existing := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Sid":%q,"Effect":"Allow","Action":["dynamodb:GetRecords"],"Resource":%q}]}`,
		tflStreamSID,
		tflStreamARN,
	)
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(tflFuncARN),
				Role:        aws.String(tflRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
		lambdamock.WithUpdateEventSourceMappingOutput(&lambda.UpdateEventSourceMappingOutput{
			UUID:                  aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID),
		}),
	)

	currentLinkState := &state.LinkState{
		Data: map[string]*core.MappingNode{
			"intermediaries": {
				Fields: map[string]*core.MappingNode{
					tableFunctionESMID: {
						Fields: map[string]*core.MappingNode{
							"uuid": core.MappingNodeFromString(tflESMUUID),
						},
					},
				},
			},
		},
	}

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "update clears function response types when reportBatchItemFailures is disabled",
		ServiceFactoryA:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeUpdate,
			InstanceName:     "test-instance",
			ResourceAInfo:    tableResourceInfoStreamsEnabled(),
			ResourceBInfo:    streamFunctionResourceInfoWithReportBatchItemFailures(false),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(tflRoleState())),
			CurrentLinkState: currentLinkState,
		},
		ExpectedOutputMatcher: matchStreamGrantOutput,
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			dynamodbservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]{testCase},
		tableFunctionLinkFactory(iamSvc),
		&s.Suite,
	)

	lambdaSvc.AssertCalledWith(
		&s.Suite,
		"UpdateEventSourceMapping",
		0,
		plugintestutils.Any,
		func(arg any) bool {
			input, ok := arg.(*lambda.UpdateEventSourceMappingInput)
			return ok &&
				input.FunctionResponseTypes != nil &&
				len(input.FunctionResponseTypes) == 0
		},
	)
}

// Regression: when all filter annotations are removed, the update must send an
// empty (non-nil) FilterCriteria so previously configured filters are removed;
// a nil field would leave them in place.
func (s *TableFunctionLinkUpdateSuite) Test_update_clears_filter_criteria_when_filter_annotations_removed() {
	loader := &testutils.MockAWSConfigLoader{}

	existing := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Sid":%q,"Effect":"Allow","Action":["dynamodb:GetRecords"],"Resource":%q}]}`,
		tflStreamSID,
		tflStreamARN,
	)
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(tflFuncARN),
				Role:        aws.String(tflRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
		lambdamock.WithUpdateEventSourceMappingOutput(&lambda.UpdateEventSourceMappingOutput{
			UUID:                  aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID),
		}),
	)

	currentLinkState := &state.LinkState{
		Data: map[string]*core.MappingNode{
			"intermediaries": {
				Fields: map[string]*core.MappingNode{
					tableFunctionESMID: {
						Fields: map[string]*core.MappingNode{
							"uuid": core.MappingNodeFromString(tflESMUUID),
						},
					},
				},
			},
		},
	}

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "update clears filter criteria when filter annotations are removed",
		ServiceFactoryA:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeUpdate,
			InstanceName:     "test-instance",
			ResourceAInfo:    tableResourceInfoStreamsEnabled(),
			ResourceBInfo:    streamFunctionResourceInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(tflRoleState())),
			CurrentLinkState: currentLinkState,
		},
		ExpectedOutputMatcher: matchStreamGrantOutput,
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			dynamodbservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]{testCase},
		tableFunctionLinkFactory(iamSvc),
		&s.Suite,
	)

	lambdaSvc.AssertCalledWith(
		&s.Suite,
		"UpdateEventSourceMapping",
		0,
		plugintestutils.Any,
		func(arg any) bool {
			input, ok := arg.(*lambda.UpdateEventSourceMappingInput)
			return ok &&
				input.FilterCriteria != nil &&
				len(input.FilterCriteria.Filters) == 0
		},
	)
}

func (s *TableFunctionLinkUpdateSuite) Test_indexed_filter_annotations_build_esm_filter_criteria() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(tflFuncARN),
				Role:        aws.String(tflRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID),
		}),
	)

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "indexed filter annotations populate the event source mapping filter criteria",
		ServiceFactoryA:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    tableResourceInfoStreamsEnabled(),
			ResourceBInfo:    streamFunctionResourceInfoWithFilters(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(tflRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchStreamGrantOutput,
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			dynamodbservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]{testCase},
		tableFunctionLinkFactory(iamSvc),
		&s.Suite,
	)

	lambdaSvc.AssertCalledWith(
		&s.Suite,
		"CreateEventSourceMapping",
		0,
		plugintestutils.Any,
		func(arg any) bool {
			input, ok := arg.(*lambda.CreateEventSourceMappingInput)
			if !ok || input.FilterCriteria == nil {
				return false
			}
			patterns := make([]string, 0, len(input.FilterCriteria.Filters))
			for _, f := range input.FilterCriteria.Filters {
				patterns = append(patterns, aws.ToString(f.Pattern))
			}
			return len(patterns) == 2 &&
				patterns[0] == `{"eventName":["INSERT"]}` &&
				patterns[1] == `{"eventName":["MODIFY"]}`
		},
	)
}

func (s *TableFunctionLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		tableFunctionEnableStreamsTestCase(loader),
		tableFunctionStreamsAlreadyEnabledTestCase(loader),
		tableFunctionDestroyNoOpTestCase(loader),
		tableFunctionEnableStreamsErrorTestCase(loader),
		tableFunctionUpdateResourceBNoOpTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		tableFunctionLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

func tableFunctionEnableStreamsTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	dynamodbSvc := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithUpdateTableOutput(&dynamodb.UpdateTableOutput{}),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "enables DynamoDB streams on the table",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) dynamodbservice.Service { return dynamodbSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &dynamodbSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      tableResourceInfoStreamsDisabled(),
			OtherResourceInfo: streamFunctionResourceInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(
				"streamSpecification",
				core.MappingNodeFields(
					"streamEnabled", core.MappingNodeFromBool(true),
					"streamViewType", core.MappingNodeFromString("NEW_AND_OLD_IMAGES"),
				),
			),
		},
		UpdateActionsCalled: map[string]any{
			"UpdateTable": &dynamodb.UpdateTableInput{
				TableName: aws.String(tflTableName),
				StreamSpecification: &dynamodbtypes.StreamSpecification{
					StreamEnabled:  aws.Bool(true),
					StreamViewType: dynamodbtypes.StreamViewType("NEW_AND_OLD_IMAGES"),
				},
			},
		},
	}
}

func tableFunctionStreamsAlreadyEnabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	dynamodbSvc := dynamodbmock.CreateDynamoDBServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "is a no-op when streams are already enabled",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) dynamodbservice.Service { return dynamodbSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &dynamodbSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      tableResourceInfoStreamsEnabled(),
			OtherResourceInfo: streamFunctionResourceInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(),
		},
	}
}

func tableFunctionDestroyNoOpTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	dynamodbSvc := dynamodbmock.CreateDynamoDBServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "does not disable streams on destroy",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) dynamodbservice.Service { return dynamodbSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &dynamodbSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      tableResourceInfoStreamsDisabled(),
			OtherResourceInfo: streamFunctionResourceInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(),
		},
	}
}

func tableFunctionEnableStreamsErrorTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	dynamodbSvc := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithUpdateTableError(errors.New("update table failed")),
	)

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "returns error when enabling streams fails",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) dynamodbservice.Service { return dynamodbSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &dynamodbSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      tableResourceInfoStreamsDisabled(),
			OtherResourceInfo: streamFunctionResourceInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectError: true,
	}
}

func tableFunctionUpdateResourceBNoOpTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	dynamodbSvc := dynamodbmock.CreateDynamoDBServiceMock()
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "does not modify the lambda function (resourceB is a no-op)",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) dynamodbservice.Service { return dynamodbSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      streamFunctionResourceInfo(),
			OtherResourceInfo: tableResourceInfoStreamsEnabled(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
	}
}

// Test_link_update_intermediary_resources exercises the create, update and
// destroy paths of UpdateIntermediaryResources, which manage the event source
// mapping and the DynamoDB stream-read permission on the function's execution
// role policy.
func (s *TableFunctionLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	// Each case wires up its own IAM mock (curried into the link constructor),
	// so we run the harness once per case with a factory bound to that case's
	// IAM service.
	createCase, createIam := tableFunctionCreateIntermediaryTestCase(loader)
	updateCase, updateIam := tableFunctionUpdateIntermediaryTestCase(loader)
	destroyCase, destroyIam := tableFunctionDestroyIntermediaryTestCase(loader)

	cases := []struct {
		testCase plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			dynamodbservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]
		iamSvc iamservice.Service
	}{
		{createCase, createIam},
		{updateCase, updateIam},
		{destroyCase, destroyIam},
	}

	for _, c := range cases {
		plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
			[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
				*aws.Config,
				dynamodbservice.Service,
				*aws.Config,
				lambdaservice.Service,
			]{c.testCase},
			tableFunctionLinkFactory(c.iamSvc),
			&s.Suite,
		)
	}
}

const (
	tflRoleName         = "process-stream-role"
	tflRoleResourceName = "processStreamFunctionRole"
)

func tflRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: tflRoleResourceName,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(tflRoleName),
			"arn", core.MappingNodeFromString(tflRoleARN),
		),
	}
}

// matchPutStreamInlineAccessPolicy verifies a PutRolePolicy targets the role's
// shared allocator inline policy and its document grants the stream-read statement.
func matchPutStreamInlineAccessPolicy(arg any) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != tflRoleName ||
		aws.ToString(input.PolicyName) != linkutils.InlineAccessPolicyName() {
		return false
	}
	var doc struct {
		Statement []struct{ Sid string }
	}
	if err := json.Unmarshal([]byte(aws.ToString(input.PolicyDocument)), &doc); err != nil {
		return false
	}
	for _, statement := range doc.Statement {
		if statement.Sid == tflStreamSID {
			return true
		}
	}
	return false
}

// matchStreamGrantOutput asserts the link preserves the event source mapping link
// data and records the stream-read statement in link data, mapping it onto the
// role's spec by Sid (so the role does not strip the grant).
func matchStreamGrantOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	esmARN := "arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID
	mappingKey := fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		tflRoleResourceName,
		linkutils.InlineAccessPolicyName(),
		tflStreamSID,
	)
	summary := map[string]any{}
	if actual != nil {
		summary["mappingValue"] = actual.ResourceDataMappings[mappingKey]
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[tflExecRole] != nil &&
			actual.LinkData.Fields[tflExecRole].Fields[linkutils.PermissionFieldName] != nil
		summary["esmUUID"] = esmLinkValue(actual, "uuid")
		summary["esmARN"] = esmLinkValue(actual, "arn")
		summary["esmEventSourceArn"] = esmLinkValue(actual, "eventSourceArn")
		summary["esmFunctionArn"] = esmLinkValue(actual, "functionArn")
	}
	expected := map[string]any{
		"mappingValue":      linkutils.PermissionFieldPath(tflExecRole),
		"hasStatement":      true,
		"esmUUID":           tflESMUUID,
		"esmARN":            esmARN,
		"esmEventSourceArn": tflStreamARN,
		"esmFunctionArn":    tflFuncARN,
	}
	return plugintestutils.EqualityCheckValues{Expected: expected, Actual: summary}, nil
}

func esmLinkValue(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
	field string,
) string {
	if actual.LinkData == nil {
		return ""
	}
	intermediaries, ok := actual.LinkData.Fields["intermediaries"]
	if !ok || intermediaries == nil {
		return ""
	}
	esm, ok := intermediaries.Fields[tableFunctionESMID]
	if !ok || esm == nil {
		return ""
	}
	return core.StringValue(esm.Fields[field])
}

func tflLambdaSvc() lambdaservice.Service {
	return lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(tflFuncARN),
				Role:        aws.String(tflRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID),
		}),
		lambdamock.WithUpdateEventSourceMappingOutput(&lambda.UpdateEventSourceMappingOutput{
			UUID:                  aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID),
		}),
		lambdamock.WithDeleteEventSourceMappingOutput(&lambda.DeleteEventSourceMappingOutput{}),
	)
}

// tableFunctionCreateIntermediaryTestCase exercises the create path: the role has
// no allocator policy yet, so the stream-read statement is packed into a new inline
// policy via PutRolePolicy, and the event source mapping is created.
func tableFunctionCreateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "creates the event source mapping and stream-read inline allocator policy on create",
		ServiceFactoryA:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return tflLambdaSvc() },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    tableResourceInfoStreamsEnabled(),
			ResourceBInfo:    streamFunctionResourceInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(tflRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchStreamGrantOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutStreamInlineAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

// tableFunctionUpdateIntermediaryTestCase exercises the update path: the inline
// allocator policy already holds the stream-read statement, which is replaced in
// place, and the event source mapping is updated using the stored UUID.
func tableFunctionUpdateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	existing := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[`+
			`{"Sid":%q,"Effect":"Allow","Action":["dynamodb:GetRecords"],"Resource":%q}]}`,
		tflStreamSID,
		"arn:aws:dynamodb:us-west-2:123456789012:table/old-orders/stream/2020-01-01T00:00:00.000",
	)
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	currentLinkState := &state.LinkState{
		Data: map[string]*core.MappingNode{
			"intermediaries": {
				Fields: map[string]*core.MappingNode{
					tableFunctionESMID: {
						Fields: map[string]*core.MappingNode{
							"uuid": core.MappingNodeFromString(tflESMUUID),
						},
					},
				},
			},
		},
	}

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "replaces the stream-read statement in the inline allocator policy on update",
		ServiceFactoryA:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return tflLambdaSvc() },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeUpdate,
			InstanceName:     "test-instance",
			ResourceAInfo:    tableResourceInfoStreamsEnabled(),
			ResourceBInfo:    streamFunctionResourceInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(tflRoleState())),
			CurrentLinkState: currentLinkState,
		},
		ExpectedOutputMatcher: matchStreamGrantOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutStreamInlineAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

// tableFunctionDestroyIntermediaryTestCase exercises the destroy path: the link's
// stream statement is the only statement in the inline allocator policy, so once it
// is removed by Sid the policy is deleted via DeleteRolePolicy. The event source
// mapping is also deleted.
func tableFunctionDestroyIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	existing := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Sid":%q,"Effect":"Allow","Action":["dynamodb:GetRecords"],"Resource":%q}]}`,
		tflStreamSID,
		tflStreamARN,
	)

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
	)

	currentLinkState := &state.LinkState{
		Data: map[string]*core.MappingNode{
			"intermediaries": {
				Fields: map[string]*core.MappingNode{
					tableFunctionESMID: {
						Fields: map[string]*core.MappingNode{
							"uuid": core.MappingNodeFromString(tflESMUUID),
						},
					},
				},
			},
		},
	}

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		dynamodbservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "deletes the event source mapping and the empty inline allocator policy on destroy",
		ServiceFactoryA:                dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return tflLambdaSvc() },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeDestroy,
			InstanceName:     "test-instance",
			ResourceAInfo:    tableResourceInfoStreamsEnabled(),
			ResourceBInfo:    streamFunctionResourceInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(tflRoleState())),
			CurrentLinkState: currentLinkState,
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
			LinkData:                   core.MappingNodeFields(),
		},
		UpdateActionsCalled: map[string]any{
			"DeleteRolePolicy": func(arg any) bool {
				input, ok := arg.(*iam.DeleteRolePolicyInput)
				return ok &&
					aws.ToString(input.RoleName) == tflRoleName &&
					aws.ToString(input.PolicyName) == linkutils.InlineAccessPolicyName()
			},
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
	}, iamSvc
}

func TestTableFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(TableFunctionLinkUpdateSuite))
}
