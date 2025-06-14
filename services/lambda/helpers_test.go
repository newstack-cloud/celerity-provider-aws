package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
)

type lambdaServiceMock struct {
	plugintestutils.MockCalls

	getFunctionOutput                  *lambda.GetFunctionOutput
	getFunctionCodeSigningOutput       *lambda.GetFunctionCodeSigningConfigOutput
	getFunctionRecursionOutput         *lambda.GetFunctionRecursionConfigOutput
	getFunctionConcurrencyOutput       *lambda.GetFunctionConcurrencyOutput
	getFunctionError                   error
	getFunctionCodeSigningError        error
	getFunctionRecursionError          error
	getFunctionConcurrencyError        error
	deleteFunctionOutput               *lambda.DeleteFunctionOutput
	deleteFunctionError                error
	updateFunctionConfigurationOutput  *lambda.UpdateFunctionConfigurationOutput
	updateFunctionConfigurationError   error
	updateFunctionCodeOutput           *lambda.UpdateFunctionCodeOutput
	updateFunctionCodeError            error
	putFunctionCodeSigningConfigOutput *lambda.PutFunctionCodeSigningConfigOutput
	putFunctionCodeSigningConfigError  error
	putFunctionConcurrencyOutput       *lambda.PutFunctionConcurrencyOutput
	putFunctionConcurrencyError        error
	putFunctionRecursionConfigOutput   *lambda.PutFunctionRecursionConfigOutput
	putFunctionRecursionConfigError    error
	putRuntimeManagementConfigOutput   *lambda.PutRuntimeManagementConfigOutput
	putRuntimeManagementConfigError    error
	tagResourceOutput                  *lambda.TagResourceOutput
	tagResourceError                   error
	untagResourceOutput                *lambda.UntagResourceOutput
	untagResourceError                 error
	createFunctionOutput               *lambda.CreateFunctionOutput
	createFunctionError                error
}

type lambdaServiceMockOption func(*lambdaServiceMock)

func createLambdaServiceMockFactory(
	opts ...lambdaServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) Service {
	mock := createLambdaServiceMock(opts...)
	return func(awsConfig *aws.Config, providerContext provider.Context) Service {
		return mock
	}
}

func createLambdaServiceMock(
	opts ...lambdaServiceMockOption,
) *lambdaServiceMock {
	mock := &lambdaServiceMock{}

	for _, opt := range opts {
		opt(mock)
	}

	return mock
}

// Mock configuration options.

func WithGetFunctionOutput(output *lambda.GetFunctionOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getFunctionOutput = output
	}
}

func WithGetFunctionError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getFunctionError = err
	}
}

func WithGetFunctionCodeSigningOutput(
	output *lambda.GetFunctionCodeSigningConfigOutput,
) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getFunctionCodeSigningOutput = output
	}
}

func WithGetFunctionCodeSigningError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getFunctionCodeSigningError = err
	}
}

func WithGetFunctionRecursionOutput(
	output *lambda.GetFunctionRecursionConfigOutput,
) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getFunctionRecursionOutput = output
	}
}

func WithGetFunctionRecursionError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getFunctionRecursionError = err
	}
}

func WithGetFunctionConcurrencyOutput(
	output *lambda.GetFunctionConcurrencyOutput,
) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getFunctionConcurrencyOutput = output
	}
}

func WithGetFunctionConcurrencyError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getFunctionConcurrencyError = err
	}
}

func WithDeleteFunctionOutput(output *lambda.DeleteFunctionOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteFunctionOutput = output
	}
}

func WithDeleteFunctionError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteFunctionError = err
	}
}

func WithUpdateFunctionConfigurationOutput(
	output *lambda.UpdateFunctionConfigurationOutput,
) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.updateFunctionConfigurationOutput = output
	}
}

func WithUpdateFunctionConfigurationError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.updateFunctionConfigurationError = err
	}
}

func WithUpdateFunctionCodeOutput(output *lambda.UpdateFunctionCodeOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.updateFunctionCodeOutput = output
	}
}

func WithUpdateFunctionCodeError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.updateFunctionCodeError = err
	}
}

func WithPutFunctionCodeSigningConfigOutput(
	output *lambda.PutFunctionCodeSigningConfigOutput,
) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putFunctionCodeSigningConfigOutput = output
	}
}

func WithPutFunctionCodeSigningConfigError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putFunctionCodeSigningConfigError = err
	}
}

func WithPutFunctionConcurrencyOutput(
	output *lambda.PutFunctionConcurrencyOutput,
) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putFunctionConcurrencyOutput = output
	}
}

func WithPutFunctionConcurrencyError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putFunctionConcurrencyError = err
	}
}

func WithPutFunctionRecursionConfigOutput(
	output *lambda.PutFunctionRecursionConfigOutput,
) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putFunctionRecursionConfigOutput = output
	}
}

func WithPutFunctionRecursionConfigError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putFunctionRecursionConfigError = err
	}
}

func WithPutRuntimeManagementConfigOutput(
	output *lambda.PutRuntimeManagementConfigOutput,
) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putRuntimeManagementConfigOutput = output
	}
}

func WithPutRuntimeManagementConfigError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putRuntimeManagementConfigError = err
	}
}

func WithTagResourceOutput(output *lambda.TagResourceOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.tagResourceOutput = output
	}
}

func WithTagResourceError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.tagResourceError = err
	}
}

func WithUntagResourceOutput(output *lambda.UntagResourceOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.untagResourceOutput = output
	}
}

func WithUntagResourceError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.untagResourceError = err
	}
}

func WithCreateFunctionOutput(output *lambda.CreateFunctionOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.createFunctionOutput = output
	}
}

func WithCreateFunctionError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.createFunctionError = err
	}
}

func (m *lambdaServiceMock) GetFunction(
	ctx context.Context,
	params *lambda.GetFunctionInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getFunctionOutput, m.getFunctionError
}

func (m *lambdaServiceMock) GetFunctionCodeSigningConfig(
	ctx context.Context,
	params *lambda.GetFunctionCodeSigningConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionCodeSigningConfigOutput, error) {
	m.RegisterCall("GetFunctionCodeSigningConfig", ctx, params)
	return m.getFunctionCodeSigningOutput, m.getFunctionCodeSigningError
}

func (m *lambdaServiceMock) GetFunctionRecursionConfig(
	ctx context.Context,
	params *lambda.GetFunctionRecursionConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionRecursionConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getFunctionRecursionOutput, m.getFunctionRecursionError
}

func (m *lambdaServiceMock) GetFunctionConcurrency(
	ctx context.Context,
	params *lambda.GetFunctionConcurrencyInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionConcurrencyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getFunctionConcurrencyOutput, m.getFunctionConcurrencyError
}

func (m *lambdaServiceMock) DeleteFunction(
	ctx context.Context,
	params *lambda.DeleteFunctionInput,
	optFns ...func(*lambda.Options),
) (*lambda.DeleteFunctionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteFunctionOutput, m.deleteFunctionError
}

func (m *lambdaServiceMock) UpdateFunctionConfiguration(
	ctx context.Context,
	params *lambda.UpdateFunctionConfigurationInput,
	optFns ...func(*lambda.Options),
) (*lambda.UpdateFunctionConfigurationOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateFunctionConfigurationOutput, m.updateFunctionConfigurationError
}

func (m *lambdaServiceMock) UpdateFunctionCode(
	ctx context.Context,
	params *lambda.UpdateFunctionCodeInput,
	optFns ...func(*lambda.Options),
) (*lambda.UpdateFunctionCodeOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateFunctionCodeOutput, m.updateFunctionCodeError
}

func (m *lambdaServiceMock) PutFunctionCodeSigningConfig(
	ctx context.Context,
	params *lambda.PutFunctionCodeSigningConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.PutFunctionCodeSigningConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putFunctionCodeSigningConfigOutput, m.putFunctionCodeSigningConfigError
}

func (m *lambdaServiceMock) PutFunctionConcurrency(
	ctx context.Context,
	params *lambda.PutFunctionConcurrencyInput,
	optFns ...func(*lambda.Options),
) (*lambda.PutFunctionConcurrencyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putFunctionConcurrencyOutput, m.putFunctionConcurrencyError
}

func (m *lambdaServiceMock) PutFunctionRecursionConfig(
	ctx context.Context,
	params *lambda.PutFunctionRecursionConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.PutFunctionRecursionConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putFunctionRecursionConfigOutput, m.putFunctionRecursionConfigError
}

func (m *lambdaServiceMock) PutRuntimeManagementConfig(
	ctx context.Context,
	params *lambda.PutRuntimeManagementConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.PutRuntimeManagementConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putRuntimeManagementConfigOutput, m.putRuntimeManagementConfigError
}

func (m *lambdaServiceMock) TagResource(
	ctx context.Context,
	params *lambda.TagResourceInput,
	optFns ...func(*lambda.Options),
) (*lambda.TagResourceOutput, error) {
	m.RegisterCall(ctx, params)
	return m.tagResourceOutput, m.tagResourceError
}

func (m *lambdaServiceMock) UntagResource(
	ctx context.Context,
	params *lambda.UntagResourceInput,
	optFns ...func(*lambda.Options),
) (*lambda.UntagResourceOutput, error) {
	m.RegisterCall(ctx, params)
	return m.untagResourceOutput, m.untagResourceError
}

func (m *lambdaServiceMock) CreateFunction(
	ctx context.Context,
	params *lambda.CreateFunctionInput,
	optFns ...func(*lambda.Options),
) (*lambda.CreateFunctionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createFunctionOutput, m.createFunctionError
}

func createBaseTestFunctionConfig(
	functionName string,
	runtime types.Runtime,
	handler string,
	role string,
) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(functionName),
			FunctionArn:  aws.String(fmt.Sprintf("arn:aws:lambda:us-east-1:123456789012:function:%s", functionName)),
			Runtime:      runtime,
			Handler:      aws.String(handler),
			Role:         aws.String(role),
			Architectures: []types.Architecture{
				types.ArchitectureX8664,
			},
		},
		Code: &types.FunctionCodeLocation{
			Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
		},
	}
}
