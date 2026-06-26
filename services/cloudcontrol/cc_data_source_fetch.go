package cloudcontrol

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	awscctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *ccDataSourceActions) Fetch(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (*provider.DataSourceFetchOutput, error) {
	service, err := a.getCloudControlService(ctx, input)
	if err != nil {
		return nil, err
	}

	var filters *provider.ResolvedDataSourceFilters
	if input.DataSourceWithResolvedSubs != nil {
		filters = input.DataSourceWithResolvedSubs.Filter
	}

	properties, err := a.resolveProperties(ctx, service, filters)
	if err != nil {
		return nil, err
	}

	exports, err := a.buildExports(properties, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	return &provider.DataSourceFetchOutput{Data: exports}, nil
}

// Returns the Cloud Control Properties JSON of the single resource matching the filters,
// via the GetResource fast path or ListResources + client-side filtering.
func (a *ccDataSourceActions) resolveProperties(
	ctx context.Context,
	service cloudcontrolservice.Service,
	filters *provider.ResolvedDataSourceFilters,
) (string, error) {
	if identifier, ok := a.fastPathIdentifier(filters); ok {
		return a.getResourceProperties(ctx, service, identifier)
	}

	return a.filterPathProperties(ctx, service, filters)
}

// Resolves the GetResource identifier directly when the filter set is a single equality on
// the primary identifier (or on `arn` when the identifier is the ARN, or derivable from
// it). Returns false to fall back to the filter path.
func (a *ccDataSourceActions) fastPathIdentifier(
	filters *provider.ResolvedDataSourceFilters,
) (string, bool) {
	if len(a.config.Meta.PrimaryIdentifierFields) > 1 {
		return "", false
	}
	field, value, ok := singleEqualityFilter(filters)
	if !ok {
		return "", false
	}
	primaryField := a.config.Meta.PrimaryIdentifierField
	switch field {
	case primaryField:
		return value, true
	case "arn":
		if primaryField == "arn" {
			return value, true
		}
		if a.config.DeriveIdentifierFromARN {
			return deriveIdentifierFromARN(value), true
		}
	}
	return "", false
}

// Lists all resources of the type and returns the Properties of the single resource
// matching the filters, completing partial list payloads with GetResource where a filtered
// field is absent.
func (a *ccDataSourceActions) filterPathProperties(
	ctx context.Context,
	service cloudcontrolservice.Service,
	filters *provider.ResolvedDataSourceFilters,
) (string, error) {
	descriptions, err := a.listResources(ctx, service)
	if err != nil {
		return "", err
	}

	matchedIdentifier := ""
	for _, description := range descriptions {
		identifier := aws.ToString(description.Identifier)
		properties := aws.ToString(description.Properties)

		fields, err := a.toFields(properties)
		if err != nil {
			return "", err
		}

		// List payloads can be partial; if a filtered field is absent, fetch the full
		// resource before evaluating.
		if !filterFieldsPresent(filters, fields) {
			properties, err = a.getResourceProperties(ctx, service, identifier)
			if err != nil {
				return "", err
			}
			fields, err = a.toFields(properties)
			if err != nil {
				return "", err
			}
		}

		if !MatchesFilters(filters, fields) {
			continue
		}
		if matchedIdentifier != "" {
			return "", fmt.Errorf(
				"the %q data source filters matched more than one resource; refine the filters to match exactly one",
				a.config.BlueprintType,
			)
		}
		matchedIdentifier = identifier
	}

	if matchedIdentifier == "" {
		return "", fmt.Errorf(
			"no %q resource matched the data source filters",
			a.config.BlueprintType,
		)
	}

	// Re-fetch the matched resource to guarantee complete properties for exports.
	return a.getResourceProperties(ctx, service, matchedIdentifier)
}

func (a *ccDataSourceActions) listResources(
	ctx context.Context,
	service cloudcontrolservice.Service,
) ([]awscctypes.ResourceDescription, error) {
	var descriptions []awscctypes.ResourceDescription
	var nextToken *string
	for {
		listResources := pluginutils.RetryableReturnValue(
			func(ctx context.Context, in *awscc.ListResourcesInput) (*awscc.ListResourcesOutput, error) {
				return service.ListResources(ctx, in)
			},
			isCCErrorRetryable,
		)
		output, err := listResources(ctx, &awscc.ListResourcesInput{
			TypeName:  aws.String(a.config.CFNType),
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		descriptions = append(descriptions, output.ResourceDescriptions...)
		if output.NextToken == nil || aws.ToString(output.NextToken) == "" {
			break
		}
		nextToken = output.NextToken
	}
	return descriptions, nil
}

func (a *ccDataSourceActions) getResourceProperties(
	ctx context.Context,
	service cloudcontrolservice.Service,
	identifier string,
) (string, error) {
	getResource := pluginutils.RetryableReturnValue(
		func(ctx context.Context, in *awscc.GetResourceInput) (*awscc.GetResourceOutput, error) {
			return service.GetResource(ctx, in)
		},
		isCCErrorRetryable,
	)

	output, err := getResource(ctx, &awscc.GetResourceInput{
		TypeName:   aws.String(a.config.CFNType),
		Identifier: aws.String(identifier),
	})
	if err != nil {
		return "", err
	}

	if output.ResourceDescription == nil {
		return "", nil
	}

	return aws.ToString(output.ResourceDescription.Properties), nil
}

func (a *ccDataSourceActions) toFields(properties string) (map[string]*core.MappingNode, error) {
	cfnNode, err := cfnJSONToMappingNode(properties)
	if err != nil {
		return nil, err
	}

	spec := CFNToSpec(cfnNode, a.config.Schema, a.config.Meta)
	return FlattenForDataSource(spec, a.config.Schema), nil
}

func (a *ccDataSourceActions) buildExports(
	properties string,
	providerContext provider.Context,
) (map[string]*core.MappingNode, error) {
	cfnNode, err := cfnJSONToMappingNode(properties)
	if err != nil {
		return nil, err
	}
	spec := CFNToSpec(cfnNode, a.config.Schema, a.config.Meta)
	if spec.Fields == nil {
		spec.Fields = map[string]*core.MappingNode{}
	}
	stripWriteOnlyTopLevelFields(spec, a.config.Meta)
	filterExternalTags(spec, a.config.Meta, providerContext)
	return selectExports(FlattenForDataSource(spec, a.config.Schema), a.config.ExportFields), nil
}

func filterFieldsPresent(
	filters *provider.ResolvedDataSourceFilters,
	fields map[string]*core.MappingNode,
) bool {
	if filters == nil {
		return true
	}

	for _, filter := range filters.Filters {
		if filter == nil {
			continue
		}

		field := core.StringValueFromScalar(filter.Field)
		if field == "" || field == regionFilterField {
			// Skip region filter, which is not a spec field.
			// Region filter is handled directly with the Cloud Control service,
			// not as a client-side filter on the resource spec.
			continue
		}

		if _, ok := fields[field]; !ok {
			return false
		}
	}

	return true
}

func singleEqualityFilter(filters *provider.ResolvedDataSourceFilters) (string, string, bool) {
	if filters == nil {
		return "", "", false
	}

	field := ""
	value := ""
	count := 0
	for _, filter := range filters.Filters {
		if filter == nil {
			continue
		}

		name := core.StringValueFromScalar(filter.Field)
		if name == "" || name == regionFilterField {
			continue
		}

		if pluginutils.GetDataSourceFilterOperator(filter) != schema.DataSourceFilterOperatorEquals {
			return "", "", false
		}

		count += 1
		field = name
		value = core.StringValue(pluginutils.GetDataSourceFilterSearchValue(filter, 0))
	}

	if count != 1 || value == "" {
		return "", "", false
	}

	return field, value, true
}

// Extracts the resource name from an ARN by taking the final "/"- or ":"-delimited segment,
// e.g. "arn:aws:dynamodb:..:table/orders" -> "orders".
func deriveIdentifierFromARN(arn string) string {
	if idx := strings.LastIndex(arn, "/"); idx >= 0 {
		return arn[idx+1:]
	}

	if idx := strings.LastIndex(arn, ":"); idx >= 0 {
		return arn[idx+1:]
	}

	return arn
}
