package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// VPCResource is an implementation of a flex resource that provides a convenient
// way to define a VPC and its associated resources based on presets
// that follow industry best practices.
func VPCResource(
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	yamlExample, _ := examples.ReadFile("resources/vpc_basic_yaml.md")
	yamlRefModeExample, _ := examples.ReadFile("resources/vpc_ref_mode.md")
	jsoncExample, _ := examples.ReadFile("resources/vpc_basic_jsonc.md")

	vpcActions := &vpcResourceActions{
		ec2ServiceFactory:                  ec2ServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
	}

	return &providerv1.ResourceDefinition{
		Type:  "aws/flex/vpc",
		Label: "AWS Flex VPC",
		PlainTextSummary: "A high-level resource for managing an AWS VPC and its " +
			"associated resources using presets that follow industry best practices.",
		FormattedDescription: "The resource type for a high-level VPC resource that " +
			"can be used to define a VPC and its associated resources based on " +
			"presets that follow industry best practices. " +
			"When using a flex VPC, you can simply link resources together and the networking " +
			"configuration required to _activate_ the link will be automatically configured.",
		Schema:  vpcResourceSchema(),
		IDField: "name",
		// A flex VPC typically needs to link out to the resources
		// that will be deployed in the VPC.
		CommonTerminal: false,
		ResourceCanLinkTo: []string{
			"aws/lambda/function",
		},
		FormattedExamples: []string{
			string(yamlExample),
			string(yamlRefModeExample),
			string(jsoncExample),
		},
		GetExternalStateFunc: vpcActions.GetExternalState,
		CreateFunc:           vpcActions.Create,
		UpdateFunc:           vpcActions.Update,
		DestroyFunc:          vpcActions.Destroy,
		StabilisedFunc:       vpcActions.Stabilised,
	}
}

type vpcResourceActions struct {
	ec2ServiceFactory                  pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *vpcResourceActions) getEC2ServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
	meta map[string]*core.MappingNode,
) (ec2service.Service, string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		meta,
	)
	if err != nil {
		return nil, "", err
	}

	return l.ec2ServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}

func (l *vpcResourceActions) getResourceGroupTaggingService(
	ctx context.Context,
	providerContext provider.Context,
	meta map[string]*core.MappingNode,
) (resgrouptagservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		meta,
	)
	if err != nil {
		return nil, err
	}

	return l.resourceGroupTaggingServiceFactory(awsConfig, providerContext), nil
}
