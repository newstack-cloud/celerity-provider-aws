package lambdassm

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionParameterTreeLink returns a link implementation for a link from a Lambda function
// to an SSM parameter tree (a prefix-scoped store of parameters managed as one resource).
// The Lambda function is granted access to every parameter beneath the tree's path prefix
// with a single statement and, optionally, an environment variable holding the prefix. When
// the function is VPC-isolated, an SSM interface VPC endpoint is activated on the linked
// flex VPC.
//
// The grant and environment variable are identical in shape to the
// aws/lambda/function::aws/ssm/parameterPath link's where the runtime contract is a path prefix
// enumerated via ssm:GetParametersByPath either way, the difference is only that the tree
// also manages the parameters beneath the prefix.
func FunctionParameterTreeLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(FunctionToParameterLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToParameterLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__parameter_tree.md")

		actions := &functionParameterTreeLinkActions{
			functionParameterLinkActions{
				lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
				iamServiceFactory:    iamServiceFactory,
				ec2ServiceFactory:    ec2ServiceFactory,
				awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
			},
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/ssm/parameterTree",
			// ReconcileLinkNetworking reads the function's live VPC attachment, so the
			// placement link that establishes it has to run first.
			Requires: linkutils.NetworkAttachedRequired(provider.LinkPriorityResourceA),
			// It doesn't matter which resource is created first as the lambda function will
			// be configured to access the parameter tree once both have been created.
			Kind:             provider.LinkKindSoft,
			PriorityResource: provider.LinkPriorityResourceNone,
			PlainTextSummary: "A link that grants a Lambda function access to all SSM parameters " +
				"in a parameter tree.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           functionParameterTreeLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

// The per-parameter link's actions are embedded for the shared service getters; every
// link hook is explicitly wired to a tree-scoped override in the definition above.
type functionParameterTreeLinkActions struct {
	functionParameterLinkActions
}
