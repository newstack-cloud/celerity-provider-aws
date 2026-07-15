package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// ParameterResource returns the resource implementation for an AWS Systems Manager
// (SSM) parameter. Unlike most resources in this provider it is a custom implementation against the
// SSM SDK rather than Cloud Control, because CloudFormation (and therefore Cloud Control)
// cannot create SecureString parameters, and SecureString is essential to low-cost
// configuration management as an alternative to Secrets Manager.
func ParameterResource(
	ssmServiceFactory pluginutils.ServiceFactory[*aws.Config, ssmservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	actions := &parameterResourceActions{
		ssmServiceFactory: ssmServiceFactory,
		awsConfigStore:    awsConfigStore,
	}

	basicExample, _ := examples.ReadFile("examples/resources/parameter_basic.md")
	secureExample, _ := examples.ReadFile("examples/resources/parameter_secure.md")
	replicaExample, _ := examples.ReadFile("examples/resources/parameter_replica.md")

	return &providerv1.ResourceDefinition{
		Type:  "aws/ssm/parameter",
		Label: "AWS SSM Parameter",
		PlainTextSummary: "A resource for managing an AWS Systems Manager parameter, " +
			"including plaintext (String/StringList) and KMS-encrypted (SecureString) values.",
		FormattedDescription: "The resource type for an AWS Systems Manager (SSM) parameter. " +
			"Supports plaintext `String`/`StringList` values as well as KMS-encrypted `SecureString` " +
			"values, providing a low-cost alternative to Secrets Manager for configuration secrets.",
		Schema:         parameterResourceSchema(),
		IDField:        "arn",
		TaggingSupport: provider.TaggingSupportFull,
		FormattedExamples: []string{
			string(basicExample),
			string(secureExample),
			string(replicaExample),
		},
		// A parameter is a leaf configuration value that other resources link to; it does
		// not link out to other resources.
		CommonTerminal:       true,
		GetExternalStateFunc: actions.GetExternalState,
		CreateFunc:           actions.Create,
		UpdateFunc:           actions.Update,
		DestroyFunc:          actions.Destroy,
		StabilisedFunc:       actions.Stabilised,
	}
}

type parameterResourceActions struct {
	ssmServiceFactory pluginutils.ServiceFactory[*aws.Config, ssmservice.Service]
	awsConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
}

func (a *parameterResourceActions) getSSMServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
	meta map[string]*core.MappingNode,
) (ssmservice.Service, string, error) {
	awsConfig, err := a.awsConfigStore.FromProviderContext(ctx, providerContext, meta)
	if err != nil {
		return nil, "", err
	}

	return a.ssmServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}

func parameterRegionMeta(specData *core.MappingNode) map[string]*core.MappingNode {
	region, hasRegion := pluginutils.GetValueByPath("$.region", specData)
	if !hasRegion || core.StringValue(region) == "" {
		return nil
	}
	return map[string]*core.MappingNode{"region": region}
}
