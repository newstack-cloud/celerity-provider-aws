//go:build unit

package cloudcontrol

import (
	"fmt"
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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CloudFormation marks read-only properties at whatever depth they occur, and the
// generated schema follows suit, but a resource type's metadata lists top-level field
// names. AWS::RDS::DBCluster is the case this was found on: the metadata lists
// "masterUserSecret" while only its "secretArn" is computed, the rest being
// author-supplied.
//
// Reporting the parent makes the engine reject the whole deployment, because a computed
// value comes back that the schema never declared. Any blueprint using
// manageMasterUserPassword on an Aurora cluster hit that.
func clusterLikeSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type: provider.ResourceDefinitionsSchemaTypeObject,
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"dbClusterArn": {
				Type:     provider.ResourceDefinitionsSchemaTypeString,
				Computed: true,
			},
			"masterUsername": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
			},
			"masterUserSecret": {
				Type: provider.ResourceDefinitionsSchemaTypeObject,
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"kmsKeyId": {
						Type: provider.ResourceDefinitionsSchemaTypeString,
					},
					"secretArn": {
						Type:     provider.ResourceDefinitionsSchemaTypeString,
						Computed: true,
					},
				},
			},
			"nothingComputed": {
				Type: provider.ResourceDefinitionsSchemaTypeObject,
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"authored": {
						Type: provider.ResourceDefinitionsSchemaTypeString,
					},
				},
			},
		},
	}
}

func TestComputedPathsForFieldExpandsToTheComputedLeaf(t *testing.T) {
	paths := computedPathsForField(clusterLikeSchema(), "masterUserSecret")

	require.Equal(t, []string{"masterUserSecret.secretArn"}, paths)
}

// A field that is itself computed keeps reporting under its own name, which is every
// other entry in every metadata list.
func TestComputedPathsForFieldKeepsAComputedFieldAsItIs(t *testing.T) {
	paths := computedPathsForField(clusterLikeSchema(), "dbClusterArn")

	require.Equal(t, []string{"dbClusterArn"}, paths)
}

// Claiming the parent is what the engine rejects, so an object with nothing computed
// inside it reports nothing at all.
func TestComputedPathsForFieldReportsNothingWhenNoDescendantIsComputed(t *testing.T) {
	paths := computedPathsForField(clusterLikeSchema(), "nothingComputed")

	require.Empty(t, paths)
}

// A resource type whose schema the engine has not augmented must behave as it did before,
// rather than silently dropping its computed fields.
func TestComputedPathsForFieldFallsBackToTheFieldNameWithoutASchema(t *testing.T) {
	require.Equal(t, []string{"someField"}, computedPathsForField(nil, "someField"))
	require.Equal(
		t,
		[]string{"unknownField"},
		computedPathsForField(clusterLikeSchema(), "unknownField"),
	)
}

// The expanded paths have to reach the deployment's computed values, not just be
// computed correctly.
//
// Everything above tests computedPathsForField directly, because the input space is
// schema shapes and a full deploy per shape would cost far more than it proves. That
// leaves one thing those tests cannot show: that the expansion is used. It has a single
// caller, so dropping it would keep all four passing while a computed leaf under an
// authored parent silently stopped being reported.
//
// The shape here is the real one that motivated the expansion: an Aurora cluster
// declaring masterUserSecret, where the parent is authored and only secretArn is
// assigned by AWS.
type CCComputedPathsWiringSuite struct {
	suite.Suite
}

func (s *CCComputedPathsWiringSuite) Test_a_computed_leaf_under_an_authored_parent_is_reported() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	const clusterID = "orders-cluster"
	const secretARN = "arn:aws:secretsmanager:::secret/orders-master"
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithCreateResourceOutput(&awscc.CreateResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				RequestToken: aws.String("tok"),
				Identifier:   aws.String(clusterID),
			},
		}),
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusSuccess,
				Identifier:      aws.String(clusterID),
			},
		}),
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(clusterID),
				Properties: aws.String(fmt.Sprintf(
					`{"DbClusterArn":"arn:aws:rds:::cluster/orders","MasterUsername":"admin",`+
						`"MasterUserSecret":{"SecretArn":%q}}`, secretARN,
				)),
			},
		}),
	)

	config := CCResourceConfig{
		BlueprintType: "aws/test/cluster",
		CFNType:       "AWS::Test::Cluster",
		Label:         "Test Cluster",
		Schema:        clusterLikeSchema(),
		Meta: CCResourceMeta{
			PrimaryIdentifierField:  "dbClusterArn",
			PrimaryIdentifierFields: []string{"dbClusterArn"},
			// The authored parent is named, not the computed leaf. Expanding it is
			// what puts masterUserSecret.secretArn into the computed values.
			ComputedFields: []string{"dbClusterArn", "masterUserSecret"},
		},
	}

	testCase := plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "a computed leaf under an authored parent is reported by its full path",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input: deployInput(providerCtx, core.MappingNodeFields(
			"masterUsername", core.MappingNodeFromString("admin"),
		), nil),
		ExpectedOutputMatcher: func(
			actual *provider.ResourceDeployOutput,
		) (plugintestutils.EqualityCheckValues, error) {
			return plugintestutils.EqualityCheckValues{
				Expected: core.MappingNodeFromString(secretARN),
				Actual:   actual.ComputedFieldValues["spec.masterUserSecret.secretArn"],
			}, nil
		},
	}

	createResource := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return CCResource(config, serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		[]plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{testCase},
		createResource,
		&s.Suite,
	)
}

func TestCCComputedPathsWiringSuite(t *testing.T) {
	suite.Run(t, new(CCComputedPathsWiringSuite))
}
