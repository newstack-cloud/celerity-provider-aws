package ssmservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS Systems
// Manager (SSM) service used by the SSM resource and data source implementations.
type Service interface {
	// PutParameter creates or updates a parameter in Systems Manager Parameter Store.
	PutParameter(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)
	// GetParameter returns a single parameter, optionally decrypting a SecureString value.
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
	// DeleteParameter deletes a parameter from Systems Manager.
	DeleteParameter(ctx context.Context, params *ssm.DeleteParameterInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParameterOutput, error)
	// DescribeParameters returns parameter metadata (type, tier, key id, description,
	// allowed pattern, policies) that GetParameter does not surface.
	DescribeParameters(ctx context.Context, params *ssm.DescribeParametersInput, optFns ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error)
	// AddTagsToResource adds or overwrites tags on a parameter.
	AddTagsToResource(ctx context.Context, params *ssm.AddTagsToResourceInput, optFns ...func(*ssm.Options)) (*ssm.AddTagsToResourceOutput, error)
	// RemoveTagsFromResource removes tags from a parameter.
	RemoveTagsFromResource(ctx context.Context, params *ssm.RemoveTagsFromResourceInput, optFns ...func(*ssm.Options)) (*ssm.RemoveTagsFromResourceOutput, error)
	// ListTagsForResource lists the tags attached to a parameter.
	ListTagsForResource(ctx context.Context, params *ssm.ListTagsForResourceInput, optFns ...func(*ssm.Options)) (*ssm.ListTagsForResourceOutput, error)
}

// NewService creates a new instance of the AWS Systems Manager service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return ssm.NewFromConfig(
		*awsConfig,
		ssm.WithEndpointResolverV2(
			&ssmEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type ssmEndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *ssmEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params ssm.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	ssmAliases := utils.Services["ssm"]
	ssmEndpoint, hasSSMEndpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"ssm",
		ssmAliases,
	)
	if hasSSMEndpoint && !core.IsScalarNil(ssmEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(ssmEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return ssm.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
