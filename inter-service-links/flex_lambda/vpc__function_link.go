package flexlambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// VPCToFunctionLinkDeps is a type alias for the core dependencies for a link from a
// flex VPC to a Lambda function. Resource A is the flex VPC (backed by the EC2 service)
// and resource B is the Lambda function.
type VPCToFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
]

// VPCFunctionLink returns a link implementation that places a Lambda function within a
// flex VPC. The function's VpcConfig is set from the VPC's computed subnets (selected by
// the aws.flexvpc.lambda.subnetType annotation) and Bluelink-managed security group.
func VPCFunctionLink() func(VPCToFunctionLinkDeps) provider.Link {
	return func(
		linkServiceDeps VPCToFunctionLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/vpc__function.md")

		actions := &vpcFunctionLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceBService.ServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceBService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/flex/vpc",
			ResourceTypeB: "aws/lambda/function",
			// The VPC must exist (and have stabilised its subnets and security group)
			// before the function can be placed into it.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceA,
			PlainTextSummary:                "A link that places a Lambda function within a flex VPC's subnets.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           vpcFunctionLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type vpcFunctionLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *vpcFunctionLinkActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}
