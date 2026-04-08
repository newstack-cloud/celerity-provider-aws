package lambdadynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionToDynamoDBTableLinkDeps is a type alias for the core
// dependencies for a link from a lambda function to a DynamoDB table.
type FunctionToDynamoDBTableLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
]

// FunctionDynamoDBTableLink returns a link implementation for
// a link from a lambda function to a DynamoDB table.
// The lambda function will be configured with environment variables
// and IAM permissions to access the DynamoDB table.
func FunctionDynamoDBTableLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
) func(FunctionToDynamoDBTableLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToDynamoDBTableLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__table.md")

		actions := &lambdaFunctionDynamoDBTableLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/dynamodb/table",
			// It doesn't matter which resource is created first,
			// the lambda function will be configured to access the DynamoDB table
			// once both have been created.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceNone,
			PlainTextSummary:                "A link that grants a Lambda function read/write access to a DynamoDB table.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           lambdaFunctionDynamoDBTableLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type lambdaFunctionDynamoDBTableLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
}

func (l *lambdaFunctionDynamoDBTableLinkActions) getLambdaService(
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

func (l *lambdaFunctionDynamoDBTableLinkActions) getIamService(
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
