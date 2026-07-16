package elasticachesecretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	elasticacheservice "github.com/newstack-cloud/bluelink-provider-aws/services/elasticache/service"
	secretsmanagerservice "github.com/newstack-cloud/bluelink-provider-aws/services/secretsmanager/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// ReplicationGroupToSecretLinkDeps is a type alias for the core dependencies for a link from an
// ElastiCache replication group to a Secrets Manager secret. Both the replication group and the
// secret are Cloud Control–backed resources, so both resource services are the Cloud Control
// service; the link reaches the ElastiCache and Secrets Manager APIs out-of-band via the
// hand-written service factories passed to ReplicationGroupSecretLink.
type ReplicationGroupToSecretLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	cloudcontrolservice.Service,
]

// ReplicationGroupSecretLink returns a link implementation for a link from an ElastiCache
// (Redis/Valkey) replication group to a Secrets Manager secret. The replication group is given a
// Redis AUTH token whose value is read from the linked secret at deploy time, worked around the
// fact that ElastiCache's authToken is a write-only field and the secret's value is never a
// computed output.
func ReplicationGroupSecretLink(
	elastiCacheServiceFactory pluginutils.ServiceFactory[*aws.Config, elasticacheservice.Service],
	secretsManagerServiceFactory pluginutils.ServiceFactory[*aws.Config, secretsmanagerservice.Service],
) func(ReplicationGroupToSecretLinkDeps) provider.Link {
	return func(
		linkServiceDeps ReplicationGroupToSecretLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/replicationgroup__secret.md")

		actions := &replicationGroupSecretLinkActions{
			elastiCacheServiceFactory:    elastiCacheServiceFactory,
			secretsManagerServiceFactory: secretsManagerServiceFactory,
			awsConfigStore:               linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/elasticache/replicationGroup",
			ResourceTypeB: "aws/secretsmanager/secret",
			// The replication group (A) is configured with the AUTH token once both it and the
			// secret have been created; ordering the group first avoids an extra modify.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceA,
			PlainTextSummary:                "A link that applies a Redis AUTH token from a Secrets Manager secret to an ElastiCache replication group.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           replicationGroupSecretLinkAnnotations(),
			ValidateFunc:                    actions.Validate,
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type replicationGroupSecretLinkActions struct {
	elastiCacheServiceFactory    pluginutils.ServiceFactory[*aws.Config, elasticacheservice.Service]
	secretsManagerServiceFactory pluginutils.ServiceFactory[*aws.Config, secretsmanagerservice.Service]
	awsConfigStore               pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *replicationGroupSecretLinkActions) getElastiCacheService(
	ctx context.Context,
	providerContext provider.Context,
) (elasticacheservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.elastiCacheServiceFactory(awsConfig, providerContext), nil
}

func (l *replicationGroupSecretLinkActions) getSecretsManagerService(
	ctx context.Context,
	providerContext provider.Context,
) (secretsmanagerservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.secretsManagerServiceFactory(awsConfig, providerContext), nil
}
