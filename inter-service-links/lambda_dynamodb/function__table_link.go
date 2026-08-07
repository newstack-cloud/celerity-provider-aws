package lambdadynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
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
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(FunctionToDynamoDBTableLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToDynamoDBTableLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__table.md")

		actions := &lambdaFunctionDynamoDBTableLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			ec2ServiceFactory:    ec2ServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/dynamodb/table",
			// ReconcileLinkNetworking reads the function's live VPC attachment, so the
			// placement link that establishes it has to run first.
			Requires: linkutils.NetworkAttachedRequired(provider.LinkPriorityResourceA),
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
	// A VPC-attached caller reaches DynamoDB through a gateway VPC endpoint, which
	// this link provisions.
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
}

func (l *lambdaFunctionDynamoDBTableLinkActions) getEC2Service(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}

	return l.ec2ServiceFactory(awsConfig, providerContext), nil
}

// The region the caller is deployed in, which the endpoint service name is built from.
func (l *lambdaFunctionDynamoDBTableLinkActions) getLambdaServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, "", err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
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
