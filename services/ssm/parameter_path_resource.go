package ssm

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// ParameterPathResource returns the resource implementation for an AWS Systems Manager
// (SSM) parameter path. It is a synthetic Bluelink resource with no AWS API object behind
// it: it represents a parameter hierarchy (a path prefix such as "/my-app/config") as a
// whole, so that links can grant IAM access and inject configuration for the entire
// namespace rather than per parameter. Individual aws/ssm/parameter resources live
// beneath the path.
//
// The service factory is accepted for signature consistency with other resources in this
// service (and the test harnesses that construct them), but no SSM API call is ever made.
func ParameterPathResource(
	ssmServiceFactory pluginutils.ServiceFactory[*aws.Config, ssmservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	actions := &parameterPathResourceActions{}

	basicExample, _ := examples.ReadFile("examples/resources/parameter_path_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/parameter_path_complete.md")

	return &providerv1.ResourceDefinition{
		Type:  "aws/ssm/parameterPath",
		Label: "AWS SSM Parameter Path",
		PlainTextSummary: "A resource representing an AWS Systems Manager parameter hierarchy " +
			"(path prefix) as a whole, for namespace-level access grants and configuration.",
		FormattedDescription: "The resource type for an AWS Systems Manager (SSM) parameter path. " +
			"It is a synthetic resource with no AWS API object behind it, representing a parameter " +
			"hierarchy (a path prefix such as `/my-app/config`) as a whole so that links can grant " +
			"access to every parameter beneath the prefix with a single statement and inject the " +
			"prefix as a single environment variable.",
		Schema:         parameterPathResourceSchema(),
		IDField:        "path",
		TaggingSupport: provider.TaggingSupportNone,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		// A parameter path is a leaf configuration namespace that other resources link to;
		// it does not link out to other resources.
		CommonTerminal:       true,
		GetExternalStateFunc: actions.GetExternalState,
		CreateFunc:           actions.Create,
		UpdateFunc:           actions.Update,
		DestroyFunc:          actions.Destroy,
		StabilisedFunc:       actions.Stabilised,
	}
}

type parameterPathResourceActions struct{}
