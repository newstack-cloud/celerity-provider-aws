package lambda

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS Lambda service
// used by the Lambda resource implementations.
type Service interface {
	// Returns information about the function or function version, with a link to
	// download the deployment package that's valid for 10 minutes. If you specify a
	// function version, only details that are specific to that version are returned.
	GetFunction(
		ctx context.Context,
		params *lambda.GetFunctionInput,
		optFns ...func(*lambda.Options),
	) (*lambda.GetFunctionOutput, error)
	// Deletes a Lambda function. To delete a specific function version, use the
	// Qualifier parameter. Otherwise, all versions and aliases are deleted. This
	// doesn't require the user to have explicit permissions for DeleteAlias.
	//
	// To delete Lambda event source mappings that invoke a function, use DeleteEventSourceMapping. For Amazon
	// Web Services services and resources that invoke your function directly, delete
	// the trigger in the service where you originally configured it.
	DeleteFunction(
		ctx context.Context,
		params *lambda.DeleteFunctionInput,
		optFns ...func(*lambda.Options),
	) (*lambda.DeleteFunctionOutput, error)
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
	// Modify the version-specific settings of a Lambda function.
	//
	// When you update a function, Lambda provisions an instance of the function and
	// its supporting resources. If your function connects to a VPC, this process can
	// take a minute. During this time, you can't modify the function, but you can
	// still invoke it. The LastUpdateStatus , LastUpdateStatusReason , and
	// LastUpdateStatusReasonCode fields in the response from GetFunctionConfiguration indicate when the
	// update is complete and the function is processing events with the new
	// configuration. For more information, see [Lambda function states].
	//
	// These settings can vary between versions of a function and are locked when you
	// publish a version. You can't modify the configuration of a published version,
	// only the unpublished version.
	//
	// To configure function concurrency, use PutFunctionConcurrency. To grant invoke permissions to an
	// Amazon Web Services account or Amazon Web Services service, use AddPermission.
	//
	// [Lambda function states]: https://docs.aws.amazon.com/lambda/latest/dg/functions-states.html
	UpdateFunctionConfiguration(
		ctx context.Context,
		params *lambda.UpdateFunctionConfigurationInput,
		optFns ...func(*lambda.Options),
	) (*lambda.UpdateFunctionConfigurationOutput, error)
	// Updates a Lambda function's code. If code signing is enabled for the function,
	// the code package must be signed by a trusted publisher. For more information,
	// see [Configuring code signing for Lambda].
	//
	// If the function's package type is Image , then you must specify the code package
	// in ImageUri as the URI of a [container image] in the Amazon ECR registry.
	//
	// If the function's package type is Zip , then you must specify the deployment
	// package as a [.zip file archive]. Enter the Amazon S3 bucket and key of the code .zip file
	// location. You can also provide the function code inline using the ZipFile field.
	//
	// The code in the deployment package must be compatible with the target
	// instruction set architecture of the function ( x86-64 or arm64 ).
	//
	// The function's code is locked when you publish a version. You can't modify the
	// code of a published version, only the unpublished version.
	//
	// For a function defined as a container image, Lambda resolves the image tag to
	// an image digest. In Amazon ECR, if you update the image tag to a new image,
	// Lambda does not automatically update the function.
	//
	// [.zip file archive]: https://docs.aws.amazon.com/lambda/latest/dg/gettingstarted-package.html#gettingstarted-package-zip
	// [Configuring code signing for Lambda]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-codesigning.html
	// [container image]: https://docs.aws.amazon.com/lambda/latest/dg/lambda-images.html
	UpdateFunctionCode(
		ctx context.Context,
		params *lambda.UpdateFunctionCodeInput,
		optFns ...func(*lambda.Options),
	) (*lambda.UpdateFunctionCodeOutput, error)
	// Update the code signing configuration for the function. Changes to the code
	// signing configuration take effect the next time a user tries to deploy a code
	// package to the function.
	PutFunctionCodeSigningConfig(
		ctx context.Context,
		params *lambda.PutFunctionCodeSigningConfigInput,
		optFns ...func(*lambda.Options),
	) (*lambda.PutFunctionCodeSigningConfigOutput, error)
	// Sets the maximum number of simultaneous executions for a function, and reserves
	// capacity for that concurrency level.
	//
	// Concurrency settings apply to the function as a whole, including all published
	// versions and the unpublished version. Reserving concurrency both ensures that
	// your function has capacity to process the specified number of events
	// simultaneously, and prevents it from scaling beyond that level. Use GetFunctionto see the
	// current setting for a function.
	//
	// Use GetAccountSettings to see your Regional concurrency limit. You can reserve concurrency for as
	// many functions as you like, as long as you leave at least 100 simultaneous
	// executions unreserved for functions that aren't configured with a per-function
	// limit. For more information, see [Lambda function scaling].
	//
	// [Lambda function scaling]: https://docs.aws.amazon.com/lambda/latest/dg/invocation-scaling.html
	PutFunctionConcurrency(
		ctx context.Context,
		params *lambda.PutFunctionConcurrencyInput,
		optFns ...func(*lambda.Options),
	) (*lambda.PutFunctionConcurrencyOutput, error)
	// Sets your function's [recursive loop detection] configuration.
	//
	// When you configure a Lambda function to output to the same service or resource
	// that invokes the function, it's possible to create an infinite recursive loop.
	// For example, a Lambda function might write a message to an Amazon Simple Queue
	// Service (Amazon SQS) queue, which then invokes the same function. This
	// invocation causes the function to write another message to the queue, which in
	// turn invokes the function again.
	//
	// Lambda can detect certain types of recursive loops shortly after they occur.
	// When Lambda detects a recursive loop and your function's recursive loop
	// detection configuration is set to Terminate , it stops your function being
	// invoked and notifies you.
	//
	// [recursive loop detection]: https://docs.aws.amazon.com/lambda/latest/dg/invocation-recursion.html
	PutFunctionRecursionConfig(
		ctx context.Context,
		params *lambda.PutFunctionRecursionConfigInput,
		optFns ...func(*lambda.Options),
	) (*lambda.PutFunctionRecursionConfigOutput, error)
	// Sets the runtime management configuration for a function's version. For more
	// information, see [Runtime updates].
	//
	// [Runtime updates]: https://docs.aws.amazon.com/lambda/latest/dg/runtimes-update.html
	PutRuntimeManagementConfig(
		ctx context.Context,
		params *lambda.PutRuntimeManagementConfigInput,
		optFns ...func(*lambda.Options),
	) (*lambda.PutRuntimeManagementConfigOutput, error)
	// Adds [tags] to a function, event source mapping, or code signing configuration.
	//
	// [tags]: https://docs.aws.amazon.com/lambda/latest/dg/tagging.html
	TagResource(
		ctx context.Context,
		params *lambda.TagResourceInput,
		optFns ...func(*lambda.Options),
	) (*lambda.TagResourceOutput, error)
	// Removes [tags] from a function, event source mapping, or code signing configuration.
	//
	// [tags]: https://docs.aws.amazon.com/lambda/latest/dg/tagging.html
	UntagResource(
		ctx context.Context,
		params *lambda.UntagResourceInput,
		optFns ...func(*lambda.Options),
	) (*lambda.UntagResourceOutput, error)
	// Creates a Lambda function. To create a function, you need a [deployment package] and an [execution role]. The
	// deployment package is a .zip file archive or container image that contains your
	// function code. The execution role grants the function permission to use Amazon
	// Web Services services, such as Amazon CloudWatch Logs for log streaming and
	// X-Ray for request tracing.
	//
	// If the deployment package is a [container image], then you set the package type to Image . For a
	// container image, the code property must include the URI of a container image in
	// the Amazon ECR registry. You do not need to specify the handler and runtime
	// properties.
	//
	// If the deployment package is a [.zip file archive], then you set the package type to Zip . For a
	// .zip file archive, the code property specifies the location of the .zip file.
	// You must also specify the handler and runtime properties. The code in the
	// deployment package must be compatible with the target instruction set
	// architecture of the function ( x86-64 or arm64 ). If you do not specify the
	// architecture, then the default value is x86-64 .
	//
	// When you create a function, Lambda provisions an instance of the function and
	// its supporting resources. If your function connects to a VPC, this process can
	// take a minute or so. During this time, you can't invoke or modify the function.
	// The State , StateReason , and StateReasonCode fields in the response from GetFunctionConfiguration
	// indicate when the function is ready to invoke. For more information, see [Lambda function states].
	//
	// A function has an unpublished version, and can have published versions and
	// aliases. The unpublished version changes when you update your function's code
	// and configuration. A published version is a snapshot of your function code and
	// configuration that can't be changed. An alias is a named resource that maps to a
	// version, and can be changed to map to a different version. Use the Publish
	// parameter to create version 1 of your function from its initial configuration.
	//
	// The other parameters let you configure version-specific and function-level
	// settings. You can modify version-specific settings later with UpdateFunctionConfiguration. Function-level
	// settings apply to both the unpublished and published versions of the function,
	// and include tags (TagResource ) and per-function concurrency limits (PutFunctionConcurrency ).
	//
	// You can use code signing if your deployment package is a .zip file archive. To
	// enable code signing for this function, specify the ARN of a code-signing
	// configuration. When a user attempts to deploy a code package with UpdateFunctionCode, Lambda
	// checks that the code package has a valid signature from a trusted publisher. The
	// code-signing configuration includes set of signing profiles, which define the
	// trusted publishers for this function.
	//
	// If another Amazon Web Services account or an Amazon Web Services service
	// invokes your function, use AddPermissionto grant permission by creating a resource-based
	// Identity and Access Management (IAM) policy. You can grant permissions at the
	// function level, on a version, or on an alias.
	//
	// To invoke your function directly, use Invoke. To invoke your function in response to
	// events in other Amazon Web Services services, create an event source mapping (CreateEventSourceMapping
	// ), or configure a function trigger in the other service. For more information,
	// see [Invoking Lambda functions].
	//
	// [Invoking Lambda functions]: https://docs.aws.amazon.com/lambda/latest/dg/lambda-invocation.html
	// [Lambda function states]: https://docs.aws.amazon.com/lambda/latest/dg/functions-states.html
	// [.zip file archive]: https://docs.aws.amazon.com/lambda/latest/dg/gettingstarted-package.html#gettingstarted-package-zip
	// [execution role]: https://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html#lambda-intro-execution-role
	// [container image]: https://docs.aws.amazon.com/lambda/latest/dg/lambda-images.html
	// [deployment package]: https://docs.aws.amazon.com/lambda/latest/dg/gettingstarted-package.html
	CreateFunction(
		ctx context.Context,
		params *lambda.CreateFunctionInput,
		optFns ...func(*lambda.Options),
	) (*lambda.CreateFunctionOutput, error)
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
