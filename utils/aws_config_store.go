package utils

import (
	"context"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/two-hundred/celerity/libs/blueprint/core"
	"github.com/two-hundred/celerity/libs/blueprint/provider"
	"github.com/two-hundred/celerity/libs/plugin-framework/sdk/pluginutils"
)

// AWSConfigCreator is a function that produces an AWS config
// from the given provider context and environment variables.
type AWSConfigCreator func(
	ctx context.Context,
	providerContext provider.Context,
	env map[string]string,
	loader AWSConfigLoader,
) (*aws.Config, error)

// AWSConfigStore is a store for AWS config that is used to derive and cache
// AWS config on a per-session basis.
type AWSConfigStore struct {
	// A copy of the environment variables for the current AWS provider process.
	env             map[string]string
	createAWSConfig AWSConfigCreator
	loader          AWSConfigLoader
	cache           map[string]*aws.Config
	mu              sync.RWMutex
}

// NewAWSConfigStore creates a new store for deriving and caching AWS config.
func NewAWSConfigStore(
	env []string,
	createAWSConfig AWSConfigCreator,
	loader AWSConfigLoader,
) *AWSConfigStore {
	envMap := envMapFromStrings(env)
	return &AWSConfigStore{
		env:             envMap,
		createAWSConfig: createAWSConfig,
		cache:           make(map[string]*aws.Config),
		mu:              sync.RWMutex{},
	}
}

// FromProviderContext creates configuration to be used to create AWS SDK clients.
func (s *AWSConfigStore) FromProviderContext(
	ctx context.Context,
	providerContext provider.Context,
) (*aws.Config, error) {
	// A session ID is passed from the client (e.g. Celerity CLI) to the host
	// and then to plugins through the context variables.
	// In the AWS provider, we use the session ID to cache AWS config
	// to avoid having to rebuild the config for each request to a plugin
	// action.
	sessionID, hasSessionID := getSessionID(ctx, providerContext)
	if hasSessionID {
		awsConfig, inCache := s.getFromCache(sessionID)
		if inCache {
			return awsConfig, nil
		}
	}

	awsConf, err := s.createAWSConfig(ctx, providerContext, s.env, s.loader)

	s.setInCache(sessionID, awsConf)
	return awsConf, err
}

func (s *AWSConfigStore) getFromCache(sessionID string) (*aws.Config, bool) {
	s.mu.RLock()
	awsConfig, ok := s.cache[sessionID]
	s.mu.RUnlock()
	return awsConfig, ok
}

func (s *AWSConfigStore) setInCache(sessionID string, awsConfig *aws.Config) {
	s.mu.Lock()
	s.cache[sessionID] = awsConfig
	s.mu.Unlock()
}

func getSessionID(
	ctx context.Context,
	providerContext provider.Context,
) (string, bool) {
	// First, try the context variables as a part of the blueprint framework
	// provider context.
	sessionID, hasSessionID := providerContext.ContextVariable(pluginutils.SessionIDKey)
	if hasSessionID {
		return core.StringValueFromScalar(sessionID), true
	}

	// If no session ID is found in the context variables, try the go context.
	goCtxSessionID, goCtxHasSessionId := ctx.Value(pluginutils.ContextSessionIDKey).(string)
	if goCtxHasSessionId {
		return goCtxSessionID, true
	}

	return "", false
}

func envMapFromStrings(env []string) map[string]string {
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		envMap[parts[0]] = parts[1]
	}
	return envMap
}
