package resourceservicemock

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
)

// A configurable mock of provider.ResourceService for use
// in link intermediary-resource tests. LookupResourceInState returns the configured
// resource state; Deploy/Destroy record their calls and return configured results.
type mockResourceService struct {
	lookupResult *state.ResourceState
	lookupError  error
	deployOutput *provider.ResourceDeployOutput
	deployError  error
	destroyError error

	// DeployCalls records every Deploy invocation, in order.
	DeployCalls []DeployCall
	// DestroyCalls records every Destroy invocation, in order.
	DestroyCalls []DestroyCall

	// lastLookupExternalID is the external ID of the most recent lookup, so a test can
	// assert what a caller identified a resource by.
	lastLookupExternalID string
}

// LastLookupExternalID returns the external ID passed to the most recent
// LookupResourceInState call.
func (m *mockResourceService) LastLookupExternalID() string {
	return m.lastLookupExternalID
}

// DeployCall captures the arguments of a single Deploy invocation.
type DeployCall struct {
	ResourceType string
	Input        *provider.ResourceDeployServiceInput
}

// DestroyCall captures the arguments of a single Destroy invocation.
type DestroyCall struct {
	ResourceType string
	Input        *provider.ResourceDestroyInput
}

// Option configures the mock resource service.
type Option func(*mockResourceService)

// WithLookupResourceInState sets the resource state returned by LookupResourceInState.
func WithLookupResourceInState(resourceState *state.ResourceState) Option {
	return func(m *mockResourceService) { m.lookupResult = resourceState }
}

// WithLookupError sets the error returned by LookupResourceInState.
func WithLookupError(err error) Option {
	return func(m *mockResourceService) { m.lookupError = err }
}

// WithDeployOutput sets the output returned by Deploy.
func WithDeployOutput(output *provider.ResourceDeployOutput) Option {
	return func(m *mockResourceService) { m.deployOutput = output }
}

// WithDeployError sets the error returned by Deploy.
func WithDeployError(err error) Option {
	return func(m *mockResourceService) { m.deployError = err }
}

// WithDestroyError sets the error returned by Destroy.
func WithDestroyError(err error) Option {
	return func(m *mockResourceService) { m.destroyError = err }
}

// Mock is the concrete mock type, exposed so tests can read the recorded calls.
type Mock = mockResourceService

// Create returns a new mock ResourceService configured with the provided options.
// The returned value satisfies provider.ResourceService, cast to *Mock to inspect
// the recorded DeployCalls / DestroyCalls.
func Create(options ...Option) *mockResourceService {
	m := &mockResourceService{}
	for _, option := range options {
		option(m)
	}
	return m
}

func (m *mockResourceService) LookupResourceInState(
	ctx context.Context,
	input *provider.ResourceLookupInput,
) (*state.ResourceState, error) {
	m.lastLookupExternalID = input.ExternalID
	return m.lookupResult, m.lookupError
}

func (m *mockResourceService) HasResourceInState(
	ctx context.Context,
	input *provider.ResourceLookupInput,
) (bool, error) {
	return m.lookupResult != nil, m.lookupError
}

func (m *mockResourceService) Deploy(
	ctx context.Context,
	resourceType string,
	input *provider.ResourceDeployServiceInput,
) (*provider.ResourceDeployOutput, error) {
	m.DeployCalls = append(m.DeployCalls, DeployCall{ResourceType: resourceType, Input: input})
	if m.deployError != nil {
		return nil, m.deployError
	}
	if m.deployOutput != nil {
		return m.deployOutput, nil
	}
	return &provider.ResourceDeployOutput{}, nil
}

func (m *mockResourceService) Destroy(
	ctx context.Context,
	resourceType string,
	input *provider.ResourceDestroyInput,
) error {
	m.DestroyCalls = append(m.DestroyCalls, DestroyCall{ResourceType: resourceType, Input: input})
	return m.destroyError
}

func (m *mockResourceService) AcquireResourceLock(
	ctx context.Context,
	input *provider.AcquireResourceLockInput,
) error {
	return nil
}
