package secretsmanagermock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagerservice "github.com/newstack-cloud/bluelink-provider-aws/services/secretsmanager/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type secretsManagerServiceMock struct {
	plugintestutils.MockCalls

	getSecretValueOutput *secretsmanager.GetSecretValueOutput
	getSecretValueError  error
}

// CreateSecretsManagerServiceMock creates a new instance of the Secrets Manager service
// mock with the provided options.
func CreateSecretsManagerServiceMock(options ...SecretsManagerServiceMockOption) *secretsManagerServiceMock {
	mock := &secretsManagerServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// CreateSecretsManagerServiceMockFactory returns a service factory bound to a single mock instance.
func CreateSecretsManagerServiceMockFactory(
	options ...SecretsManagerServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) secretsmanagerservice.Service {
	mock := CreateSecretsManagerServiceMock(options...)
	return func(awsConfig *aws.Config, providerContext provider.Context) secretsmanagerservice.Service {
		return mock
	}
}

// SecretsManagerServiceMockOption is a function type for configuring the Secrets Manager service mock.
type SecretsManagerServiceMockOption func(*secretsManagerServiceMock)

func WithGetSecretValueOutput(output *secretsmanager.GetSecretValueOutput) SecretsManagerServiceMockOption {
	return func(mock *secretsManagerServiceMock) { mock.getSecretValueOutput = output }
}

func WithGetSecretValueError(err error) SecretsManagerServiceMockOption {
	return func(mock *secretsManagerServiceMock) { mock.getSecretValueError = err }
}

func (m *secretsManagerServiceMock) GetSecretValue(
	ctx context.Context,
	params *secretsmanager.GetSecretValueInput,
	optFns ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	m.RegisterCall(ctx, params)
	if m.getSecretValueError != nil {
		return nil, m.getSecretValueError
	}
	return m.getSecretValueOutput, nil
}
