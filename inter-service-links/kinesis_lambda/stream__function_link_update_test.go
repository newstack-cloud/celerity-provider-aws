//go:build unit

package kinesislambda

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
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
	tflStreamARN = "arn:aws:kinesis:us-west-2:123456789012:stream/events-stream"
	tflFuncARN   = "arn:aws:lambda:us-west-2:123456789012:function:process-stream"
	tflRoleARN   = "arn:aws:iam::123456789012:role/process-stream-role"
	tflExecRole  = "processStreamFunctionExecutionRole"
	tflStreamSID = "KinesisStreamReadeventsStream"
	tflESMUUID   = "esm-uuid-1234"
)

type StreamFunctionLinkUpdateSuite struct {
	suite.Suite
}

func streamFunctionLinkFactory(
	iamSvc iamservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, cloudcontrolservice.Service, *aws.Config, lambdaservice.Service],
) provider.Link {
	build := StreamFunctionLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, cloudcontrolservice.Service, *aws.Config, lambdaservice.Service],
	) provider.Link {
		return build(StreamToFunctionLinkDeps(deps))
	}
}

func streamResourceInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "eventsStream",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"name", core.MappingNodeFromString("events-stream"),
				"arn", core.MappingNodeFromString(tflStreamARN),
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

func streamFunctionResourceInfoWithFilters() *provider.ResourceInfo {
	info := streamFunctionResourceInfo()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.kinesis.lambda.filter.0",
				core.MappingNodeFromString(`{"data":{"eventType":["created"]}}`),
				"aws.kinesis.lambda.filter.1",
				core.MappingNodeFromString(`{"data":{"eventType":["updated"]}}`),
			),
		},
	}
	return info
}

func (s *StreamFunctionLinkUpdateSuite) Test_indexed_filter_annotations_build_esm_filter_criteria() {
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
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "indexed filter annotations populate the event source mapping filter criteria",
		ServiceFactoryA:                noopCloudControlServiceFactory,
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    streamResourceInfo(),
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
			cloudcontrolservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]{testCase},
		streamFunctionLinkFactory(iamSvc),
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
				patterns[0] == `{"data":{"eventType":["created"]}}` &&
				patterns[1] == `{"data":{"eventType":["updated"]}}`
		},
	)
}

func (s *StreamFunctionLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		streamFunctionUpdateResourceANoOpTestCase(loader),
		streamFunctionDestroyNoOpTestCase(loader),
		streamFunctionUpdateResourceBNoOpTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		streamFunctionLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

func streamFunctionUpdateResourceANoOpTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "does not modify the Kinesis stream (resourceA is a no-op)",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         noopCloudControlServiceFactory,
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      streamResourceInfo(),
			OtherResourceInfo: streamFunctionResourceInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(),
		},
	}
}

func streamFunctionDestroyNoOpTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "is a no-op on destroy for the Kinesis stream",
		Resource:                plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA:         noopCloudControlServiceFactory,
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      streamResourceInfo(),
			OtherResourceInfo: streamFunctionResourceInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(),
		},
	}
}

func streamFunctionUpdateResourceBNoOpTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                    "does not modify the lambda function (resourceB is a no-op)",
		Resource:                plugintestutils.LinkUpdateResourceB,
		ServiceFactoryA:         noopCloudControlServiceFactory,
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      streamFunctionResourceInfo(),
			OtherResourceInfo: streamResourceInfo(),
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
// mapping and the Kinesis stream-read permission on the function's execution
// role policy.
func (s *StreamFunctionLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	// Each case wires up its own IAM mock (curried into the link constructor),
	// so we run the harness once per case with a factory bound to that case's
	// IAM service.
	createCase, createIam := streamFunctionCreateIntermediaryTestCase(loader)
	updateCase, updateIam := streamFunctionUpdateIntermediaryTestCase(loader)
	destroyCase, destroyIam := streamFunctionDestroyIntermediaryTestCase(loader)

	cases := []struct {
		testCase plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			cloudcontrolservice.Service,
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
				cloudcontrolservice.Service,
				*aws.Config,
				lambdaservice.Service,
			]{c.testCase},
			streamFunctionLinkFactory(c.iamSvc),
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

// Asserts the link preserves the event source mapping link
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

	return plugintestutils.EqualityCheckValues{
		Expected: expected,
		Actual:   summary,
	}, nil
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
	esm, ok := intermediaries.Fields[streamFunctionESMID]
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
			UUID: aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String(
				"arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID,
			),
		}),
		lambdamock.WithUpdateEventSourceMappingOutput(&lambda.UpdateEventSourceMappingOutput{
			UUID: aws.String(tflESMUUID),
			EventSourceMappingArn: aws.String(
				"arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + tflESMUUID,
			),
		}),
		lambdamock.WithDeleteEventSourceMappingOutput(&lambda.DeleteEventSourceMappingOutput{}),
	)
}

func streamFunctionCreateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
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
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "creates the event source mapping and stream-read inline allocator policy on create",
		ServiceFactoryA:                noopCloudControlServiceFactory,
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return tflLambdaSvc() },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    streamResourceInfo(),
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

func streamFunctionUpdateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	existing := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[`+
			`{"Sid":%q,"Effect":"Allow","Action":["kinesis:GetRecords"],"Resource":%q}]}`,
		tflStreamSID,
		"arn:aws:kinesis:us-west-2:123456789012:stream/old-events-stream",
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
					streamFunctionESMID: {
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
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "replaces the stream-read statement in the inline allocator policy on update",
		ServiceFactoryA:                noopCloudControlServiceFactory,
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return tflLambdaSvc() },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeUpdate,
			InstanceName:     "test-instance",
			ResourceAInfo:    streamResourceInfo(),
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

func streamFunctionDestroyIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	existing := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Sid":%q,"Effect":"Allow","Action":["kinesis:GetRecords"],"Resource":%q}]}`,
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
					streamFunctionESMID: {
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
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "deletes the event source mapping and the empty inline allocator policy on destroy",
		ServiceFactoryA:                noopCloudControlServiceFactory,
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return tflLambdaSvc() },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeDestroy,
			InstanceName:     "test-instance",
			ResourceAInfo:    streamResourceInfo(),
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

func TestStreamFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(StreamFunctionLinkUpdateSuite))
}
