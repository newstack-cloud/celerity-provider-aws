package iam

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func iamManagedPolicyResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "IAMManagedPolicyDefinition",
		Description: "The definition of an AWS IAM managed policy.",
		Required:    []string{"policyName", "policyDocument"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"policyName": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "A name for the IAM managed policy, up to 128 characters in length.",
				FormattedDescription: "A name for the IAM managed policy, up to 128 characters in length. " +
					"The policy name must be unique within the account. Policy names are not distinguished by case. " +
					"If you don't specify a name, the provider generates a unique physical ID and uses that ID for the policy name.",
				Pattern:      `[\w+=,.@-]+`,
				MinLength:    1,
				MaxLength:    128,
				MustRecreate: true,
				Nullable:     true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("MyPolicy"),
					core.MappingNodeFromString("S3ReadOnlyPolicy"),
					core.MappingNodeFromString("EC2FullAccessPolicy"),
				},
			},
			"policyDocument": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "The policy document that is associated with this managed policy.",
				FormattedDescription: "The policy document that is associated with this managed policy. " +
					"For more information about the elements that you can use in an IAM policy, " +
					"see [IAM Policy Elements Reference](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements.html) in the IAM User Guide.",
				Label:    "PolicyDocument",
				Required: []string{"Version", "Statement"},
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"Version": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The policy language version. The only valid value is '2012-10-17'.",
						Pattern:     `^2012-10-17$`,
						Examples: []*core.MappingNode{
							core.MappingNodeFromString("2012-10-17"),
						},
					},
					"Statement": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "An array of individual statements in the policy document.",
						Items: &provider.ResourceDefinitionsSchema{
							Type:     provider.ResourceDefinitionsSchemaTypeObject,
							Label:    "Statement",
							Required: []string{"Effect", "Action"},
							Attributes: map[string]*provider.ResourceDefinitionsSchema{
								"Effect": {
									Type:        provider.ResourceDefinitionsSchemaTypeString,
									Description: "The effect of the statement. Valid values are 'Allow' and 'Deny'.",
									Pattern:     `^(Allow|Deny)$`,
									Examples: []*core.MappingNode{
										core.MappingNodeFromString("Allow"),
										core.MappingNodeFromString("Deny"),
									},
								},
								"Action": {
									Type:        provider.ResourceDefinitionsSchemaTypeArray,
									Description: "The action or actions that will be allowed or denied.",
									Items: &provider.ResourceDefinitionsSchema{
										Type: provider.ResourceDefinitionsSchemaTypeString,
										Examples: []*core.MappingNode{
											core.MappingNodeFromString("s3:GetObject"),
											core.MappingNodeFromString("ec2:DescribeInstances"),
											core.MappingNodeFromString("lambda:InvokeFunction"),
										},
									},
								},
								"Resource": {
									Type:        provider.ResourceDefinitionsSchemaTypeArray,
									Description: "The resource or resources that the statement covers.",
									Items: &provider.ResourceDefinitionsSchema{
										Type: provider.ResourceDefinitionsSchemaTypeString,
										Examples: []*core.MappingNode{
											core.MappingNodeFromString("*"),
											core.MappingNodeFromString("arn:aws:s3:::my-bucket/*"),
											core.MappingNodeFromString("arn:aws:ec2:us-west-2:123456789012:instance/*"),
										},
									},
									Nullable: true,
								},
								"Condition": {
									Type:        provider.ResourceDefinitionsSchemaTypeObject,
									Description: "The conditions under which the statement is in effect.",
									Nullable:    true,
								},
							},
						},
					},
				},
				Examples: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"Version": core.MappingNodeFromString("2012-10-17"),
							"Statement": {
								Items: []*core.MappingNode{
									{
										Fields: map[string]*core.MappingNode{
											"Effect": core.MappingNodeFromString("Allow"),
											"Action": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("s3:GetObject"),
													core.MappingNodeFromString("s3:ListBucket"),
												},
											},
											"Resource": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("arn:aws:s3:::my-bucket"),
													core.MappingNodeFromString("arn:aws:s3:::my-bucket/*"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"description": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "A description of the managed policy that you provide.",
				Pattern:     `[\u0009\u000A\u000D\u0020-\u007E\u00A1-\u00FF]*`,
				MaxLength:   1000,
				Nullable:    true,
			},
			"path": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The path to the managed policy. For more information about paths, see IAM Identifiers in the IAM User Guide. " +
					"This parameter is optional. If it is not included, it defaults to a slash (/).",
				FormattedDescription: "The path to the managed policy. For more information about paths, see " +
					"[IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html) in the IAM User Guide. " +
					"This parameter is optional. If it is not included, it defaults to a slash (/).",
				Pattern:      `(\u002F)|(\u002F[\u0021-\u007E]+\u002F)`,
				MinLength:    1,
				MaxLength:    512,
				Default:      core.MappingNodeFromString("/"),
				MustRecreate: true,
				Nullable:     true,
			},
			"tags": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of tags that are attached to the managed policy.",
				FormattedDescription: "A list of tags that are attached to the managed policy. For more information about tagging, see " +
					"[Tagging IAM resources](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html) in the IAM User Guide.",
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
				Description: "The Amazon Resource Name (ARN) of the IAM managed policy.",
				FormattedDescription: "The Amazon Resource Name (ARN) of the IAM managed policy. " +
					"This is a computed field that is automatically set after the managed policy is created.",
				Computed: true,
			},
			"id": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The stable and unique string identifying the managed policy.",
				FormattedDescription: "The stable and unique string identifying the managed policy. " +
					"This is a computed field that is automatically set after the managed policy is created.",
				Computed: true,
			},
			"attachmentCount": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The number of entities (users, groups, and roles) that the managed policy is attached to.",
				FormattedDescription: "The number of entities (users, groups, and roles) that the managed policy is attached to. " +
					"This is a computed field that is automatically set after the managed policy is created.",
				Computed: true,
			},
			"createDate": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The date and time, in ISO 8601 date-time format, when the managed policy was created.",
				FormattedDescription: "The date and time, in ISO 8601 date-time format, when the managed policy was created. " +
					"This is a computed field that is automatically set after the managed policy is created.",
				Computed: true,
			},
			"defaultVersionId": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The identifier for the version of the policy that is set as the default version.",
				FormattedDescription: "The identifier for the version of the policy that is set as the default version. " +
					"This is a computed field that is automatically set after the managed policy is created.",
				Computed: true,
			},
			"isAttachable": {
				Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
				Description: "Specifies whether the policy can be attached to an IAM user, group, or role.",
				FormattedDescription: "Specifies whether the policy can be attached to an IAM user, group, or role. " +
					"This is a computed field that is automatically set after the managed policy is created.",
				Computed: true,
			},
			"permissionsBoundaryUsageCount": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The number of entities (users and roles) for which the managed policy is used to set the permissions boundary.",
				FormattedDescription: "The number of entities (users and roles) for which the managed policy is used to set the permissions boundary. " +
					"This is a computed field that is automatically set after the managed policy is created.",
				Computed: true,
			},
			"updateDate": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The date and time, in ISO 8601 date-time format, when the managed policy was last updated.",
				FormattedDescription: "The date and time, in ISO 8601 date-time format, when the managed policy was last updated. " +
					"This is a computed field that is automatically set after the managed policy is created.",
				Computed: true,
			},
		},
	}
}
