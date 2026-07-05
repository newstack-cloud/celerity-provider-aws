//go:build integration

// Package e2e contains end-to-end integration tests that deploy real blueprints
// (authored as blueprintlang files under testdata/) against a real AWS account by
// driving the blueprint framework's container in-process with the AWS provider
// registered directly. Each test loads a .blueprint file, stages changes, deploys,
// asserts on the resulting state/exports (and real AWS side effects), then destroys
// the instance. Build-tagged `integration`; skips when AWS_REGION is not set.
package e2e

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/integration"
	awsprovider "github.com/newstack-cloud/bluelink-provider-aws/provider"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	kmsservice "github.com/newstack-cloud/bluelink-provider-aws/services/kms/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	s3service "github.com/newstack-cloud/bluelink-provider-aws/services/s3/service"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint-state/memfile"
	"github.com/newstack-cloud/bluelink/libs/blueprint/changes"
	"github.com/newstack-cloud/bluelink/libs/blueprint/container"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/blueprint/transform"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/providerserverv1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var idCounter atomic.Uint64

// Controls how many blueprint instances are deployed or destroyed against AWS
// concurrently. Subtests run in parallel (t.Parallel), so without a cap they would fire
// enough simultaneous Cloud Control operations to hit account/region concurrency limits.
// The gate enforces the bound regardless of how `go test` is invoked (i.e. independent of
// -parallel); override the size with E2E_CONCURRENCY (default 6).
var e2eOpGate = make(chan struct{}, e2eConcurrency())

func e2eConcurrency() int {
	if v := os.Getenv("E2E_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 6
}

// Blocks until a concurrency slot is free and returns a release func
// (call via defer) to free it.
func acquireE2ESlot() func() {
	e2eOpGate <- struct{}{}
	return func() { <-e2eOpGate }
}

// Harness drives in-process blueprint deployments against
// a real AWS account.
type Harness struct {
	T          *testing.T
	Ctx        context.Context
	Loader     container.Loader
	State      state.Container
	AWSConfig  aws.Config
	Account    *integration.AWSAccountInfo
	Region     string
	NamePrefix string
	sessionID  string
}

// Setup builds the AWS provider, an in-memory state container, and a blueprint
// loader wired together in-process. It skips the test when AWS_REGION is not set.
func Setup(t *testing.T) *Harness {
	t.Helper()

	region := os.Getenv("AWS_REGION")
	if region == "" {
		t.Skip("AWS_REGION not set; skipping integration test")
	}

	ctx := context.Background()
	loader := &utils.DefaultAWSConfigLoader{}

	account, err := integration.GetAWSAccountInfo(ctx, loader)
	require.NoError(t, err, "resolve AWS account info")

	awsConfig, err := loader.LoadDefaultConfig(ctx)
	require.NoError(t, err, "load AWS config")

	configStore := utils.NewAWSConfigStore(
		os.Environ(),
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	prov := awsprovider.NewProvider(
		iamservice.NewService,
		lambdaservice.NewService,
		ec2service.NewService,
		resgrouptagservice.NewService,
		sqsservice.NewService,
		dynamodbservice.NewService,
		eventsservice.NewService,
		cloudcontrolservice.NewService,
		s3service.NewService,
		ssmservice.NewService,
		kmsservice.NewService,
		configStore,
	)

	// Mirror the production host (plugin launch): derive each resource's
	// ResourceCanLinkTo from the provider's registered link types. Without this, the
	// resources report an empty "can link to" list and link selectors never match,
	// which will mean linked resources are dropped from deployment.
	allLinkTypes, err := prov.ListLinkTypes(ctx)
	require.NoError(t, err, "list provider link types")
	wrappedProv := providerserverv1.WrapProviderWithDerivedCanLinkTo(prov, allLinkTypes)

	stateContainer, err := memfile.LoadStateContainer(t.TempDir(), afero.NewMemMapFs(), core.NewNopLogger())
	require.NoError(t, err, "create state container")

	bpLoader := container.NewDefaultLoader(
		map[string]provider.Provider{"aws": wrappedProv},
		map[string]transform.SpecTransformer{},
		stateContainer,
		&fsChildResolver{fs: afero.NewOsFs()},
		container.WithLoaderLogger(core.NewNopLogger()),
	)

	suffix := uniqueSuffix()
	return &Harness{
		T:          t,
		Ctx:        ctx,
		Loader:     bpLoader,
		State:      stateContainer,
		AWSConfig:  awsConfig,
		Account:    account,
		Region:     region,
		NamePrefix: "bluelink-it-" + suffix,
		sessionID:  "it-session-" + suffix,
	}
}

// Name returns a run-unique resource name from the harness prefix and a label,
// e.g. "bluelink-it-<suffix>-<label>".
func (h *Harness) Name(label string) string {
	return fmt.Sprintf("%s-%s", h.NamePrefix, label)
}

// DeployedInstance is the result of deploying a blueprint: the instance ID and its
// final state (resources + exports).
type DeployedInstance struct {
	InstanceID string
	State      state.InstanceState
}

// Deploy loads the named blueprintlang file from testdata/, stages and deploys it
// as a new instance (injecting the run's unique namePrefix plus any extra blueprint
// variables), registers a t.Cleanup that destroys it, and returns the deployed
// instance state. The blueprint should reference `${variables.namePrefix}` in its
// resource names to stay unique per run.
func (h *Harness) Deploy(
	t *testing.T,
	blueprintFile string,
	vars map[string]*core.ScalarValue,
) *DeployedInstance {
	t.Helper()

	params := h.params(vars)
	bp, err := h.Loader.Load(h.Ctx, filepath.Join("testdata", blueprintFile), params)
	require.NoErrorf(t, err, "load blueprint %s", blueprintFile)

	instanceName := "it-instance-" + uniqueSuffix()

	changeSet := h.stage(t, bp, instanceName, params, false)

	// Register teardown before deploying so partial deployments are still cleaned up.
	t.Cleanup(func() {
		h.destroy(instanceName, bp)
	})

	// Bound concurrent deploys so parallel subtests stay within AWS operation limits.
	defer acquireE2ESlot()()

	deployChannels := container.CreateDeployChannels()
	err = bp.Deploy(h.Ctx, &container.DeployInput{
		InstanceName: instanceName,
		Changes:      changeSet,
	}, deployChannels, params)
	require.NoErrorf(t, err, "start deploy of %s", blueprintFile)

	finished := consumeDeploy(t, deployChannels)
	require.Equalf(t, core.InstanceStatusDeployed, finished.Status,
		"deploy of %s did not succeed: %v", blueprintFile, finished.FailureReasons)

	instanceState, err := h.State.Instances().Get(h.Ctx, finished.InstanceID)
	require.NoError(t, err, "read deployed instance state")
	return &DeployedInstance{InstanceID: finished.InstanceID, State: instanceState}
}

// ResourceSpec returns the deployed spec/external state of a resource by its
// logical (blueprint) name.
func (d *DeployedInstance) ResourceSpec(t *testing.T, name string) *core.MappingNode {
	t.Helper()
	resourceID, ok := d.State.ResourceIDs[name]
	require.Truef(t, ok, "resource %q not found in deployed instance", name)
	resourceState, ok := d.State.Resources[resourceID]
	require.Truef(t, ok, "resource state for %q not found", name)
	return resourceState.SpecData
}

// Export returns a blueprint export value by name.
func (d *DeployedInstance) Export(t *testing.T, name string) *core.MappingNode {
	t.Helper()
	exportState, ok := d.State.Exports[name]
	require.Truef(t, ok, "export %q not found in deployed instance", name)
	return exportState.Value
}

func (h *Harness) stage(
	t *testing.T,
	bp container.BlueprintContainer,
	instanceName string,
	params core.BlueprintParams,
	destroy bool,
) *changes.BlueprintChanges {
	t.Helper()
	channels := newChangeStagingChannels()
	err := bp.StageChanges(h.Ctx, &container.StageChangesInput{
		InstanceName: instanceName,
		Destroy:      destroy,
	}, channels, params)
	require.NoError(t, err, "start change staging")
	return consumeStage(t, channels)
}

func (h *Harness) destroy(instanceName string, bp container.BlueprintContainer) {
	// Bound concurrent destroys alongside deploys (same AWS operation limits).
	defer acquireE2ESlot()()

	params := h.params(nil)
	channels := newChangeStagingChannels()
	if err := bp.StageChanges(h.Ctx, &container.StageChangesInput{
		InstanceName: instanceName,
		Destroy:      true,
	}, channels, params); err != nil {
		h.T.Errorf("cleanup: stage destroy for %s failed: %v", instanceName, err)
		return
	}
	changeSet, err := consumeStageErr(channels)
	if err != nil {
		h.T.Errorf("cleanup: stage destroy for %s failed: %v", instanceName, err)
		return
	}

	deployChannels := container.CreateDeployChannels()
	bp.Destroy(h.Ctx, &container.DestroyInput{
		InstanceName: instanceName,
		Changes:      changeSet,
	}, deployChannels, params)
	finished, err := consumeDeployErr(deployChannels)
	if err != nil {
		h.T.Errorf("cleanup: destroy of %s failed: %v", instanceName, err)
		return
	}
	if finished.Status != core.InstanceStatusDestroyed {
		h.T.Errorf("cleanup: destroy of %s ended in status %v: %v", instanceName, finished.Status, finished.FailureReasons)
	}
}

func (h *Harness) params(vars map[string]*core.ScalarValue) core.BlueprintParams {
	blueprintVars := map[string]*core.ScalarValue{
		"namePrefix": core.ScalarFromString(h.NamePrefix),
	}
	maps.Copy(blueprintVars, vars)
	return core.NewDefaultParams(
		map[string]map[string]*core.ScalarValue{
			"aws": {"region": core.ScalarFromString(h.Region)},
		},
		map[string]map[string]*core.ScalarValue{},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString(h.sessionID)},
		blueprintVars,
	)
}

func uniqueSuffix() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), idCounter.Add(1))
}
