//go:build unit

package sqslambda

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
	"github.com/stretchr/testify/suite"
)

const (
	qflQueueName = "orders-queue"
	qflQueueARN  = "arn:aws:sqs:us-west-2:123456789012:orders-queue"
	qflFuncARN   = "arn:aws:lambda:us-west-2:123456789012:function:process-queue"
	qflRoleARN   = "arn:aws:iam::123456789012:role/process-queue-role"
	qflExecRole  = "processQueueFunctionExecutionRole"
	qflQueueSID  = "SQSQueueConsumeordersQueue"
	qflESMUUID   = "esm-uuid-1234"
)

type QueueFunctionLinkUpdateSuite struct {
	suite.Suite
}

func queueResourceInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersQueue",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"queueName", core.MappingNodeFromString(qflQueueName),
				"arn", core.MappingNodeFromString(qflQueueARN),
			),
		},
	}
}

func queueFunctionResourceInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "processQueueFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(qflFuncARN),
			),
		},
	}
}

func queueFunctionResourceInfoWithFilters() *provider.ResourceInfo {
	info := queueFunctionResourceInfo()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.sqs.lambda.filter.0",
				core.MappingNodeFromString(`{"body":{"action":["create"]}}`),
				"aws.sqs.lambda.filter.1",
				core.MappingNodeFromString(`{"body":{"action":["update"]}}`),
			),
		},
	}
	return info
}

func (s *QueueFunctionLinkUpdateSuite) Test_indexed_filter_annotations_build_esm_filter_criteria() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(qflFuncARN),
				Role:        aws.String(qflRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(qflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + qflESMUUID),
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
			ResourceAInfo:    queueResourceInfo(),
			ResourceBInfo:    queueFunctionResourceInfoWithFilters(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(qflRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchConsumeGrantOutput,
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			cloudcontrolservice.Service,
			*aws.Config,
			lambdaservice.Service,
		]{testCase},
		queueFunctionLinkFactory(iamSvc),
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
				patterns[0] == `{"body":{"action":["create"]}}` &&
				patterns[1] == `{"body":{"action":["update"]}}`
		},
	)
}

func (s *QueueFunctionLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		queueFunctionUpdateResourceANoOpTestCase(loader),
		queueFunctionUpdateResourceADestroyNoOpTestCase(loader),
		queueFunctionUpdateResourceBNoOpTestCase(loader),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		queueFunctionLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

func queueFunctionUpdateResourceANoOpTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:            "does not modify the queue (resourceA is a no-op)",
		Resource:        plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA: noopCloudControlServiceFactory,
		ConfigStoreA:    testConfigStore(loader),
		ServiceFactoryB: lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStoreB:    testConfigStore(loader),
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeCreate,
			ResourceInfo:      queueResourceInfo(),
			OtherResourceInfo: queueFunctionResourceInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(),
		},
	}
}

func queueFunctionUpdateResourceADestroyNoOpTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:            "does not modify the queue on destroy (resourceA is a no-op)",
		Resource:        plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA: noopCloudControlServiceFactory,
		ConfigStoreA:    testConfigStore(loader),
		ServiceFactoryB: lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStoreB:    testConfigStore(loader),
		Input: &provider.LinkUpdateResourceInput{
			LinkUpdateType:    provider.LinkUpdateTypeDestroy,
			ResourceInfo:      queueResourceInfo(),
			OtherResourceInfo: queueFunctionResourceInfo(),
			LinkContext:       testLinkContext(),
		},
		ExpectedOutput: &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(),
		},
	}
}

func queueFunctionUpdateResourceBNoOpTestCase(
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
			ResourceInfo:      queueFunctionResourceInfo(),
			OtherResourceInfo: queueResourceInfo(),
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
// mapping and the SQS queue-consume permission on the function's execution
// role policy.
func (s *QueueFunctionLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	// Each case wires up its own IAM mock (curried into the link constructor),
	// so we run the harness once per case with a factory bound to that case's
	// IAM service.
	createCase, createIam := queueFunctionCreateIntermediaryTestCase(loader)
	updateCase, updateIam := queueFunctionUpdateIntermediaryTestCase(loader)
	destroyCase, destroyIam := queueFunctionDestroyIntermediaryTestCase(loader)

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
			queueFunctionLinkFactory(c.iamSvc),
			&s.Suite,
		)
	}
}

const (
	qflRoleName         = "process-queue-role"
	qflRoleResourceName = "processQueueFunctionRole"
)

func qflRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: qflRoleResourceName,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(qflRoleName),
			"arn", core.MappingNodeFromString(qflRoleARN),
		),
	}
}

// matchPutConsumeInlineAccessPolicy verifies a PutRolePolicy targets the role's
// shared allocator inline policy and its document grants the queue-consume statement.
func matchPutConsumeInlineAccessPolicy(arg any) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != qflRoleName ||
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
		if statement.Sid == qflQueueSID {
			return true
		}
	}
	return false
}

// Asserts that the link preserves the event source mapping link
// data and records the queue-consume statement in link data, mapping it onto the
// role's spec by Sid (so the role does not strip the grant).
func matchConsumeGrantOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	esmARN := "arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + qflESMUUID
	mappingKey := fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		qflRoleResourceName,
		linkutils.InlineAccessPolicyName(),
		qflQueueSID,
	)
	summary := map[string]any{}
	if actual != nil {
		summary["mappingValue"] = actual.ResourceDataMappings[mappingKey]
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[qflExecRole] != nil &&
			actual.LinkData.Fields[qflExecRole].Fields[linkutils.PermissionFieldName] != nil
		summary["esmUUID"] = esmLinkValue(actual, "uuid")
		summary["esmARN"] = esmLinkValue(actual, "arn")
		summary["esmEventSourceArn"] = esmLinkValue(actual, "eventSourceArn")
		summary["esmFunctionArn"] = esmLinkValue(actual, "functionArn")
	}
	expected := map[string]any{
		"mappingValue":      linkutils.PermissionFieldPath(qflExecRole),
		"hasStatement":      true,
		"esmUUID":           qflESMUUID,
		"esmARN":            esmARN,
		"esmEventSourceArn": qflQueueARN,
		"esmFunctionArn":    qflFuncARN,
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
	esm, ok := intermediaries.Fields[queueFunctionESMID]
	if !ok || esm == nil {
		return ""
	}
	return core.StringValue(esm.Fields[field])
}

func qflLambdaSvc() lambdaservice.Service {
	return lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &lambdatypes.FunctionConfiguration{
				FunctionArn: aws.String(qflFuncARN),
				Role:        aws.String(qflRoleARN),
				Environment: &lambdatypes.EnvironmentResponse{Variables: map[string]string{}},
			},
		}),
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(qflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + qflESMUUID),
		}),
		lambdamock.WithUpdateEventSourceMappingOutput(&lambda.UpdateEventSourceMappingOutput{
			UUID:                  aws.String(qflESMUUID),
			EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:" + qflESMUUID),
		}),
		lambdamock.WithDeleteEventSourceMappingOutput(&lambda.DeleteEventSourceMappingOutput{}),
	)
}

func queueFunctionCreateIntermediaryTestCase(
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
		Name:                           "creates the event source mapping and queue-consume inline allocator policy on create",
		ServiceFactoryA:                noopCloudControlServiceFactory,
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return qflLambdaSvc() },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    queueResourceInfo(),
			ResourceBInfo:    queueFunctionResourceInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(qflRoleState())),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchConsumeGrantOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutConsumeInlineAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

func queueFunctionUpdateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	existing := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[`+
			`{"Sid":%q,"Effect":"Allow","Action":["sqs:ReceiveMessage"],"Resource":%q}]}`,
		qflQueueSID,
		"arn:aws:sqs:us-west-2:123456789012:old-orders-queue",
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
					queueFunctionESMID: {
						Fields: map[string]*core.MappingNode{
							"uuid": core.MappingNodeFromString(qflESMUUID),
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
		Name:                           "replaces the queue-consume statement in the inline allocator policy on update",
		ServiceFactoryA:                noopCloudControlServiceFactory,
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return qflLambdaSvc() },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeUpdate,
			InstanceName:     "test-instance",
			ResourceAInfo:    queueResourceInfo(),
			ResourceBInfo:    queueFunctionResourceInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(qflRoleState())),
			CurrentLinkState: currentLinkState,
		},
		ExpectedOutputMatcher: matchConsumeGrantOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutConsumeInlineAccessPolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

func queueFunctionDestroyIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
], iamservice.Service) {
	existing := fmt.Sprintf(
		`{"Version":"2012-10-17","Statement":[{"Sid":%q,"Effect":"Allow","Action":["sqs:ReceiveMessage"],"Resource":%q}]}`,
		qflQueueSID,
		qflQueueARN,
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
					queueFunctionESMID: {
						Fields: map[string]*core.MappingNode{
							"uuid": core.MappingNodeFromString(qflESMUUID),
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
		ServiceFactoryB:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return qflLambdaSvc() },
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeDestroy,
			InstanceName:     "test-instance",
			ResourceAInfo:    queueResourceInfo(),
			ResourceBInfo:    queueFunctionResourceInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(qflRoleState())),
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
					aws.ToString(input.RoleName) == qflRoleName &&
					aws.ToString(input.PolicyName) == linkutils.InlineAccessPolicyName()
			},
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
	}, iamSvc
}

func TestQueueFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(QueueFunctionLinkUpdateSuite))
}
