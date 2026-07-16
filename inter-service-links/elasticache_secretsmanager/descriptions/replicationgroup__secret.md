## ElastiCache Replication Group to Secrets Manager Secret (Redis AUTH token)

Applies a Redis AUTH token, stored in an AWS Secrets Manager secret, to an ElastiCache
(Redis/Valkey) replication group.

ElastiCache treats `authToken` as a write-only field with no computed output, and a Secrets
Manager secret's generated value is likewise never exposed as a computed attribute. As a result
a generated AUTH token cannot be referenced from the secret into the replication group's
`authToken` field through the blueprint. This hand-written link closes that gap by reading the
secret value at deploy time and applying it out-of-band.

When a replication group (resource A) links to a secret (resource B), the link:

1. **Reads the secret value** at deploy time via Secrets Manager `GetSecretValue`, using the
   secret's primary identifier (its `id`, which is the secret ARN) as the secret id.

2. **Applies the value as the replication group's AUTH token** via ElastiCache
   `ModifyReplicationGroup`, setting `AuthToken` together with an `AuthTokenUpdateStrategy`.
   On first configuration the strategy is `SET`; on subsequent updates (for example when the
   secret is rotated) the strategy is `ROTATE`, so an out-of-band secret change is reapplied
   without drift thrash. The strategy can be overridden with the
   `aws.elasticache.secretsmanager.authTokenUpdateStrategy` annotation.

The AUTH token value is **never** written into blueprint state or exposed as a computed field;
only the secret ARN and a boolean marker recording that the token was applied are retained in
link data.

Redis AUTH requires in-transit encryption. The replication group must set
`transitEncryptionEnabled` to `true`; if it does not, the link surfaces a validation
diagnostic at blueprint validation time rather than failing at deploy time.
