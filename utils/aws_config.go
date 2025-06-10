package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

// AWSConfigLoader defines the interface for loading AWS configurations.
type AWSConfigLoader interface {
	LoadDefaultConfig(
		ctx context.Context,
		optFns ...func(*config.LoadOptions) error,
	) (aws.Config, error)
}

// DefaultAWSConfigLoader implements AWSConfigLoader using the AWS SDK.
type DefaultAWSConfigLoader struct{}

func (l *DefaultAWSConfigLoader) LoadDefaultConfig(
	ctx context.Context,
	optFns ...func(*config.LoadOptions) error,
) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, optFns...)
}

// AWSConfigFromProviderContext creates an AWS config from the given
// provider context and environment variables.
func AWSConfigFromProviderContext(
	ctx context.Context,
	providerContext provider.Context,
	env map[string]string,
	loader AWSConfigLoader,
) (*aws.Config, error) {
	if loader == nil {
		loader = &DefaultAWSConfigLoader{}
	}

	opts := []func(*config.LoadOptions) error{}
	opts = append(opts, RegionOptions(providerContext)...)
	opts = append(opts, RetryConfigOptions(providerContext, env)...)
	opts = append(opts, CredentialOptions(providerContext)...)
	opts = append(opts, SharedEndpointOptions(providerContext)...)
	opts = append(opts, EC2MetadataServiceOptions(providerContext, env)...)

	certOpts, err := CertOptions(providerContext, env)
	if err != nil {
		return nil, err
	}
	opts = append(opts, certOpts...)

	opts = append(opts, HTTPClientOptions(providerContext)...)
	opts = append(opts, AssumeRoleOptions(providerContext)...)
	opts = append(opts, AssumeRoleWithWebIdentityOptions(providerContext)...)

	cfg, err := loader.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// RegionOptions returns the region options derived from the given provider context.
func RegionOptions(
	providerContext provider.Context,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	region, hasRegion := providerContext.ProviderConfigVariable("region")
	if hasRegion && !core.IsScalarNil(region) {
		opts = append(opts, config.WithRegion(core.StringValueFromScalar(region)))
	}

	return opts
}

// RetryConfigOptions returns the retry config options derived from the
// given provider context and environment variables.
func RetryConfigOptions(
	providerContext provider.Context,
	env map[string]string,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	retryMode, hasRetryMode := getProviderConfigValueFallbackToEnv(
		providerContext,
		env,
		"retryMode",
		"AWS_RETRY_MODE",
	)
	if hasRetryMode && !core.IsScalarNil(retryMode) {
		retryModeValue := core.StringValueFromScalar(retryMode)
		opts = append(opts, config.WithRetryMode(aws.RetryMode(retryModeValue)))
	}

	maxRetries, hasMaxRetries := providerContext.ProviderConfigVariable("maxRetries")
	if hasMaxRetries && !core.IsScalarNil(maxRetries) {
		maxRetriesValue := core.IntValueFromScalar(maxRetries)
		opts = append(opts, config.WithRetryMaxAttempts(maxRetriesValue))
	}

	return opts
}

// CredentialOptions returns the credential options derived from the given provider context.
func CredentialOptions(
	providerContext provider.Context,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	accessKeyID, hasAccessKeyID := providerContext.ProviderConfigVariable(
		"accessKeyId",
	)
	secretAccessKey, hasSecretAccessKey := providerContext.ProviderConfigVariable(
		"secretAccessKey",
	)
	sessionToken, _ := providerContext.ProviderConfigVariable(
		"sessionToken",
	)

	if hasAccessKeyID && !core.IsScalarNil(accessKeyID) &&
		hasSecretAccessKey && !core.IsScalarNil(secretAccessKey) {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				core.StringValueFromScalar(accessKeyID),
				core.StringValueFromScalar(secretAccessKey),
				core.StringValueFromScalar(sessionToken),
			),
		))
	}

	sharedCredentialsFiles, hasSharedCredentialsFiles := providerContext.ProviderConfigVariable(
		"sharedCredentialsFiles",
	)
	if hasSharedCredentialsFiles && !core.IsScalarNil(sharedCredentialsFiles) {
		credentialsFilesStr := core.StringValueFromScalar(sharedCredentialsFiles)
		credentialsFiles := strings.Split(credentialsFilesStr, ",")
		opts = append(opts, config.WithSharedCredentialsFiles(
			credentialsFiles,
		))
	}

	sharedConfigFiles, hasSharedConfigFiles := providerContext.ProviderConfigVariable(
		"sharedConfigFiles",
	)
	if hasSharedConfigFiles && !core.IsScalarNil(sharedConfigFiles) {
		configFilesStr := core.StringValueFromScalar(sharedConfigFiles)
		configFiles := strings.Split(configFilesStr, ",")
		opts = append(opts, config.WithSharedConfigFiles(
			configFiles,
		))
	}

	profile, hasProfile := providerContext.ProviderConfigVariable(
		"profile",
	)
	if hasProfile && !core.IsScalarNil(profile) {
		opts = append(opts, config.WithSharedConfigProfile(
			core.StringValueFromScalar(profile),
		))
	}

	return opts
}

// SharedEndpointOptions returns the shared endpoint options derived from the given provider context.
func SharedEndpointOptions(
	providerContext provider.Context,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	useFIPSEndpoint, hasUseFIPSEndpoint := providerContext.ProviderConfigVariable(
		"useFIPSEndpoint",
	)
	if hasUseFIPSEndpoint && !core.IsScalarNil(useFIPSEndpoint) {
		useFIPSEndpointValue := core.BoolValueFromScalar(useFIPSEndpoint)
		fipsEndpointState := aws.FIPSEndpointStateDisabled
		if useFIPSEndpointValue {
			fipsEndpointState = aws.FIPSEndpointStateEnabled
		}
		opts = append(opts, config.WithUseFIPSEndpoint(fipsEndpointState))
	}

	useDualStackEndpoint, hasUseDualStackEndpoint := providerContext.ProviderConfigVariable(
		"useDualStackEndpoint",
	)
	if hasUseDualStackEndpoint && !core.IsScalarNil(useDualStackEndpoint) {
		useDualStackEndpointValue := core.BoolValueFromScalar(useDualStackEndpoint)
		dualStackEndpointState := aws.DualStackEndpointStateDisabled
		if useDualStackEndpointValue {
			dualStackEndpointState = aws.DualStackEndpointStateEnabled
		}
		opts = append(opts, config.WithUseDualStackEndpoint(dualStackEndpointState))
	}

	return opts
}

// CertOptions returns the cert options derived from the given provider context and environment variables.
func CertOptions(
	providerContext provider.Context,
	env map[string]string,
) ([]func(*config.LoadOptions) error, error) {
	opts := []func(*config.LoadOptions) error{}

	customCABundle, hasCustomCABundle := getProviderConfigValueFallbackToEnv(
		providerContext,
		env,
		"customCABundle",
		"AWS_CA_BUNDLE",
	)
	if hasCustomCABundle && !core.IsScalarNil(customCABundle) {
		path := core.StringValueFromScalar(customCABundle)
		bundleData, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		opts = append(opts, config.WithCustomCABundle(
			bytes.NewReader(bundleData),
		))
	}

	return opts, nil
}

// HTTPClientOptions returns the http client options derived from the given provider context.
func HTTPClientOptions(
	providerContext provider.Context,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	insecureScalarValue, hasInsecure := providerContext.ProviderConfigVariable(
		"insecure",
	)
	insecure := core.BoolValueFromScalar(insecureScalarValue)

	httpProxyScalarValue, hasHTTPProxy := providerContext.ProviderConfigVariable(
		"httpProxy",
	)
	httpProxy := core.StringValueFromScalar(httpProxyScalarValue)

	httpsProxyScalarValue, hasHTTPSProxy := providerContext.ProviderConfigVariable(
		"httpsProxy",
	)
	httpsProxy := core.StringValueFromScalar(httpsProxyScalarValue)

	// The AWS SDK will automatically pick up the HTTP_PROXY and HTTPS_PROXY
	// environment variables, we only need to configure http proxies if defined
	// in the provider config.
	if (hasInsecure && insecure) ||
		(hasHTTPProxy && !core.IsScalarNil(httpProxyScalarValue)) ||
		(hasHTTPSProxy && !core.IsScalarNil(httpsProxyScalarValue)) {
		customClient := awshttp.NewBuildableClient().WithTransportOptions(
			func(t *http.Transport) {
				if insecure {
					t.TLSClientConfig = &tls.Config{
						InsecureSkipVerify: true,
					}
				}

				finalProxyURL := strings.TrimSpace(httpProxy)
				if finalProxyURL == "" {
					finalProxyURL = strings.TrimSpace(httpsProxy)
				}
				if finalProxyURL != "" {
					parsedProxyURL, err := url.Parse(finalProxyURL)
					if err != nil {
						log.Fatalf("Failed to parse proxy URL: %v", err)
					}

					t.Proxy = http.ProxyURL(parsedProxyURL)
				}
			},
		)
		opts = append(opts, config.WithHTTPClient(customClient))
	}

	return opts
}

// EC2MetadataServiceOptions returns the ec2 metadata service options
// derived from the given provider context and environment variables.
func EC2MetadataServiceOptions(
	providerContext provider.Context,
	env map[string]string,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	ec2MetadataServiceEndpoint, hasEC2MetadataServiceEndpoint := getProviderConfigValueFallbackToEnv(
		providerContext,
		env,
		"ec2MetadataServiceEndpoint",
		"AWS_EC2_METADATA_SERVICE_ENDPOINT",
	)
	ec2MetadataServiceEndpointMode, hasEC2MetadataServiceEndpointMode := getProviderConfigValueFallbackToEnv(
		providerContext,
		env,
		"ec2MetadataServiceEndpointMode",
		"AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE",
	)

	if hasEC2MetadataServiceEndpoint && !core.IsScalarNil(ec2MetadataServiceEndpoint) {
		opts = append(
			opts,
			config.WithEC2IMDSEndpoint(
				core.StringValueFromScalar(ec2MetadataServiceEndpoint),
			),
		)
	}

	if hasEC2MetadataServiceEndpointMode && !core.IsScalarNil(ec2MetadataServiceEndpointMode) {
		opts = append(
			opts,
			config.WithEC2IMDSEndpointMode(
				imds.EndpointModeState(
					imdsEndpointModeStateFromString(
						core.StringValueFromScalar(ec2MetadataServiceEndpointMode),
					),
				),
			),
		)
	}

	return opts
}

func imdsEndpointModeStateFromString(s string) imds.EndpointModeState {
	switch strings.ToLower(s) {
	case "ipv4":
		return imds.EndpointModeStateIPv4
	case "ipv6":
		return imds.EndpointModeStateIPv6
	default:
		return imds.EndpointModeStateUnset
	}
}

// AssumeRoleOptions returns the assume role options derived from the given provider context.
func AssumeRoleOptions(
	providerContext provider.Context,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	assumeRoleARN, hasAssumeRoleARN := providerContext.ProviderConfigVariable(
		"assumeRole.roleArn",
	)

	assumeRoleExternalID, hasAssumeRoleExternalID := providerContext.ProviderConfigVariable(
		"assumeRole.externalId",
	)

	assumeRoleDuration, hasAssumeRoleDuration := providerContext.ProviderConfigVariable(
		"assumeRole.duration",
	)

	assumeRolePolicy, hasAssumeRolePolicy := providerContext.ProviderConfigVariable(
		"assumeRole.policy",
	)

	// Wrap in plugin config so we can use convenience helpers to get dynamic keys
	// based on prefix that emulate complex structures such as arrays.
	pluginConfig := core.PluginConfig(
		providerContext.ProviderConfigVariables(),
	)
	policyARNConfigValues := pluginConfig.SliceFromPrefix("assumeRole.policyArns")

	sessionName, hasSessionName := providerContext.ProviderConfigVariable(
		"assumeRole.sessionName",
	)

	sourceIdentity, hasSourceIdentity := providerContext.ProviderConfigVariable(
		"assumeRole.sourceIdentity",
	)

	tagValues := pluginConfig.MapFromPrefix("assumeRole.tags")

	transitiveTagKeys, hasTransitiveTagKeys := providerContext.ProviderConfigVariable(
		"assumeRole.transitiveTagKeys",
	)

	if hasAssumeRoleARN && !core.IsScalarNil(assumeRoleARN) {
		config.WithAssumeRoleCredentialOptions(
			func(o *stscreds.AssumeRoleOptions) {
				o.RoleARN = core.StringValueFromScalar(assumeRoleARN)

				if hasAssumeRoleExternalID && !core.IsScalarNil(assumeRoleExternalID) {
					o.ExternalID = aws.String(core.StringValueFromScalar(assumeRoleExternalID))
				}

				if hasAssumeRoleDuration && !core.IsScalarNil(assumeRoleDuration) {
					// Validation in the provider config definition will make sure
					// that the duration is a valid duration string so it's safe to ignore
					// the error here.
					duration, _ := time.ParseDuration(core.StringValueFromScalar(assumeRoleDuration))
					o.Duration = duration
				}

				if hasAssumeRolePolicy && !core.IsScalarNil(assumeRolePolicy) {
					o.Policy = aws.String(core.StringValueFromScalar(assumeRolePolicy))
				}

				if len(policyARNConfigValues) > 0 {
					o.PolicyARNs = toSTSPolicyARNs(policyARNConfigValues)
				}

				if hasSessionName && !core.IsScalarNil(sessionName) {
					o.RoleSessionName = core.StringValueFromScalar(sessionName)
				}

				if hasSourceIdentity && !core.IsScalarNil(sourceIdentity) {
					o.SourceIdentity = aws.String(core.StringValueFromScalar(sourceIdentity))
				}

				if len(tagValues) > 0 {
					o.Tags = toSTSTags(tagValues)
				}

				if hasTransitiveTagKeys && !core.IsScalarNil(transitiveTagKeys) {
					o.TransitiveTagKeys = strings.Split(
						core.StringValueFromScalar(transitiveTagKeys),
						",",
					)
				}
			},
		)
	}

	return opts
}

// AssumeRoleWithWebIdentityOptions returns the assume role with web identity
// options derived from the given provider context.
func AssumeRoleWithWebIdentityOptions(
	providerContext provider.Context,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	assumeRoleWebIdentityARN, hasAssumeRoleWebIdentityARN := providerContext.ProviderConfigVariable(
		"assumeRoleWithWebIdentity.roleArn",
	)

	assumeRoleWebIdentityToken, hasAssumeRoleWebIdentityToken := providerContext.ProviderConfigVariable(
		"assumeRoleWithWebIdentity.webIdentityToken",
	)

	assumeRoleWebIdentityTokenFile, hasAssumeRoleWebIdentityTokenFile := providerContext.ProviderConfigVariable(
		"assumeRoleWithWebIdentity.webIdentityTokenFile",
	)

	assumeRoleWebIdentityDuration, hasAssumeRoleWebIdentityDuration := providerContext.ProviderConfigVariable(
		"assumeRoleWithWebIdentity.duration",
	)

	assumeRoleWebIdentityPolicy, hasAssumeRoleWebIdentityPolicy := providerContext.ProviderConfigVariable(
		"assumeRoleWithWebIdentity.policy",
	)

	sessionName, hasSessionName := providerContext.ProviderConfigVariable(
		"assumeRoleWithWebIdentity.sessionName",
	)

	pluginConfig := core.PluginConfig(
		providerContext.ProviderConfigVariables(),
	)

	policyARNConfigValues := pluginConfig.SliceFromPrefix("assumeRoleWithWebIdentity.policyArns")

	if hasAssumeRoleWebIdentityARN && !core.IsScalarNil(assumeRoleWebIdentityARN) {
		config.WithWebIdentityRoleCredentialOptions(
			func(o *stscreds.WebIdentityRoleOptions) {
				o.RoleARN = core.StringValueFromScalar(assumeRoleWebIdentityARN)

				if hasAssumeRoleWebIdentityToken && !core.IsScalarNil(assumeRoleWebIdentityToken) {
					o.TokenRetriever = staticTokenRetriever(
						core.StringValueFromScalar(assumeRoleWebIdentityToken),
					)
				}

				if hasAssumeRoleWebIdentityTokenFile && !core.IsScalarNil(assumeRoleWebIdentityTokenFile) {
					o.TokenRetriever = stscreds.IdentityTokenFile(
						core.StringValueFromScalar(assumeRoleWebIdentityTokenFile),
					)
				}

				if hasAssumeRoleWebIdentityDuration && !core.IsScalarNil(assumeRoleWebIdentityDuration) {
					duration, _ := time.ParseDuration(
						core.StringValueFromScalar(assumeRoleWebIdentityDuration),
					)
					o.Duration = duration
				}

				if hasAssumeRoleWebIdentityPolicy && !core.IsScalarNil(assumeRoleWebIdentityPolicy) {
					o.Policy = aws.String(core.StringValueFromScalar(assumeRoleWebIdentityPolicy))
				}

				if len(policyARNConfigValues) > 0 {
					o.PolicyARNs = toSTSPolicyARNs(policyARNConfigValues)
				}

				if hasSessionName && !core.IsScalarNil(sessionName) {
					o.RoleSessionName = core.StringValueFromScalar(sessionName)
				}
			},
		)
	}

	return opts
}

func getProviderConfigValueFallbackToEnv(
	providerContext provider.Context,
	env map[string]string,
	key string,
	envKey string,
) (*core.ScalarValue, bool) {
	providerConfigValue, hasProviderConfigValue := providerContext.ProviderConfigVariable(key)
	if hasProviderConfigValue && !core.IsScalarNil(providerConfigValue) {
		return providerConfigValue, true
	}

	envValue, hasEnvValue := env[envKey]
	if hasEnvValue {
		return core.ScalarFromString(envValue), true
	}

	return nil, false
}

func toSTSTags(tagValues map[string]*core.ScalarValue) []types.Tag {
	tags := make([]types.Tag, 0, len(tagValues))

	for key := range tagValues {
		tags = append(tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(core.StringValueFromScalar(tagValues[key])),
		})
	}

	return tags
}

func toSTSPolicyARNs(policyARNConfigValues []*core.ScalarValue) []types.PolicyDescriptorType {
	policyARNs := make([]types.PolicyDescriptorType, 0, len(policyARNConfigValues))

	for _, policyARN := range policyARNConfigValues {
		policyARNs = append(policyARNs, types.PolicyDescriptorType{
			Arn: aws.String(core.StringValueFromScalar(policyARN)),
		})
	}

	return policyARNs
}

type staticTokenRetriever string

func (s staticTokenRetriever) GetIdentityToken() ([]byte, error) {
	return []byte(s), nil
}
