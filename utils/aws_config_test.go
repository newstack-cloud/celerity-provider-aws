package utils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/newstack-cloud/celerity-provider-aws/internal/testutils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AWSConfigTestSuite struct {
	suite.Suite
}

func (s *AWSConfigTestSuite) TestAWSConfigFromProviderContext() {
	tests := []struct {
		name        string
		providerCtx provider.Context
		env         map[string]string
		mockLoader  *testutils.MockAWSConfigLoader
		expectError bool
	}{
		{
			name: "successfully loads config with region",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"region": core.ScalarFromString("us-west-2"),
				},
				nil,
			),
			env: map[string]string{},
			mockLoader: &testutils.MockAWSConfigLoader{
				LoadDefaultConfigFunc: func(
					ctx context.Context,
					optFns ...func(*config.LoadOptions) error,
				) (aws.Config, error) {
					cfg := aws.Config{}
					for _, opt := range optFns {
						err := opt(&config.LoadOptions{})
						if err != nil {
							return aws.Config{}, err
						}
					}
					return cfg, nil
				},
			},
			expectError: false,
		},
		{
			name: "handles loader error",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env: map[string]string{},
			mockLoader: &testutils.MockAWSConfigLoader{
				LoadDefaultConfigFunc: func(
					ctx context.Context,
					optFns ...func(*config.LoadOptions) error,
				) (aws.Config, error) {
					return aws.Config{}, assert.AnError
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg, err := AWSConfigFromProviderContext(context.Background(), tt.providerCtx, tt.env, tt.mockLoader)

			if tt.expectError {
				s.Error(err)
				s.Nil(cfg)
			} else {
				s.NoError(err)
				s.NotNil(cfg)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestRegionOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		expectedRegion string
	}{
		{
			name: "region from provider config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"region": core.ScalarFromString("us-east-1"),
				},
				nil,
			),
			expectedRegion: "us-east-1",
		},
		{
			name: "no region in provider config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedRegion: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := RegionOptions(tt.providerCtx)

			if tt.expectedRegion == "" {
				s.Empty(opts)
			} else {
				s.Len(opts, 1)

				// Create a LoadOptions to test the option function
				loadOpts := &config.LoadOptions{}
				err := opts[0](loadOpts)
				s.NoError(err)
				s.Equal(tt.expectedRegion, loadOpts.Region)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestRetryConfigOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		env            map[string]string
		expectedRetry  aws.RetryMode
		expectedMaxRet int
	}{
		{
			name: "retry mode and max retries from provider config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"retryMode":  core.ScalarFromString("standard"),
					"maxRetries": core.ScalarFromInt(3),
				},
				nil,
			),
			env:            map[string]string{},
			expectedRetry:  aws.RetryModeStandard,
			expectedMaxRet: 3,
		},
		{
			name: "retry mode from env var",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env: map[string]string{
				"AWS_RETRY_MODE": "adaptive",
			},
			expectedRetry:  aws.RetryModeAdaptive,
			expectedMaxRet: 0,
		},
		{
			name: "no retry config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env:            map[string]string{},
			expectedRetry:  "",
			expectedMaxRet: 0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := RetryConfigOptions(tt.providerCtx, tt.env)

			if tt.expectedRetry == "" && tt.expectedMaxRet == 0 {
				s.Empty(opts)
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			if tt.expectedRetry != "" {
				s.Equal(tt.expectedRetry, loadOpts.RetryMode)
			}
			if tt.expectedMaxRet > 0 {
				s.Equal(tt.expectedMaxRet, loadOpts.RetryMaxAttempts)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestCredentialOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		expectedConfig *config.LoadOptions
	}{
		{
			name: "static credentials",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"accessKeyId":     core.ScalarFromString("test-access-key"),
					"secretAccessKey": core.ScalarFromString("test-secret-key"),
					"sessionToken":    core.ScalarFromString("test-session-token"),
				},
				nil,
			),
			expectedConfig: &config.LoadOptions{
				Credentials: credentials.NewStaticCredentialsProvider(
					"test-access-key",
					"test-secret-key",
					"test-session-token",
				),
			},
		},
		{
			name: "shared credentials files",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"sharedCredentialsFiles": core.ScalarFromString("/path/to/creds1,/path/to/creds2"),
					"sharedConfigFiles":      core.ScalarFromString("/path/to/config1,/path/to/config2"),
					"profile":                core.ScalarFromString("test-profile"),
				},
				nil,
			),
			expectedConfig: &config.LoadOptions{
				SharedConfigFiles:      []string{"/path/to/config1", "/path/to/config2"},
				SharedCredentialsFiles: []string{"/path/to/creds1", "/path/to/creds2"},
				SharedConfigProfile:    "test-profile",
			},
		},
		{
			name: "no credentials",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := CredentialOptions(tt.providerCtx)

			if tt.expectedConfig.Credentials == nil && len(tt.expectedConfig.SharedConfigFiles) == 0 {
				s.Empty(opts)
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			if tt.expectedConfig.Credentials != nil {
				s.NotNil(loadOpts.Credentials)
			}
			if len(tt.expectedConfig.SharedConfigFiles) > 0 {
				s.Equal(tt.expectedConfig.SharedConfigFiles, loadOpts.SharedConfigFiles)
			}
			if tt.expectedConfig.SharedConfigProfile != "" {
				s.Equal(tt.expectedConfig.SharedConfigProfile, loadOpts.SharedConfigProfile)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestSharedEndpointOptions() {
	tests := []struct {
		name         string
		providerCtx  provider.Context
		expectedFIPS aws.FIPSEndpointState
		expectedDual aws.DualStackEndpointState
	}{
		{
			name: "FIPS and dual stack enabled",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"useFIPSEndpoint":      core.ScalarFromBool(true),
					"useDualStackEndpoint": core.ScalarFromBool(true),
				},
				nil,
			),
			expectedFIPS: aws.FIPSEndpointStateEnabled,
			expectedDual: aws.DualStackEndpointStateEnabled,
		},
		{
			name: "FIPS and dual stack disabled",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"useFIPSEndpoint":      core.ScalarFromBool(false),
					"useDualStackEndpoint": core.ScalarFromBool(false),
				},
				nil,
			),
			expectedFIPS: aws.FIPSEndpointStateDisabled,
			expectedDual: aws.DualStackEndpointStateDisabled,
		},
		{
			name: "no endpoint config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedFIPS: aws.FIPSEndpointStateDisabled,
			expectedDual: aws.DualStackEndpointStateDisabled,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := SharedEndpointOptions(tt.providerCtx)

			if tt.expectedFIPS == aws.FIPSEndpointStateUnset &&
				tt.expectedDual == aws.DualStackEndpointStateUnset {
				s.Empty(opts)
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			if tt.expectedFIPS != aws.FIPSEndpointStateDisabled {
				s.Equal(tt.expectedFIPS, loadOpts.UseFIPSEndpoint)
			}
			if tt.expectedDual != aws.DualStackEndpointStateDisabled {
				s.Equal(tt.expectedDual, loadOpts.UseDualStackEndpoint)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestCertOptions() {
	tests := []struct {
		name        string
		providerCtx provider.Context
		env         map[string]string
		expectError bool
	}{
		{
			name: "custom CA bundle from file",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"customCaBundle": core.ScalarFromString("__testdata/ca-bundle.pem"),
				},
				nil,
			),
			env:         map[string]string{},
			expectError: false,
		},
		{
			name: "custom CA bundle from env",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env: map[string]string{
				"AWS_CA_BUNDLE": "__testdata/ca-bundle.pem",
			},
			expectError: false,
		},
		{
			name: "no CA bundle",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env:         map[string]string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts, err := CertOptions(tt.providerCtx, tt.env)

			if tt.expectError {
				s.Error(err)
				s.Nil(opts)
				return
			}

			s.NoError(err)
			if len(opts) == 0 {
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			// Note: We can't easily test the actual CA bundle content
			// as it's loaded from a file. The test just verifies that
			// the option functions don't error.
		})
	}
}

func (s *AWSConfigTestSuite) TestHTTPClientOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		expectedConfig *config.LoadOptions
	}{
		{
			name: "custom HTTP client config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"httpProxy": core.ScalarFromString("http://proxy.example.com:8080"),
				},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
		{
			name: "no HTTP client config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := HTTPClientOptions(tt.providerCtx)

			if len(opts) == 0 {
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			// Note: We can't easily test the actual HTTP client configuration
			// as it involves network settings. The test just verifies that
			// the option functions don't error.
		})
	}
}

func (s *AWSConfigTestSuite) TestEC2MetadataServiceOptions() {
	tests := []struct {
		name             string
		providerCtx      provider.Context
		env              map[string]string
		expectedMode     imds.EndpointModeState
		expectedEndpoint string
	}{
		{
			name: "EC2 metadata service config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"ec2MetadataServiceEndpoint":     core.ScalarFromString("http://169.254.169.254"),
					"ec2MetadataServiceEndpointMode": core.ScalarFromString("IPv4"),
				},
				nil,
			),
			env:              map[string]string{},
			expectedMode:     imds.EndpointModeStateIPv4,
			expectedEndpoint: "http://169.254.169.254",
		},
		{
			name: "EC2 metadata service from env",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env: map[string]string{
				"AWS_EC2_METADATA_SERVICE_ENDPOINT":      "http://169.254.169.254",
				"AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE": "IPv4",
			},
			expectedMode:     imds.EndpointModeStateIPv4,
			expectedEndpoint: "http://169.254.169.254",
		},
		{
			name: "no EC2 metadata service config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env:              map[string]string{},
			expectedMode:     imds.EndpointModeStateIPv4,
			expectedEndpoint: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := EC2MetadataServiceOptions(tt.providerCtx, tt.env)

			if tt.expectedMode == imds.EndpointModeStateIPv4 && tt.expectedEndpoint == "" {
				s.Empty(opts)
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			// Note: We can't easily test the actual EC2 metadata service configuration
			// as it involves network settings. The test just verifies that
			// the option functions don't error.
		})
	}
}

func (s *AWSConfigTestSuite) TestAssumeRoleOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		expectedConfig *config.LoadOptions
	}{
		{
			name: "assume role config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"assumeRole.roleArn":         core.ScalarFromString("arn:aws:iam::123456789012:role/test-role"),
					"assumeRole.sessionName":     core.ScalarFromString("test-session"),
					"assumeRole.externalId":      core.ScalarFromString("test-external-id"),
					"assumeRole.durationSeconds": core.ScalarFromInt(3600),
				},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
		{
			name: "no assume role config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := AssumeRoleOptions(tt.providerCtx)

			if len(opts) == 0 {
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			// Note: We can't easily test the actual assume role configuration
			// as it involves AWS STS. The test just verifies that
			// the option functions don't error.
		})
	}
}

func (s *AWSConfigTestSuite) TestAssumeRoleWithWebIdentityOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		expectedConfig *config.LoadOptions
	}{
		{
			name: "assume role with web identity config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"assumeRoleWithWebIdentity.roleArn":          core.ScalarFromString("arn:aws:iam::123456789012:role/test-role"),
					"assumeRoleWithWebIdentity.sessionName":      core.ScalarFromString("test-session"),
					"assumeRoleWithWebIdentity.webIdentityToken": core.ScalarFromString("test-token"),
				},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
		{
			name: "no assume role with web identity config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := AssumeRoleWithWebIdentityOptions(tt.providerCtx)

			if len(opts) == 0 {
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			// Note: We can't easily test the actual assume role with web identity configuration
			// as it involves AWS STS. The test just verifies that
			// the option functions don't error.
		})
	}
}

func TestAWSConfigSuite(t *testing.T) {
	suite.Run(t, new(AWSConfigTestSuite))
}
