package cloudcontrol

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *ccResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	service, err := a.getCloudControlService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	identifier := primaryIdentifier(input.CurrentResourceSpec, a.config.Meta)
	if identifier == "" {
		identifier, err = a.lookupIdentifierByTags(ctx, input)
		if err != nil {
			return nil, err
		}
	}
	if identifier == "" {
		// The resource could not be located, so it is treated as non-existent.
		return &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
		}, nil
	}

	specState, err := a.fetchResourceSpecState(ctx, service, identifier, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: specState,
	}, nil
}

func (a *ccResourceActions) fetchResourceSpecState(
	ctx context.Context,
	service cloudcontrolservice.Service,
	identifier string,
	providerContext provider.Context,
) (*core.MappingNode, error) {
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
		return nil, err
	}

	properties := ""
	if output.ResourceDescription != nil {
		properties = aws.ToString(output.ResourceDescription.Properties)
	}

	return a.cfnPropertiesToSpecState(properties, identifier, providerContext)
}

// The properties JSON may come from either a GetResource result or a request-status
// ResourceModel. The conversion strips write-only fields, filters out Bluelink provenance
// tags, records the primary identifier, and applies any behaviour post-read hook.
func (a *ccResourceActions) cfnPropertiesToSpecState(
	properties string,
	identifier string,
	providerContext provider.Context,
) (*core.MappingNode, error) {
	cfnNode, err := cfnJSONToMappingNode(properties)
	if err != nil {
		return nil, err
	}

	specState := CFNToSpec(cfnNode, a.config.Schema, a.config.Meta)
	if specState.Fields == nil {
		specState.Fields = map[string]*core.MappingNode{}
	}
	a.stripWriteOnlyFields(specState)
	a.filterExternalTags(specState, providerContext)
	specState.Fields[fieldPrimaryIdentifier] = core.MappingNodeFromString(identifier)

	if a.behaviour != nil && a.behaviour.AfterReadExternalState != nil {
		if err := a.behaviour.AfterReadExternalState(specState); err != nil {
			return nil, err
		}
	}

	return specState, nil
}

func (a *ccResourceActions) lookupIdentifierByTags(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (string, error) {
	filters := utils.BuildBluelinkTagFiltersForLookup(input)
	if filters == nil {
		return "", nil
	}

	service, err := a.getResourceGroupTaggingService(ctx, input.ProviderContext)
	if err != nil {
		return "", err
	}

	output, err := service.GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters:          filters,
		ResourceTypeFilters: []string{cfnServiceName(a.config.CFNType)},
	})
	if err != nil {
		return "", err
	}

	for _, mapping := range output.ResourceTagMappingList {
		if arn := aws.ToString(mapping.ResourceARN); arn != "" {
			return arn, nil
		}
	}
	return "", nil
}

func (a *ccResourceActions) stripWriteOnlyFields(specState *core.MappingNode) {
	stripWriteOnlyTopLevelFields(specState, a.config.Meta)
}

func (a *ccResourceActions) filterExternalTags(
	specState *core.MappingNode,
	providerContext provider.Context,
) {
	filterExternalTags(specState, a.config.Meta, providerContext)
}

// Removes top-level write-only properties from an external-state spec, since Cloud Control
// never returns them. Shared by the resource and data source engines.
func stripWriteOnlyTopLevelFields(specState *core.MappingNode, meta CCResourceMeta) {
	for _, field := range meta.WriteOnlyFields {
		if !strings.Contains(field, ".") {
			delete(specState.Fields, field)
		}
	}
}

// Strips Bluelink provenance tags from an external-state spec's tag property, leaving only
// user tags. Shared by the resource and data source engines.
func filterExternalTags(
	specState *core.MappingNode,
	meta CCResourceMeta,
	providerContext provider.Context,
) {
	if !meta.IsTaggable() {
		return
	}
	tagProperty := meta.TagPropertyName
	tagNode, exists := specState.Fields[tagProperty]
	if !exists {
		return
	}

	prefix := utils.GetBluelinkTagPrefix(providerContext.TaggingConfig())
	filtered := utils.FilterTagsMap(
		extractUserTags(tagNode, meta.TagShape),
		prefix,
	)
	if len(filtered) == 0 {
		delete(specState.Fields, tagProperty)
		return
	}
	specState.Fields[tagProperty] = buildTagNode(filtered, meta.TagShape)
}

// The lowercase AWS service segment of a CFN type name (e.g. "AWS::SQS::Queue" -> "sqs"),
// used for Resource Groups Tagging type filters.
func cfnServiceName(cfnType string) string {
	parts := strings.Split(cfnType, "::")
	if len(parts) >= 2 {
		return strings.ToLower(parts[1])
	}
	return ""
}
