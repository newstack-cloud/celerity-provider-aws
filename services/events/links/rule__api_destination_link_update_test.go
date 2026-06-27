//go:build unit

package eventslinks

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	eventsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/events_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

const (
	tadRoleARN        = "arn:aws:iam::123456789012:role/events-target-role"
	tadRoleName       = "events-target-role"
	tadDestinationARN = "arn:aws:events:us-west-2:123456789012:api-destination/order-webhook/abc123"
	tadRoleResource   = "orderWebhookRole"
	tadLinkDataRole   = "orderWebhookRuleRole"
	tadInvokeSID      = "InvokeApiDestinationorderWebhookDestination"
)

type RuleAPIDestinationLinkUpdateSuite struct {
	suite.Suite
}

func ruleAPIDestinationLinkFactory(
	iamSvc iamservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, eventsservice.Service, *aws.Config, eventsservice.Service],
) provider.Link {
	build := RuleAPIDestinationLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, eventsservice.Service, *aws.Config, eventsservice.Service],
	) provider.Link {
		return build(RuleToAPIDestinationLinkDeps(deps))
	}
}

func testLinkContext() provider.LinkContext {
	return plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {"region": core.ScalarFromString("us-west-2")},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func testConfigStore(loader *testutils.MockAWSConfigLoader) pluginutils.ServiceConfigStore[*aws.Config] {
	return utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
}

// The EventBridge rule (resource A). Its targets[] holds the
// entry referencing the API destination, including the roleArn EventBridge assumes
// to invoke it.
func ruleResourceInfoA() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "orderWebhookRule",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"name", core.MappingNodeFromString("order-created-rule"),
				"targets", &core.MappingNode{
					Items: []*core.MappingNode{
						core.MappingNodeFields(
							"id", core.MappingNodeFromString("order-webhook"),
							"arn", core.MappingNodeFromString(tadDestinationARN),
							"roleArn", core.MappingNodeFromString(tadRoleARN),
						),
					},
				},
			),
		},
	}
}

func apiDestinationResourceInfoB() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "orderWebhookDestination",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"name", core.MappingNodeFromString("order-webhook-destination"),
				"arn", core.MappingNodeFromString(tadDestinationARN),
			),
		},
	}
}

func tadRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: tadRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(tadRoleName),
			"arn", core.MappingNodeFromString(tadRoleARN),
		),
	}
}

func tadRoleService() provider.ResourceService {
	return resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(tadRoleState()),
	)
}

func (s *RuleAPIDestinationLinkUpdateSuite) Test_link_update_resources() {
	loader := &testutils.MockAWSConfigLoader{}
	iamSvc := iammock.CreateIamServiceMock()

	emptyData := &core.MappingNode{Fields: map[string]*core.MappingNode{}}

	testCases := []plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		eventsservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		{
			Name:            "does not modify the rule resource (A)",
			Resource:        plugintestutils.LinkUpdateResourceA,
			ServiceFactoryA: eventsmock.CreateEventsServiceMockFactory(),
			ConfigStoreA:    testConfigStore(loader),
			ServiceFactoryB: eventsmock.CreateEventsServiceMockFactory(),
			ConfigStoreB:    testConfigStore(loader),
			Input: &provider.LinkUpdateResourceInput{
				LinkUpdateType:    provider.LinkUpdateTypeCreate,
				ResourceInfo:      ruleResourceInfoA(),
				OtherResourceInfo: apiDestinationResourceInfoB(),
				LinkContext:       testLinkContext(),
			},
			ExpectedOutput: &provider.LinkUpdateResourceOutput{LinkData: emptyData},
		},
		{
			Name:            "does not modify the API destination resource (B)",
			Resource:        plugintestutils.LinkUpdateResourceB,
			ServiceFactoryA: eventsmock.CreateEventsServiceMockFactory(),
			ConfigStoreA:    testConfigStore(loader),
			ServiceFactoryB: eventsmock.CreateEventsServiceMockFactory(),
			ConfigStoreB:    testConfigStore(loader),
			Input: &provider.LinkUpdateResourceInput{
				LinkUpdateType:    provider.LinkUpdateTypeCreate,
				ResourceInfo:      apiDestinationResourceInfoB(),
				OtherResourceInfo: ruleResourceInfoA(),
				LinkContext:       testLinkContext(),
			},
			ExpectedOutput: &provider.LinkUpdateResourceOutput{LinkData: emptyData},
		},
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		testCases,
		ruleAPIDestinationLinkFactory(iamSvc),
		&s.Suite,
	)
}

func (s *RuleAPIDestinationLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}

	createCase, createIam := createIntermediaryTestCase(loader)
	updateCase, updateIam := updateIntermediaryTestCase(loader)
	destroyCase, destroyIam := destroyIntermediaryTestCase(loader)

	cases := []struct {
		testCase plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config,
			eventsservice.Service,
			*aws.Config,
			eventsservice.Service,
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
				eventsservice.Service,
				*aws.Config,
				eventsservice.Service,
			]{c.testCase},
			ruleAPIDestinationLinkFactory(c.iamSvc),
			&s.Suite,
		)
	}
}

// Verifies a PutRolePolicy targets the role's shared
// allocator inline policy and its document grants the link's invoke statement.
func matchPutInlineInvokePolicy(arg any) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != tadRoleName ||
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
		if statement.Sid == tadInvokeSID {
			return true
		}
	}
	return false
}

// Asserts the link records its statement in link data and
// maps it onto the role's spec by Sid (so the role does not strip the grant).
func matchInvokeLinkOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	mappingKey := fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		tadRoleResource,
		linkutils.InlineAccessPolicyName(),
		tadInvokeSID,
	)
	summary := map[string]any{}
	if actual != nil {
		summary["mappingValue"] = actual.ResourceDataMappings[mappingKey]
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[tadLinkDataRole] != nil &&
			actual.LinkData.Fields[tadLinkDataRole].Fields[linkutils.PermissionFieldName] != nil
	}
	expected := map[string]any{
		"mappingValue": linkutils.PermissionFieldPath(tadLinkDataRole),
		"hasStatement": true,
	}
	return plugintestutils.EqualityCheckValues{Expected: expected, Actual: summary}, nil
}

func createIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	eventsservice.Service,
	*aws.Config,
	eventsservice.Service,
], iamservice.Service) {
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		eventsservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		Name:                           "grants invoke access via a new inline allocator policy on create",
		ServiceFactoryA:                eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    ruleResourceInfoA(),
			ResourceBInfo:    apiDestinationResourceInfoB(),
			LinkContext:      testLinkContext(),
			ResourceService:  tadRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchInvokeLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutInlineInvokePolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

func updateIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	eventsservice.Service,
	*aws.Config,
	eventsservice.Service,
], iamservice.Service) {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"` + tadInvokeSID + `","Effect":"Allow","Action":["events:InvokeApiDestination"],"Resource":"arn:old"}]}`
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		eventsservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		Name:                           "replaces the invoke statement in the inline allocator policy on update",
		ServiceFactoryA:                eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeUpdate,
			InstanceName:     "test-instance",
			ResourceAInfo:    ruleResourceInfoA(),
			ResourceBInfo:    apiDestinationResourceInfoB(),
			LinkContext:      testLinkContext(),
			ResourceService:  tadRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: matchInvokeLinkOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return matchPutInlineInvokePolicy(arg) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}, iamSvc
}

func destroyIntermediaryTestCase(
	loader *testutils.MockAWSConfigLoader,
) (plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	eventsservice.Service,
	*aws.Config,
	eventsservice.Service,
], iamservice.Service) {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"` + tadInvokeSID + `","Effect":"Allow","Action":["events:InvokeApiDestination"],"Resource":"` + tadDestinationARN + `"}]}`
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{linkutils.InlineAccessPolicyName()}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
	)

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		eventsservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		Name:                           "removes the invoke statement and deletes the empty inline policy on destroy",
		ServiceFactoryA:                eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                eventsmock.CreateEventsServiceMockFactory(),
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeDestroy,
			InstanceName:     "test-instance",
			ResourceAInfo:    ruleResourceInfoA(),
			ResourceBInfo:    apiDestinationResourceInfoB(),
			LinkContext:      testLinkContext(),
			ResourceService:  tadRoleService(),
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		},
		UpdateActionsCalled: map[string]any{
			"DeleteRolePolicy": func(arg any) bool {
				input, ok := arg.(*iam.DeleteRolePolicyInput)
				return ok &&
					aws.ToString(input.RoleName) == tadRoleName &&
					aws.ToString(input.PolicyName) == linkutils.InlineAccessPolicyName()
			},
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
	}, iamSvc
}

func TestRuleAPIDestinationLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(RuleAPIDestinationLinkUpdateSuite))
}
