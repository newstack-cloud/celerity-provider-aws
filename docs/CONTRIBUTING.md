# Contributing to Bluelink Provider for AWS

## Setup

Ensure git uses the custom directory for git hooks so the pre-commit and commit-msg linting hooks
kick in.

```bash
git config core.hooksPath .githooks
```

### Prerequisites

- [Go](https://golang.org/dl/) >=1.23
- [GolangCI-Lint](https://golangci-lint.run/welcome/install/#local-installation) - Used for linting and formatting.
- [Node.js](https://nodejs.org/en/download/) - Used for running scripts for commit message linting.
- [Yarn](https://yarnpkg.com/getting-started/install) - Used for managing dependencies for commit message linting.

Dependencies are managed with Go modules (go.mod) and will be installed automatically when you first
run tests.

If you want to install dependencies manually you can run:

```bash
go mod download
```

### Node dependencies

There are node.js dependencies that provide tools that are used in git hooks and scripting for the provider.

Install dependencies from the root directory by simply running:
```bash
yarn
```

## Running Tests

This project uses Go build tags to separate unit and integration tests. All test files include a
build tag at the top of the file:

- `//go:build unit` for unit tests
- `//go:build integration` for integration tests

Test caching is disabled (`-count=1`) for all test modes since this project is heavily
integration-focused and cached test results can mask real issues.

### Unit Tests

Unit tests mock AWS service clients and do not require AWS credentials. They test individual
resource operations (create, update, destroy, get external state, stabilised) and link operations
(update resource A/B, update intermediary resources, stage changes).

Run all unit tests with the [test runner script](#test-runner-script):

```bash
bash scripts/run-tests.sh --unit
```

The script runs the whole module. To run a focused subset while iterating, invoke `go test`
directly with the `unit` tag:

```bash
# A specific service
go test -tags=unit -count=1 ./services/lambda/...

# Inter-service links
go test -tags=unit -count=1 ./inter-service-links/lambda_dynamodb/...
```

### Integration Tests

Integration tests make real API calls to AWS and require valid AWS credentials. These tests create
and destroy actual AWS resources.

**Prerequisites:**

1. Create a `.env.test` file in the project root with your AWS configuration:

   ```bash
   AWS_PROFILE=your-profile
   AWS_REGION=eu-west-2
   ```

2. Ensure your AWS credentials are configured (via `aws configure` or environment variables).

Run integration tests with the [test runner script](#test-runner-script), which sources `.env.test`
automatically:

```bash
bash scripts/run-tests.sh --integration
```

To run a focused subset while iterating, source `.env.test` yourself and invoke `go test` directly
with the `integration` tag:

```bash
set -o allexport && source .env.test && set +o allexport
go test -tags=integration -count=1 -timeout 90s ./services/sqs/...
```

Integration test files follow the naming convention `*_int_test.go`.

#### End-to-end (e2e) integration tests

The primary integration mechanism is the **end-to-end suite** in [`tests/e2e/`](../tests/e2e). It
deploys real **blueprintlang** blueprints (under `tests/e2e/testdata/*.blueprint`) against a real AWS
account by driving the blueprint framework's container **in process** with the AWS provider
registered directly (no gRPC, no CLI, no plugin install). This exercises the real
parse → stage → deploy → destroy path, the same core mechanism that practitioners use where resources, data
sources and links are all covered through one harness, with `testify/suite` tables.

How it works (`tests/e2e/harness.go`):

```go
//go:build integration

package e2e

func (s *ResourceE2ESuite) Test_resources() {
    h := Setup(s.T())                                  // builds provider + in-memory state + loader;
                                                       // skips if AWS_REGION is unset
    inst := h.Deploy(s.T(), "queue_resource.blueprint", nil) // stage -> deploy -> read state;
                                                       // registers a t.Cleanup that destroys it
    s.Require().Equal(expectedARN, core.StringValue(inst.Export(s.T(), "queueArn")))
    s.Require().Equal(expectedARN, core.StringValue(inst.ResourceSpec(s.T(), "queue").Fields["arn"]))
}
```

Conventions:

- **Blueprints are data, not Go.** Each case is a `.blueprint` file in `testdata/`. Blueprints use a
  `namePrefix` blueprint variable (injected per run by the harness) so resource names are unique
  across runs while the files stay static.
- **Guaranteed teardown.** `h.Deploy` registers a `t.Cleanup` that stages a destroy and tears the
  instance down (the real destroy path); a failed destroy fails the test.
- **Data sources** are tested via **change staging** (`h.Stage`), which resolves data sources against
  real AWS without deploying. The resolved values are read from the staged exports
  (`ResolvedExport`). A blueprint must declare at least one resource, so data source blueprints include a
  placeholder resource that is never deployed.
- **Links** are established with `select by label { ... }` on one resource matching `metadata.labels`
  on the other; the framework runs the link's resource/intermediary updates during deploy. Side
  effects (redrive/queue/function policies, role inline policies, env vars, event source mappings)
  are asserted via the raw AWS SDK helpers in `tests/e2e/aws_assertions.go`.

### Running All Tests

To run both unit and integration tests together, use the [test runner script](#test-runner-script):

```bash
bash scripts/run-tests.sh
```

### Test Runner Script

A convenience script is provided that handles environment setup, coverage, and reporting:

```bash
# Run all tests (unit + integration)
bash scripts/run-tests.sh

# Run only unit tests
bash scripts/run-tests.sh --unit

# Run only integration tests
bash scripts/run-tests.sh --integration
```

The script will:
- Source `.env.test` automatically when running integration or all tests
- Generate a `coverage.txt` coverage profile
- Generate `coverage.html` for local visual coverage inspection
- In CI, generate a `report.json` for test reporting

### Test Structure

Most resources and data sources are **generated** and served by a single generic Cloud Control
engine, so their behaviour is tested once at the engine level rather than per resource. See
[CLOUD_CONTROL_RESOURCES.md](./agent-guidance/CLOUD_CONTROL_RESOURCES.md) for how generation works.

**Generic Cloud Control engine** — the shared CRUD and data-source behaviour every generated type
inherits is unit-tested here:

```
services/cloudcontrol/
  cc_resource_create_test.go               # Create
  cc_resource_update_test.go               # Update
  cc_resource_destroy_test.go              # Destroy
  cc_resource_get_external_state_test.go   # State fetch (incl. tag filtering)
  cc_resource_stabilised_test.go           # Stabilisation
  cc_data_source_fetch_test.go             # Data source fetch (fast + filter paths)
  cc_data_source_filter_test.go            # Client-side filter operators
  cc_data_source_flatten_test.go           # Schema-aware export flattening
  cc_test_helpers_test.go                  # Shared engine test fixtures
```

**Generated per-type files** - schema/spec-definition assertions for each generated type, regenerated
by the codegen (do not hand-edit):

```
services/cloudcontrol/gen/
  {service}_{type}.go                      # Generated resource/data source (e.g. sqs_queue.go)
  {service}_{type}_test.go                 # Type/spec-definition tests
  {service}_{type}_validation_test.go      # Schema attribute assertions
```

**Codegen** - `tools/awsgen` is covered by golden tests (`TestGolden`, `TestGoldenDataSources`) that
compare emitted Go against `tools/awsgen/testdata/*.golden`. After an intentional codegen change,
refresh them with `go test ./tools/awsgen -update` and commit the updated goldens.

**Hand-written resources/data sources** - the flex VPC resource (`flex/`) and the Lambda data sources
(`services/lambda/`) are not generated and keep their own unit tests plus a `test_helpers_test.go`
with shared fixtures.

**Links** are hand-written and tested per relationship:

```
services/{service}/links/                  # Intra-service links
inter-service-links/{serviceA}_{serviceB}/ # Inter-service links
  {resourceA}__{resourceB}_link_update_test.go         # Update operation tests
  {resourceA}__{resourceB}_link_stage_changes_test.go  # Stage changes tests
```


#### Integration Tests

Most integration tests are the blueprint-driven end-to-end suite in [`tests/e2e/`](../tests/e2e)
(blueprintlang fixtures under `tests/e2e/testdata/`).
There are some exceptions where integration tests are written per resource or link, e.g. for the flex VPC resource and the inter-service links.

### Linting

Run the linter:

```bash
golangci-lint run ./...
```

## Further documentation

- [Commit Guidelines](./COMMIT_GUIDELINES.md)
- [Source Control and Release Strategy](./SOURCE_CONTROL_RELEASE_STRATEGY.md)
