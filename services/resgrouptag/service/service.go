package resgrouptagservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the
// Amazon Resource Groups & Tagging service used by resource implementations.
type Service interface {
	// Returns all the tagged or previously tagged resources that are located in the
	// specified Amazon Web Services Region for the account.
	//
	// Depending on what information you want returned, you can also specify the
	// following:
	//
	//   - Filters that specify what tags and resource types you want returned. The
	//     response includes all tags that are associated with the requested resources.
	//
	//   - Information about compliance with the account's effective tag policy. For
	//     more information on tag policies, see [Tag Policies]in the Organizations User Guide.
	//
	// This operation supports pagination, where the response can be sent in multiple
	// pages. You should check the PaginationToken response parameter to determine if
	// there are additional results available to return. Repeat the query, passing the
	// PaginationToken response parameter value as an input to the next request until
	// you recieve a null value. A null value for PaginationToken indicates that there
	// are no more results waiting to be returned.
	//
	// [Tag Policies]: https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_tag-policies.html
	GetResources(ctx context.Context, params *resourcegroupstaggingapi.GetResourcesInput, optFns ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error)
}

// NewService creates a new instance of the Amazon Resource Groups & Tagging service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return resourcegroupstaggingapi.NewFromConfig(
		*awsConfig,
		resourcegroupstaggingapi.WithEndpointResolverV2(
			&resgrouptagEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type resgrouptagEndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *resgrouptagEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params resourcegroupstaggingapi.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	resourcegroupstaggingapiAliases := utils.Services["resourcegroupstaggingapi"]
	resourcegroupstaggingapiEndpoint, hasResourcegroupstaggingapiEndpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"resourcegroupstaggingapi",
		resourcegroupstaggingapiAliases,
	)
	if hasResourcegroupstaggingapiEndpoint && !core.IsScalarNil(resourcegroupstaggingapiEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(resourcegroupstaggingapiEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return resourcegroupstaggingapi.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
