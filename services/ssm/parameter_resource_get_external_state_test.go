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

type SSMParameterResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *SSMParameterResourceGetExternalStateSuite) Test_get_external_state() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ssmservice.Service]{
		getStringParameterExternalStateTestCase(providerCtx, loader),
		getSecureStringParameterExternalStateTestCase(providerCtx, loader),
		getMissingParameterExternalStateTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		ParameterResource,
		&s.Suite,
	)
}

func getStringParameterExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithGetParameterOutput(&ssm.GetParameterOutput{
			Parameter: &ssmtypes.Parameter{
				ARN:      aws.String(testParameterARN),
				Name:     aws.String(testParameterName),
				Type:     ssmtypes.ParameterTypeString,
				Value:    aws.String("db.internal.example.com"),
				Version:  5,
				DataType: aws.String("text"),
			},
		}),
		ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{
			Parameters: []ssmtypes.ParameterMetadata{
				{
					Name:        aws.String(testParameterName),
					Type:        ssmtypes.ParameterTypeString,
					Tier:        ssmtypes.ParameterTierStandard,
					Description: aws.String("database host"),
					Version:     5,
				},
			},
		}),
		ssmmock.WithListTagsForResourceOutput(&ssm.ListTagsForResourceOutput{
			TagList: []ssmtypes.Tag{
				{Key: aws.String("Environment"), Value: aws.String("production")},
				// Bluelink provenance tags must be filtered out of external state.
				{Key: aws.String("bluelink:blueprint-instance:name"), Value: aws.String("orders")},
			},
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ssmservice.Service]{
		Name:           "reads a String parameter",
		ServiceFactory: func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: getExternalStateInput(providerCtx),
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"name":        core.MappingNodeFromString(testParameterName),
					"type":        core.MappingNodeFromString("String"),
					"value":       core.MappingNodeFromString("db.internal.example.com"),
					"arn":         core.MappingNodeFromString(testParameterARN),
					"version":     core.MappingNodeFromInt(5),
					"dataType":    core.MappingNodeFromString("text"),
					"description": core.MappingNodeFromString("database host"),
					"tier":        core.MappingNodeFromString("Standard"),
					"tags": core.MappingNodeFields(
						"Environment", core.MappingNodeFromString("production"),
					),
				},
			},
		},
	}
}

func getSecureStringParameterExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithGetParameterOutput(&ssm.GetParameterOutput{
			Parameter: &ssmtypes.Parameter{
				ARN:     aws.String(testParameterARN),
				Name:    aws.String(testParameterName),
				Type:    ssmtypes.ParameterTypeSecureString,
				Value:   aws.String("decrypted-secret"),
				Version: 2,
			},
		}),
		ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{
			Parameters: []ssmtypes.ParameterMetadata{
				{
					Name:  aws.String(testParameterName),
					Type:  ssmtypes.ParameterTypeSecureString,
					Tier:  ssmtypes.ParameterTierStandard,
					KeyId: aws.String("alias/my-key"),
				},
			},
		}),
		ssmmock.WithListTagsForResourceOutput(&ssm.ListTagsForResourceOutput{}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ssmservice.Service]{
		Name:           "reads a SecureString parameter with decryption",
		ServiceFactory: func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: getExternalStateInput(providerCtx),
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"name":        core.MappingNodeFromString(testParameterName),
					"type":        core.MappingNodeFromString("SecureString"),
					"secureValue": core.MappingNodeFromString("decrypted-secret"),
					"arn":         core.MappingNodeFromString(testParameterARN),
					"version":     core.MappingNodeFromInt(2),
					"tier":        core.MappingNodeFromString("Standard"),
					"keyId":       core.MappingNodeFromString("alias/my-key"),
				},
			},
		},
	}
}

func getMissingParameterExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithGetParameterError(&ssmtypes.ParameterNotFound{}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ssmservice.Service]{
		Name:           "returns empty state when the parameter does not exist",
		ServiceFactory: func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: getExternalStateInput(providerCtx),
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
		},
	}
}

func getExternalStateInput(providerCtx provider.Context) *provider.ResourceGetExternalStateInput {
	return &provider.ResourceGetExternalStateInput{
		ProviderContext: providerCtx,
		CurrentResourceSpec: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"name": core.MappingNodeFromString(testParameterName),
			},
		},
	}
}

func TestSSMParameterResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterResourceGetExternalStateSuite))
}
