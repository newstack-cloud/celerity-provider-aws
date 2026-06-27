# Link Design Guidance for Bluelink AWS Provider

This document outlines the approach to modelling links for AWS services in the Bluelink AWS Provider plugin.

A few things to note around the behaviour and limitations of links:

- For one-to-many relationships where annotations to “fine-tune” the relationship are on the side of the “one”, all annotations support a dynamic annotation key
that includes the name of the resource in the “many” that the annotation should apply to.
For example, in a link `aws/sqs/queue` → `aws/lambda/function` where there are multiple queues targeting a single lambda function, instead of using the
`aws.lambda.sqs.batchSize` annotation for all queues, you can use `aws.lambda.sqs.queue1.batchSize` so the annotation is only used for that specific queue.
- When deciding whether there should be a link implementation to connect resources together is that when a resource depends on another directly
(e.g. AWS resource references ARN for another resource, which is a required field), a link **MUST** not be implemented,
for this, practitioners can use implicit dependency via a reference in same way they would with other IaC tools like CloudFormation or Terraform.
***The exception to the rule is for IAM and networking resources as they are often the glue which enables a link between two resources.
For example, a Lambda function requires a Role ARN for the function’s execution role, the permissions of the role will often depend on links
between functions and other resources in a blueprint. For this case, a placeholder role should be created with minimal permissions that
link implementations can update with the permissions that “activate” the link between the function and other resources.***
- Another rule to take into account when deciding whether there should be a link implementation to connect two resources together is that if intermediary resources are needed
to connect the two given resource types together, if the intermediary resources have a dual purpose of linking two resources together and configuring the behaviour of
one of the two resources that is unrelated to the connection between the two, these links **MUST** not be implemented.
- No need to create data sources for “mapping”, “glue” or intermediary resources (e.g. event source mappings for lambda functions).
- Existing resources in the same blueprint can be used as intermediary resources, (e.g. an IAM role for a lambda execution role). These resources are usually dependencies of other resources,
they only need to be defined in the most minimal form as a blueprint resource and the details will be populated by the link implementations. For example, a lambda function depends on an IAM role resource,
however, the resource does not need to be fully implemented with all the permissions, policies will be updated by links from lambdas to other resource types. Custom permissions can still be defined in the role
defined in the blueprint and these will be combined with the link-populated permissions. Existing intermediary resources MUST not be included in the `IntermediaryResourceStates` property in the response for the
`UpdateIntermediaryResources` method.

## Links never carry structured user input

Links **MUST NOT** carry rich, structured user input. The data attached to a link relationship falls into exactly three kinds, each with a fixed home:

- **Scalar fine-tuning** (e.g. `batchSize`, `maxReceiveCount`, `accessLevel`, `populateEnvVars`) — modelled as **annotations**. Annotations are scalar only (`boolean | string | integer | enum`) and MUST NOT be expanded to objects, maps or arrays. An opaque string that AWS itself treats as a single atomic value (e.g. a JSON `Input` blob or a `filterCriteria` pattern passed through verbatim) is acceptable as a `string` annotation.
- **Derived/computed relationship data** (e.g. assembled IAM policy documents, statement IDs, ARNs the link composes from the context of resource A and resource B). This is produced by the link implementation and surfaced via the link data projection and any managed-intermediary state. This may be arbitrarily complex; it is derived, not user-supplied. **A link that owns an intermediary resource (a managed Cloud Control resource, or an SDK-created one such as an event source mapping) MUST project that intermediary into its link data so the intermediary's create/update/destroy is visible in plans/previews**. An intermediary-only link must never stage as "no changes". The projection lives under an `intermediaries` map keyed by the intermediary's deterministic id; see **Surfacing Intermediary Changes at Stage Time** in `CREATE_LINK_INSTRUCTIONS.md` for the `linkutils` helpers and pattern.
- **Rich structured user input** (e.g. an EventBridge target's `inputTransformer` / `httpParameters` / `ecsParameters`, an SNS subscription's `filterPolicy`). This is the signature of a **join-resource**. It MUST be modelled as a **resource** in its own right, with the structured input on the resource's spec.

When a relationship requires rich structured user input, model it as a resource: put the configuration on the resource spec, express the connections to the two related resources as substitution references (per the direct-dependency rule above), and use links only for the derived side-effects that "activate" the relationship (permissions, resource policies, IAM role policies).

This keeps the resource/link boundary crisp and reuses the resource machinery (schema validation, substitution, drift detection, state) instead of reinventing it inside links. Precedent: AWS exposes an EventBridge target only as a sub-element of `PutTargets` (no standalone ARN), yet it carries rich per-target configuration, this allows it to be modelled as an `aws/events/target` resource (CRUD via `PutTargets`/`RemoveTargets`), mirroring how `aws/lambda/eventSourceMapping` is a resource that can also serve as a managed intermediary. The links from a target to its destination (Lambda, SQS, API destination) then carry no user input at all, only the derived permission/policy side-effects.

## Two link shapes and reference-implied activation

Links fall into two shapes depending on whether the relationship is also expressed as a field reference:

- **No-field-reference links** (e.g. `aws/lambda/function → aws/dynamodb/table` access, `aws/dynamodb/table → aws/lambda/function` stream trigger): neither resource holds a field referencing the other; the connection lives entirely in an intermediary (event source mapping) or IAM policy. The `linkSelector` is the only thing the author writes, and the link manufactures everything.
- **Field-reference links** (e.g. EventBridge `aws/events/rule → aws/lambda/function | aws/sqs/queue | aws/events/apiDestination`): the relationship is a required join field on the source, the rule's `targets[].arn` is wired with a `${...}` reference. That field is also the link's correlation key, so it cannot be removed.

For a field-reference link, the `${...}` reference and a `linkSelector` would otherwise state the same relationship twice. To avoid that, mark the source field as a **wiring slot** by setting `ActivatesLinkOnReference: true` on its schema (for Cloud Control–generated resources, do this in a `services/cloudcontrol/overlays/{type}.go` overlay, see `events_rule.go`). When a reference to another resource is placed at a wiring slot, the framework activates the registered link for that type pair **without** a `linkSelector`.

Use a wiring slot only for fields that *mean* "wire a resource here". A plain reference to another resource's output value (e.g. an endpoint string) at a non-slot field must never activate a link, the activation signal comes from the source slot and not from the referenced field.
