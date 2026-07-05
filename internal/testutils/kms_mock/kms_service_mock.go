package kmsmock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmsservice "github.com/newstack-cloud/bluelink-provider-aws/services/kms/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type kmsServiceMock struct {
	plugintestutils.MockCalls

	createGrantOutput *kms.CreateGrantOutput
	createGrantError  error

	revokeGrantOutput *kms.RevokeGrantOutput
	revokeGrantError  error

	listGrantsOutput *kms.ListGrantsOutput
	listGrantsError  error
}

// CreateKMSServiceMock creates a new instance of the KMS service mock with the provided options.
func CreateKMSServiceMock(options ...KMSServiceMockOption) *kmsServiceMock {
	mock := &kmsServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// CreateKMSServiceMockFactory returns a service factory bound to a single mock instance.
func CreateKMSServiceMockFactory(
	options ...KMSServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) kmsservice.Service {
	mock := CreateKMSServiceMock(options...)
	return func(awsConfig *aws.Config, providerContext provider.Context) kmsservice.Service {
		return mock
	}
}

// KMSServiceMockOption is a function type for configuring the KMS service mock.
type KMSServiceMockOption func(*kmsServiceMock)

func WithCreateGrantOutput(output *kms.CreateGrantOutput) KMSServiceMockOption {
	return func(mock *kmsServiceMock) { mock.createGrantOutput = output }
}

func WithCreateGrantError(err error) KMSServiceMockOption {
	return func(mock *kmsServiceMock) { mock.createGrantError = err }
}

func WithRevokeGrantOutput(output *kms.RevokeGrantOutput) KMSServiceMockOption {
	return func(mock *kmsServiceMock) { mock.revokeGrantOutput = output }
}

func WithRevokeGrantError(err error) KMSServiceMockOption {
	return func(mock *kmsServiceMock) { mock.revokeGrantError = err }
}

func WithListGrantsOutput(output *kms.ListGrantsOutput) KMSServiceMockOption {
	return func(mock *kmsServiceMock) { mock.listGrantsOutput = output }
}

func WithListGrantsError(err error) KMSServiceMockOption {
	return func(mock *kmsServiceMock) { mock.listGrantsError = err }
}

func (m *kmsServiceMock) CreateGrant(
	ctx context.Context,
	params *kms.CreateGrantInput,
	optFns ...func(*kms.Options),
) (*kms.CreateGrantOutput, error) {
	m.RegisterCall(ctx, params)
	if m.createGrantError != nil {
		return nil, m.createGrantError
	}
	return m.createGrantOutput, nil
}

func (m *kmsServiceMock) RevokeGrant(
	ctx context.Context,
	params *kms.RevokeGrantInput,
	optFns ...func(*kms.Options),
) (*kms.RevokeGrantOutput, error) {
	m.RegisterCall(ctx, params)
	if m.revokeGrantError != nil {
		return nil, m.revokeGrantError
	}
	return m.revokeGrantOutput, nil
}

func (m *kmsServiceMock) ListGrants(
	ctx context.Context,
	params *kms.ListGrantsInput,
	optFns ...func(*kms.Options),
) (*kms.ListGrantsOutput, error) {
	m.RegisterCall(ctx, params)
	if m.listGrantsError != nil {
		return nil, m.listGrantsError
	}
	return m.listGrantsOutput, nil
}
