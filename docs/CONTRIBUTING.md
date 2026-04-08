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

Run unit tests for a specific service:

```bash
go test -tags=unit -count=1 ./services/lambda/...
```

Run unit tests for inter-service links:

```bash
go test -tags=unit -count=1 ./inter-service-links/lambda_dynamodb/...
```

Run all unit tests:

```bash
go test -tags=unit -count=1 ./...
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

Run integration tests:

```bash
# Source the environment file first
set -o allexport && source .env.test && set +o allexport

go test -tags=integration -count=1 -timeout 90s ./...
```

Integration test files follow the naming convention `*_int_test.go`.

### Running All Tests

To run both unit and integration tests together:

```bash
go test -tags=unit,integration -count=1 -timeout 90s ./...
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

Each resource implementation includes tests for all lifecycle operations:

```
services/{service}/
  {resource}_resource_create_test.go       # Unit tests for create
  {resource}_resource_update_test.go       # Unit tests for update
  {resource}_resource_destroy_test.go      # Unit tests for destroy
  {resource}_resource_get_external_state_test.go  # Unit tests for state fetch
  {resource}_resource_stabilised_test.go   # Unit tests for stabilisation
  {resource}_resource_create_int_test.go   # Integration tests (where applicable)
  test_helpers_test.go                     # Shared test fixtures and helpers
```

Link tests follow a similar pattern:

```
services/{service}/links/
  {resourceA}__{resourceB}_link_update_test.go         # Update operation tests
  {resourceA}__{resourceB}_link_stage_changes_test.go  # Stage changes tests
```

### Test Mocks

Service mocks are in `internal/testutils/` and provide mock implementations of AWS service
interfaces:

- `dynamodb_mock/` - DynamoDB service mock
- `ec2_mock/` - EC2 service mock
- `iam_mock/` - IAM service mock
- `lambda_mock/` - Lambda service mock
- `sqs_mock/` - SQS service mock
- `aws_config_loader_mock.go` - AWS config loader mock

Each service directory also has a `test_helpers_test.go` file that sets up common test fixtures
used across that service's tests.

### Linting

Run the linter:

```bash
golangci-lint run ./...
```

## Further documentation

- [Commit Guidelines](./COMMIT_GUIDELINES.md)
- [Source Control and Release Strategy](./SOURCE_CONTROL_RELEASE_STRATEGY.md)
