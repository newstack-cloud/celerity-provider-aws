package cloudcontrolservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS Cloud Control API
// used by the generic Cloud Control–backed resource engine.
type Service interface {
	// CreateResource creates the specified resource from its desired state.
	CreateResource(
		ctx context.Context,
		params *cloudcontrol.CreateResourceInput,
		optFns ...func(*cloudcontrol.Options),
	) (*cloudcontrol.CreateResourceOutput, error)

	// UpdateResource updates the specified resource by applying a JSON Patch document.
	UpdateResource(
		ctx context.Context,
		params *cloudcontrol.UpdateResourceInput,
		optFns ...func(*cloudcontrol.Options),
	) (*cloudcontrol.UpdateResourceOutput, error)

	// GetResource returns the current state of the specified resource.
	GetResource(
		ctx context.Context,
		params *cloudcontrol.GetResourceInput,
		optFns ...func(*cloudcontrol.Options),
	) (*cloudcontrol.GetResourceOutput, error)

	// ListResources returns the resource descriptions for the resources of the
	// specified type. Cloud Control offers no server-side property filtering, so
	// callers paginate and filter client-side.
	ListResources(
		ctx context.Context,
		params *cloudcontrol.ListResourcesInput,
		optFns ...func(*cloudcontrol.Options),
	) (*cloudcontrol.ListResourcesOutput, error)

	// DeleteResource deletes the specified resource.
	DeleteResource(
		ctx context.Context,
		params *cloudcontrol.DeleteResourceInput,
		optFns ...func(*cloudcontrol.Options),
	) (*cloudcontrol.DeleteResourceOutput, error)

	// GetResourceRequestStatus returns the current status of a resource operation request.
	GetResourceRequestStatus(
		ctx context.Context,
		params *cloudcontrol.GetResourceRequestStatusInput,
		optFns ...func(*cloudcontrol.Options),
	) (*cloudcontrol.GetResourceRequestStatusOutput, error)
}

// NewService creates a new instance of the AWS Cloud Control API service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return cloudcontrol.NewFromConfig(
		*awsConfig,
		cloudcontrol.WithEndpointResolverV2(
			&cloudControlEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type cloudControlEndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *cloudControlEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params cloudcontrol.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	aliases := utils.Services["cloudcontrolapi"]
	endpoint, hasEndpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"cloudcontrolapi",
		aliases,
	)
	if hasEndpoint && !core.IsScalarNil(endpoint) {
		u, err := url.Parse(core.StringValueFromScalar(endpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return cloudcontrol.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
