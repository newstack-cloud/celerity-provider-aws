# Cloud Control–Backed Resources & Data Sources

Most AWS resources **and data sources** in this provider are produced by a single
generic **AWS Cloud Control API** engine plus **generated, per-type schemas**, instead
of being hand-written. Resources go through the engine's CRUD + stabilisation path;
data sources go through a shared fetch/filter/flatten path backed by Cloud Control
`GetResource`/`ListResources`.

The flex VPC (`aws/flex/vpc`) is hand-rolled as it is a higher-level multi-resource
abstraction, not a 1:1 Cloud Control resource. EC2 is onboarded **data-source-only**
(VPC/Subnet/SecurityGroup lookups) because the flex VPC owns the networking fabric.

## Layout

```
services/cloudcontrol/                # Component A: the shared engine (hand-written once)
  service/service.go                  #   Service interface over the Cloud Control client
  cc_resource*.go                     #   Resource: Create/Update/Destroy/GetExternalState/Stabilised
  cc_deploy.go                        #   async deploy / await-computed-state orchestration
  cc_data_source*.go                  #   Data source: fetch / filter / flatten
  cc_names.go cc_spec.go cc_tags.go   #   camelCase<->CFN translation, computed-strip, tags
  cc_meta.go cc_converters.go cc_retry.go
  overlays/                           # hand-written, per-type adjustments (survive regeneration)
    overlays.go                       #   schema overlay + meta-adjustment registries
    behaviour.go                      #   behaviour overlay registry (validate / name-gen / transforms)
    stabilisation.go                  #   stabilise-required registry (slow resources)
    <type>.go                         #   per-type registrations (sqs_queue, events_rule, rds_db_instance, …)
  gen/                                # GENERATED: one file per type + registries + examples (DO NOT EDIT)
    <service>_<type>.go               #   resource schema builder + CCResource registration
    <service>_<type>_data_source.go   #   data source schema builder + registration
    registry.go                       #   GeneratedResources(...)  -> merged in provider.go
    data_sources.go                   #   GeneratedDataSources(...) -> merged in provider.go
    examples_embed.go                 #   //go:embed examples
    examples/resources/*.md           #   generated resource examples (blueprintlang/yaml/jsonc)
    examples/datasources/*.md         #   generated data source examples

tools/awsgen/                         # Component B: the code generator
  service_config.go                   #   services onboarded + per-type overrides + data source config
  schemas/*.json                      #   vendored CloudFormation registry schemas
  examplevalues.go                    #   Layer 1 example value seeds
  curated_examples/                   #   Layer 2 hand-authored example overrides
  descriptions.go                     #   description sanitisation
```

## Commands

Sync the vendored CloudFormation schemas for the onboarded services (optionally a subset
of types). Requires AWS credentials, since it calls the CloudFormation schema registry API:

```bash
go run ./tools/awsgen -sync            # add -region <region> to use a different registry
```

Generate the Go resources, data sources and examples from the vendored schemas:

```bash
go run ./tools/awsgen
```

## Adding a new resource type

1. Add or extend an entry in `services` in [tools/awsgen/service_config.go](../tools/awsgen/service_config.go).
   A `serviceEntry` has:
   - `Name`: the CloudFormation service segment (e.g. `"SQS"`, `"DynamoDB"`).
   - `Include`: optional allowlist of fully-qualified CFN types (for large services where
     only a few types are wanted); empty means the whole service.
   - `Exclude`: types to skip (e.g. schemas malformed upstream).
   - `TypeOverrides`: maps a CFN type to an explicit Bluelink type where the derived name is
     wrong (e.g. acronym casing: `AWS::IAM::OIDCProvider` becomes `aws/iam/oidcProvider`).
   - `DataSourceOnly`: emit only the opted-in data sources, no managed resources.
2. `go run ./tools/awsgen -sync` to vendor the schema (requires AWS credentials).
3. `go run ./tools/awsgen` to generate.
4. `go build ./... && go test -tags unit ./services/cloudcontrol/... ./provider/...`.
5. Review the generation warnings (free-form objects encoded as JSON strings, composite
   identifiers, etc.) and add an overlay if the generated schema needs refinement.

The Bluelink type is derived from the CFN type by `deriveBlueprintType`
(`AWS::SQS::Queue` → `aws/sqs/queue`); use `TypeOverrides` when the derived name is
awkward. There is **no `Cc` suffix**; the generated resources are the canonical
implementations (the only hand-written resource is `aws/flex/vpc`).

## Data sources

A type opted into `dataSourceConfigs` in [tools/awsgen/service_config.go](../tools/awsgen/service_config.go)
also gets a generated data source. Each config declares:
- `FilterFields`: the practitioner-facing filterable fields (matching the hand-written data
  sources they replace). `"region"` selects the AWS region.
- `DeriveIdentifierFromARN`: enables a `GetResource` fast path where a single `arn` equality
  filter resolves to the primary identifier by ARN suffix. Set it to `false` where the
  identifier is the ARN itself (events rule) or non-derivable (sqs queue URL), in which case
  the engine falls back to `ListResources` plus filtering.

## Overlays

Some CloudFormation constructs do not translate cleanly (free-form `object`-or-`string`
properties, gnarly `$ref` graphs, constraints documented only in prose), and some behaviour
cannot be derived from a schema at all. Generation emits a base; an overlay refines it and is
re-applied on every regeneration. Overlays live in
[services/cloudcontrol/overlays/](../services/cloudcontrol/overlays/), registered from per-type
`init()` functions. There are four kinds:

- **Schema overlays:** `Register(type, fn)` refines the generated schema in place. The
  generated constructor wraps its schema in `overlays.Apply(type, schema)`. See
  `overlays/sqs_queue.go` (re-types the free-form `RedrivePolicy`, restores numeric bounds).
- **Meta adjustments:** `RegisterMetaAdjustment(type, ...)` tells the engine about fields a
  schema overlay re-typed. `RemoveJSONStringFields` stops treating a field as a JSON string at
  the Cloud Control boundary, and `AddFieldNameOverrides` maps camelCase paths to CloudFormation
  property names a first-character flip cannot recover (e.g. `principal.aws` becomes `AWS`).
  These are expressed in primitives to avoid importing the `cloudcontrol` package (cycle).
- **Behaviour overlays:** `RegisterBehaviour(type, ...)` carries code-level behaviour the
  engine cannot infer: `CustomValidate` (plan-time, literal values only), `ValidateResolvedSpec`
  (deploy-time, runs before both create and update on the fully resolved spec, the guard for
  constraints on reference-wired values, e.g. minimum subnet counts; see `overlays/subnet_groups.go`),
  unique-name generation (`Name`), and spec transforms (`BeforeCreate`, `AfterReadExternalState`).
- **Stabilise-required:** `RegisterStabiliseRequired(type)` marks a slow resource (see below).

### Example overlays: two layers, applied during generation

- *Layer 1 (value seeds):* per-field representative values in
  [tools/awsgen/examplevalues.go](../tools/awsgen/examplevalues.go), used when the generator's
  generic placeholders are not semantically valid (an ARN, a region, an enum).
- *Layer 2 (curated markdown):* drop hand-authored `<stem>_<variant>.md` files in
  [tools/awsgen/curated_examples/](../tools/awsgen/curated_examples/); the generator embeds
  them verbatim instead of the generated baseline (e.g. `iam_role_*`, `events_connection_*`).

Every generated example's blueprintlang is round-trip-parsed during generation, so generated
docs can never be syntactically invalid.

## Slow / stabilise-required resources

Some resources take minutes to provision and expose computed fields that only become available
once fully provisioned (e.g. an RDS DB instance endpoint, a flex VPC's NAT gateways).
`overlays.RegisterStabiliseRequired(type)` marks these, and both ends of a deployment key off
the set:

- **Producers:** a stabilise-required resource's `Create` returns as soon as its identifier is
  known (it does not block for the whole provision); its computed fields are finalised at
  stabilisation.
- **Consumers:** every generated resource declares this set as its stabilised dependencies
  (`StabilisedDependenciesFunc`), so anything that depends on a slow type waits for it to
  stabilise before deploying, ensuring it resolves the late computed fields.

Registered today: `aws/rds/dbInstance` (`overlays/rds_db_instance.go`) and `aws/flex/vpc`
(`flex/stabilisation.go`).

## Description sanitisation

CloudFormation registry descriptions are written for CloudFormation and leak detail we do not
want in Bluelink docs. During generation, every resource and field description is passed through
`sanitiseDescription` in [tools/awsgen/descriptions.go](../tools/awsgen/descriptions.go), which:

- drops pure boilerplate (`"Resource Type definition for AWS::…"`) and substitutes a synthesised
  one-liner (`"Manages an Events archive."`);
- rewrites inline `AWS::Svc::Type` identifiers to readable prose (`AWS::DynamoDB::Table` →
  `DynamoDB table`) and CloudFormation pseudo-parameters (`AWS::Region` → `the region`);
- strips CloudFormation doc links (keeping other AWS doc links) and expands AWS doc macros
  (`CFNlong`/`CFN` → `Bluelink`, `DDB` → `DynamoDB`, …);
- rebrands CloudFormation framing onto Bluelink equivalents (`template` → `blueprint`,
  `stack` → `deployment`) and removes RST double-backtick markup.

The functional `CFNType` field (which the engine uses to talk to Cloud Control) is unaffected;
only documentation strings are sanitised.

## Additional notes

- Resources Cloud Control cannot serve well can always be hand-written with the AWS SDK using
  the same `provider.ResourceDefinition` shape (as `aws/flex/vpc` is).
