package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"

	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// UserResource returns a resource implementation for an AWS IAM User.
func UserResource(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/iam_user_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/iam_user_complete.md")

	iamUserActions := &iamUserResourceActions{
		iamServiceFactory:                  iamServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
		uniqueNameGenerator:                utils.IAMUserNameGenerator,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/iam/user",
		Label:            "AWS IAM User",
		PlainTextSummary: "A resource for managing an AWS IAM user.",
		FormattedDescription: "The resource type used to define an [IAM user](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users.html) " +
			"that is deployed to AWS.",
		Schema:         iamUserResourceSchema(),
		IDField:        "arn",
		TaggingSupport: provider.TaggingSupportFull,
		// An IAM user is commonly referenced by other resources that need to grant permissions to the user
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		GetExternalStateFunc: iamUserActions.GetExternalState,
		CreateFunc:           iamUserActions.Create,
		UpdateFunc:           iamUserActions.Update,
		DestroyFunc:          iamUserActions.Destroy,
		StabilisedFunc:       iamUserActions.Stabilised,
	}
}

type iamUserResourceActions struct {
	iamServiceFactory                  pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
	uniqueNameGenerator                utils.UniqueNameGenerator
}

func (i *iamUserResourceActions) getIamService(
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

func extractUserNameFromARN(arn string) (string, error) {
	if arn == "" {
		return "", fmt.Errorf("ARN cannot be empty")
	}

	parts := strings.Split(arn, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid IAM user ARN format: %s", arn)
	}

	userName := parts[len(parts)-1]
	if userName == "" {
		return "", fmt.Errorf("user name cannot be empty in ARN: %s", arn)
	}

	return userName, nil
}
