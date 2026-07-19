//go:build unit

package linkutils

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

// This mocks only the calls mutateLambdaEnvironmentVariables makes,
// scripted per attempt so tests can drive the conflict-then-converge sequence that
// two concurrent links to the same function produce. All other Service methods
// panic via the nil embedded interface.
type seqLambdaService struct {
	lambdaservice.Service
	updateResponses []error
	updateInputs    []*lambda.UpdateFunctionConfigurationInput
	getResponses    []*lambda.GetFunctionOutput
	getCalls        int
}

func (s *seqLambdaService) UpdateFunctionConfiguration(
	_ context.Context,
	input *lambda.UpdateFunctionConfigurationInput,
	_ ...func(*lambda.Options),
) (*lambda.UpdateFunctionConfigurationOutput, error) {
	s.updateInputs = append(s.updateInputs, input)
	err := s.updateResponses[len(s.updateInputs)-1]
	if err != nil {
		return nil, err
	}
	return &lambda.UpdateFunctionConfigurationOutput{}, nil
}

func (s *seqLambdaService) GetFunction(
	_ context.Context,
	_ *lambda.GetFunctionInput,
	_ ...func(*lambda.Options),
) (*lambda.GetFunctionOutput, error) {
	output := s.getResponses[s.getCalls]
	s.getCalls++
	return output, nil
}

func configWithEnv(envVars map[string]string) *types.FunctionConfiguration {
	return &types.FunctionConfiguration{
		Environment: &types.EnvironmentResponse{Variables: envVars},
	}
}

var updateConflictErr = apiError(
	"ResourceConflictException",
	"The operation cannot be performed at this time. An update is in progress for resource: arn:aws:lambda:eu-west-2:123456789012:function:orders-api",
)

type LambdaEnvUpdateSuite struct {
	suite.Suite
}

func (s *LambdaEnvUpdateSuite) withFastBackoff(run func()) {
	original := transientRetryBackoff
	transientRetryBackoff = []time.Duration{time.Millisecond, time.Millisecond}
	defer func() { transientRetryBackoff = original }()
	run()
}

func (s *LambdaEnvUpdateSuite) Test_update_merges_and_writes_once_without_conflict() {
	service := &seqLambdaService{updateResponses: []error{nil}}

	err := UpdateLambdaEnvironmentVariables(
		context.Background(),
		service,
		"arn:aws:lambda:eu-west-2:123456789012:function:orders-api",
		configWithEnv(map[string]string{"EXISTING": "kept"}),
		map[string]string{"NEW_VAR": "added"},
	)

	s.Require().NoError(err)
	s.Require().Len(service.updateInputs, 1)
	s.Equal(
		map[string]string{"EXISTING": "kept", "NEW_VAR": "added"},
		service.updateInputs[0].Environment.Variables,
	)
	s.Zero(service.getCalls, "no conflict, so no re-read")
}

// The core concurrent-links scenario: this link's first write is rejected because
// another link's update is in flight on the same function. The retry must re-read the
// configuration and merge on top of what the winning link wrote, retrying with the
// original merge base would silently erase the other link's variables.
func (s *LambdaEnvUpdateSuite) Test_update_conflict_rereads_before_remerging() {
	s.withFastBackoff(func() {
		service := &seqLambdaService{
			updateResponses: []error{updateConflictErr, nil},
			getResponses: []*lambda.GetFunctionOutput{
				{Configuration: configWithEnv(map[string]string{
					"EXISTING":       "kept",
					"OTHER_LINK_VAR": "written-by-winner",
				})},
			},
		}

		err := UpdateLambdaEnvironmentVariables(
			context.Background(),
			service,
			"arn:aws:lambda:eu-west-2:123456789012:function:orders-api",
			configWithEnv(map[string]string{"EXISTING": "kept"}),
			map[string]string{"THIS_LINK_VAR": "added"},
		)

		s.Require().NoError(err)
		s.Require().Len(service.updateInputs, 2)
		s.Equal(1, service.getCalls)
		s.Equal(
			map[string]string{
				"EXISTING":       "kept",
				"OTHER_LINK_VAR": "written-by-winner",
				"THIS_LINK_VAR":  "added",
			},
			service.updateInputs[1].Environment.Variables,
			"the retry merge base must include the concurrent link's variables",
		)
	})
}

func (s *LambdaEnvUpdateSuite) Test_remove_conflict_rereads_before_remerging() {
	s.withFastBackoff(func() {
		service := &seqLambdaService{
			updateResponses: []error{updateConflictErr, nil},
			getResponses: []*lambda.GetFunctionOutput{
				{Configuration: configWithEnv(map[string]string{
					"REMOVE_ME":      "stale",
					"OTHER_LINK_VAR": "written-by-winner",
				})},
			},
		}

		err := RemoveLambdaEnvironmentVariables(
			context.Background(),
			service,
			"arn:aws:lambda:eu-west-2:123456789012:function:orders-api",
			configWithEnv(map[string]string{"REMOVE_ME": "stale"}),
			[]string{"REMOVE_ME"},
		)

		s.Require().NoError(err)
		s.Require().Len(service.updateInputs, 2)
		s.Equal(
			map[string]string{"OTHER_LINK_VAR": "written-by-winner"},
			service.updateInputs[1].Environment.Variables,
			"removal must only drop its own variables, keeping the concurrent link's",
		)
	})
}

func (s *LambdaEnvUpdateSuite) Test_update_defers_to_engine_when_window_exhausted() {
	s.withFastBackoff(func() {
		service := &seqLambdaService{
			// Initial attempt plus one per backoff step, all conflicting.
			updateResponses: []error{updateConflictErr, updateConflictErr, updateConflictErr},
			getResponses: []*lambda.GetFunctionOutput{
				{Configuration: configWithEnv(nil)},
				{Configuration: configWithEnv(nil)},
			},
		}

		err := UpdateLambdaEnvironmentVariables(
			context.Background(),
			service,
			"arn:aws:lambda:eu-west-2:123456789012:function:orders-api",
			nil,
			map[string]string{"NEW_VAR": "added"},
		)

		var retryErr *provider.RetryableError
		s.Require().True(
			errors.As(err, &retryErr),
			"an exhausted in-call window should defer to the engine's retry policy",
		)
		s.ErrorIs(retryErr.ChildError, updateConflictErr)
	})
}

func (s *LambdaEnvUpdateSuite) Test_update_returns_non_conflict_errors_immediately() {
	terminal := apiError("InvalidParameterValueException", "Environment variables are not supported")
	service := &seqLambdaService{updateResponses: []error{terminal}}

	err := UpdateLambdaEnvironmentVariables(
		context.Background(),
		service,
		"arn:aws:lambda:eu-west-2:123456789012:function:orders-api",
		nil,
		map[string]string{"NEW_VAR": "added"},
	)

	s.ErrorIs(err, terminal)
	s.Len(service.updateInputs, 1)
	s.Zero(service.getCalls)
}

func (s *LambdaEnvUpdateSuite) Test_update_stops_waiting_on_context_cancellation() {
	service := &seqLambdaService{updateResponses: []error{updateConflictErr}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := UpdateLambdaEnvironmentVariables(
		ctx,
		service,
		"arn:aws:lambda:eu-west-2:123456789012:function:orders-api",
		nil,
		map[string]string{"NEW_VAR": "added"},
	)

	s.ErrorIs(err, context.Canceled)
}

func TestLambdaEnvUpdateSuite(t *testing.T) {
	suite.Run(t, new(LambdaEnvUpdateSuite))
}
