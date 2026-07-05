package ssm

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

func parameterDataSourceSchema() map[string]*provider.DataSourceSpecSchema {
	return map[string]*provider.DataSourceSpecSchema{
		"name": {
			Label:       "Name",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The name of the parameter.",
			Nullable:    false,
		},
		"arn": {
			Label:       "ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the parameter.",
			Nullable:    false,
		},
		"type": {
			Label:       "Type",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The type of the parameter (String, StringList or SecureString).",
			Nullable:    false,
		},
		"value": {
			Label: "Value",
			Type:  provider.DataSourceSpecTypeString,
			Description: "The value of the parameter. Only populated for plaintext String and " +
				"StringList parameters; SecureString values are never returned by this data source.",
			Nullable: true,
		},
		"version": {
			Label:       "Version",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The version of the parameter.",
			Nullable:    false,
		},
		"tier": {
			Label:       "Tier",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The tier of the parameter (Standard, Advanced or Intelligent-Tiering).",
			Nullable:    true,
		},
		"dataType": {
			Label:       "Data Type",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The data type of the parameter.",
			Nullable:    true,
		},
		"keyId": {
			Label:       "KMS Key ID",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The KMS key id used to encrypt a SecureString parameter.",
			Nullable:    true,
		},
		"description": {
			Label:       "Description",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The description of the parameter.",
			Nullable:    true,
		},
		"allowedPattern": {
			Label:       "Allowed Pattern",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The regular expression used to validate the parameter value.",
			Nullable:    true,
		},
	}
}
