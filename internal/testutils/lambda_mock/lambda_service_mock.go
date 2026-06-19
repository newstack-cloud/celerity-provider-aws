package lambdamock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type lambdaServiceMock struct {
	plugintestutils.MockCalls

	addPermissionOutput    *lambda.AddPermissionOutput
	addPermissionError     error
	removePermissionOutput *lambda.RemovePermissionOutput
	removePermissionError  error

	getFunctionOutput *lambda.GetFunctionOutput
	getFunctionError  error

	getFunctionCodeSigningOutput *lambda.GetFunctionCodeSigningConfigOutput
	getFunctionCodeSigningError  error

	getFunctionRecursionOutput *lambda.GetFunctionRecursionConfigOutput
	getFunctionRecursionError  error

	getFunctionConcurrencyOutput *lambda.GetFunctionConcurrencyOutput
	getFunctionConcurrencyError  error

	// Get Provisioned Concurrency-related mock fields
	getProvisionedConcurrencyOutput *lambda.GetProvisionedConcurrencyConfigOutput
	getProvisionedConcurrencyError  error

	// Delete Function-related mock fields
	deleteFunctionOutput *lambda.DeleteFunctionOutput
	deleteFunctionError  error

	// Update Function Configuration-related mock fields
	updateFunctionConfigurationOutput *lambda.UpdateFunctionConfigurationOutput
	updateFunctionConfigurationError  error
	updateFunctionCodeOutput          *lambda.UpdateFunctionCodeOutput
	updateFunctionCodeError           error

	// Function Code Signing Config-related mock fields
	putFunctionCodeSigningConfigOutput    *lambda.PutFunctionCodeSigningConfigOutput
	putFunctionCodeSigningConfigError     error
	deleteFunctionCodeSigningConfigOutput *lambda.DeleteFunctionCodeSigningConfigOutput
	deleteFunctionCodeSigningConfigError  error

	// Function Concurrency-related mock fields
	putFunctionConcurrencyOutput     *lambda.PutFunctionConcurrencyOutput
	putFunctionConcurrencyError      error
	putFunctionRecursionConfigOutput *lambda.PutFunctionRecursionConfigOutput
	putFunctionRecursionConfigError  error
	putRuntimeManagementConfigOutput *lambda.PutRuntimeManagementConfigOutput
	putRuntimeManagementConfigError  error

	// Tag Resource-related mock fields
	tagResourceOutput   *lambda.TagResourceOutput
	tagResourceError    error
	untagResourceOutput *lambda.UntagResourceOutput
	untagResourceError  error
	listTagsOutput      *lambda.ListTagsOutput
	listTagsError       error

	// Create Function-related mock fields
	createFunctionOutput *lambda.CreateFunctionOutput
	createFunctionError  error

	// Publish Version-related mock fields
	publishVersionOutput *lambda.PublishVersionOutput
	publishVersionError  error

	// Provisioned Concurrency-related mock fields
	putProvisionedConcurrencyOutput *lambda.PutProvisionedConcurrencyConfigOutput
	putProvisionedConcurrencyError  error

	// Alias-related mock fields
	createAliasOutput *lambda.CreateAliasOutput
	createAliasError  error
	getAliasOutput    *lambda.GetAliasOutput
	getAliasError     error
	updateAliasOutput *lambda.UpdateAliasOutput
	updateAliasError  error
	deleteAliasOutput *lambda.DeleteAliasOutput
	deleteAliasError  error

	// Code signing config-related mock fields
	createCodeSigningConfigOutput *lambda.CreateCodeSigningConfigOutput
	createCodeSigningConfigError  error
	getCodeSigningConfigOutput    *lambda.GetCodeSigningConfigOutput
	getCodeSigningConfigError     error
	updateCodeSigningConfigOutput *lambda.UpdateCodeSigningConfigOutput
	updateCodeSigningConfigError  error
	deleteCodeSigningConfigOutput *lambda.DeleteCodeSigningConfigOutput
	deleteCodeSigningConfigError  error

	// Event Source Mapping fields
	createEventSourceMappingOutput *lambda.CreateEventSourceMappingOutput
	createEventSourceMappingError  error
	getEventSourceMappingOutput    *lambda.GetEventSourceMappingOutput
	getEventSourceMappingError     error
	updateEventSourceMappingOutput *lambda.UpdateEventSourceMappingOutput
	updateEventSourceMappingError  error
	deleteEventSourceMappingOutput *lambda.DeleteEventSourceMappingOutput
	deleteEventSourceMappingError  error

	// Function URL fields
	createFunctionUrlConfigOutput *lambda.CreateFunctionUrlConfigOutput
	createFunctionUrlConfigError  error
	getFunctionUrlConfigOutput    *lambda.GetFunctionUrlConfigOutput
	getFunctionUrlConfigError     error
	updateFunctionUrlConfigOutput *lambda.UpdateFunctionUrlConfigOutput
	updateFunctionUrlConfigError  error
	deleteFunctionUrlConfigOutput *lambda.DeleteFunctionUrlConfigOutput
	deleteFunctionUrlConfigError  error

	// Layer Version fields
	publishLayerVersionOutput *lambda.PublishLayerVersionOutput
	publishLayerVersionError  error
	getLayerVersionOutput     *lambda.GetLayerVersionOutput
	getLayerVersionError      error
	deleteLayerVersionOutput  *lambda.DeleteLayerVersionOutput
	deleteLayerVersionError   error

	// Event Invoke Config fields
	putFunctionEventInvokeConfigOutput    *lambda.PutFunctionEventInvokeConfigOutput
	putFunctionEventInvokeConfigError     error
	getFunctionEventInvokeConfigOutput    *lambda.GetFunctionEventInvokeConfigOutput
	getFunctionEventInvokeConfigError     error
	deleteFunctionEventInvokeConfigOutput *lambda.DeleteFunctionEventInvokeConfigOutput
	deleteFunctionEventInvokeConfigError  error
	updateFunctionEventInvokeConfigOutput *lambda.UpdateFunctionEventInvokeConfigOutput
	updateFunctionEventInvokeConfigError  error

	// Layer Version Permissions fields
	addLayerVersionPermissionOutput    *lambda.AddLayerVersionPermissionOutput
	addLayerVersionPermissionError     error
	getLayerVersionPolicyOutput        *lambda.GetLayerVersionPolicyOutput
	getLayerVersionPolicyError         error
	removeLayerVersionPermissionOutput *lambda.RemoveLayerVersionPermissionOutput
	removeLayerVersionPermissionError  error

	// Event Source Mapping mock methods
	MockCreateEventSourceMapping func(ctx context.Context, input *lambda.CreateEventSourceMappingInput) (*lambda.CreateEventSourceMappingOutput, error)
	MockGetEventSourceMapping    func(ctx context.Context, input *lambda.GetEventSourceMappingInput) (*lambda.GetEventSourceMappingOutput, error)
	MockUpdateEventSourceMapping func(ctx context.Context, input *lambda.UpdateEventSourceMappingInput) (*lambda.UpdateEventSourceMappingOutput, error)
	MockDeleteEventSourceMapping func(ctx context.Context, input *lambda.DeleteEventSourceMappingInput) (*lambda.DeleteEventSourceMappingOutput, error)
}

type lambdaServiceMockOption func(*lambdaServiceMock)

func CreateLambdaServiceMockFactory(
	opts ...lambdaServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
	mock := CreateLambdaServiceMock(opts...)
	return func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return mock
	}
}

func CreateLambdaServiceMock(
	opts ...lambdaServiceMockOption,
) *lambdaServiceMock {
	mock := &lambdaServiceMock{}

	for _, opt := range opts {
		opt(mock)
	}

	return mock
}

// NewMockService creates a new mock Service for testing.
func NewMockService(t any) *lambdaServiceMock {
	return CreateLambdaServiceMock()
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

func WithGetProvisionedConcurrencyOutput(
	output *lambda.GetProvisionedConcurrencyConfigOutput,
) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getProvisionedConcurrencyOutput = output
	}
}

func WithGetProvisionedConcurrencyError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getProvisionedConcurrencyError = err
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

func WithDeleteFunctionCodeSigningConfigOutput(output *lambda.DeleteFunctionCodeSigningConfigOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteFunctionCodeSigningConfigOutput = output
	}
}

func WithDeleteFunctionCodeSigningConfigError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteFunctionCodeSigningConfigError = err
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

func WithListTagsOutput(output *lambda.ListTagsOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.listTagsOutput = output
	}
}

func WithListTagsError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.listTagsError = err
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

func WithPublishVersionOutput(output *lambda.PublishVersionOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.publishVersionOutput = output
	}
}

func WithPublishVersionError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.publishVersionError = err
	}
}

func WithPutProvisionedConcurrencyConfigOutput(output *lambda.PutProvisionedConcurrencyConfigOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putProvisionedConcurrencyOutput = output
	}
}

func WithPutProvisionedConcurrencyConfigError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.putProvisionedConcurrencyError = err
	}
}

// Alias mock options

func WithCreateAliasOutput(output *lambda.CreateAliasOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.createAliasOutput = output
	}
}

func WithCreateAliasError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.createAliasError = err
	}
}

func WithGetAliasOutput(output *lambda.GetAliasOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getAliasOutput = output
	}
}

func WithGetAliasError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getAliasError = err
	}
}

func WithUpdateAliasOutput(output *lambda.UpdateAliasOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.updateAliasOutput = output
	}
}

func WithUpdateAliasError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.updateAliasError = err
	}
}

func WithDeleteAliasOutput(output *lambda.DeleteAliasOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteAliasOutput = output
	}
}

func WithDeleteAliasError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteAliasError = err
	}
}

// Code signing config mock options

func WithCreateCodeSigningConfigOutput(output *lambda.CreateCodeSigningConfigOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.createCodeSigningConfigOutput = output
	}
}

func WithCreateCodeSigningConfigError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.createCodeSigningConfigError = err
	}
}

func WithGetCodeSigningConfigOutput(output *lambda.GetCodeSigningConfigOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getCodeSigningConfigOutput = output
	}
}

func WithGetCodeSigningConfigError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getCodeSigningConfigError = err
	}
}

func WithUpdateCodeSigningConfigOutput(output *lambda.UpdateCodeSigningConfigOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.updateCodeSigningConfigOutput = output
	}
}

func WithUpdateCodeSigningConfigError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.updateCodeSigningConfigError = err
	}
}

func WithDeleteCodeSigningConfigOutput(output *lambda.DeleteCodeSigningConfigOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteCodeSigningConfigOutput = output
	}
}

func WithDeleteCodeSigningConfigError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteCodeSigningConfigError = err
	}
}

// Event Source Mapping mock helpers.
func WithCreateEventSourceMappingOutput(output *lambda.CreateEventSourceMappingOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.createEventSourceMappingOutput = output
	}
}

func WithGetEventSourceMappingOutput(output *lambda.GetEventSourceMappingOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getEventSourceMappingOutput = output
	}
}

func WithGetEventSourceMappingError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.getEventSourceMappingError = err
	}
}

func WithUpdateEventSourceMappingOutput(output *lambda.UpdateEventSourceMappingOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.updateEventSourceMappingOutput = output
	}
}

func WithDeleteEventSourceMappingOutput(output *lambda.DeleteEventSourceMappingOutput) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteEventSourceMappingOutput = output
	}
}

func WithDeleteEventSourceMappingError(err error) lambdaServiceMockOption {
	return func(m *lambdaServiceMock) {
		m.deleteEventSourceMappingError = err
	}
}

// Function URL mock methods

func (m *lambdaServiceMock) DeleteEventSourceMapping(ctx context.Context, input *lambda.DeleteEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.DeleteEventSourceMappingOutput, error) {
	m.RegisterCall(ctx, input)
	if m.MockDeleteEventSourceMapping != nil {
		return m.MockDeleteEventSourceMapping(ctx, input)
	}
	return m.deleteEventSourceMappingOutput, m.deleteEventSourceMappingError
}

func (m *lambdaServiceMock) CreateFunctionUrlConfig(
	ctx context.Context,
	params *lambda.CreateFunctionUrlConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.CreateFunctionUrlConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createFunctionUrlConfigOutput, m.createFunctionUrlConfigError
}

func (m *lambdaServiceMock) GetFunctionUrlConfig(
	ctx context.Context,
	params *lambda.GetFunctionUrlConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionUrlConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getFunctionUrlConfigOutput, m.getFunctionUrlConfigError
}

func (m *lambdaServiceMock) UpdateFunctionUrlConfig(
	ctx context.Context,
	params *lambda.UpdateFunctionUrlConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.UpdateFunctionUrlConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateFunctionUrlConfigOutput, m.updateFunctionUrlConfigError
}

func (m *lambdaServiceMock) DeleteFunctionUrlConfig(
	ctx context.Context,
	params *lambda.DeleteFunctionUrlConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.DeleteFunctionUrlConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteFunctionUrlConfigOutput, m.deleteFunctionUrlConfigError
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

func (m *lambdaServiceMock) GetProvisionedConcurrencyConfig(
	ctx context.Context,
	params *lambda.GetProvisionedConcurrencyConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetProvisionedConcurrencyConfigOutput, error) {
	m.RegisterCall(ctx, params)

	if m.getProvisionedConcurrencyError != nil {
		return nil, m.getProvisionedConcurrencyError
	}

	return m.getProvisionedConcurrencyOutput, nil
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

func (m *lambdaServiceMock) DeleteFunctionCodeSigningConfig(
	ctx context.Context,
	params *lambda.DeleteFunctionCodeSigningConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.DeleteFunctionCodeSigningConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteFunctionCodeSigningConfigOutput, m.deleteFunctionCodeSigningConfigError
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

func (m *lambdaServiceMock) PublishVersion(
	ctx context.Context,
	params *lambda.PublishVersionInput,
	optFns ...func(*lambda.Options),
) (*lambda.PublishVersionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.publishVersionOutput, m.publishVersionError
}

func (m *lambdaServiceMock) PutProvisionedConcurrencyConfig(
	ctx context.Context,
	params *lambda.PutProvisionedConcurrencyConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.PutProvisionedConcurrencyConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putProvisionedConcurrencyOutput, m.putProvisionedConcurrencyError
}

// Alias mock methods

func (m *lambdaServiceMock) CreateAlias(
	ctx context.Context,
	params *lambda.CreateAliasInput,
	optFns ...func(*lambda.Options),
) (*lambda.CreateAliasOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createAliasOutput, m.createAliasError
}

func (m *lambdaServiceMock) GetAlias(
	ctx context.Context,
	params *lambda.GetAliasInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetAliasOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getAliasOutput, m.getAliasError
}

func (m *lambdaServiceMock) UpdateAlias(
	ctx context.Context,
	params *lambda.UpdateAliasInput,
	optFns ...func(*lambda.Options),
) (*lambda.UpdateAliasOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateAliasOutput, m.updateAliasError
}

func (m *lambdaServiceMock) DeleteAlias(
	ctx context.Context,
	params *lambda.DeleteAliasInput,
	optFns ...func(*lambda.Options),
) (*lambda.DeleteAliasOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteAliasOutput, m.deleteAliasError
}

// Code signing config mock methods

func (m *lambdaServiceMock) CreateCodeSigningConfig(
	ctx context.Context,
	params *lambda.CreateCodeSigningConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.CreateCodeSigningConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createCodeSigningConfigOutput, m.createCodeSigningConfigError
}

func (m *lambdaServiceMock) GetCodeSigningConfig(
	ctx context.Context,
	params *lambda.GetCodeSigningConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetCodeSigningConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getCodeSigningConfigOutput, m.getCodeSigningConfigError
}

func (m *lambdaServiceMock) UpdateCodeSigningConfig(
	ctx context.Context,
	params *lambda.UpdateCodeSigningConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.UpdateCodeSigningConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateCodeSigningConfigOutput, m.updateCodeSigningConfigError
}

func (m *lambdaServiceMock) DeleteCodeSigningConfig(
	ctx context.Context,
	params *lambda.DeleteCodeSigningConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.DeleteCodeSigningConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteCodeSigningConfigOutput, m.deleteCodeSigningConfigError
}

func (m *lambdaServiceMock) ListTags(
	ctx context.Context,
	params *lambda.ListTagsInput,
	optFns ...func(*lambda.Options),
) (*lambda.ListTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listTagsOutput, m.listTagsError
}

// Event Source Mapping mock implementations.
func (m *lambdaServiceMock) CreateEventSourceMapping(ctx context.Context, input *lambda.CreateEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.CreateEventSourceMappingOutput, error) {
	m.RegisterCall(ctx, input)
	if m.MockCreateEventSourceMapping != nil {
		return m.MockCreateEventSourceMapping(ctx, input)
	}
	return m.createEventSourceMappingOutput, m.createEventSourceMappingError
}

func (m *lambdaServiceMock) GetEventSourceMapping(ctx context.Context, input *lambda.GetEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.GetEventSourceMappingOutput, error) {
	m.RegisterCall(ctx, input)
	if m.MockGetEventSourceMapping != nil {
		return m.MockGetEventSourceMapping(ctx, input)
	}
	return m.getEventSourceMappingOutput, m.getEventSourceMappingError
}

func (m *lambdaServiceMock) UpdateEventSourceMapping(ctx context.Context, input *lambda.UpdateEventSourceMappingInput, optFns ...func(*lambda.Options)) (*lambda.UpdateEventSourceMappingOutput, error) {
	m.RegisterCall(ctx, input)
	if m.MockUpdateEventSourceMapping != nil {
		return m.MockUpdateEventSourceMapping(ctx, input)
	}
	return m.updateEventSourceMappingOutput, m.updateEventSourceMappingError
}

// Layer Version mock methods

func (m *lambdaServiceMock) PublishLayerVersion(
	ctx context.Context,
	params *lambda.PublishLayerVersionInput,
	optFns ...func(*lambda.Options),
) (*lambda.PublishLayerVersionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.publishLayerVersionOutput, m.publishLayerVersionError
}

func (m *lambdaServiceMock) GetLayerVersion(
	ctx context.Context,
	params *lambda.GetLayerVersionInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetLayerVersionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getLayerVersionOutput, m.getLayerVersionError
}

func (m *lambdaServiceMock) DeleteLayerVersion(
	ctx context.Context,
	params *lambda.DeleteLayerVersionInput,
	optFns ...func(*lambda.Options),
) (*lambda.DeleteLayerVersionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteLayerVersionOutput, m.deleteLayerVersionError
}

// Function URL helper functions

func WithCreateFunctionUrlConfigOutput(output *lambda.CreateFunctionUrlConfigOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.createFunctionUrlConfigOutput = output
	}
}

func WithCreateFunctionUrlConfigError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.createFunctionUrlConfigError = err
	}
}

func WithGetFunctionUrlConfigOutput(output *lambda.GetFunctionUrlConfigOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.getFunctionUrlConfigOutput = output
	}
}

func WithGetFunctionUrlConfigError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.getFunctionUrlConfigError = err
	}
}

func WithUpdateFunctionUrlConfigOutput(output *lambda.UpdateFunctionUrlConfigOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.updateFunctionUrlConfigOutput = output
	}
}

func WithUpdateFunctionUrlConfigError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.updateFunctionUrlConfigError = err
	}
}

func WithDeleteFunctionUrlConfigOutput(output *lambda.DeleteFunctionUrlConfigOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.deleteFunctionUrlConfigOutput = output
	}
}

func WithDeleteFunctionUrlConfigError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.deleteFunctionUrlConfigError = err
	}
}

// Layer Version helper functions

func WithPublishLayerVersionOutput(output *lambda.PublishLayerVersionOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.publishLayerVersionOutput = output
	}
}

func WithPublishLayerVersionError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.publishLayerVersionError = err
	}
}

func WithGetLayerVersionOutput(output *lambda.GetLayerVersionOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.getLayerVersionOutput = output
	}
}

func WithGetLayerVersionError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.getLayerVersionError = err
	}
}

func WithDeleteLayerVersionOutput(output *lambda.DeleteLayerVersionOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.deleteLayerVersionOutput = output
	}
}

func WithDeleteLayerVersionError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.deleteLayerVersionError = err
	}
}

// Event Invoke Config helper functions

func WithPutFunctionEventInvokeConfigOutput(output *lambda.PutFunctionEventInvokeConfigOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.putFunctionEventInvokeConfigOutput = output
	}
}

func WithPutFunctionEventInvokeConfigError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.putFunctionEventInvokeConfigError = err
	}
}

func WithGetFunctionEventInvokeConfigOutput(output *lambda.GetFunctionEventInvokeConfigOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.getFunctionEventInvokeConfigOutput = output
	}
}

func WithGetFunctionEventInvokeConfigError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.getFunctionEventInvokeConfigError = err
	}
}

func WithDeleteFunctionEventInvokeConfigOutput(output *lambda.DeleteFunctionEventInvokeConfigOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.deleteFunctionEventInvokeConfigOutput = output
	}
}

func WithDeleteFunctionEventInvokeConfigError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.deleteFunctionEventInvokeConfigError = err
	}
}

func WithUpdateFunctionEventInvokeConfigOutput(output *lambda.UpdateFunctionEventInvokeConfigOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.updateFunctionEventInvokeConfigOutput = output
	}
}

func WithUpdateFunctionEventInvokeConfigError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.updateFunctionEventInvokeConfigError = err
	}
}

// Event Invoke Config mock methods

func (m *lambdaServiceMock) PutFunctionEventInvokeConfig(
	ctx context.Context,
	params *lambda.PutFunctionEventInvokeConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.PutFunctionEventInvokeConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putFunctionEventInvokeConfigOutput, m.putFunctionEventInvokeConfigError
}

func (m *lambdaServiceMock) GetFunctionEventInvokeConfig(
	ctx context.Context,
	params *lambda.GetFunctionEventInvokeConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetFunctionEventInvokeConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getFunctionEventInvokeConfigOutput, m.getFunctionEventInvokeConfigError
}

func (m *lambdaServiceMock) DeleteFunctionEventInvokeConfig(
	ctx context.Context,
	params *lambda.DeleteFunctionEventInvokeConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.DeleteFunctionEventInvokeConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteFunctionEventInvokeConfigOutput, m.deleteFunctionEventInvokeConfigError
}

func (m *lambdaServiceMock) UpdateFunctionEventInvokeConfig(
	ctx context.Context,
	params *lambda.UpdateFunctionEventInvokeConfigInput,
	optFns ...func(*lambda.Options),
) (*lambda.UpdateFunctionEventInvokeConfigOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateFunctionEventInvokeConfigOutput, m.updateFunctionEventInvokeConfigError
}

// Layer Version Permissions mock methods

func (m *lambdaServiceMock) AddLayerVersionPermission(
	ctx context.Context,
	params *lambda.AddLayerVersionPermissionInput,
	optFns ...func(*lambda.Options),
) (*lambda.AddLayerVersionPermissionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.addLayerVersionPermissionOutput, m.addLayerVersionPermissionError
}

func (m *lambdaServiceMock) GetLayerVersionPolicy(
	ctx context.Context,
	params *lambda.GetLayerVersionPolicyInput,
	optFns ...func(*lambda.Options),
) (*lambda.GetLayerVersionPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getLayerVersionPolicyOutput, m.getLayerVersionPolicyError
}

func (m *lambdaServiceMock) RemoveLayerVersionPermission(
	ctx context.Context,
	params *lambda.RemoveLayerVersionPermissionInput,
	optFns ...func(*lambda.Options),
) (*lambda.RemoveLayerVersionPermissionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.removeLayerVersionPermissionOutput, m.removeLayerVersionPermissionError
}

// Layer Version Permissions mock helpers

func WithAddLayerVersionPermissionOutput(output *lambda.AddLayerVersionPermissionOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.addLayerVersionPermissionOutput = output
	}
}

func WithAddLayerVersionPermissionError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.addLayerVersionPermissionError = err
	}
}

func WithGetLayerVersionPolicyOutput(output *lambda.GetLayerVersionPolicyOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.getLayerVersionPolicyOutput = output
	}
}

func WithGetLayerVersionPolicyError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.getLayerVersionPolicyError = err
	}
}

func WithRemoveLayerVersionPermissionOutput(output *lambda.RemoveLayerVersionPermissionOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.removeLayerVersionPermissionOutput = output
	}
}

func WithRemoveLayerVersionPermissionError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.removeLayerVersionPermissionError = err
	}
}

func WithAddPermissionOutput(output *lambda.AddPermissionOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.addPermissionOutput = output
	}
}

func WithAddPermissionError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.addPermissionError = err
	}
}

func WithRemovePermissionOutput(output *lambda.RemovePermissionOutput) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.removePermissionOutput = output
	}
}

func WithRemovePermissionError(err error) func(*lambdaServiceMock) {
	return func(m *lambdaServiceMock) {
		m.removePermissionError = err
	}
}

func (m *lambdaServiceMock) AddPermission(
	ctx context.Context,
	params *lambda.AddPermissionInput,
	optFns ...func(*lambda.Options),
) (*lambda.AddPermissionOutput, error) {
	m.RegisterCall("AddPermission", params)
	if m.addPermissionError != nil {
		return nil, m.addPermissionError
	}
	return m.addPermissionOutput, nil
}

func (m *lambdaServiceMock) RemovePermission(
	ctx context.Context,
	params *lambda.RemovePermissionInput,
	optFns ...func(*lambda.Options),
) (*lambda.RemovePermissionOutput, error) {
	m.RegisterCall("RemovePermission", params)
	if m.removePermissionError != nil {
		return nil, m.removePermissionError
	}
	return m.removePermissionOutput, nil
}
