package cloudcontrol

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// CCDataSourceConfig is the per-type configuration for a Cloud Control–backed data
// source. Codegen emits one of these per included CloudFormation type.
type CCDataSourceConfig struct {
	// BlueprintType is the Bluelink data source type, e.g. "aws/dynamodb/table".
	BlueprintType string
	// CFNType is the CloudFormation type name, e.g. "AWS::DynamoDB::Table".
	CFNType string
	Label   string
	// PlainTextSummary and FormattedDescription document the data source type.
	PlainTextSummary     string
	FormattedDescription string
	// Schema is the typed resource schema, reused to translate the
	// fetched Cloud Control properties into typed exports.
	// (overlays can make alterations to the schema)
	Schema *provider.ResourceDefinitionsSchema
	// Meta carries the per-type information the engine needs for identifier resolution
	// and CFN<->spec name translation.
	Meta CCResourceMeta
	// FilterFields declares the filterable fields exposed to practitioners.
	FilterFields map[string]*provider.DataSourceFilterSchema
	// ExportFields declares the (flattened) exportable fields.
	ExportFields map[string]*provider.DataSourceSpecSchema
	// DeriveIdentifierFromARN enables a fast path where a single `arn` equality filter
	// is resolved to the primary identifier by taking the ARN's final segment. Set for
	// name-style identifiers (e.g. eventBus name, table name); leave false where the
	// identifier is the ARN itself or a non-derivable form (e.g. an SQS queue URL).
	DeriveIdentifierFromARN bool
	// FormattedExamples holds rendered example markdown for the data source type.
	FormattedExamples []string
}

// CCDataSource builds a provider.DataSource backed by the generic Cloud Control
// engine. The fetch reads the resource via GetResource (fast path for an exact
// identifier/ARN filter) or ListResources + client-side filtering, then translates the
// CFN properties into typed, flattened exports.
func CCDataSource(
	config CCDataSourceConfig,
	cloudControlServiceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.DataSource {
	actions := &ccDataSourceActions{
		config:                     config,
		cloudControlServiceFactory: cloudControlServiceFactory,
		awsConfigStore:             awsConfigStore,
	}

	return &providerv1.DataSourceDefinition{
		Type:                 config.BlueprintType,
		Label:                config.Label,
		PlainTextSummary:     config.PlainTextSummary,
		FormattedDescription: config.FormattedDescription,
		MarkdownExamples:     config.FormattedExamples,
		Fields:               config.ExportFields,
		FilterFields:         config.FilterFields,
		FetchFunc:            actions.Fetch,
	}
}

type ccDataSourceActions struct {
	config                     CCDataSourceConfig
	cloudControlServiceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service]
	awsConfigStore             pluginutils.ServiceConfigStore[*aws.Config]
}

func (a *ccDataSourceActions) getCloudControlService(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (cloudcontrolservice.Service, error) {
	var filters *provider.ResolvedDataSourceFilters
	if input.DataSourceWithResolvedSubs != nil {
		filters = input.DataSourceWithResolvedSubs.Filter
	}
	meta := map[string]*core.MappingNode{
		"region": pluginutils.ExtractMatchFromFilters(filters, regionFilterField),
	}
	awsConfig, err := a.awsConfigStore.FromProviderContext(ctx, input.ProviderContext, meta)
	if err != nil {
		return nil, err
	}
	return a.cloudControlServiceFactory(awsConfig, input.ProviderContext), nil
}
