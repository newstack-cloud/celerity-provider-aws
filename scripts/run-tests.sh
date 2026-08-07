#!/usr/bin/env bash


# Default to running all tests
TEST_MODE="all"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --unit)
    TEST_MODE="unit"
    shift # past argument
    ;;
    --integration)
    TEST_MODE="integration"
    shift # past argument
    ;;
    --all)
    TEST_MODE="all"
    shift # past argument
    ;;
    -r|--run)
    TEST_RUN="$2"
    shift # past argument
    shift # past value
    ;;
    --run=*)
    TEST_RUN="${key#*=}"
    shift # past argument
    ;;
    -h|--help)
    HELP=yes
    shift # past argument
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

function help {
  cat << EOF
Test runner
Runs tests for the AWS provider:
  bash scripts/run-tests.sh [--unit|--integration|--all] [--run <pattern>]

Options:
  --unit                 Run unit tests only.
  --integration          Run integration (e2e) tests only.
  --all                  Run unit and integration tests (default).
  -r, --run <pattern>    Pass a -run regex to go test to target specific tests,
                         e.g. --run 'TestResources/lambda_function'.
  -h, --help             Show this help.

Environment:
  E2E_CONCURRENCY        Max concurrent e2e AWS operations (default 6).
  E2E_SLOW               Set to 1 to include the slow e2e fixtures that are
                         skipped by default. Currently the direct-author parity
                         fixture, which stands up an Aurora cluster and adds
                         several minutes plus RDS cost to the run:
                           E2E_SLOW=1 bash scripts/run-tests.sh --integration
EOF
}

if [ -n "$HELP" ]; then
  help
  exit 0
fi

set -e
echo "" > coverage.txt

if [ "$TEST_MODE" == "unit" ]; then
  TEST_TAGS="unit"
elif [ "$TEST_MODE" == "integration" ]; then
  TEST_TAGS="integration"
elif [ "$TEST_MODE" == "all" ]; then
  TEST_TAGS="unit,integration"
fi

# Unit tests are fast (mocked AWS); integration tests deploy, stabilise and destroy
# real AWS resources end-to-end, which is far slower, so use a much longer timeout.
TEST_TIMEOUT="90000ms"
# E2E subtests run in parallel (t.Parallel). -parallel caps how many run at once; the
# harness also enforces an AWS-operation gate (E2E_CONCURRENCY, default 6) so concurrency
# stays within Cloud Control limits regardless of this flag.
TEST_PARALLEL=""
if [[ "$TEST_MODE" == "integration" || "$TEST_MODE" == "all" ]]; then
  TEST_TIMEOUT="60m"
  TEST_PARALLEL="-parallel ${E2E_CONCURRENCY:-6}"
fi

# Optionally target specific tests via a -run regex (e.g. a single e2e subtest).
TEST_RUN_FLAG=""
if [ -n "$TEST_RUN" ]; then
  TEST_RUN_FLAG="-run $TEST_RUN"
fi

if [[ "$TEST_MODE" == "integration" || "$TEST_MODE" == "all" ]] && [ -f ".env.test" ]; then
  echo "Exporting environment variables from .env.test for integration tests ..."
  set -o allexport
  source .env.test
  set +o allexport
fi

# List packages with the active tags so integration-only packages (e.g. tests/e2e,
# whose files are all build-tagged "integration") are included in the run.
go test -tags="$TEST_TAGS" -count=1 -timeout "$TEST_TIMEOUT" $TEST_PARALLEL $TEST_RUN_FLAG -race -coverprofile=coverage.txt -coverpkg=./... -covermode=atomic `go list -tags="$TEST_TAGS" ./... | egrep -v '/testutils(/|$)'`

if [ -z "$GITHUB_ACTION" ]; then
  # We are on a dev machine so produce html output of coverage
  # to get a visual to better reveal uncovered lines.
  go tool cover -html=coverage.txt -o coverage.html
fi

if [ -n "$GITHUB_ACTION" ]; then
  # We are in a CI environment so run tests again to generate JSON report.
  go test -count=1 -timeout "$TEST_TIMEOUT" $TEST_PARALLEL $TEST_RUN_FLAG -json -tags "$TEST_TAGS" `go list -tags="$TEST_TAGS" ./... | egrep -v '/testutils(/|$)'` > report.json
fi
