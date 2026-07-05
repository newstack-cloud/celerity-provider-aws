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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type SSMParameterDataSourceSuite struct {
	suite.Suite
}

type parameterFetchTestCase struct {
	Name           string
	ServiceFactory func(awsConfig *aws.Config, providerContext provider.Context) ssmservice.Service
	ConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
	Input          *provider.DataSourceFetchInput
	ExpectedOutput *provider.DataSourceFetchOutput
	ExpectError    bool
}

func (s *SSMParameterDataSourceSuite) Test_fetch() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			pluginutils.SessionIDKey: core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []parameterFetchTestCase{
		fetchStringParameterTestCase(providerCtx, loader),
		fetchSecureStringOmitsValueTestCase(providerCtx, loader),
		fetchMissingNameTestCase(providerCtx, loader),
		fetchGetParameterErrorTestCase(providerCtx, loader),
	}

	for _, tc := range testCases {
		s.Run(tc.Name, func() {
			dataSource := ParameterDataSource(tc.ServiceFactory, tc.ConfigStore)
			output, err := dataSource.Fetch(context.Background(), tc.Input)
			if tc.ExpectError {
				s.Error(err)
				s.Nil(output)
			} else {
				s.NoError(err)
				s.NotNil(output)
				s.Equal(tc.ExpectedOutput.Data, output.Data)
			}
		})
	}
}

func configStore(loader *testutils.MockAWSConfigLoader) pluginutils.ServiceConfigStore[*aws.Config] {
	return utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
}

func fetchStringParameterTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) parameterFetchTestCase {
	return parameterFetchTestCase{
		Name: "fetches a String parameter including its value",
		ServiceFactory: ssmmock.CreateSSMServiceMockFactory(
			ssmmock.WithGetParameterOutput(&ssm.GetParameterOutput{
				Parameter: &ssmtypes.Parameter{
					ARN:      aws.String(testParameterARN),
					Name:     aws.String(testParameterName),
					Type:     ssmtypes.ParameterTypeString,
					Value:    aws.String("db.internal.example.com"),
					Version:  4,
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
					},
				},
			}),
		),
		ConfigStore: configStore(loader),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("name", testParameterName),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"name":        core.MappingNodeFromString(testParameterName),
				"arn":         core.MappingNodeFromString(testParameterARN),
				"type":        core.MappingNodeFromString("String"),
				"value":       core.MappingNodeFromString("db.internal.example.com"),
				"version":     core.MappingNodeFromInt(4),
				"dataType":    core.MappingNodeFromString("text"),
				"tier":        core.MappingNodeFromString("Standard"),
				"description": core.MappingNodeFromString("database host"),
			},
		},
	}
}

func fetchSecureStringOmitsValueTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) parameterFetchTestCase {
	return parameterFetchTestCase{
		Name: "omits the value for a SecureString parameter",
		ServiceFactory: ssmmock.CreateSSMServiceMockFactory(
			ssmmock.WithGetParameterOutput(&ssm.GetParameterOutput{
				Parameter: &ssmtypes.Parameter{
					ARN:     aws.String(testParameterARN),
					Name:    aws.String(testParameterName),
					Type:    ssmtypes.ParameterTypeSecureString,
					Value:   aws.String("AQICAHencryptedblob"),
					Version: 1,
				},
			}),
			ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{
				Parameters: []ssmtypes.ParameterMetadata{
					{
						Name:  aws.String(testParameterName),
						Type:  ssmtypes.ParameterTypeSecureString,
						Tier:  ssmtypes.ParameterTierStandard,
						KeyId: aws.String("alias/aws/ssm"),
					},
				},
			}),
		),
		ConfigStore: configStore(loader),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("name", testParameterName),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"name":    core.MappingNodeFromString(testParameterName),
				"arn":     core.MappingNodeFromString(testParameterARN),
				"type":    core.MappingNodeFromString("SecureString"),
				"version": core.MappingNodeFromInt(1),
				"tier":    core.MappingNodeFromString("Standard"),
				"keyId":   core.MappingNodeFromString("alias/aws/ssm"),
			},
		},
	}
}

func fetchMissingNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) parameterFetchTestCase {
	return parameterFetchTestCase{
		Name:           "returns error when name filter is missing",
		ServiceFactory: ssmmock.CreateSSMServiceMockFactory(),
		ConfigStore:    configStore(loader),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("region", "us-west-2"),
			},
		},
		ExpectError: true,
	}
}

func fetchGetParameterErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) parameterFetchTestCase {
	return parameterFetchTestCase{
		Name: "returns error when GetParameter fails",
		ServiceFactory: ssmmock.CreateSSMServiceMockFactory(
			ssmmock.WithGetParameterError(errTestGetParameter),
		),
		ConfigStore: configStore(loader),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("name", testParameterName),
			},
		},
		ExpectError: true,
	}
}

func TestSSMParameterDataSourceSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterDataSourceSuite))
}
