package linkutils

import (
	"context"
	"fmt"
	"maps"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type LambdaLinkSetupData struct {
	LambdaFuncResourceInfo *provider.ResourceInfo
	LoadRoleInfo           bool
}

type LambdaLinkSetupContext struct {
	FunctionARN  string
	LambdaOutput *types.FunctionConfiguration
	// The name of the role in AWS.
	RoleName string
	// The name of the role resource in the blueprint.
	RoleResourceName string
	// The resource state of the role in the blueprint.
	RoleResourceState *state.ResourceState
	LambdaService     lambdaservice.Service
}

// SetupLinkFromLambdaFunction sets up a link from a Lambda function to another resource.
func SetupLinkFromLambdaFunction(
	ctx context.Context,
	setupData *LambdaLinkSetupData,
	lambdaService lambdaservice.Service,
	resourceService provider.ResourceService,
	providerCtx provider.Context,
) (*LambdaLinkSetupContext, error) {
	functionARN, hasFunctionARN := utils.ExtractARNFromResourceInfo(
		setupData.LambdaFuncResourceInfo,
	)
	if !hasFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked from %q function resource",
			pluginutils.GetResourceName(setupData.LambdaFuncResourceInfo),
		)
	}

	lambdaOutput, err := lambdaService.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(functionARN),
	})
	if err != nil {
		return nil, err
	}

	setupCtx := &LambdaLinkSetupContext{
		FunctionARN:   functionARN,
		LambdaOutput:  lambdaOutput.Configuration,
		LambdaService: lambdaService,
	}

	if setupData.LoadRoleInfo {
		roleARN := aws.ToString(lambdaOutput.Configuration.Role)
		roleResourceState, err := resourceService.LookupResourceInState(
			ctx,
			&provider.ResourceLookupInput{
				InstanceID:      pluginutils.GetInstanceID(setupData.LambdaFuncResourceInfo),
				ResourceType:    "aws/iam/role",
				ExternalID:      roleARN,
				ProviderContext: providerCtx,
			},
		)
		if err != nil {
			return nil, err
		}

		if roleResourceState == nil {
			return nil, fmt.Errorf(
				"the lambda execution role %q was not created as a part of the "+
					"same blueprint, when linking a lambda function to another resource,"+
					" the execution role must be created as a part of the same blueprint",
				roleARN,
			)
		}

		roleName, hasRoleName := pluginutils.GetValueByPath(
			"$.name",
			roleResourceState.SpecData,
		)
		if !hasRoleName {
			return nil, fmt.Errorf(
				"role name could not be retrieved from the execution role of the %q function resource",
				pluginutils.GetResourceName(setupData.LambdaFuncResourceInfo),
			)
		}

		setupCtx.RoleName = core.StringValue(roleName)
		setupCtx.RoleResourceName = roleResourceState.Name
		setupCtx.RoleResourceState = roleResourceState
	}

	return setupCtx, nil
}

// UpdateLambdaEnvironmentVariables updates the environment variables for a Lambda function
// by merging the current environment variables with the new ones.
// This is mostly useful for links that connect lambda functions to other resources.
func UpdateLambdaEnvironmentVariables(
	ctx context.Context,
	lambdaService lambdaservice.Service,
	functionARN string,
	currentConfig *types.FunctionConfiguration,
	envVarsToSet map[string]string,
) error {
	finalEnvVars := map[string]string{}
	maps.Copy(finalEnvVars, currentConfig.Environment.Variables)
	maps.Copy(finalEnvVars, envVarsToSet)

	_, err := lambdaService.UpdateFunctionConfiguration(
		ctx,
		&lambda.UpdateFunctionConfigurationInput{
			FunctionName: aws.String(functionARN),
			Environment: &types.Environment{
				Variables: finalEnvVars,
			},
		},
	)
	return err
}

// RemoveLambdaEnvironmentVariables removes the environment variables for a Lambda function
// by removing the specified environment variables from the current environment variables.
// This is mostly useful for links that connect lambda functions to other resources.
func RemoveLambdaEnvironmentVariables(
	ctx context.Context,
	lambdaService lambdaservice.Service,
	functionARN string,
	currentConfig *types.FunctionConfiguration,
	envVarsToRemove []string,
) error {
	finalEnvVars := map[string]string{}
	maps.Copy(finalEnvVars, currentConfig.Environment.Variables)

	for _, envVarName := range envVarsToRemove {
		delete(finalEnvVars, envVarName)
	}

	_, err := lambdaService.UpdateFunctionConfiguration(
		ctx,
		&lambda.UpdateFunctionConfigurationInput{
			FunctionName: aws.String(functionARN),
			Environment: &types.Environment{
				Variables: finalEnvVars,
			},
		},
	)
	return err
}
