package provider

import (
	"regexp"

	"github.com/two-hundred/celerity-provider-aws/services/lambda"
	"github.com/two-hundred/celerity-provider-aws/types"
	"github.com/two-hundred/celerity-provider-aws/utils"
	"github.com/two-hundred/celerity/libs/blueprint/core"
	"github.com/two-hundred/celerity/libs/blueprint/provider"
	"github.com/two-hundred/celerity/libs/plugin-framework/sdk/providerv1"
	"github.com/two-hundred/celerity/libs/plugin-framework/sdk/validation"
)

func NewProvider(
	lambdaServiceFactory types.ServiceFactory[lambda.Service],
	awsConfigStore *utils.AWSConfigStore,
) provider.Provider {
	return &providerv1.ProviderPluginDefinition{
		ProviderNamespace:        "aws",
		ProviderConfigDefinition: providerConfigDefinition(),
		Resources: map[string]provider.Resource{
			"aws/lambda/function": lambda.FunctionResource(
				lambdaServiceFactory,
				awsConfigStore,
			),
		},
		DataSources:         map[string]provider.DataSource{},
		Links:               map[string]provider.Link{},
		CustomVariableTypes: map[string]provider.CustomVariableType{},
		Functions:           map[string]provider.Function{},
	}
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
