package elasticachesecretsmanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// UpdateResourceA reads the Redis AUTH token from the linked Secrets Manager secret at deploy
// time and applies it to the ElastiCache replication group via ModifyReplicationGroup. The token
// value is never written into blueprint state or a computed field.
func (l *replicationGroupSecretLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// On destroy, we do not remove the AUTH token: deleting the token (DELETE strategy) is
	// destructive to live clients, and the replication group is typically being torn down anyway.
	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")

	replicationGroupID, hasReplicationGroupID := extractReplicationGroupID(input.ResourceInfo)
	if !hasReplicationGroupID {
		return nil, fmt.Errorf(
			"replication group id could not be retrieved from the ElastiCache replication group %q",
			pluginutils.GetResourceName(input.ResourceInfo),
		)
	}

	// The Secrets Manager secret's primary identifier (its "id" field) is the secret ARN.
	secretARN, hasSecretARN := extractSecretID(input.OtherResourceInfo)
	if !hasSecretARN {
		return nil, fmt.Errorf(
			"secret ARN could not be retrieved from the linked Secrets Manager secret %q",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	authToken, err := l.readAuthToken(ctx, providerCtx, secretARN)
	if err != nil {
		return nil, err
	}

	strategy := resolveAuthTokenUpdateStrategy(input)

	elastiCacheService, err := l.getElastiCacheService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	_, err = elastiCacheService.ModifyReplicationGroup(ctx, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:      aws.String(replicationGroupID),
		AuthToken:               aws.String(authToken),
		AuthTokenUpdateStrategy: strategy,
		ApplyImmediately:        aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf(
			"failed to apply Redis AUTH token to ElastiCache replication group %q: %w",
			replicationGroupID,
			err,
		)
	}

	// Record only that the token was applied and from which secret (an ARN, not a secret
	// value) plus the strategy used. The AUTH token value itself is never persisted.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			"authToken",
			core.MappingNodeFields(
				"configured", core.MappingNodeFromBool(true),
				"secretArn", core.MappingNodeFromString(secretARN),
				"updateStrategy", core.MappingNodeFromString(string(strategy)),
			),
		),
	}, nil
}

// Reads the secret's string value to use as the Redis AUTH token. The value is held only in
// memory and passed straight to ModifyReplicationGroup; it is never returned in link data.
func (l *replicationGroupSecretLinkActions) readAuthToken(
	ctx context.Context,
	providerCtx provider.Context,
	secretARN string,
) (string, error) {
	secretsManagerService, err := l.getSecretsManagerService(ctx, providerCtx)
	if err != nil {
		return "", err
	}

	secretValueOutput, err := secretsManagerService.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretARN),
	})
	if err != nil {
		return "", fmt.Errorf(
			"failed to read the Redis AUTH token from Secrets Manager secret %q: %w",
			secretARN,
			err,
		)
	}

	authToken := aws.ToString(secretValueOutput.SecretString)
	if authToken == "" {
		return "", fmt.Errorf(
			"secrets Manager secret %q has no string value to use as a Redis AUTH token",
			secretARN,
		)
	}

	return authToken, nil
}

// UpdateResourceB is a no-op: the secret is not modified by this link, it is only read.
func (l *replicationGroupSecretLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

// UpdateIntermediaryResources is a no-op: this link manages no intermediary resources, the AUTH
// token is applied directly to the replication group via UpdateResourceA.
func (l *replicationGroupSecretLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData:                   core.MappingNodeFields(),
	}, nil
}

// Determines the AuthTokenUpdateStrategy for ModifyReplicationGroup. An explicit annotation wins;
// otherwise the strategy is SET on first configuration (create) and ROTATE on subsequent updates,
// so a rotated secret is reapplied without drift thrash.
func resolveAuthTokenUpdateStrategy(
	input *provider.LinkUpdateResourceInput,
) elasticachetypes.AuthTokenUpdateStrategyType {
	strategy, hasStrategy := pluginutils.GetStringAnnotation(
		input.ResourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: "aws.elasticache.secretsmanager.authTokenUpdateStrategy",
		},
	)
	if hasStrategy && strategy != "" {
		return elasticachetypes.AuthTokenUpdateStrategyType(strategy)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeUpdate {
		return elasticachetypes.AuthTokenUpdateStrategyTypeRotate
	}
	return elasticachetypes.AuthTokenUpdateStrategyTypeSet
}

func extractReplicationGroupID(resourceInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(resourceInfo)
	idNode, has := pluginutils.GetValueByPath("$.replicationGroupId", spec)
	if !has || idNode == nil {
		return "", false
	}
	value := core.StringValue(idNode)
	return value, value != ""
}

func extractSecretID(secretInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(secretInfo)
	arnNode, has := pluginutils.GetValueByPath("$.id", spec)
	if !has || arnNode == nil {
		return "", false
	}
	value := core.StringValue(arnNode)
	return value, value != ""
}
