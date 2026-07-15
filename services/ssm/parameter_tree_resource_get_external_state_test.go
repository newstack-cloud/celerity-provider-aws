//go:build unit

package ssm

import (
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

type SSMParameterTreeResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *SSMParameterTreeResourceGetExternalStateSuite) Test_get_external_state() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	// The mock is shared across cases so that value read-back can be asserted globally:
	// the tree must never call GetParameter (with or without decryption).
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithDescribeParametersOutput(describeTreeParametersOutput()),
		ssmmock.WithListTagsForResourceOutput(&ssm.ListTagsForResourceOutput{
			TagList: []ssmtypes.Tag{
				{Key: aws.String("Environment"), Value: aws.String("production")},
				{Key: aws.String("bluelink:instanceId"), Value: aws.String("test-instance-id")},
			},
		}),
	)

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ssmservice.Service]{
		{
			Name:           "returns structural metadata without values",
			ServiceFactory: func(*aws.Config, provider.Context) ssmservice.Service { return service },
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceGetExternalStateInput{
				ProviderContext:     providerCtx,
				CurrentResourceSpec: treeSpecData(),
			},
			ExpectedOutput: &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"path":   core.MappingNodeFromString(testTreePath),
						"region": core.MappingNodeFromString("us-west-2"),
						"parameters": {
							Fields: map[string]*core.MappingNode{
								"apiToken": core.MappingNodeFields(
									"arn", core.MappingNodeFromString(treeParameterARN("apiToken")),
									"type", core.MappingNodeFromString("SecureString"),
								),
								"db/host": core.MappingNodeFields(
									"arn", core.MappingNodeFromString(treeParameterARN("db/host")),
									"type", core.MappingNodeFromString("String"),
								),
								"logLevel": core.MappingNodeFields(
									"arn", core.MappingNodeFromString(treeParameterARN("logLevel")),
									"type", core.MappingNodeFromString("String"),
								),
							},
						},
						// The Bluelink provenance tag is filtered out.
						"tags": core.MappingNodeFields(
							"Environment", core.MappingNodeFromString("production"),
						),
					},
				},
			},
			ExpectError: false,
		},
		{
			Name:           "returns empty external state when no managed parameters exist",
			ServiceFactory: func(*aws.Config, provider.Context) ssmservice.Service { return service },
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceGetExternalStateInput{
				ProviderContext: providerCtx,
				CurrentResourceSpec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"path": core.MappingNodeFromString(testTreePath),
						"values": core.MappingNodeFields(
							"unprovisioned", core.MappingNodeFromString("value"),
						),
					},
				},
			},
			ExpectedOutput: emptyExternalState(),
			ExpectError:    false,
		},
		{
			Name:           "returns error when path is missing",
			ServiceFactory: func(*aws.Config, provider.Context) ssmservice.Service { return service },
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceGetExternalStateInput{
				ProviderContext: providerCtx,
				CurrentResourceSpec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				},
			},
			ExpectError: true,
		},
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		ParameterTreeResource,
		&s.Suite,
	)

	// The no-read-back guarantee: values must never be fetched or decrypted when
	// reporting external state.
	service.MockCalls.AssertNotCalled(&s.Suite, "GetParameter")
}

func TestSSMParameterTreeResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterTreeResourceGetExternalStateSuite))
}
