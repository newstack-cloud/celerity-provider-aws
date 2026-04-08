package iam

import (
	"context"
	"embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

//go:embed examples/resources/*.md
var accessKeyExamples embed.FS

// AccessKeyResource returns a resource implementation for an AWS IAM Access Key.
func AccessKeyResource(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := accessKeyExamples.ReadFile("examples/resources/iam_access_key_basic.md")
	completeExample, _ := accessKeyExamples.ReadFile("examples/resources/iam_access_key_complete.md")
	jsoncExample, _ := accessKeyExamples.ReadFile("examples/resources/iam_access_key_jsonc.md")

	iamAccessKeyActions := &iamAccessKeyResourceActions{
		iamServiceFactory: iamServiceFactory,
		awsConfigStore:    awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/iam/accessKey",
		Label:            "AWS IAM Access Key",
		PlainTextSummary: "A resource for managing an AWS IAM access key.",
		FormattedDescription: "The resource type used to define an [IAM access key](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html) " +
			"that is deployed to AWS.",
		Schema:  iamAccessKeyResourceSchema(),
		IDField: "id",
		// An IAM access key is commonly referenced by other resources that need to use the access key
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
			string(jsoncExample),
		},
		GetExternalStateFunc: iamAccessKeyActions.GetExternalState,
		CreateFunc:           iamAccessKeyActions.Create,
		UpdateFunc:           iamAccessKeyActions.Update,
		DestroyFunc:          iamAccessKeyActions.Destroy,
		StabilisedFunc:       iamAccessKeyActions.Stabilised,
	}
}

type iamAccessKeyResourceActions struct {
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	awsConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
}

func (i *iamAccessKeyResourceActions) getIamService(
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
