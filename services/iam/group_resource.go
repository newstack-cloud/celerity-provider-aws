package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"

	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// GroupResource returns a resource implementation for an AWS IAM Group.
func GroupResource(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/iam_group_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/iam_group_complete.md")

	iamGroupActions := &iamGroupResourceActions{
		iamServiceFactory:   iamServiceFactory,
		awsConfigStore:      awsConfigStore,
		uniqueNameGenerator: utils.IAMGroupNameGenerator,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/iam/group",
		Label:            "AWS IAM Group",
		PlainTextSummary: "A resource for managing an AWS IAM group.",
		FormattedDescription: "The resource type used to define an [IAM group](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_groups.html) " +
			"that is deployed to AWS.",
		Schema:  iamGroupResourceSchema(),
		IDField: "arn",
		// An IAM group is commonly referenced by other resources that need to grant permissions to the group
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		GetExternalStateFunc: iamGroupActions.GetExternalState,
		CreateFunc:           iamGroupActions.Create,
		UpdateFunc:           iamGroupActions.Update,
		DestroyFunc:          iamGroupActions.Destroy,
		StabilisedFunc:       iamGroupActions.Stabilised,
	}
}

type iamGroupResourceActions struct {
	iamServiceFactory   pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	awsConfigStore      pluginutils.ServiceConfigStore[*aws.Config]
	uniqueNameGenerator utils.UniqueNameGenerator
}

func (i *iamGroupResourceActions) getIamService(
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

func extractGroupNameFromARN(arn string) (string, error) {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return "", fmt.Errorf("invalid ARN format: %s", arn)
	}

	resourcePart := parts[5]
	resourceParts := strings.SplitN(resourcePart, "/", 2)
	if len(resourceParts) < 2 {
		return "", fmt.Errorf("invalid resource format in ARN: %s", resourcePart)
	}

	if resourceParts[0] != "group" {
		return "", fmt.Errorf("ARN is not for an IAM group: %s", arn)
	}

	return resourceParts[1], nil
}
