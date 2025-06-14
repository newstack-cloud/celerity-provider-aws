package utils

import (
	"context"
	"math/rand"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/celerity-provider-aws/internal/testutils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type AWSConfigStoreTestSuite struct {
	suite.Suite
	mockConfigCreator AWSConfigCreator
}

func (s *AWSConfigStoreTestSuite) SetupTest() {
	// Create a mock config creator that returns a simple config
	s.mockConfigCreator = func(
		ctx context.Context,
		providerContext provider.Context,
		env map[string]string,
		loader AWSConfigLoader,
	) (*aws.Config, error) {
		var sessionID string
		if v := ctx.Value(pluginutils.ContextSessionIDKey); v != nil {
			sessionID, _ = v.(string)
		} else {
			sessionID = "no-session"
		}
		appID := "test-config-" + sessionID
		if sessionID == "no-session" {
			// Return a unique config each time
			appID += "-" + strconv.Itoa(rand.Int())
		}
		return &aws.Config{
			Region: "us-west-2",
			AppID:  appID,
		}, nil
	}
}

func (s *AWSConfigStoreTestSuite) Test_from_provider_context() {
	env := []string{
		"AWS_EC2_METADATA_SERVICE_ENDPOINT=http://169.254.169.254",
	}
	providerConfig := map[string]*core.ScalarValue{
		"region":               core.ScalarFromString("us-west-2"),
		"retryMode":            core.ScalarFromString("standard"),
		"maxRetries":           core.ScalarFromInt(3),
		"accessKeyId":          core.ScalarFromString("test-access-key"),
		"secretAccessKey":      core.ScalarFromString("test-secret-key"),
		"useFIPSEndpoint":      core.ScalarFromBool(true),
		"useDualStackEndpoint": core.ScalarFromBool(true),
	}
	sessionID := "test-session-1"

	store := NewAWSConfigStore(env, s.mockConfigCreator, &testutils.MockAWSConfigLoader{})
	providerContext := plugintestutils.NewTestProviderContext("aws", providerConfig, nil)

	// Create a context with the session ID
	ctx := context.WithValue(context.Background(), pluginutils.ContextSessionIDKey, sessionID)

	// First call should not be cached
	cfg, err := store.FromProviderContext(ctx, providerContext)
	s.NoError(err)
	s.NotNil(cfg)

	// Second call should be cached
	cachedCfg, err := store.FromProviderContext(ctx, providerContext)
	s.NoError(err)
	s.NotNil(cachedCfg)
	s.Equal(cfg, cachedCfg, "Cached config should be identical to original config")

	// Verify cache behavior with different session IDs
	otherSessionCtx := context.WithValue(
		context.Background(),
		pluginutils.ContextSessionIDKey,
		sessionID+"-other",
	)
	otherCfg, err := store.FromProviderContext(otherSessionCtx, providerContext)
	s.NoError(err)
	s.NotNil(otherCfg)
	s.NotEqual(cfg, otherCfg, "Configs for different sessions should be different")
}

func (s *AWSConfigStoreTestSuite) Test_from_provider_context_no_session_id() {
	store := NewAWSConfigStore([]string{}, s.mockConfigCreator, &testutils.MockAWSConfigLoader{})
	providerContext := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		nil,
	)

	// Call without session ID in context
	cfg, err := store.FromProviderContext(context.Background(), providerContext)
	s.NoError(err)
	s.NotNil(cfg)

	// Second call should still work but won't use cache
	cfg2, err := store.FromProviderContext(context.Background(), providerContext)
	s.NoError(err)
	s.NotNil(cfg2)
	s.NotEqual(cfg, cfg2, "Configs without session ID should not be cached")
}

func TestAWSConfigStoreSuite(t *testing.T) {
	suite.Run(t, new(AWSConfigStoreTestSuite))
}
