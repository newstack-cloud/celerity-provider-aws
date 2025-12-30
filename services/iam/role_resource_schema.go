package iam

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func iamRoleResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "IAMRoleDefinition",
		Description: "The definition of an AWS IAM role.",
		Required:    []string{"assumeRolePolicyDocument"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"assumeRolePolicyDocument": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "The trust policy that is associated with this role. Trust policies define which entities can assume the role.",
				FormattedDescription: "The trust policy that is associated with this role. Trust policies define which entities can assume the role. " +
					"You can associate only one trust policy with a role. For more information about the elements that you can use in an IAM policy, " +
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
							Required: []string{"Effect", "Principal", "Action"},
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
								"Principal": {
									Type:        provider.ResourceDefinitionsSchemaTypeObject,
									Description: "The principal element specifies the user, account, service, or other entity that is allowed or denied access to the resource.",
									Attributes: map[string]*provider.ResourceDefinitionsSchema{
										"Service": {
											Type:        provider.ResourceDefinitionsSchemaTypeArray,
											Description: "The AWS service that can assume the role.",
											Items: &provider.ResourceDefinitionsSchema{
												Type: provider.ResourceDefinitionsSchemaTypeString,
												Examples: []*core.MappingNode{
													core.MappingNodeFromString("lambda.amazonaws.com"),
													core.MappingNodeFromString("ec2.amazonaws.com"),
													core.MappingNodeFromString("s3.amazonaws.com"),
												},
											},
											Nullable: true,
										},
										"AWS": {
											Type:        provider.ResourceDefinitionsSchemaTypeArray,
											Description: "The AWS account or user that can assume the role.",
											Items: &provider.ResourceDefinitionsSchema{
												Type:    provider.ResourceDefinitionsSchemaTypeString,
												Pattern: `^(arn:aws:iam::\d{12}:user/[\w+=,.@-]+|\d{12}|\*)$`,
												Examples: []*core.MappingNode{
													core.MappingNodeFromString("123456789012"),
													core.MappingNodeFromString("arn:aws:iam::123456789012:user/MyUser"),
													core.MappingNodeFromString("*"),
												},
											},
											Nullable: true,
										},
									},
								},
								"Action": {
									Type:        provider.ResourceDefinitionsSchemaTypeArray,
									Description: "The action or actions that will be allowed or denied.",
									Items: &provider.ResourceDefinitionsSchema{
										Type: provider.ResourceDefinitionsSchemaTypeString,
										Examples: []*core.MappingNode{
											core.MappingNodeFromString("sts:AssumeRole"),
											core.MappingNodeFromString("s3:GetObject"),
											core.MappingNodeFromString("ec2:DescribeInstances"),
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
											"Principal": {
												Fields: map[string]*core.MappingNode{
													"Service": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("lambda.amazonaws.com"),
														},
													},
												},
											},
											"Action": {
												Items: []*core.MappingNode{
													core.MappingNodeFromString("sts:AssumeRole"),
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
				Description: "A description of the role that you provide.",
				Pattern:     `[\u0009\u000A\u000D\u0020-\u007E\u00A1-\u00FF]*`,
				MaxLength:   1000,
				Nullable:    true,
			},
			"managedPolicyArns": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of Amazon Resource Names (ARNs) of the IAM managed policies that you want to attach to the role.",
				FormattedDescription: "A list of Amazon Resource Names (ARNs) of the IAM managed policies that you want to attach to the role. " +
					"For more information about ARNs, see [Amazon Resource Names (ARNs) and AWS Service Namespaces](https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html) " +
					"in the AWS General Reference.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The ARN of an IAM managed policy.",
					Pattern:     `^arn:(aws[a-zA-Z-]*)?:iam::(aws|\d{12}):policy/[\w+=,.@-]+$`,
				},
				Nullable: true,
			},
			"maxSessionDuration": {
				Type: provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The maximum session duration (in seconds) that you want to set for the specified role. " +
					"If you do not specify a value for this setting, the default value of one hour is applied. " +
					"This setting can have a value from 1 hour to 12 hours.",
				FormattedDescription: "The maximum session duration (in seconds) that you want to set for the specified role. " +
					"If you do not specify a value for this setting, the default value of one hour is applied. " +
					"This setting can have a value from 1 hour to 12 hours. For more information, see " +
					"[Using IAM roles](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use.html) in the IAM User Guide.",
				Minimum:  core.ScalarFromInt(3600),      // 1 hour
				Maximum:  core.ScalarFromInt(43200),     // 12 hours
				Default:  core.MappingNodeFromInt(3600), // 1 hour
				Nullable: true,
			},
			"path": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The path to the role. For more information about paths, see IAM Identifiers in the IAM User Guide. " +
					"This parameter is optional. If it is not included, it defaults to a slash (/).",
				FormattedDescription: "The path to the role. For more information about paths, see " +
					"[IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html) in the IAM User Guide. " +
					"This parameter is optional. If it is not included, it defaults to a slash (/).",
				Pattern:      `(\u002F)|(\u002F[\u0021-\u007E]+\u002F)`,
				MinLength:    1,
				MaxLength:    512,
				Default:      core.MappingNodeFromString("/"),
				MustRecreate: true,
				Nullable:     true,
			},
			"permissionsBoundary": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The ARN of the policy used to set the permissions boundary for the role. " +
					"For more information about permissions boundaries, see Permissions boundaries for IAM identities in the IAM User Guide.",
				FormattedDescription: "The ARN of the policy used to set the permissions boundary for the role. " +
					"For more information about permissions boundaries, see " +
					"[Permissions boundaries for IAM identities](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_boundaries.html) " +
					"in the IAM User Guide.",
				Pattern:  `^arn:(aws[a-zA-Z-]*)?:iam::(aws|\d{12}):policy/[\w+=,.@-]+$`,
				Nullable: true,
			},
			"policies": {
				Type: provider.ResourceDefinitionsSchemaTypeArray,
				Description: "Adds or updates an inline policy document that is embedded in the specified IAM role. " +
					"When you embed an inline policy in a role, the inline policy is used as part of the role's access (permissions) policy.",
				FormattedDescription: "Adds or updates an inline policy document that is embedded in the specified IAM role. " +
					"When you embed an inline policy in a role, the inline policy is used as part of the role's access (permissions) policy. " +
					"For more information about policies, see [Managed Policies and Inline Policies](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_managed-vs-inline.html) " +
					"in the IAM User Guide.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:  provider.ResourceDefinitionsSchemaTypeObject,
					Label: "Policy",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"policyName": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The name of the policy document.",
							Pattern:     `[\w+=,.@-]+`,
							MinLength:   1,
							MaxLength:   128,
						},
						"policyDocument": {
							Type:        provider.ResourceDefinitionsSchemaTypeObject,
							Description: "The policy document.",
							Label:       "PolicyDocument",
							Required:    []string{"Version", "Statement"},
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
						},
					},
				},
				Required: []string{"policyName", "policyDocument"},
			},
			"roleName": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "A name for the IAM role, up to 64 characters in length. " +
					"The role name must be unique within the account. Role names are not distinguished by case.",
				FormattedDescription: "A name for the IAM role, up to 64 characters in length. " +
					"The role name must be unique within the account. Role names are not distinguished by case. " +
					"If you don't specify a name, the provider generates a unique physical ID and uses that ID for the role name.",
				Pattern:      `[\w+=,.@-]+`,
				MinLength:    1,
				MaxLength:    64,
				MustRecreate: true,
				Nullable:     true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("MyRole"),
					core.MappingNodeFromString("EC2-Role"),
					core.MappingNodeFromString("lambda-execution-role"),
				},
			},
			"tags": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of tags that are attached to the role.",
				FormattedDescription: "A list of tags that are attached to the role. For more information about tagging, see " +
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
				Description: "The Amazon Resource Name (ARN) of the IAM role.",
				FormattedDescription: "The Amazon Resource Name (ARN) of the IAM role. " +
					"This is a computed field that is automatically set after the role is created.",
				Computed: true,
			},
			"roleId": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The stable and unique string identifying the role.",
				FormattedDescription: "The stable and unique string identifying the role. " +
					"This is a computed field that is automatically set after the role is created.",
				Computed: true,
			},
		},
	}
}
