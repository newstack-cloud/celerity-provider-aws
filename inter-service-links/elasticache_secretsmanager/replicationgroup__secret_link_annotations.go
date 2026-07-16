package elasticachesecretsmanager

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func replicationGroupSecretLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		// resourceA (ElastiCache replication group) annotation - controls how the AUTH token
		// is applied via ModifyReplicationGroup. It stays on resourceA because it configures
		// the modification of the replication group itself.
		"aws/elasticache/replicationGroup::aws.elasticache.secretsmanager.authTokenUpdateStrategy": {
			Name:  "aws.elasticache.secretsmanager.authTokenUpdateStrategy",
			Label: "Auth Token Update Strategy",
			Type:  core.ScalarTypeString,
			Description: "The strategy ElastiCache uses when applying the AUTH token via ModifyReplicationGroup. " +
				"ROTATE adds the new token while keeping the previous one valid for a rotation window, which is " +
				"safer for live clients. SET retires the previous tokens, leaving a single valid token, and is " +
				"only accepted by ElastiCache after a previous ROTATE; on first configuration there is nothing " +
				"to retire, so the link falls back to ROTATE. When unset, the link uses ROTATE for both first " +
				"configuration and subsequent updates.",
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("SET"),
				core.ScalarFromString("ROTATE"),
			},
			Required: false,
		},
	}
}
