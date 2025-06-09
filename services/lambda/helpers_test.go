package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/two-hundred/celerity-provider-aws/utils"
	"github.com/two-hundred/celerity/libs/blueprint/provider"
)

type getExternalStateTestCase struct {
	name                 string
	lambdaServiceFactory func(awsConfig *aws.Config, providerContext provider.Context) Service
	awsConfigStore       *utils.AWSConfigStore
	input                *provider.ResourceGetExternalStateInput
	expectedOutput       *provider.ResourceGetExternalStateOutput
	expectError          bool
}

type destroyTestCase struct {
	name                 string
	lambdaServiceFactory func(awsConfig *aws.Config, providerContext provider.Context) Service
	awsConfigStore       *utils.AWSConfigStore
	input                *provider.ResourceDestroyInput
	expectError          bool
}

type lambdaServiceMock struct {
	getFunctionOutput            *lambda.GetFunctionOutput
	getFunctionCodeSigningOutput *lambda.GetFunctionCodeSigningConfigOutput
	getFunctionRecursionOutput   *lambda.GetFunctionRecursionConfigOutput
	getFunctionConcurrencyOutput *lambda.GetFunctionConcurrencyOutput
	getFunctionError             error
	getFunctionCodeSigningError  error
	getFunctionRecursionError    error
	getFunctionConcurrencyError  error
	deleteFunctionOutput         *lambda.DeleteFunctionOutput
	deleteFunctionError          error
}

type lambdaServiceMockOption func(*lambdaServiceMock)

func createLambdaServiceMockFactory(
	opts ...lambdaServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) Service {
	mock := &lambdaServiceMock{}
	for _, opt := range opts {
		opt(mock)
	}
	return func(awsConfig *aws.Config, providerContext provider.Context) Service {
		return mock
	}
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

func (m *lambdaServiceMock) GetFunction(
	ctx context.Context,
	params *lambda.GetFunctionInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionOutput, error) {
	return m.getFunctionOutput, m.getFunctionError
}

func (m *lambdaServiceMock) GetFunctionCodeSigningConfig(
	ctx context.Context,
	params *lambda.GetFunctionCodeSigningConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionCodeSigningConfigOutput, error) {
	return m.getFunctionCodeSigningOutput, m.getFunctionCodeSigningError
}

func (m *lambdaServiceMock) GetFunctionRecursionConfig(
	ctx context.Context,
	params *lambda.GetFunctionRecursionConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionRecursionConfigOutput, error) {
	return m.getFunctionRecursionOutput, m.getFunctionRecursionError
}

func (m *lambdaServiceMock) GetFunctionConcurrency(
	ctx context.Context,
	params *lambda.GetFunctionConcurrencyInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionConcurrencyOutput, error) {
	return m.getFunctionConcurrencyOutput, m.getFunctionConcurrencyError
}

func (m *lambdaServiceMock) DeleteFunction(
	ctx context.Context,
	params *lambda.DeleteFunctionInput,
	optFns ...func(*lambda.Options),
) (*lambda.DeleteFunctionOutput, error) {
	return m.deleteFunctionOutput, m.deleteFunctionError
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
