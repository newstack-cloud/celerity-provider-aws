package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// SAMLProviderResource returns a resource implementation for an AWS IAM SAML Provider.
func SAMLProviderResource(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/iam_saml_provider_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/iam_saml_provider_complete.md")
	jsoncExample, _ := examples.ReadFile("examples/resources/iam_saml_provider_jsonc.md")

	iamSAMLProviderActions := &iamSAMLProviderResourceActions{
		iamServiceFactory:                  iamServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
		uniqueNameGenerator:                utils.IAMSAMLProviderNameGenerator,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/iam/samlProvider",
		Label:            "AWS IAM SAML Provider",
		PlainTextSummary: "A resource for managing an AWS IAM SAML provider.",
		FormattedDescription: "The resource type used to define an [IAM SAML provider](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_saml.html) " +
			"that is deployed to AWS. SAML providers are entities in IAM that describe an external identity provider (IdP) service that supports the " +
			"Security Assertion Markup Language (SAML) 2.0 standard.",
		Schema:         iamSAMLProviderResourceSchema(),
		IDField:        "arn",
		TaggingSupport: provider.TaggingSupportFull,
		// An IAM SAML provider is commonly referenced by other resources that need to use it
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
			string(jsoncExample),
		},
		GetExternalStateFunc: iamSAMLProviderActions.GetExternalState,
		CreateFunc:           iamSAMLProviderActions.Create,
		UpdateFunc:           iamSAMLProviderActions.Update,
		DestroyFunc:          iamSAMLProviderActions.Destroy,
		StabilisedFunc:       iamSAMLProviderActions.Stabilised,
	}
}

type iamSAMLProviderResourceActions struct {
	iamServiceFactory                  pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
	uniqueNameGenerator                utils.UniqueNameGenerator
}

func (i *iamSAMLProviderResourceActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := i.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return i.iamServiceFactory(awsConfig, providerContext), nil
}

func (i *iamSAMLProviderResourceActions) getResourceGroupTaggingService(
	ctx context.Context,
	providerContext provider.Context,
) (resgrouptagservice.Service, error) {
	// IAM is a global service - Resource Groups Tagging API must use us-east-1
	// to query IAM resources regardless of the configured region.
	meta := map[string]*core.MappingNode{
		"region": core.MappingNodeFromString("us-east-1"),
	}
	awsConfig, err := i.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		meta,
	)
	if err != nil {
		return nil, err
	}

	return i.resourceGroupTaggingServiceFactory(awsConfig, providerContext), nil
}

// extractNameFromArn extracts the name from a SAML provider ARN
// ARN format: arn:aws:iam::account-id:saml-provider/provider-name.
func extractNameFromArn(arn string) (string, error) {
	parts := strings.Split(arn, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid SAML provider ARN format: %s", arn)
	}

	// The name is everything after the last slash
	return parts[1], nil
}
