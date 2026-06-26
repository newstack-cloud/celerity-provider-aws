//go:build integration

package e2e

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/stretchr/testify/require"
)

// PublishAlias publishes a version of an existing function and creates an alias
// pointing to it (via the raw Lambda API), registering cleanup. Returns the alias
// name and version. Used to fixture the alias data source, which reads existing
// infrastructure during change staging.
func (h *Harness) PublishAlias(t *testing.T, functionName, aliasName string) string {
	t.Helper()
	client := lambda.NewFromConfig(h.AWSConfig)
	published, err := client.PublishVersion(h.Ctx, &lambda.PublishVersionInput{
		FunctionName: aws.String(functionName),
	})
	require.NoError(t, err)
	version := aws.ToString(published.Version)

	_, err = client.CreateAlias(h.Ctx, &lambda.CreateAliasInput{
		FunctionName:    aws.String(functionName),
		Name:            aws.String(aliasName),
		FunctionVersion: aws.String(version),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = client.DeleteAlias(h.Ctx, &lambda.DeleteAliasInput{
			FunctionName: aws.String(functionName),
			Name:         aws.String(aliasName),
		})
	})
	return version
}

// CreateFunctionURL creates a function URL config on an existing function (raw
// Lambda API), registering cleanup. Used to fixture the function URL data source.
func (h *Harness) CreateFunctionURL(t *testing.T, functionName string) {
	t.Helper()
	client := lambda.NewFromConfig(h.AWSConfig)
	_, err := client.CreateFunctionUrlConfig(h.Ctx, &lambda.CreateFunctionUrlConfigInput{
		FunctionName: aws.String(functionName),
		AuthType:     lambdatypes.FunctionUrlAuthTypeNone,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = client.DeleteFunctionUrlConfig(h.Ctx, &lambda.DeleteFunctionUrlConfigInput{
			FunctionName: aws.String(functionName),
		})
	})
}

// SigningProfileVersionARN creates an AWS Signer signing profile (raw Signer API),
// registering cleanup, and returns its version ARN. This is used both to fixture the code
// signing config data source and, passed as a blueprint variable, to create a Code
// Signing Config resource in the codeSigningConfig link blueprint.
func (h *Harness) SigningProfileVersionARN(t *testing.T, label string) string {
	t.Helper()
	// Signer profile names allow only [0-9a-zA-Z_].
	profileName := strings.ReplaceAll(h.Name(label), "-", "_")
	client := signer.NewFromConfig(h.AWSConfig)
	out, err := client.PutSigningProfile(h.Ctx, &signer.PutSigningProfileInput{
		ProfileName: aws.String(profileName),
		PlatformId:  aws.String("AWSLambda-SHA384-ECDSA"),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = client.CancelSigningProfile(h.Ctx, &signer.CancelSigningProfileInput{
			ProfileName: aws.String(profileName),
		})
	})
	return aws.ToString(out.ProfileVersionArn)
}

// CreateCodeSigningConfig creates a Code Signing Config from a signing profile
// version ARN (raw Lambda API), registering cleanup, and returns its ARN. Used to
// fixture the code signing config data source.
func (h *Harness) CreateCodeSigningConfig(t *testing.T, profileVersionARN string) string {
	t.Helper()
	client := lambda.NewFromConfig(h.AWSConfig)
	out, err := client.CreateCodeSigningConfig(h.Ctx, &lambda.CreateCodeSigningConfigInput{
		AllowedPublishers: &lambdatypes.AllowedPublishers{
			SigningProfileVersionArns: []string{profileVersionARN},
		},
	})
	require.NoError(t, err)
	cscARN := aws.ToString(out.CodeSigningConfig.CodeSigningConfigArn)
	t.Cleanup(func() {
		_, _ = client.DeleteCodeSigningConfig(h.Ctx, &lambda.DeleteCodeSigningConfigInput{
			CodeSigningConfigArn: aws.String(cscARN),
		})
	})
	return cscARN
}
