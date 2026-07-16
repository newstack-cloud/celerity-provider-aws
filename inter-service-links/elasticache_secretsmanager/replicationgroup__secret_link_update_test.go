//go:build unit

package elasticachesecretsmanager

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	elasticachemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/elasticache_mock"
	secretsmanagermock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/secretsmanager_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type ReplicationGroupSecretLinkUpdateSuite struct {
	suite.Suite
}

func matchModifyReplicationGroup(strategy elasticachetypes.AuthTokenUpdateStrategyType) func(arg any) bool {
	return func(arg any) bool {
		input, ok := arg.(*elasticache.ModifyReplicationGroupInput)
		if !ok {
			return false
		}
		return aws.ToString(input.ReplicationGroupId) == rgReplicationGroupID &&
			aws.ToString(input.AuthToken) == rgAuthTokenValue &&
			input.AuthTokenUpdateStrategy == strategy &&
			aws.ToBool(input.ApplyImmediately)
	}
}

func (s *ReplicationGroupSecretLinkUpdateSuite) harnessCase(
	name string,
	input *provider.LinkUpdateResourceInput,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	cloudcontrolservice.Service,
] {
	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		cloudcontrolservice.Service,
	]{
		Name:            name,
		Resource:        plugintestutils.LinkUpdateResourceA,
		ServiceFactoryA: noopCloudControlServiceFactory,
		ConfigStoreA:    testConfigStore(loader),
		ServiceFactoryB: noopCloudControlServiceFactory,
		ConfigStoreB:    testConfigStore(loader),
		Input:           input,
		ExpectedOutputMatcher: func(
			actual *provider.LinkUpdateResourceOutput,
		) (plugintestutils.EqualityCheckValues, error) {
			return plugintestutils.EqualityCheckValues{Expected: true, Actual: actual != nil}, nil
		},
	}
}

func (s *ReplicationGroupSecretLinkUpdateSuite) Test_applies_auth_token_with_rotate_strategy_on_create() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock(
		elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
	)
	secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock(
		secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(rgAuthTokenValue),
		}),
	)

	input := &provider.LinkUpdateResourceInput{
		LinkUpdateType:    provider.LinkUpdateTypeCreate,
		ResourceInfo:      replicationGroupInfo(nil),
		OtherResourceInfo: secretInfo(),
		LinkContext:       testLinkContext(),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		[]plugintestutils.LinkUpdateResourceTestCase[
			*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{s.harnessCase("applies auth token on create", input, loader)},
		replicationGroupSecretLinkFactory(elastiCacheSvc, secretsManagerSvc),
		&s.Suite,
	)

	secretsManagerSvc.AssertCalledWith(
		&s.Suite, "GetSecretValue", 0, plugintestutils.Any,
		func(arg any) bool {
			in, ok := arg.(*secretsmanager.GetSecretValueInput)
			return ok && aws.ToString(in.SecretId) == rgSecretARN
		},
	)
	// ROTATE is the only strategy ElastiCache accepts for a server without an existing AUTH
	// token, so it is the default on first configuration.
	elastiCacheSvc.AssertCalledWith(
		&s.Suite, "ModifyReplicationGroup", 0, plugintestutils.Any,
		matchModifyReplicationGroup(elasticachetypes.AuthTokenUpdateStrategyTypeRotate),
	)
}

func (s *ReplicationGroupSecretLinkUpdateSuite) Test_rotates_auth_token_on_update() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock(
		elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
	)
	secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock(
		secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(rgAuthTokenValue),
		}),
	)

	input := &provider.LinkUpdateResourceInput{
		LinkUpdateType:    provider.LinkUpdateTypeUpdate,
		ResourceInfo:      replicationGroupInfo(nil),
		OtherResourceInfo: secretInfo(),
		LinkContext:       testLinkContext(),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		[]plugintestutils.LinkUpdateResourceTestCase[
			*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{s.harnessCase("rotates auth token on update", input, loader)},
		replicationGroupSecretLinkFactory(elastiCacheSvc, secretsManagerSvc),
		&s.Suite,
	)

	elastiCacheSvc.AssertCalledWith(
		&s.Suite, "ModifyReplicationGroup", 0, plugintestutils.Any,
		matchModifyReplicationGroup(elasticachetypes.AuthTokenUpdateStrategyTypeRotate),
	)
}

func (s *ReplicationGroupSecretLinkUpdateSuite) Test_annotation_overrides_update_strategy() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock(
		elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
	)
	secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock(
		secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(rgAuthTokenValue),
		}),
	)

	annotations := map[string]*core.MappingNode{
		"aws.elasticache.secretsmanager.authTokenUpdateStrategy": core.MappingNodeFromString("SET"),
	}
	input := &provider.LinkUpdateResourceInput{
		LinkUpdateType:    provider.LinkUpdateTypeUpdate,
		ResourceInfo:      replicationGroupInfo(annotations),
		OtherResourceInfo: secretInfo(),
		LinkContext:       testLinkContext(),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		[]plugintestutils.LinkUpdateResourceTestCase[
			*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{s.harnessCase("annotation overrides strategy", input, loader)},
		replicationGroupSecretLinkFactory(elastiCacheSvc, secretsManagerSvc),
		&s.Suite,
	)

	// The annotation forces SET on an update (which would otherwise default to ROTATE),
	// retiring the previous token after a rotation window.
	elastiCacheSvc.AssertCalledWith(
		&s.Suite, "ModifyReplicationGroup", 0, plugintestutils.Any,
		matchModifyReplicationGroup(elasticachetypes.AuthTokenUpdateStrategyTypeSet),
	)
}

// A SET annotation is safe as a static value: on first configuration there is no previous
// token to retire, so the link falls back to ROTATE instead of failing.
func (s *ReplicationGroupSecretLinkUpdateSuite) Test_set_strategy_falls_back_to_rotate_on_first_configuration() {
	annotations := map[string]*core.MappingNode{
		"aws.elasticache.secretsmanager.authTokenUpdateStrategy": core.MappingNodeFromString("SET"),
	}

	strategy := resolveAuthTokenUpdateStrategy(&provider.LinkUpdateResourceInput{
		LinkUpdateType: provider.LinkUpdateTypeCreate,
		ResourceInfo:   replicationGroupInfo(annotations),
	})
	s.Equal(elasticachetypes.AuthTokenUpdateStrategyTypeRotate, strategy)
}

func (s *ReplicationGroupSecretLinkUpdateSuite) Test_destroy_is_a_no_op() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock()
	secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock()

	input := &provider.LinkUpdateResourceInput{
		LinkUpdateType:    provider.LinkUpdateTypeDestroy,
		ResourceInfo:      replicationGroupInfo(nil),
		OtherResourceInfo: secretInfo(),
		LinkContext:       testLinkContext(),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		[]plugintestutils.LinkUpdateResourceTestCase[
			*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{s.harnessCase("destroy is a no-op", input, loader)},
		replicationGroupSecretLinkFactory(elastiCacheSvc, secretsManagerSvc),
		&s.Suite,
	)

	elastiCacheSvc.AssertNotCalled(&s.Suite, "ModifyReplicationGroup")
	secretsManagerSvc.AssertNotCalled(&s.Suite, "GetSecretValue")
}

// Test_auth_token_value_never_appears_in_output_state asserts that the link data returned from a
// successful create records only the secret ARN and the strategy, never the token value.
func (s *ReplicationGroupSecretLinkUpdateSuite) Test_auth_token_value_never_appears_in_output_state() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheFactory := elasticachemock.CreateElastiCacheServiceMockFactory(
		elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
	)
	secretsManagerFactory := secretsmanagermock.CreateSecretsManagerServiceMockFactory(
		secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(rgAuthTokenValue),
		}),
	)

	actions := &replicationGroupSecretLinkActions{
		elastiCacheServiceFactory:    elastiCacheFactory,
		secretsManagerServiceFactory: secretsManagerFactory,
		awsConfigStore:               testConfigStore(loader),
	}

	output, err := actions.UpdateResourceA(context.Background(), &provider.LinkUpdateResourceInput{
		LinkUpdateType:    provider.LinkUpdateTypeCreate,
		ResourceInfo:      replicationGroupInfo(nil),
		OtherResourceInfo: secretInfo(),
		LinkContext:       testLinkContext(),
	})
	s.Require().NoError(err)
	s.Require().NotNil(output)

	// The secret ARN and strategy marker are recorded, but never the token value.
	authTokenData := output.LinkData.Fields["authToken"]
	s.Require().NotNil(authTokenData)
	s.Equal(rgSecretARN, core.StringValue(authTokenData.Fields["secretArn"]))
	s.True(core.BoolValue(authTokenData.Fields["configured"]))
	assertNoTokenValue(&s.Suite, output.LinkData)
}

func assertNoTokenValue(s *suite.Suite, node *core.MappingNode) {
	if node == nil {
		return
	}
	s.NotEqual(rgAuthTokenValue, core.StringValue(node))
	for _, field := range node.Fields {
		assertNoTokenValue(s, field)
	}
	for _, item := range node.Items {
		assertNoTokenValue(s, item)
	}
}

func (s *ReplicationGroupSecretLinkUpdateSuite) Test_validate_auth_token() {
	testCases := []struct {
		name          string
		token         string
		errorContains string
	}{
		{
			name:  "accepts a valid token",
			token: rgAuthTokenValue,
		},
		{
			name:          "rejects a token shorter than 16 characters",
			token:         "too-short",
			errorContains: "between 16 and 128 characters",
		},
		{
			name:          "rejects a token longer than 128 characters",
			token:         strings.Repeat("a", 129),
			errorContains: "between 16 and 128 characters",
		},
		{
			name:          "rejects non-printable-ASCII characters",
			token:         "valid-length-token-with-é-in-it",
			errorContains: "printable ASCII",
		},
		{
			name:          "rejects forbidden special characters",
			token:         "valid-length-token-with-@-in-it",
			errorContains: `must not contain the characters '/', '"', '@' or '%'`,
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			err := validateAuthToken(testCase.token, rgSecretARN)
			if testCase.errorContains == "" {
				s.NoError(err)
				return
			}
			s.Require().Error(err)
			s.Contains(err.Error(), testCase.errorContains)
			s.Contains(err.Error(), rgSecretARN)
			s.NotContains(err.Error(), testCase.token)
		})
	}
}

func TestReplicationGroupSecretLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(ReplicationGroupSecretLinkUpdateSuite))
}
