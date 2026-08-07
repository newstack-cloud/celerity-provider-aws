//go:build unit

package ssm

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ssmmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ssm_mock"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

const (
	testReplicaRegion       = "eu-west-1"
	testReplicaParameterARN = "arn:aws:ssm:eu-west-1:123456789012:parameter/my-app/prod/db-host"
)

// SSMParameterResourceRegionSuite verifies that the optional "region" spec field
// re-targets the SDK client for every lifecycle operation, enabling per-region
// parameter replicas.
type SSMParameterResourceRegionSuite struct {
	suite.Suite
	providerCtx provider.Context
}

func (s *SSMParameterResourceRegionSuite) SetupTest() {
	s.providerCtx = plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func (s *SSMParameterResourceRegionSuite) Test_create_targets_spec_region() {
	service := ssmmock.CreateSSMServiceMock(replicaParameterMockOptions()...)
	var capturedRegion string
	resource := resourceWithRegionCapture(service, &capturedRegion)

	_, err := resource.Deploy(
		context.Background(),
		deployInput(s.providerCtx, replicaParameterSpec()),
	)

	s.NoError(err)
	s.Equal(testReplicaRegion, capturedRegion)
	service.AssertCalled(&s.Suite, "PutParameter")
}

func (s *SSMParameterResourceRegionSuite) Test_create_defaults_to_provider_region() {
	service := ssmmock.CreateSSMServiceMock(replicaParameterMockOptions()...)
	var capturedRegion string
	resource := resourceWithRegionCapture(service, &capturedRegion)

	spec := replicaParameterSpec()
	delete(spec.Fields, "region")
	_, err := resource.Deploy(context.Background(), deployInput(s.providerCtx, spec))

	s.NoError(err)
	s.Equal("us-west-2", capturedRegion)
}

func (s *SSMParameterResourceRegionSuite) Test_update_targets_spec_region() {
	service := ssmmock.CreateSSMServiceMock(replicaParameterMockOptions()...)
	var capturedRegion string
	resource := resourceWithRegionCapture(service, &capturedRegion)

	currentState := replicaParameterSpec()
	currentState.Fields["arn"] = core.MappingNodeFromString(testReplicaParameterARN)
	_, err := resource.Deploy(
		context.Background(),
		updateInput(s.providerCtx, currentState, replicaParameterSpec()),
	)

	s.NoError(err)
	s.Equal(testReplicaRegion, capturedRegion)
}

func (s *SSMParameterResourceRegionSuite) Test_destroy_targets_region_recorded_in_state() {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithDeleteParameterOutput(&ssm.DeleteParameterOutput{}),
	)
	var capturedRegion string
	resource := resourceWithRegionCapture(service, &capturedRegion)

	err := resource.Destroy(
		context.Background(),
		destroyInput(s.providerCtx, replicaParameterSpec()),
	)

	s.NoError(err)
	s.Equal(testReplicaRegion, capturedRegion)
	service.AssertCalled(&s.Suite, "DeleteParameter")
}

func (s *SSMParameterResourceRegionSuite) Test_get_external_state_targets_and_reports_spec_region() {
	service := ssmmock.CreateSSMServiceMock(replicaParameterMockOptions()...)
	var capturedRegion string
	resource := resourceWithRegionCapture(service, &capturedRegion)

	output, err := resource.GetExternalState(
		context.Background(),
		&provider.ResourceGetExternalStateInput{
			ProviderContext:     s.providerCtx,
			CurrentResourceSpec: replicaParameterSpec(),
		},
	)

	s.NoError(err)
	s.Equal(testReplicaRegion, capturedRegion)
	s.Equal(
		core.MappingNodeFromString(testReplicaRegion),
		output.ResourceSpecState.Fields["region"],
	)
}

func resourceWithRegionCapture(
	service ssmservice.Service,
	capturedRegion *string,
) provider.Resource {
	factory := func(cfg *aws.Config, _ provider.Context) ssmservice.Service {
		*capturedRegion = cfg.Region
		return service
	}
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		&testutils.MockAWSConfigLoader{},
		utils.AWSConfigCacheKey,
	)
	return ParameterResource(factory, configStore)
}

func replicaParameterSpec() *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":   core.MappingNodeFromString(testParameterName),
			"type":   core.MappingNodeFromString("String"),
			"value":  core.MappingNodeFromString("db.internal.example.com"),
			"region": core.MappingNodeFromString(testReplicaRegion),
		},
	}
}

func replicaParameterMockOptions() []ssmmock.SSMServiceMockOption {
	return []ssmmock.SSMServiceMockOption{
		ssmmock.WithPutParameterOutput(&ssm.PutParameterOutput{Version: 1}),
		ssmmock.WithGetParameterOutput(&ssm.GetParameterOutput{
			Parameter: &ssmtypes.Parameter{
				ARN:     aws.String(testReplicaParameterARN),
				Name:    aws.String(testParameterName),
				Type:    ssmtypes.ParameterTypeString,
				Value:   aws.String("db.internal.example.com"),
				Version: 1,
			},
		}),
		ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{
			Parameters: []ssmtypes.ParameterMetadata{
				{
					Name: aws.String(testParameterName),
					Type: ssmtypes.ParameterTypeString,
					Tier: ssmtypes.ParameterTierStandard,
				},
			},
		}),
		ssmmock.WithListTagsForResourceOutput(&ssm.ListTagsForResourceOutput{}),
	}
}

func TestSSMParameterResourceRegionSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterResourceRegionSuite))
}
