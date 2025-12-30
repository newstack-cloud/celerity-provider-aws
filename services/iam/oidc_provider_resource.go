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

// OIDCProviderResource returns a resource implementation for an AWS IAM OIDC Provider.
func OIDCProviderResource(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/iam_oidc_provider_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/iam_oidc_provider_complete.md")
	jsoncExample, _ := examples.ReadFile("examples/resources/iam_oidc_provider_jsonc.md")

	iamOIDCProviderActions := &iamOIDCProviderResourceActions{
		iamServiceFactory:                  iamServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
		uniqueNameGenerator:                utils.IAMOIDCProviderUrlGenerator,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/iam/oidcProvider",
		Label:            "AWS IAM OIDC Provider",
		PlainTextSummary: "A resource for managing an AWS IAM OIDC provider.",
		FormattedDescription: "The resource type used to define an [IAM OIDC provider](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html) " +
			"that is deployed to AWS. OIDC providers are entities in IAM that describe an external identity provider (IdP) service that supports the " +
			"OpenID Connect (OIDC) standard.",
		Schema:         iamOIDCProviderResourceSchema(),
		IDField:        "arn",
		TaggingSupport: provider.TaggingSupportFull,
		// An IAM OIDC provider is commonly referenced by other resources that need to use it
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
			string(jsoncExample),
		},
		ResourceCanLinkTo:    []string{},
		GetExternalStateFunc: iamOIDCProviderActions.GetExternalState,
		CreateFunc:           iamOIDCProviderActions.Create,
		UpdateFunc:           iamOIDCProviderActions.Update,
		DestroyFunc:          iamOIDCProviderActions.Destroy,
		StabilisedFunc:       iamOIDCProviderActions.Stabilised,
	}
}

type iamOIDCProviderResourceActions struct {
	iamServiceFactory                  pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
	uniqueNameGenerator                utils.UniqueNameGenerator
}

func (i *iamOIDCProviderResourceActions) getIamService(
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

func (i *iamOIDCProviderResourceActions) getResourceGroupTaggingService(
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

// extractUrlFromArn extracts the URL from an OIDC provider ARN
// ARN format: arn:aws:iam::account-id:oidc-provider/oidc.example.com.
func extractUrlFromArn(arn string) (string, error) {
	parts := strings.Split(arn, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid OIDC provider ARN format: %s", arn)
	}

	// The URL is everything after the last slash
	url := parts[1]

	// Add https:// prefix if not present (AWS stores it without the protocol)
	if !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	return url, nil
}
