package provider

import (
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/flex"
	eventslambda "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/events_lambda"
	eventssqs "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/events_sqs"
	flexlambda "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/flex_lambda"
	kinesislambda "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/kinesis_lambda"
	lambdadynamodb "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/lambda_dynamodb"
	lambdaelasticache "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/lambda_elasticache"
	lambdakms "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/lambda_kms"
	lambdards "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/lambda_rds"
	lambdas3 "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/lambda_s3"
	lambdasecretsmanager "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/lambda_secretsmanager"
	lambdasns "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/lambda_sns"
	lambdassm "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/lambda_ssm"
	s3lambda "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/s3_lambda"
	s3sns "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/s3_sns"
	s3sqs "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/s3_sqs"
	snslambda "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/sns_lambda"
	snssqs "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/sns_sqs"
	sqslambda "github.com/newstack-cloud/bluelink-provider-aws/inter-service-links/sqs_lambda"
	cloudcontrolgen "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/gen"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	eventslinks "github.com/newstack-cloud/bluelink-provider-aws/services/events/links"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	kmsservice "github.com/newstack-cloud/bluelink-provider-aws/services/kms/service"
	"github.com/newstack-cloud/bluelink-provider-aws/services/lambda"
	lambdalinks "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/links"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	s3service "github.com/newstack-cloud/bluelink-provider-aws/services/s3/service"
	sqslinks "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/links"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink-provider-aws/services/ssm"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/validation"
)

func NewProvider(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	sqsServiceFactory pluginutils.ServiceFactory[*aws.Config, sqsservice.Service],
	dynamodbServiceFactory pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service],
	eventsServiceFactory pluginutils.ServiceFactory[*aws.Config, eventsservice.Service],
	cloudControlServiceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
	s3ServiceFactory pluginutils.ServiceFactory[*aws.Config, s3service.Service],
	ssmServiceFactory pluginutils.ServiceFactory[*aws.Config, ssmservice.Service],
	kmsServiceFactory pluginutils.ServiceFactory[*aws.Config, kmsservice.Service],
	awsConfigStore *utils.AWSConfigStore,
) provider.Provider {
	return &providerv1.ProviderPluginDefinition{
		ProviderNamespace:        "aws",
		ProviderConfigDefinition: providerConfigDefinition(),
		Resources: mergeResources(
			map[string]provider.Resource{
				"aws/flex/vpc": flex.VPCResource(
					ec2ServiceFactory,
					resourceGroupTaggingServiceFactory,
					awsConfigStore,
				),
				"aws/ssm/parameter": ssm.ParameterResource(
					ssmServiceFactory,
					awsConfigStore,
				),
			},
			cloudcontrolgen.GeneratedResources(
				cloudControlServiceFactory,
				resourceGroupTaggingServiceFactory,
				awsConfigStore,
			),
		),
		// Lambda data sources stay hand-written (they surface deployment/runtime
		// metadata beyond the CloudFormation model); every other data source is served
		// by the generated Cloud Control engine (cloudcontrolgen.GeneratedDataSources).
		DataSources: mergeDataSources(map[string]provider.DataSource{
			"aws/lambda/function": lambda.FunctionDataSource(
				lambdaServiceFactory,
				awsConfigStore,
			),
			"aws/lambda/functionUrl": lambda.FunctionUrlDataSource(
				lambdaServiceFactory,
				awsConfigStore,
			),
			"aws/lambda/alias": lambda.AliasDataSource(
				lambdaServiceFactory,
				awsConfigStore,
			),
			"aws/lambda/codeSigningConfig": lambda.CodeSigningConfigDataSource(
				lambdaServiceFactory,
				awsConfigStore,
			),
			"aws/lambda/layerVersion": lambda.LayerVersionDataSource(
				lambdaServiceFactory,
				awsConfigStore,
			),
			// aws/ssm/parameter is custom to be consistent with its resource.
			"aws/ssm/parameter": ssm.ParameterDataSource(
				ssmServiceFactory,
				awsConfigStore,
			),
		}, cloudcontrolgen.GeneratedDataSources(
			cloudControlServiceFactory,
			awsConfigStore,
		)),
		Links: map[string]provider.Link{
			"aws/lambda/function::aws/lambda/function": lambdalinks.FunctionFunctionLink(
				iamServiceFactory,
				ec2ServiceFactory,
			)(
				lambdalinks.FunctionToFunctionLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			"aws/lambda/function::aws/lambda/codeSigningConfig": lambdalinks.FunctionCodeSigningConfigLink(
				pluginutils.NewSingleLinkServiceDeps(
					lambdaServiceFactory,
					awsConfigStore,
				),
			),
			"aws/sqs/queue::aws/sqs/queue": sqslinks.QueueQueueLink(
				pluginutils.NewSingleLinkServiceDeps(
					sqsServiceFactory,
					awsConfigStore,
				),
			),
			// Inter-service links: Lambda <-> DynamoDB
			"aws/lambda/function::aws/dynamodb/table": lambdadynamodb.FunctionDynamoDBTableLink(
				iamServiceFactory,
			)(
				lambdadynamodb.FunctionToDynamoDBTableLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, dynamodbservice.Service]{
						ServiceFactory: dynamodbServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			"aws/dynamodb/table::aws/lambda/function": lambdadynamodb.DynamoDBTableLambdaFunctionLink(
				iamServiceFactory,
			)(
				lambdadynamodb.DynamoDBTableToLambdaFunctionLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, dynamodbservice.Service]{
						ServiceFactory: dynamodbServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Consumer triggers: SQS queue / Kinesis stream -> Lambda function (event source mappings).
			"aws/sqs/queue::aws/lambda/function": sqslambda.SQSQueueLambdaFunctionLink(
				iamServiceFactory,
			)(
				sqslambda.QueueToFunctionLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			"aws/kinesis/stream::aws/lambda/function": kinesislambda.StreamFunctionLink(
				iamServiceFactory,
			)(
				kinesislambda.StreamToFunctionLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// S3 bucket notifications: the bucket notifies a Lambda function on object events.
			"aws/s3/bucket::aws/lambda/function": s3lambda.BucketFunctionLink()(
				s3lambda.BucketToFunctionLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, s3service.Service]{
						ServiceFactory: s3ServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			"aws/s3/bucket::aws/sqs/queue": s3sqs.BucketQueueLink()(
				s3sqs.BucketToQueueLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, s3service.Service]{
						ServiceFactory: s3ServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, sqsservice.Service]{
						ServiceFactory: sqsServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			"aws/s3/bucket::aws/sns/topic": s3sns.BucketTopicLink()(
				s3sns.BucketToTopicLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, s3service.Service]{
						ServiceFactory: s3ServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Lambda access: the function reads/writes the bucket.
			"aws/lambda/function::aws/s3/bucket": lambdas3.FunctionBucketLink(
				iamServiceFactory,
				ec2ServiceFactory,
			)(
				lambdas3.FunctionToBucketLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Inter-service link: Lambda -> Secrets Manager (config secrets access)
			"aws/lambda/function::aws/secretsmanager/secret": lambdasecretsmanager.FunctionSecretLink(
				iamServiceFactory,
				ec2ServiceFactory,
			)(
				lambdasecretsmanager.FunctionToSecretLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Inter-service link: Lambda -> RDS Proxy (database access via connection pooling)
			"aws/lambda/function::aws/rds/dbProxy": lambdards.FunctionProxyLink(
				iamServiceFactory,
				ec2ServiceFactory,
			)(
				lambdards.FunctionToProxyLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Inter-service link: Lambda -> ElastiCache replication group (cache access)
			"aws/lambda/function::aws/elasticache/replicationGroup": lambdaelasticache.FunctionCacheLink(
				iamServiceFactory,
				ec2ServiceFactory,
			)(
				lambdaelasticache.FunctionToCacheLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Inter-service link: Lambda -> Aurora DB cluster (direct database access)
			"aws/lambda/function::aws/rds/dbCluster": lambdards.FunctionClusterLink(
				iamServiceFactory,
				ec2ServiceFactory,
			)(
				lambdards.FunctionToClusterLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Inter-service link: Lambda -> KMS key (cryptographic use)
			"aws/lambda/function::aws/kms/key": lambdakms.FunctionKeyLink(
				iamServiceFactory,
				ec2ServiceFactory,
				kmsServiceFactory,
			)(
				lambdakms.FunctionToKeyLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Inter-service link: Lambda -> SSM parameter (config access)
			"aws/lambda/function::aws/ssm/parameter": lambdassm.FunctionParameterLink(
				iamServiceFactory,
				ec2ServiceFactory,
			)(
				lambdassm.FunctionToParameterLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, ssmservice.Service]{
						ServiceFactory: ssmServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Flex VPC placement: the VPC places the Lambda function in its subnets.
			"aws/flex/vpc::aws/lambda/function": flexlambda.VPCFunctionLink()(
				flexlambda.VPCToFunctionLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, ec2service.Service]{
						ServiceFactory: ec2ServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Inter-service links: EventBridge <-> Lambda / SQS
			"aws/lambda/function::aws/events/eventBus": eventslambda.FunctionEventBusLink(
				iamServiceFactory,
			)(
				eventslambda.FunctionToEventBusLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, eventsservice.Service]{
						ServiceFactory: eventsServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Rule -> Lambda function: the invoke permission is a self-contained
			// managed intermediary deployed through the framework's ResourceService,
			// so this link needs no AWS service factories.
			"aws/events/rule::aws/lambda/function": eventslambda.RuleFunctionLink(),
			// Rule -> SQS queue: same managed-intermediary model (queue inline policy).
			"aws/events/rule::aws/sqs/queue": eventssqs.RuleQueueLink(),
			// SNS subscription -> SQS queue: reference-implied (subscription endpoint),
			// grants the queue policy allowing SNS to deliver.
			"aws/sns/subscription::aws/sqs/queue": snssqs.SubscriptionQueueLink(),
			// SNS subscription -> Lambda function: reference-implied (subscription
			// endpoint), grants SNS permission to invoke the function.
			"aws/sns/subscription::aws/lambda/function": snslambda.SubscriptionFunctionLink(),
			// Lambda function -> SNS topic: grants sns:Publish on the topic and populates
			// topic env vars; activates the SNS VPC endpoint for VPC-isolated functions.
			"aws/lambda/function::aws/sns/topic": lambdasns.FunctionTopicLink(
				iamServiceFactory,
				ec2ServiceFactory,
			)(
				lambdasns.FunctionToTopicLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
						ServiceFactory: lambdaServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
						ServiceFactory: cloudControlServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
			// Intra-service links: EventBridge rule <-> API destination
			"aws/events/rule::aws/events/apiDestination": eventslinks.RuleAPIDestinationLink(
				iamServiceFactory,
			)(
				eventslinks.RuleToAPIDestinationLinkDeps{
					ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, eventsservice.Service]{
						ServiceFactory: eventsServiceFactory,
						ConfigStore:    awsConfigStore,
					},
					ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, eventsservice.Service]{
						ServiceFactory: eventsServiceFactory,
						ConfigStore:    awsConfigStore,
					},
				},
			),
		},
		CustomVariableTypes: map[string]provider.CustomVariableType{},
		Functions:           map[string]provider.Function{},
	}
}

func mergeResources(
	base map[string]provider.Resource,
	generated map[string]provider.Resource,
) map[string]provider.Resource {
	for resourceType, resource := range generated {
		if _, exists := base[resourceType]; !exists {
			base[resourceType] = resource
		}
	}
	return base
}

func mergeDataSources(
	base map[string]provider.DataSource,
	generated map[string]provider.DataSource,
) map[string]provider.DataSource {
	for dataSourceType, dataSource := range generated {
		if _, exists := base[dataSourceType]; !exists {
			base[dataSourceType] = dataSource
		}
	}
	return base
}

func providerConfigDefinition() *core.ConfigDefinition {
	return &core.ConfigDefinition{
		Fields: map[string]*core.ConfigFieldDefinition{
			"accessKeyId": {
				Type:  core.ScalarTypeString,
				Label: "Access Key ID",
				Description: "The access key ID for API operations. " +
					"This can be retrieved from the 'Security & Credentials' section of the AWS console.",
			},
			"secretAccessKey": {
				Type:  core.ScalarTypeString,
				Label: "Secret Access Key",
				Description: "The secret access key for API operations. " +
					"This can be retrieved from the 'Security & Credentials' section of the AWS console.",
				Secret: true,
			},
			"customCABundle": {
				Type:  core.ScalarTypeString,
				Label: "Custom CA Bundle",
				Description: "The path to a custom CA bundle file to use for " +
					"TLS connections to AWS services. This can also be " +
					"configured using the `AWS_CA_BUNDLE` environment variable.",
			},
			"ec2MetadataServiceEndpoint": {
				Type:  core.ScalarTypeString,
				Label: "EC2 Metadata Service Endpoint",
				Description: "The address of the EC2 metadata service endpoint to use. " +
					"This can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.",
			},
			"ec2MetadataServiceEndpointMode": {
				Type:  core.ScalarTypeString,
				Label: "EC2 Metadata Service Endpoint Mode",
				Description: "The protocol to use for the EC2 metadata service endpoint. " +
					"This can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE` environment variable.",
				AllowedValues: []*core.ScalarValue{
					core.ScalarFromString("IPv4"),
					core.ScalarFromString("IPv6"),
				},
			},
			"endpoint.<serviceOrAlias>": {
				Type:  core.ScalarTypeString,
				Label: "Custom Service Endpoints",
				Description: "The addresses to use to override the default service endpoint URL. " +
					"<serviceOrAlias> must match a valid service string or an alias for a service. " +
					"The following is a list of all the supported services and their aliases:\n" +
					utils.AWSServiceList(),
			},
			"httpProxy": {
				Type:  core.ScalarTypeString,
				Label: "HTTP Proxy",
				Description: "URL of a proxy to use for HTTP requests when accessing the AWS API. " +
					"This can also be set using the `HTTP_PROXY` environment variable.",
			},
			"httpsProxy": {
				Type:  core.ScalarTypeString,
				Label: "HTTPS Proxy",
				Description: "URL of a proxy to use for HTTPS requests when accessing the AWS API. " +
					"This can also be set using the `HTTPS_PROXY` environment variable.",
			},
			"insecure": {
				Type:  core.ScalarTypeBool,
				Label: "Insecure",
				Description: "If true, the provider will not verify the TLS " +
					"certificate of the AWS API. If omitted, the default value is `false`.",
			},
			"maxRetries": {
				Type:  core.ScalarTypeInteger,
				Label: "Max Retries",
				Description: "The maximum number of retries to attempt when a request to an AWS API fails. " +
					"If not set, the AWS SDK defaults will be used.",
			},
			"profile": {
				Type:  core.ScalarTypeString,
				Label: "Profile",
				Description: "The name of the AWS profile to use for API operations. If not set, " +
					"the default profile created with `aws configure` will be used.",
			},
			"region": {
				Type:  core.ScalarTypeString,
				Label: "Region",
				Description: "The AWS region to use for API operations. If not set, " +
					"the default region will be used based on the environment.",
			},
			"retryMode": {
				Type:  core.ScalarTypeString,
				Label: "Retry Mode",
				Description: "Determines how retries are attempted. " +
					"Valid values are `standard` and `adaptive`. " +
					"This can also be configured using the `AWS_RETRY_MODE` environment variable.",
				AllowedValues: []*core.ScalarValue{
					core.ScalarFromString("standard"),
					core.ScalarFromString("adaptive"),
				},
			},
			"s3UsePathStyle": {
				Type:  core.ScalarTypeBool,
				Label: "S3 Use Path Style",
				Description: "If true, the provider will use the path-style addressing " +
					"for S3 URLs. If false, the virtual hosted-style addressing will be used.\n" +
					"Path style addresses are of the form https://s3.amazonaws.com/<bucket>/<key>, " +
					"while virtual hosted-style addresses are of the form https://<bucket>.s3.amazonaws.com/<key>.",
			},
			"sharedConfigFiles": {
				Type:         core.ScalarTypeString,
				Label:        "Shared Config Files",
				Description:  "A comma-separated list of paths to shared AWS config files to use for API operations.",
				DefaultValue: core.ScalarFromString("~/.aws/config"),
			},
			"sharedCredentialsFiles": {
				Type:         core.ScalarTypeString,
				Label:        "Shared Credentials Files",
				Description:  "A comma-separated list of paths to shared AWS credentials files to use for API operations.",
				DefaultValue: core.ScalarFromString("~/.aws/credentials"),
			},
			"sessionToken": {
				Type:        core.ScalarTypeString,
				Label:       "Session Token",
				Description: "The session token. This is only required if you are using temporary security credentials.",
				Secret:      true,
			},
			"useDualStackEndpoint": {
				Type:        core.ScalarTypeBool,
				Label:       "Use Dual Stack Endpoint",
				Description: "If true, the provider will resolve and endpoint with DualStack capability.",
			},
			"useFIPSEndpoint": {
				Type:        core.ScalarTypeBool,
				Label:       "Use FIPS Endpoint",
				Description: "If true, the provider will resolve and endpoint with FIPS capability.",
			},
			"assumeRole.duration": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role Duration",
				Description: "The duration between 15 minutes and 12 hours for which the assumed role session will be valid. " +
					"Valid units of time are ns, us (or μs), ms, s, m, h.",
				DefaultValue: core.ScalarFromString("1h"),
				Examples: []*core.ScalarValue{
					core.ScalarFromString("15m"),
					core.ScalarFromString("1h"),
					core.ScalarFromString("12h"),
				},
				ValidateFunc: validateAssumeRoleDuration,
			},
			"assumeRole.externalId": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role External ID",
				Description: "An optional unique identifier that may be required " +
					"when assuming a role in another account.",
				ValidateFunc: validation.WrapForPluginConfig(
					validation.AllOf(
						validation.StringLengthRange(2, 1224),
						validation.StringMatchesPattern(
							regexp.MustCompile(`[\w+=,.@:\/\-]*`),
						),
					),
				),
			},
			"assumeRole.roleArn": {
				Type:         core.ScalarTypeString,
				Label:        "Assume Role ARN",
				Description:  "The ARN of the IAM role to assume for API operations.",
				ValidateFunc: validateARN,
			},
			"assumeRole.policy": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role Policy",
				Description: "The IAM policy JSON document containing further" +
					" restrictions for the IAM role being assumed.",
				ValidateFunc: validation.WrapForPluginConfig(
					validation.StringIsJSON(),
				),
			},
			"assumeRole.policyArns.<index>": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role Policy ARNs",
				Description: "Amazon Resource Names (ARNs) of IAM Policies " +
					"for further restricting permissions for the IAM Role being assumed.",
				ValidateFunc: validateARN,
			},
			"assumeRole.sessionName": {
				Type:         core.ScalarTypeString,
				Label:        "Assume Role Session Name",
				Description:  "A unique identifier for the assumed role session.",
				ValidateFunc: validateAssumeRoleSessionName,
			},
			"assumeRole.sourceIdentity": {
				Type:         core.ScalarTypeString,
				Label:        "Assume Role Source Identity",
				Description:  "Source identity defined by the principal that is assuming the role.",
				ValidateFunc: validateAssumeRoleSourceIdentity,
			},
			"assumeRole.tags.<tagName>": {
				Type:        core.ScalarTypeString,
				Label:       "Assume Role Tags",
				Description: "Tags to apply to the assumed role session.",
			},
			"assumeRole.transitiveTagKeys": {
				Type:        core.ScalarTypeString,
				Label:       "Assume Role Transitive Tag Keys",
				Description: "A comma-separated list of tag keys to pass to any subsequent sessions.",
			},
			"assumeRoleWithWebIdentity.duration": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role With Web Identity Duration",
				Description: "The duration between 15 minutes and 12 hours for which the assumed role session will be valid. " +
					"Valid units of time are ns, us (or μs), ms, s, m, h.",
				DefaultValue: core.ScalarFromString("1h"),
				Examples: []*core.ScalarValue{
					core.ScalarFromString("15m"),
					core.ScalarFromString("1h"),
					core.ScalarFromString("12h"),
				},
				ValidateFunc: validateAssumeRoleDuration,
			},
			"assumeRoleWithWebIdentity.policy": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role With Web Identity Policy",
				Description: "The IAM policy JSON document containing further " +
					"restrictions for the IAM role being assumed with web identity.",
				ValidateFunc: validation.WrapForPluginConfig(
					validation.StringIsJSON(),
				),
			},
			"assumeRoleWithWebIdentity.policyArns.<index>": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role With Web Identity Policy ARNs",
				Description: "Amazon Resource Names (ARNs) of IAM Policies " +
					"for further restricting permissions for the IAM Role being assumed with web identity.",
				ValidateFunc: validateARN,
			},
			"assumeRoleWithWebIdentity.roleArn": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role With Web Identity ARN",
				Description: "The ARN of the IAM role to assume for API operations " +
					"when using web identity.",
				ValidateFunc: validateARN,
			},
			"assumeRoleWithWebIdentity.sessionName": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role With Web Identity Session Name",
				Description: "A unique identifier for the assumed role session " +
					"when using web identity.",
				ValidateFunc: validateAssumeRoleSessionName,
			},
			"assumeRoleWithWebIdentity.webIdentityToken": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role With Web Identity Web Identity Token",
				Description: "The web identity token to use when assuming the role " +
					"when using web identity.",
				ValidateFunc: validation.WrapForPluginConfig(
					validation.StringLengthRange(4, 20000),
				),
			},
			"assumeRoleWithWebIdentity.webIdentityTokenFile": {
				Type:  core.ScalarTypeString,
				Label: "Assume Role With Web Identity Web Identity Token File",
				Description: "The path to the file containing the web identity token " +
					"to use when assuming the role when using web identity.",
			},
		},
	}
}
