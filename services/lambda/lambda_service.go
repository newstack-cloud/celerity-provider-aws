package lambda

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/two-hundred/celerity-provider-aws/utils"
	"github.com/two-hundred/celerity/libs/blueprint/core"
	"github.com/two-hundred/celerity/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS Lambda service
// used by the Lambda resource implementation.
type Service interface {
	// Returns information about the function or function version, with a link to
	// download the deployment package that's valid for 10 minutes. If you specify a
	// function version, only details that are specific to that version are returned.
	GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
	// Deletes a Lambda function. To delete a specific function version, use the
	// Qualifier parameter. Otherwise, all versions and aliases are deleted. This
	// doesn't require the user to have explicit permissions for DeleteAlias.
	//
	// To delete Lambda event source mappings that invoke a function, use DeleteEventSourceMapping. For Amazon
	// Web Services services and resources that invoke your function directly, delete
	// the trigger in the service where you originally configured it.
	DeleteFunction(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error)
	// Returns the code signing configuration for the specified function.
	GetFunctionCodeSigningConfig(
		ctx context.Context,
		params *lambda.GetFunctionCodeSigningConfigInput,
		optFns ...func(*lambda.Options),
	) (*lambda.GetFunctionCodeSigningConfigOutput, error)
	// Returns your function's [recursive loop detection] configuration.
	//
	// [recursive loop detection]: https://docs.aws.amazon.com/lambda/latest/dg/invocation-recursion.html
	GetFunctionRecursionConfig(
		ctx context.Context,
		params *lambda.GetFunctionRecursionConfigInput,
		optFns ...func(*lambda.Options),
	) (*lambda.GetFunctionRecursionConfigOutput, error)
	// Returns details about the reserved concurrency configuration for a function. To
	// set a concurrency limit for a function, use PutFunctionConcurrency.
	GetFunctionConcurrency(
		ctx context.Context,
		params *lambda.GetFunctionConcurrencyInput,
		optFns ...func(*lambda.Options),
	) (*lambda.GetFunctionConcurrencyOutput, error)
}

// NewService creates a new instance of the AWS Lambda service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return lambda.NewFromConfig(
		*awsConfig,
		lambda.WithEndpointResolverV2(
			&lambdaEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type lambdaEndpointResolverV2 struct {
	providerContext provider.Context
}

func (l *lambdaEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params lambda.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	lambdaAliases := utils.Services["lambda"]
	lambdaEndpoint, hasLambdaEndpoint := utils.GetEndpointFromProviderConfig(
		l.providerContext,
		"lambda",
		lambdaAliases,
	)
	if hasLambdaEndpoint && !core.IsScalarNil(lambdaEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(lambdaEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return lambda.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
