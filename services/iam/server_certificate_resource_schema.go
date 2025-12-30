package iam

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func iamServerCertificateResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "IAMServerCertificateDefinition",
		Description: "The definition of an AWS IAM server certificate.",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"serverCertificateName": {
				Type:                 provider.ResourceDefinitionsSchemaTypeString,
				Description:          "The name for the server certificate. Do not include the path in this value. The name cannot contain spaces.",
				FormattedDescription: "The name for the server certificate. Do not include the path in this value. The name cannot contain spaces. Allowed characters: upper/lowercase alphanumeric, _+=,.@-. If not specified, a unique name will be generated.",
				Pattern:              `[\w+=,.@-]+`,
				MinLength:            1,
				MaxLength:            128,
				Nullable:             true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("MyServerCertificate"),
					core.MappingNodeFromString("prod-cert"),
				},
			},
			"certificateBody": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The contents of the public key certificate in PEM-encoded format.",
				Pattern:      `[\u0009\u000A\u000D\u0020-\u00FF]+`,
				MinLength:    1,
				MaxLength:    16384,
				MustRecreate: true,
			},
			"privateKey": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The contents of the private key in PEM-encoded format.",
				Pattern:      `[\u0009\u000A\u000D\u0020-\u00FF]+`,
				MinLength:    1,
				MaxLength:    16384,
				Sensitive:    true,
				MustRecreate: true,
				// There is no way to access a private key via the AWS IAM service
				// so drift detection is not relevant.
				IgnoreDrift: true,
			},
			"certificateChain": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The contents of the public key certificate chain in PEM-encoded format.",
				Pattern:      `[\u0009\u000A\u000D\u0020-\u00FF]+`,
				MinLength:    1,
				MaxLength:    2097152,
				Nullable:     true,
				MustRecreate: true,
			},
			"path": {
				Type:                 provider.ResourceDefinitionsSchemaTypeString,
				Description:          "The path for the server certificate. Defaults to '/'. For CloudFront, must begin with '/cloudfront' and end with a slash.",
				FormattedDescription: "The path for the server certificate. Defaults to '/'. For CloudFront, must begin with '/cloudfront' and end with a slash. Pattern: (/) or (/.../).",
				Pattern:              `(\u002F)|(\u002F[\u0021-\u007F]+\u002F)`,
				MinLength:            1,
				MaxLength:            512,
				Default:              core.MappingNodeFromString("/"),
				Nullable:             true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("/"),
					core.MappingNodeFromString("/cloudfront/test/"),
				},
			},
			"tags": {
				Type:                 provider.ResourceDefinitionsSchemaTypeArray,
				Description:          "A list of tags that are attached to the server certificate.",
				FormattedDescription: "A list of tags that are attached to the server certificate. For more information about tagging, see [Tagging IAM resources](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html) in the IAM User Guide.",
				SortArrayByField:     "key",
				Items: &provider.ResourceDefinitionsSchema{
					Type:  provider.ResourceDefinitionsSchemaTypeObject,
					Label: "Tag",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"key": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The key name of the tag.",
							MinLength:   1,
							MaxLength:   128,
							Pattern:     `[\w+=,.@-]+`,
						},
						"value": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The value for the tag.",
							MinLength:   0,
							MaxLength:   256,
							Pattern:     `[\w+=,.@-]*`,
						},
					},
					Required: []string{"key", "value"},
				},
				MaxLength: 50,
				Nullable:  true,
			},

			// Computed fields
			"arn": {
				Type:                 provider.ResourceDefinitionsSchemaTypeString,
				Description:          "The Amazon Resource Name (ARN) of the IAM server certificate.",
				FormattedDescription: "The Amazon Resource Name (ARN) of the IAM server certificate. This is a computed field that is automatically set after the certificate is created.",
				Computed:             true,
			},
		},
		Examples: []*core.MappingNode{
			{
				Fields: map[string]*core.MappingNode{
					"serverCertificateName": core.MappingNodeFromString("MyServerCertificate"),
					"certificateBody":       core.MappingNodeFromString("-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"),
					"privateKey":            core.MappingNodeFromString("-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"),
					"certificateChain":      core.MappingNodeFromString("-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"),
					"path":                  core.MappingNodeFromString("/"),
					"tags": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"key":   core.MappingNodeFromString("Environment"),
									"value": core.MappingNodeFromString("Production"),
								},
							},
						},
					},
				},
			},
		},
	}
}
