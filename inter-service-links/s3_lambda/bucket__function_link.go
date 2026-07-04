package s3lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	s3service "github.com/newstack-cloud/bluelink-provider-aws/services/s3/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// BucketToFunctionLinkDeps is the core dependency set for a link from an S3 bucket to a
// Lambda function (object event notification). Only the S3 service (resource A) is used;
// the Lambda permission is managed as a Cloud Control intermediary via the resource
// service.
type BucketToFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	s3service.Service,
	*aws.Config,
	lambdaservice.Service,
]

// BucketFunctionLink returns a link implementation that configures an S3 bucket to invoke
// a Lambda function on object events. The link grants S3 permission to invoke the
// function and merges a lambda entry into the bucket's notification configuration.
func BucketFunctionLink() func(BucketToFunctionLinkDeps) provider.Link {
	return func(deps BucketToFunctionLinkDeps) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/bucket__function.md")

		actions := &bucketFunctionLinkActions{
			s3ServiceFactory: deps.ResourceAService.ServiceFactory,
			awsConfigStore:   deps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA:                   "aws/s3/bucket",
			ResourceTypeB:                   "aws/lambda/function",
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceA,
			PlainTextSummary:                "A link that configures an S3 bucket to invoke a Lambda function on object event notifications.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           bucketFunctionLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type bucketFunctionLinkActions struct {
	s3ServiceFactory pluginutils.ServiceFactory[*aws.Config, s3service.Service]
	awsConfigStore   pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *bucketFunctionLinkActions) getS3Service(
	ctx context.Context,
	providerContext provider.Context,
) (s3service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.s3ServiceFactory(awsConfig, providerContext), nil
}
