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

// RoleResource returns a resource implementation for an AWS IAM Role.
func RoleResource(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/iam_role_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/iam_role_complete.md")
	jsoncExample, _ := examples.ReadFile("examples/resources/iam_role_jsonc.md")

	iamRoleActions := &iamRoleResourceActions{
		iamServiceFactory:                  iamServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
		uniqueNameGenerator:                utils.IAMRoleNameGenerator,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/iam/role",
		Label:            "AWS IAM Role",
		PlainTextSummary: "A resource for managing an AWS IAM role.",
		FormattedDescription: "The resource type used to define an [IAM role](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html) " +
			"that is deployed to AWS.",
		Schema:         iamRoleResourceSchema(),
		IDField:        "arn",
		TaggingSupport: provider.TaggingSupportFull,
		// An IAM role is commonly used by other resources that need to assume permissions
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
			string(jsoncExample),
		},
		GetExternalStateFunc: iamRoleActions.GetExternalState,
		CreateFunc:           iamRoleActions.Create,
		UpdateFunc:           iamRoleActions.Update,
		DestroyFunc:          iamRoleActions.Destroy,
		StabilisedFunc:       iamRoleActions.Stabilised,
	}
}

type iamRoleResourceActions struct {
	iamServiceFactory                  pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
	uniqueNameGenerator                utils.UniqueNameGenerator
}

func (i *iamRoleResourceActions) getIamService(
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

// extractRoleNameFromARN extracts the role name from an IAM role ARN
// ARN format: arn:aws:iam::123456789012:role/role-name.
func extractRoleNameFromARN(arn string) (string, error) {
	if arn == "" {
		return "", fmt.Errorf("ARN cannot be empty")
	}

	parts := strings.Split(arn, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid IAM role ARN format: %s", arn)
	}

	roleName := parts[len(parts)-1]
	if roleName == "" {
		return "", fmt.Errorf("role name cannot be empty in ARN: %s", arn)
	}

	return roleName, nil
}
