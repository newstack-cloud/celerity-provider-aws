package iam

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func iamOIDCProviderResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "IAMOIDCProviderDefinition",
		Description: "The definition of an AWS IAM OIDC provider.",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"url": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The URL of the identity provider. The URL must begin with https:// and should correspond to the iss claim in the provider's OpenID Connect ID tokens.",
				FormattedDescription: "The URL of the identity provider. The URL must begin with https:// and should correspond to the iss claim in the provider's OpenID Connect ID tokens. " +
					"Per the OIDC standard, path components are allowed but query parameters are not. " +
					"Typically the URL consists of only a hostname, like https://server.example.org or https://example.com.",
				Pattern:      `^https://[a-zA-Z0-9\.\-\/]+$`,
				MinLength:    1,
				MaxLength:    255,
				MustRecreate: true,
				Nullable:     true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("https://accounts.google.com"),
					core.MappingNodeFromString("https://token.actions.githubusercontent.com"),
					core.MappingNodeFromString("https://example.com"),
				},
			},
			"clientIdList": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of client IDs (also known as audiences) for the IAM OIDC provider.",
				FormattedDescription: "A list of client IDs (also known as audiences) for the IAM OIDC provider. " +
					"When a mobile or web app registers with an OpenID Connect provider, they establish a value that identifies the application. " +
					"This is the value that's sent as the client_id parameter on OAuth requests.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "A client ID (audience) for the OIDC provider.",
					MinLength:   1,
					MaxLength:   255,
					Examples: []*core.MappingNode{
						core.MappingNodeFromString("sts.amazonaws.com"),
						core.MappingNodeFromString("my-app-id"),
						core.MappingNodeFromString("123456789012"),
					},
				},
				MinLength: 0,
				MaxLength: 100,
				Nullable:  true,
			},
			"thumbprintList": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of server certificate thumbprints for the OpenID Connect (OIDC) identity provider's server certificates.",
				FormattedDescription: "A list of server certificate thumbprints for the OpenID Connect (OIDC) identity provider's server certificates. " +
					"Typically this list includes only one entry. However, IAM lets you have up to five thumbprints for an OIDC provider. " +
					"This lets you maintain multiple thumbprints if the identity provider is rotating certificates.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "A 40-character SHA-1 hash of the server certificate.",
					Pattern:     `^[0-9a-fA-F]{40}$`,
					MinLength:   40,
					MaxLength:   40,
					Examples: []*core.MappingNode{
						core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
						core.MappingNodeFromString("9e99a48a9960b14926bb7f3b02e22da2b0ab7280"),
					},
				},
				MinLength: 0,
				MaxLength: 5,
				Nullable:  true,
			},
			"tags": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of tags that are attached to the specified IAM OIDC provider.",
				FormattedDescription: "A list of tags that are attached to the specified IAM OIDC provider. " +
					"For more information about tagging, see [Tagging IAM resources](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html) in the IAM User Guide.",
				SortArrayByField: "key",
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
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the IAM OIDC provider.",
				FormattedDescription: "The Amazon Resource Name (ARN) of the IAM OIDC provider. " +
					"This is a computed field that is automatically set after the OIDC provider is created.",
				Computed: true,
			},
		},
	}
}
