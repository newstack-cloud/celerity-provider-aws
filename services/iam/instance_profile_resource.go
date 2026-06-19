package iam

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"

	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

//go:embed examples/resources/*.md
var instanceProfileExamples embed.FS

// InstanceProfileResource returns a resource implementation for an AWS IAM Instance Profile.
func InstanceProfileResource(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := instanceProfileExamples.ReadFile("examples/resources/iam_instance_profile_basic.md")
	completeExample, _ := instanceProfileExamples.ReadFile("examples/resources/iam_instance_profile_complete.md")

	iamInstanceProfileActions := &iamInstanceProfileResourceActions{
		iamServiceFactory:   iamServiceFactory,
		awsConfigStore:      awsConfigStore,
		uniqueNameGenerator: utils.IAMInstanceProfileNameGenerator,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/iam/instanceProfile",
		Label:            "AWS IAM Instance Profile",
		PlainTextSummary: "A resource for managing an AWS IAM instance profile.",
		FormattedDescription: "The resource type used to define an [IAM instance profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2_instance-profiles.html) " +
			"that is deployed to AWS.",
		Schema:  iamInstanceProfileResourceSchema(),
		IDField: "arn",
		// An IAM instance profile is commonly used by EC2 instances
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		GetExternalStateFunc: iamInstanceProfileActions.GetExternalState,
		CreateFunc:           iamInstanceProfileActions.Create,
		UpdateFunc:           iamInstanceProfileActions.Update,
		DestroyFunc:          iamInstanceProfileActions.Destroy,
		StabilisedFunc:       iamInstanceProfileActions.Stabilised,
	}
}

type iamInstanceProfileResourceActions struct {
	iamServiceFactory   pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	awsConfigStore      pluginutils.ServiceConfigStore[*aws.Config]
	uniqueNameGenerator utils.UniqueNameGenerator
}

func (i *iamInstanceProfileResourceActions) getIamService(
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

// Extracts the instance profile name from an IAM instance profile ARN
// ARN format: arn:aws:iam::123456789012:instance-profile/instance-profile-name.
func extractInstanceProfileNameFromARN(arn string) (string, error) {
	if arn == "" {
		return "", fmt.Errorf("ARN cannot be empty")
	}

	parts := strings.Split(arn, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid IAM instance profile ARN format: %s", arn)
	}

	instanceProfileName := parts[len(parts)-1]
	if instanceProfileName == "" {
		return "", fmt.Errorf("instance profile name cannot be empty in ARN: %s", arn)
	}

	return instanceProfileName, nil
}

// Extracts the role name from a role specification
// which can be either a role name or a role ARN.
func extractRoleNameFromRoleSpec(roleSpec string) (string, error) {
	if roleSpec == "" {
		return "", fmt.Errorf("role specification cannot be empty")
	}

	// If it's an ARN, extract the role name
	if strings.HasPrefix(roleSpec, "arn:aws:iam::") {
		parts := strings.Split(roleSpec, "/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid IAM role ARN format: %s", roleSpec)
		}
		return parts[len(parts)-1], nil
	}

	// Otherwise, it's already a role name
	return roleSpec, nil
}
