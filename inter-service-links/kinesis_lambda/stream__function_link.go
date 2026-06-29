package kinesislambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// StreamToFunctionLinkDeps is a type alias for the core dependencies for a link
// from a Kinesis data stream to a Lambda function (stream trigger). The stream is a
// Cloud Control–backed resource, so resource A's service is the Cloud Control service;
// the link only reads the stream ARN from state and does not call it.
type StreamToFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
]

// StreamFunctionLink returns a link implementation for a link from a Kinesis data
// stream to a Lambda function.
// This creates an Event Source Mapping that triggers the Lambda function when records
// are written to the Kinesis stream. It also adds IAM permissions to the Lambda
// function's execution role to read from the stream.
func StreamFunctionLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
) func(StreamToFunctionLinkDeps) provider.Link {
	return func(
		linkServiceDeps StreamToFunctionLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/stream__function.md")

		actions := &kinesisStreamLambdaFunctionLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceBService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceBService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/kinesis/stream",
			ResourceTypeB: "aws/lambda/function",
			// The Kinesis stream must exist before the Event Source Mapping can be
			// created, so the stream is the priority resource.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceA,
			PlainTextSummary:                "A link that configures a Kinesis data stream to trigger a Lambda function via an event source mapping.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           kinesisStreamLambdaFunctionLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type kinesisStreamLambdaFunctionLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *kinesisStreamLambdaFunctionLinkActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}

func (l *kinesisStreamLambdaFunctionLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.iamServiceFactory(awsConfig, providerContext), nil
}
