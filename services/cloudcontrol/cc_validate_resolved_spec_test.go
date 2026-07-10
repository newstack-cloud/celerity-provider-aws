//go:build unit

package cloudcontrol

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	cloudcontrolmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/cloudcontrol_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

// Covers the ValidateResolvedSpec behaviour overlay wiring: a registered
// deploy-time constraint (the RDS DB subnet group minimum here) must abort
// both create and update before any Cloud Control call is made.
type CCValidateResolvedSpecSuite struct {
	suite.Suite
}

func testDBSubnetGroupConfig() CCResourceConfig {
	return CCResourceConfig{
		BlueprintType: "aws/rds/dbSubnetGroup",
		CFNType:       "AWS::RDS::DBSubnetGroup",
		Label:         "AWS RDS DBSubnetGroup",
		Schema: &provider.ResourceDefinitionsSchema{
			Type:  provider.ResourceDefinitionsSchemaTypeObject,
			Label: "AWS RDS DBSubnetGroup",
			Attributes: map[string]*provider.ResourceDefinitionsSchema{
				"dbSubnetGroupName": {
					Type:         provider.ResourceDefinitionsSchemaTypeString,
					Label:        "DB Subnet Group Name",
					MustRecreate: true,
					Nullable:     true,
				},
				"dbSubnetGroupDescription": {
					Type:  provider.ResourceDefinitionsSchemaTypeString,
					Label: "Description",
				},
				"subnetIds": {
					Type:  provider.ResourceDefinitionsSchemaTypeArray,
					Label: "Subnet Ids",
					Items: &provider.ResourceDefinitionsSchema{
						Type: provider.ResourceDefinitionsSchemaTypeString,
					},
				},
			},
		},
		Meta: CCResourceMeta{
			PrimaryIdentifierField: "dbSubnetGroupName",
			CreateOnlyFields:       []string{"dbSubnetGroupName"},
		},
	}
}

func newTestDBSubnetGroupResource(
	serviceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
	configStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	return CCResource(
		testDBSubnetGroupConfig(),
		serviceFactory,
		mockResourceGroupTaggingServiceFactory,
		configStore,
	)
}

func (s *CCValidateResolvedSpecSuite) Test_deploy_aborts_when_resolved_spec_fails_validation() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	// No Cloud Control stubs: validation must fail before any API call.
	service := cloudcontrolmock.CreateCloudControlServiceMock()

	invalidSpec := core.MappingNodeFields(
		"dbSubnetGroupName",
		core.MappingNodeFromString("orders-db-subnets"),
		"dbSubnetGroupDescription",
		core.MappingNodeFromString("Orders DB subnet group"),
		"subnetIds",
		core.MappingNodeItems(core.MappingNodeFromString("subnet-1")),
	)

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		{
			Name: "create aborts on a single subnet ID",
			ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) cloudcontrolservice.Service {
				return service
			},
			ServiceMockCalls: &service.MockCalls,
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceDeployInput{
				InstanceID: "test-instance-id",
				ResourceID: "test-resource-id",
				Changes: &provider.Changes{
					AppliedResourceInfo: provider.ResourceInfo{
						ResourceID:   "test-resource-id",
						ResourceName: "OrdersDBSubnetGroup",
						InstanceID:   "test-instance-id",
						ResourceWithResolvedSubs: &provider.ResolvedResource{
							Type: &schema.ResourceTypeWrapper{
								Value: "aws/rds/dbSubnetGroup",
							},
							Spec: invalidSpec,
						},
					},
				},
				ProviderContext: providerCtx,
			},
			ExpectError: true,
			SaveActionsNotCalled: []string{
				"CreateResource",
			},
		},
		{
			Name: "update aborts on a single subnet ID",
			ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) cloudcontrolservice.Service {
				return service
			},
			ServiceMockCalls: &service.MockCalls,
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceDeployInput{
				InstanceID: "test-instance-id",
				ResourceID: "test-resource-id",
				Changes: &provider.Changes{
					AppliedResourceInfo: provider.ResourceInfo{
						ResourceID:   "test-resource-id",
						ResourceName: "OrdersDBSubnetGroup",
						InstanceID:   "test-instance-id",
						CurrentResourceState: &state.ResourceState{
							ResourceID: "test-resource-id",
							Name:       "OrdersDBSubnetGroup",
							InstanceID: "test-instance-id",
							SpecData: core.MappingNodeFields(
								"dbSubnetGroupName",
								core.MappingNodeFromString("orders-db-subnets"),
							),
						},
						ResourceWithResolvedSubs: &provider.ResolvedResource{
							Type: &schema.ResourceTypeWrapper{
								Value: "aws/rds/dbSubnetGroup",
							},
							Spec: invalidSpec,
						},
					},
					ModifiedFields: []provider.FieldChange{
						{
							FieldPath: "spec.subnetIds",
						},
					},
				},
				ProviderContext: providerCtx,
			},
			ExpectError: true,
			SaveActionsNotCalled: []string{
				"UpdateResource",
				"GetResource",
			},
		},
	}

	plugintestutils.RunResourceDeployTestCases(testCases, newTestDBSubnetGroupResource, &s.Suite)
}

func TestCCValidateResolvedSpecSuite(t *testing.T) {
	suite.Run(t, new(CCValidateResolvedSpecSuite))
}
