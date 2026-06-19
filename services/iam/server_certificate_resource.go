package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// ServerCertificateResource returns a resource implementation for an AWS IAM Server Certificate.
func ServerCertificateResource(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	return serverCertificateResourceWithNameGen(
		iamServiceFactory,
		resourceGroupTaggingServiceFactory,
		awsConfigStore,
		utils.IAMServerCertificateNameGenerator,
	)
}

func serverCertificateResourceWithNameGen(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
	uniqueNameGenerator utils.UniqueNameGenerator,
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/iam_server_certificate_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/iam_server_certificate_complete.md")

	actions := &iamServerCertificateResourceActions{
		iamServiceFactory:                  iamServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
		uniqueNameGenerator:                uniqueNameGenerator,
	}
	return &providerv1.ResourceDefinition{
		Type:                 "aws/iam/serverCertificate",
		Label:                "AWS IAM Server Certificate",
		PlainTextSummary:     "A resource for managing an AWS IAM server certificate.",
		FormattedDescription: "The resource type used to define an [IAM server certificate](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_server-certs.html) that is deployed to AWS.",
		Schema:               iamServerCertificateResourceSchema(),
		IDField:              "arn",
		TaggingSupport:       provider.TaggingSupportFull,
		// To avoid trying to create a server certificate with the same name as the existing one,
		// certificates should be deleted before the replacement is created.
		DestroyBeforeCreate:  true,
		CommonTerminal:       false,
		FormattedExamples:    []string{string(basicExample), string(completeExample)},
		GetExternalStateFunc: actions.GetExternalState,
		CreateFunc:           actions.Create,
		UpdateFunc:           actions.Update,
		DestroyFunc:          actions.Destroy,
		StabilisedFunc:       actions.Stabilised,
	}
}

type iamServerCertificateResourceActions struct {
	iamServiceFactory                  pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
	uniqueNameGenerator                utils.UniqueNameGenerator // Not used for server certs, but included for consistency
}

func (a *iamServerCertificateResourceActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := a.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return a.iamServiceFactory(awsConfig, providerContext), nil
}

func (a *iamServerCertificateResourceActions) getResourceGroupTaggingService(
	ctx context.Context,
	providerContext provider.Context,
) (resgrouptagservice.Service, error) {
	// IAM is a global service - Resource Groups Tagging API must use us-east-1
	// to query IAM resources regardless of the configured region.
	meta := map[string]*core.MappingNode{
		"region": core.MappingNodeFromString("us-east-1"),
	}
	awsConfig, err := a.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		meta,
	)
	if err != nil {
		return nil, err
	}
	return a.resourceGroupTaggingServiceFactory(awsConfig, providerContext), nil
}
