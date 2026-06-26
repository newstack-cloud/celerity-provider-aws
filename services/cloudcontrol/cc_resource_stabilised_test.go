//go:build unit

package cloudcontrol

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	cloudcontrolmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/cloudcontrol_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type CCResourceStabilisedSuite struct {
	suite.Suite
}

func (s *CCResourceStabilisedSuite) Test_has_stabilised() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, cloudcontrolservice.Service]{
		stabilisedSuccessCase(providerCtx, loader),
		stabilisedStatusCase(providerCtx, loader, "not stable when IN_PROGRESS", cctypes.OperationStatusInProgress, false, false),
		stabilisedStatusCase(providerCtx, loader, "errors when FAILED", cctypes.OperationStatusFailed, false, true),
		stabilisedNoTokenCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(testCases, newTestResource, &s.Suite)
}

// Covers the success path: once the operation completes, the
// resource model carries the finalised computed fields, which Stabilised returns so the
// engine persists them (finalising any that were not available at config-complete).
func stabilisedSuccessCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusSuccess,
				Identifier:      aws.String(testQueueURL),
				ResourceModel:   aws.String(`{"QueueUrl":"` + testQueueURL + `","Arn":"` + testQueueARN + `"}`),
			},
		}),
	)

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:           "stable when SUCCESS, finalising computed fields",
		ServiceFactory: func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ConfigStore:    newAWSConfigStore(loader),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID:      "test-instance-id",
			ResourceID:      "test-resource-id",
			ResourceSpec:    specWithToken("status-token"),
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec." + fieldPrimaryIdentifier: core.MappingNodeFromString(testQueueURL),
				"spec.queueUrl":                  core.MappingNodeFromString(testQueueURL),
				"spec.arn":                       core.MappingNodeFromString(testQueueARN),
			},
		},
	}
}

func stabilisedStatusCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
	name string,
	status cctypes.OperationStatus,
	expectedStable bool,
	expectError bool,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{OperationStatus: status},
		}),
	)

	tc := plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:           name,
		ServiceFactory: func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ConfigStore:    newAWSConfigStore(loader),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID:      "test-instance-id",
			ResourceID:      "test-resource-id",
			ResourceSpec:    specWithToken("status-token"),
			ProviderContext: providerCtx,
		},
		ExpectError: expectError,
	}
	if !expectError {
		tc.ExpectedOutput = &provider.ResourceHasStabilisedOutput{Stabilised: expectedStable}
	}
	return tc
}

// Proves the RDS/ElastiCache shape: a
// resource whose endpoint is only assigned once provisioning completes. The endpoint is
// absent at config-complete (early capture) but present in the success event's resource
// model, and Stabilised returns it (alongside every other computed field) so the engine
// finalises it in state.
func (s *CCResourceStabilisedSuite) Test_captures_late_endpoint_on_stabilisation() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	const dbID = "orders-db"
	const dbARN = "arn:aws:rds:us-west-2:123456789012:db:orders-db"
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusSuccess,
				Identifier:      aws.String(dbID),
				ResourceModel: aws.String(
					`{"DBInstanceIdentifier":"` + dbID + `","Arn":"` + dbARN +
						`","Endpoint":{"Address":"orders-db.abc.us-west-2.rds.amazonaws.com","Port":5432}}`,
				),
			},
		}),
	)

	testCase := plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:           "captures the endpoint resolved on stabilisation",
		ServiceFactory: func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ConfigStore:    newAWSConfigStore(loader),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID:      "test-instance-id",
			ResourceID:      "test-resource-id",
			ResourceSpec:    specWithToken("db-token"),
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec." + fieldPrimaryIdentifier: core.MappingNodeFromString(dbID),
				"spec.arn":                       core.MappingNodeFromString(dbARN),
				"spec.endpoint": core.MappingNodeFields(
					"address", core.MappingNodeFromString("orders-db.abc.us-west-2.rds.amazonaws.com"),
					"port", core.MappingNodeFromInt(5432),
				),
			},
		},
	}

	newDatabaseResource := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return CCResource(testDatabaseConfig(), serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		[]plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, cloudcontrolservice.Service]{testCase},
		newDatabaseResource,
		&s.Suite,
	)
}

// An RDS-shaped config whose endpoint is a computed (CFN
// read-only) object only populated once the instance is available.
func testDatabaseConfig() CCResourceConfig {
	return CCResourceConfig{
		BlueprintType: "aws/rds/dbInstance",
		CFNType:       "AWS::RDS::DBInstance",
		Label:         "Test RDS DB Instance",
		Schema: &provider.ResourceDefinitionsSchema{
			Type:  provider.ResourceDefinitionsSchemaTypeObject,
			Label: "DB Instance",
			Attributes: map[string]*provider.ResourceDefinitionsSchema{
				"dbInstanceIdentifier": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "DB Instance Identifier"},
				"arn":                  {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "ARN", Computed: true},
				"endpoint": {
					Type:     provider.ResourceDefinitionsSchemaTypeObject,
					Label:    "Endpoint",
					Computed: true,
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"address": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Address", Computed: true},
						"port":    {Type: provider.ResourceDefinitionsSchemaTypeInteger, Label: "Port", Computed: true},
					},
				},
			},
		},
		Meta: CCResourceMeta{
			PrimaryIdentifierField: "dbInstanceIdentifier",
			ComputedFields:         []string{"arn", "endpoint"},
		},
	}
}

func stabilisedNoTokenCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock()
	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:           "stable when no in-flight request token",
		ServiceFactory: func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ConfigStore:    newAWSConfigStore(loader),
		Input: &provider.ResourceHasStabilisedInput{
			InstanceID:      "test-instance-id",
			ResourceID:      "test-resource-id",
			ResourceSpec:    &core.MappingNode{Fields: map[string]*core.MappingNode{}},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{Stabilised: true},
	}
}

func specWithToken(token string) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			fieldRequestToken: core.MappingNodeFromString(token),
		},
	}
}

func TestCCResourceStabilisedSuite(t *testing.T) {
	suite.Run(t, new(CCResourceStabilisedSuite))
}
