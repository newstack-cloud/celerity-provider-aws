package kmsservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS Key Management
// Service (KMS) used by the KMS link implementations. Only the grant operations are needed:
// KMS keys themselves are Cloud Control–backed, so this interface covers the key-policy-side
// authorisation that links manage via KMS grants.
type Service interface {
	// CreateGrant adds a grant to a KMS key, delegating use of the key to a grantee
	// principal for a set of operations independently of the key policy.
	CreateGrant(ctx context.Context, params *kms.CreateGrantInput, optFns ...func(*kms.Options)) (*kms.CreateGrantOutput, error)
	// RevokeGrant deletes a grant from a KMS key.
	RevokeGrant(ctx context.Context, params *kms.RevokeGrantInput, optFns ...func(*kms.Options)) (*kms.RevokeGrantOutput, error)
	// ListGrants lists the grants on a KMS key, optionally filtered by grantee principal.
	ListGrants(ctx context.Context, params *kms.ListGrantsInput, optFns ...func(*kms.Options)) (*kms.ListGrantsOutput, error)
}

// NewService creates a new instance of the AWS Key Management Service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return kms.NewFromConfig(
		*awsConfig,
		kms.WithEndpointResolverV2(
			&kmsEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type kmsEndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *kmsEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params kms.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	kmsAliases := utils.Services["kms"]
	kmsEndpoint, hasKMSEndpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"kms",
		kmsAliases,
	)
	if hasKMSEndpoint && !core.IsScalarNil(kmsEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(kmsEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return kms.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
