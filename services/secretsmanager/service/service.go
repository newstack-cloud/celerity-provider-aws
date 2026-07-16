package secretsmanagerservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS Secrets
// Manager service used by the hand-written link implementations. Secrets themselves
// are Cloud Control–backed; this interface covers reading a secret's value at deploy
// time (which Cloud Control cannot surface, as generateSecretString/secretString are
// write-only and never exposed as a computed output).
type Service interface {
	// GetSecretValue retrieves the contents of the encrypted SecretString (or
	// SecretBinary) field for a secret, identified by name or ARN.
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// NewService creates a new instance of the AWS Secrets Manager service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return secretsmanager.NewFromConfig(
		*awsConfig,
		secretsmanager.WithEndpointResolverV2(
			&secretsManagerEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type secretsManagerEndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *secretsManagerEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params secretsmanager.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	secretsManagerAliases := utils.Services["secretsmanager"]
	secretsManagerEndpoint, hasSecretsManagerEndpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"secretsmanager",
		secretsManagerAliases,
	)
	if hasSecretsManagerEndpoint && !core.IsScalarNil(secretsManagerEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(secretsManagerEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return secretsmanager.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
