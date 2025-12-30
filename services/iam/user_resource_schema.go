package iam

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func iamUserResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "IAMUserDefinition",
		Description: "The definition of an AWS IAM user.",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"groups": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of group names to which you want to add the user.",
				FormattedDescription: "A list of group names to which you want to add the user. " +
					"Each group must already exist in the account.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The name of the group.",
					Pattern:     `[\w+=,.@-]+`,
					MinLength:   1,
					MaxLength:   128,
					Examples: []*core.MappingNode{
						core.MappingNodeFromString("Developers"),
						core.MappingNodeFromString("Administrators"),
						core.MappingNodeFromString("ReadOnlyUsers"),
					},
				},
				Nullable: true,
			},
			"loginProfile": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "Creates a password for the specified IAM user. A password allows an IAM user to access AWS services through the AWS Management Console.",
				FormattedDescription: "Creates a password for the specified IAM user. A password allows an IAM user to access AWS services through the AWS Management Console. " +
					"For more information about managing passwords, see " +
					"[Managing passwords](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_passwords.html) in the IAM User Guide.",
				Label:    "LoginProfile",
				Required: []string{"password"},
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"password": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The user's password. The password must meet the account's password policy, if one exists.",
						MinLength:   1,
						MaxLength:   128,
					},
					"passwordResetRequired": {
						Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
						Description: "Specifies whether the user is required to set a new password on next sign-in.",
						Default:     core.MappingNodeFromBool(false),
						Nullable:    true,
					},
				},
				Nullable: true,
			},
			"managedPolicyArns": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of Amazon Resource Names (ARNs) of the IAM managed policies that you want to attach to the user.",
				FormattedDescription: "A list of Amazon Resource Names (ARNs) of the IAM managed policies that you want to attach to the user. " +
					"For more information about ARNs, see [Amazon Resource Names (ARNs) and AWS Service Namespaces](https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html) " +
					"in the AWS General Reference.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The ARN of an IAM managed policy.",
					Pattern:     `^arn:(aws[a-zA-Z-]*)?:iam::(aws|\d{12}):policy/[\w+=,.@-]+$`,
					Examples: []*core.MappingNode{
						core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
						core.MappingNodeFromString("arn:aws:iam::aws:policy/PowerUserAccess"),
						core.MappingNodeFromString("arn:aws:iam::123456789012:policy/MyCustomPolicy"),
					},
				},
				Nullable: true,
			},
			"path": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The path for the user name. For more information about paths, see IAM identifiers in the IAM User Guide. " +
					"This parameter is optional. If it is not included, it defaults to a slash (/).",
				FormattedDescription: "The path for the user name. For more information about paths, see " +
					"[IAM identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html) in the IAM User Guide. " +
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
				Description: "The ARN of the managed policy that is used to set the permissions boundary for the user. " +
					"For more information about permissions boundaries, see Permissions boundaries for IAM identities in the IAM User Guide.",
				FormattedDescription: "The ARN of the managed policy that is used to set the permissions boundary for the user. " +
					"For more information about permissions boundaries, see " +
					"[Permissions boundaries for IAM identities](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_boundaries.html) " +
					"in the IAM User Guide.",
				Pattern:  `^arn:(aws[a-zA-Z-]*)?:iam::(aws|\d{12}):policy/[\w+=,.@-]+$`,
				Nullable: true,
			},
			"policies": {
				Type: provider.ResourceDefinitionsSchemaTypeArray,
				Description: "Adds or updates an inline policy document that is embedded in the specified IAM user. " +
					"When you embed an inline policy in a user, the inline policy is used as part of the user's access (permissions) policy.",
				FormattedDescription: "Adds or updates an inline policy document that is embedded in the specified IAM user. " +
					"When you embed an inline policy in a user, the inline policy is used as part of the user's access (permissions) policy. " +
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
														core.MappingNodeFromString("iam:ListUsers"),
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
				Nullable: true,
			},
			"userName": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name of the user to create. Do not include the path in this value. " +
					"The user name must be unique within the account. User names are not distinguished by case.",
				FormattedDescription: "The name of the user to create. Do not include the path in this value. " +
					"The user name must be unique within the account. User names are not distinguished by case. " +
					"If you don't specify a name, the provider generates a unique physical ID and uses that ID for the user name.",
				Pattern:      `[\w+=,.@-]+`,
				MinLength:    1,
				MaxLength:    64,
				MustRecreate: true,
				Nullable:     true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("MyUser"),
					core.MappingNodeFromString("developer-user"),
					core.MappingNodeFromString("service-account"),
				},
			},
			"tags": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of tags that are attached to the user.",
				FormattedDescription: "A list of tags that are attached to the user. For more information about tagging, see " +
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
				Description: "The Amazon Resource Name (ARN) of the IAM user.",
				FormattedDescription: "The Amazon Resource Name (ARN) of the IAM user. " +
					"This is a computed field that is automatically set after the user is created.",
				Computed: true,
			},
			"userId": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The stable and unique string identifying the user.",
				FormattedDescription: "The stable and unique string identifying the user. " +
					"This is a computed field that is automatically set after the user is created.",
				Computed: true,
			},
		},
	}
}
