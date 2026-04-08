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

// DynamoDBTableToLambdaFunctionLinkDeps is a type alias for the core
// dependencies for a link from a DynamoDB table to a Lambda function (stream trigger).
type DynamoDBTableToLambdaFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	dynamodbservice.Service,
	*aws.Config,
	lambdaservice.Service,
]

// DynamoDBTableLambdaFunctionLink returns a link implementation for
// a link from a DynamoDB table to a Lambda function.
// This creates an Event Source Mapping that triggers the Lambda function
// when records are written to the DynamoDB table's stream.
// It also adds IAM permissions to the Lambda function's execution role
// to read from the DynamoDB stream, and enables DynamoDB Streams on the
// table if not already enabled.
func DynamoDBTableLambdaFunctionLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
) func(DynamoDBTableToLambdaFunctionLinkDeps) provider.Link {
	return func(
		linkServiceDeps DynamoDBTableToLambdaFunctionLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/table__function.md")

		actions := &dynamoDBTableLambdaFunctionLinkActions{
			dynamodbServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
			lambdaServiceFactory:   linkServiceDeps.ResourceBService.ServiceFactory,
			iamServiceFactory:      iamServiceFactory,
			awsConfigStore:         linkServiceDeps.ResourceBService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/dynamodb/table",
			ResourceTypeB: "aws/lambda/function",
			// The DynamoDB table must be created first with streams enabled
			// before the Event Source Mapping can be created.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceA,
			PlainTextSummary:                "A link that configures a DynamoDB table to trigger a Lambda function via DynamoDB Streams.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           dynamoDBTableLambdaFunctionLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type dynamoDBTableLambdaFunctionLinkActions struct {
	dynamodbServiceFactory pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service]
	lambdaServiceFactory   pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	iamServiceFactory      pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	awsConfigStore         pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *dynamoDBTableLambdaFunctionLinkActions) getLambdaService(
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

func (l *dynamoDBTableLambdaFunctionLinkActions) getIamService(
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

func (l *dynamoDBTableLambdaFunctionLinkActions) getDynamoDBService(
	ctx context.Context,
	providerContext provider.Context,
) (dynamodbservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.dynamodbServiceFactory(awsConfig, providerContext), nil
}
