package cloudcontrol

import (
	"context"
	"maps"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/overlays"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// CCResourceConfig is the per-type configuration for a Cloud Control–backed resource.
// Codegen emits one of these per included CloudFormation type.
type CCResourceConfig struct {
	// BlueprintType is the Bluelink resource type, e.g. "aws/sqs/queue".
	BlueprintType string
	// CFNType is the CloudFormation type name, e.g. "AWS::SQS::Queue".
	CFNType              string
	Label                string
	PlainTextSummary     string
	FormattedDescription string
	// Schema is the generated resource spec schema (the engine augments it with
	// internal Cloud Control bookkeeping fields).
	Schema *provider.ResourceDefinitionsSchema
	// Meta carries per-type information the generic engine cannot infer at runtime.
	Meta CCResourceMeta
	// FormattedExamples holds rendered example markdown for the resource type.
	FormattedExamples []string
	// CommonTerminal hints whether the resource is not expected to link out.
	CommonTerminal bool
}

// CCResource builds a provider.Resource backed by the generic Cloud Control engine,
// with all CRUD funcs parameterised by the resource type's CFN name and metadata.
func CCResource(
	config CCResourceConfig,
	cloudControlServiceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	config = applyMetaAdjustment(config)

	actions := &ccResourceActions{
		config:                             config,
		behaviour:                          overlays.BehaviourFor(config.BlueprintType),
		cloudControlServiceFactory:         cloudControlServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
	}

	taggingSupport := provider.TaggingSupportNone
	if config.Meta.IsTaggable() {
		taggingSupport = provider.TaggingSupportFull
	}

	definition := &providerv1.ResourceDefinition{
		Type:                 config.BlueprintType,
		Label:                config.Label,
		PlainTextSummary:     config.PlainTextSummary,
		FormattedDescription: config.FormattedDescription,
		Schema:               withInternalFields(config.Schema),
		IDField:              config.Meta.PrimaryIdentifierField,
		TaggingSupport:       taggingSupport,
		CommonTerminal:       config.CommonTerminal,
		FormattedExamples:    config.FormattedExamples,
		GetExternalStateFunc: actions.GetExternalState,
		CreateFunc:           actions.Create,
		UpdateFunc:           actions.Update,
		DestroyFunc:          actions.Destroy,
		StabilisedFunc:       actions.Stabilised,
		// Every generated resource declares the slow-to-stabilise types as its
		// stabilised dependencies, so a resource that depends on one of them (e.g. an
		// RDS instance whose endpoint is only assigned once provisioned) waits for it to
		// stabilise before deploying, and resolves its late computed fields correctly.
		StabilisedDependenciesFunc: stabiliseRequiredDependencies,
	}

	if actions.behaviour != nil && actions.behaviour.CustomValidate != nil {
		definition.CustomValidateFunc = actions.behaviour.CustomValidate
	}

	return definition
}

func applyMetaAdjustment(config CCResourceConfig) CCResourceConfig {
	adjustment := overlays.MetaAdjustmentFor(config.BlueprintType)
	if adjustment == nil {
		return config
	}
	config.Meta.JSONStringFields = removeStrings(
		config.Meta.JSONStringFields,
		adjustment.RemoveJSONStringFields,
	)
	config.Meta.FieldNameOverrides = mergeStringMaps(
		config.Meta.FieldNameOverrides,
		adjustment.AddFieldNameOverrides,
	)
	return config
}

func stabiliseRequiredDependencies(
	ctx context.Context,
	input *provider.ResourceStabilisedDependenciesInput,
) (*provider.ResourceStabilisedDependenciesOutput, error) {
	return &provider.ResourceStabilisedDependenciesOutput{
		StabilisedDependencies: overlays.StabiliseRequiredTypes(),
	}, nil
}

// Adds the engine's bookkeeping fields to the schema as Computed + IgnoreDrift strings so
// they persist in state without being reported as drift.
func withInternalFields(schema *provider.ResourceDefinitionsSchema) *provider.ResourceDefinitionsSchema {
	if schema == nil {
		return schema
	}
	if schema.Attributes == nil {
		schema.Attributes = map[string]*provider.ResourceDefinitionsSchema{}
	}
	addInternalField(
		schema,
		fieldRequestToken,
		"Internal: the Cloud Control operation request token used for stabilisation polling.",
	)
	addInternalField(
		schema,
		fieldPrimaryIdentifier,
		"Internal: the Cloud Control primary identifier used for get, update and delete operations.",
	)
	return schema
}

func mergeStringMaps(base, additions map[string]string) map[string]string {
	if len(additions) == 0 {
		return base
	}
	if base == nil {
		base = make(map[string]string, len(additions))
	}
	maps.Copy(base, additions)
	return base
}

func removeStrings(values, remove []string) []string {
	if len(values) == 0 || len(remove) == 0 {
		return values
	}

	excluded := make(map[string]bool, len(remove))
	for _, value := range remove {
		excluded[value] = true
	}

	var out []string
	for _, value := range values {
		if !excluded[value] {
			out = append(out, value)
		}
	}

	return out
}

func addInternalField(
	schema *provider.ResourceDefinitionsSchema,
	name string,
	description string,
) {
	if _, exists := schema.Attributes[name]; exists {
		return
	}
	schema.Attributes[name] = &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeString,
		Label:       name,
		Description: description,
		Computed:    true,
		IgnoreDrift: true,
		Nullable:    true,
	}
}

type ccResourceActions struct {
	config                             CCResourceConfig
	behaviour                          *overlays.Behaviour
	cloudControlServiceFactory         pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
}

func (a *ccResourceActions) getCloudControlService(
	ctx context.Context,
	providerContext provider.Context,
) (cloudcontrolservice.Service, error) {
	awsConfig, err := a.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return a.cloudControlServiceFactory(awsConfig, providerContext), nil
}

func (a *ccResourceActions) getResourceGroupTaggingService(
	ctx context.Context,
	providerContext provider.Context,
) (resgrouptagservice.Service, error) {
	awsConfig, err := a.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return a.resourceGroupTaggingServiceFactory(awsConfig, providerContext), nil
}
