package lambdalinks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	lambdaService, err := l.getLambdaService(
		ctx,
		provider.NewProviderContextFromLinkContext(
			input.LinkContext,
			"aws",
		),
	)
	if err != nil {
		return nil, err
	}

	annotations := getLambdaFunctionLinkAnnotations(
		input.ResourceInfo,
		input.OtherResourceInfo,
	)
	if !annotations.populateEnvVars {
		return &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields(),
			ResourceDataMappings: map[string]string{},
		}, nil
	}

	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceInfo,
			LoadRoleInfo:           false,
		},
		lambdaService,
		// Resource service is only needed if we need to load the role info.
		/* resourceService */
		nil,
		provider.NewProviderContextFromLinkContext(
			input.LinkContext,
			"aws",
		),
	)
	if err != nil {
		return nil, err
	}

	otherFunctionARN, hasOtherFunctionARN := utils.ExtractARNFromResourceInfo(
		input.OtherResourceInfo,
	)
	if !hasOtherFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked to %q function resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeCallerFunctionEnvVars(
			ctx,
			input,
			setupCtx.FunctionARN,
			annotations.envVarName,
			setupCtx.LambdaOutput,
			lambdaService,
		)
	}

	return l.addCallerFunctionEnvVars(
		ctx,
		input,
		setupCtx.FunctionARN,
		otherFunctionARN,
		annotations.envVarName,
		setupCtx.LambdaOutput,
		lambdaService,
	)
}

func (l *lambdaFunctionFunctionLinkActions) addCallerFunctionEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	otherFunctionARN string,
	envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := invokeLambdaFunctionEnvVarName(
		envVarName,
		input.OtherResourceInfo,
	)
	dataMappingKey := fmt.Sprintf(
		"%s::spec.environment.variables[\"%s\"]",
		input.ResourceInfo.ResourceName,
		finalEnvVarName,
	)
	linkDataFieldPath := fmt.Sprintf(
		"%s.environmentVariables[\"%s\"]",
		input.ResourceInfo.ResourceName,
		finalEnvVarName,
	)

	err := linkutils.UpdateLambdaEnvironmentVariables(
		ctx,
		lambdaService,
		functionARN,
		currentFunctionConfig,
		map[string]string{
			finalEnvVarName: otherFunctionARN,
		},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			pluginutils.GetResourceName(input.ResourceInfo),
			core.MappingNodeFields(
				"environmentVariables",
				core.MappingNodeFields(
					envVarName,
					core.MappingNodeFromString(otherFunctionARN),
				),
			),
		),
		ResourceDataMappings: map[string]string{
			dataMappingKey: linkDataFieldPath,
		},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) removeCallerFunctionEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := invokeLambdaFunctionEnvVarName(
		envVarName,
		input.OtherResourceInfo,
	)

	err := linkutils.RemoveLambdaEnvironmentVariables(
		ctx,
		lambdaService,
		functionARN,
		currentFunctionConfig,
		[]string{finalEnvVarName},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			pluginutils.GetResourceName(input.ResourceInfo),
			core.MappingNodeFields(),
		),
		ResourceDataMappings: map[string]string{},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The target function is not updated as a part of the link update,
	// the link only requires updates to the linked from function
	// and its execution role that will allow it to invoke the target function.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(
		input.LinkContext,
		"aws",
	)
	lambdaService, err := l.getLambdaService(
		ctx,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceAInfo,
			LoadRoleInfo:           true,
		},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	// Acquire a lock on the role resource to prevent concurrent updates
	// to the same role from multiple links.
	err = input.ResourceService.AcquireResourceLock(
		ctx,
		&provider.AcquireResourceLockInput{
			InstanceID:      pluginutils.GetInstanceID(input.ResourceAInfo),
			ResourceName:    setupCtx.RoleResourceName,
			ProviderContext: providerCtx,
		},
	)
	if err != nil {
		return nil, err
	}

	iamService, err := l.getIamService(
		ctx,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	rolePolicies, err := iamService.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: aws.String(setupCtx.RoleName),
	})
	if err != nil {
		return nil, err
	}

	otherFunctionARN, hasOtherFunctionARN := utils.ExtractARNFromResourceInfo(
		input.ResourceBInfo,
	)
	if !hasOtherFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked to %q function resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	sid := createLambdaFunctionInvokeSID(input.ResourceBInfo)
	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeRolePolicyPermissions(
			ctx,
			input,
			setupCtx,
			rolePolicies,
			iamService,
		)
	}

	output, err := l.setRolePolicyPermissions(
		ctx,
		input,
		otherFunctionARN,
		sid,
		setupCtx,
		rolePolicies,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	return l.updateNetworkingElements(
		ctx,
		input,
		output,
	)
}

func (l *lambdaFunctionFunctionLinkActions) setRolePolicyPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	otherFunctionARN string,
	sid string,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	invokePermission := map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   "lambda:InvokeFunction",
		"Resource": otherFunctionARN,
	}

	if len(rolePolicies.PolicyNames) > 0 {
		return l.updateExistingRolePolicy(
			ctx,
			input,
			&roleInfo{
				roleName:          setupCtx.RoleName,
				roleResourceState: setupCtx.RoleResourceState,
				rolePolicies:      rolePolicies,
				invokePermission:  invokePermission,
				sid:               sid,
			},
			iamService,
		)
	}

	return l.addNewRolePolicy(
		ctx,
		input,
		&roleInfo{
			roleName:          setupCtx.RoleName,
			roleResourceState: setupCtx.RoleResourceState,
			rolePolicies:      rolePolicies,
			invokePermission:  invokePermission,
			sid:               sid,
		},
		iamService,
	)
}

func (l *lambdaFunctionFunctionLinkActions) addNewRolePolicy(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleData *roleInfo,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	policyName, err := linkutils.CreateNewRolePolicy(
		ctx,
		iamService,
		input.InstanceName,
		roleData.roleName,
		[]map[string]any{roleData.invokePermission},
	)
	if err != nil {
		return nil, err
	}

	roleResourcePath := fmt.Sprintf(
		"%s::spec.policies[@.name=\"%q\"].Statement[@.Sid=\"%q\"]",
		roleData.roleResourceState.Name,
		policyName,
		roleData.sid,
	)
	linkDataExecRoleName := createLinkDataExecutionRoleName(input.ResourceAInfo)
	linkDataFieldPath := linkutils.PermissionFieldPath(linkDataExecRoleName)

	invokePermissionNode, err := pluginutils.AnyToMappingNode(roleData.invokePermission)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: core.MappingNodeFields(
			linkDataExecRoleName,
			core.MappingNodeFields(
				linkutils.PermissionFieldName,
				invokePermissionNode,
				linkutils.PolicyNameFieldName,
				policyName,
			),
		),
		ResourceDataMappings: map[string]string{
			roleResourcePath: linkDataFieldPath,
		},
	}, nil
}

type roleInfo struct {
	roleName          string
	roleResourceState *state.ResourceState
	rolePolicies      *iam.ListRolePoliciesOutput
	invokePermission  map[string]any
	sid               string
}

func (l *lambdaFunctionFunctionLinkActions) updateExistingRolePolicy(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleData *roleInfo,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	executionRoleName := fmt.Sprintf(
		"%sExecutionRole",
		input.ResourceAInfo.ResourceName,
	)
	linkDataFieldPath := linkutils.PermissionFieldPath(executionRoleName)

	policyName, existingPermSID := linkutils.ExtractPolicyNameAndCurrentPermissionSID(
		input.CurrentLinkState,
		executionRoleName,
		roleData.rolePolicies,
	)

	err := linkutils.UpdateExistingRolePolicy(
		ctx,
		iamService,
		roleData.roleName,
		policyName,
		[]map[string]any{roleData.invokePermission},
		[]string{existingPermSID},
	)
	if err != nil {
		return nil, err
	}

	roleResourcePath := fmt.Sprintf(
		"%s::spec.policies[@.name=\"%q\"].Statement[@.Sid=\"%q\"]",
		roleData.roleResourceState.Name,
		policyName,
		roleData.sid,
	)

	invokePermissionNode, err := pluginutils.AnyToMappingNode(roleData.invokePermission)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: core.MappingNodeFields(
			executionRoleName,
			core.MappingNodeFields(
				linkutils.PermissionFieldName,
				invokePermissionNode,
				linkutils.PolicyNameFieldName,
				policyName,
			),
		),
		ResourceDataMappings: map[string]string{
			roleResourcePath: linkDataFieldPath,
		},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) removeRolePolicyPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	executionRoleName := fmt.Sprintf(
		"%sExecutionRole",
		input.ResourceAInfo.ResourceName,
	)

	policyName, existingPermSID := linkutils.ExtractPolicyNameAndCurrentPermissionSID(
		input.CurrentLinkState,
		executionRoleName,
		rolePolicies,
	)

	err := linkutils.RemoveExistingRolePolicyPermissions(
		ctx,
		iamService,
		setupCtx.RoleName,
		policyName,
		[]string{existingPermSID},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: core.MappingNodeFields(
			executionRoleName,
			core.MappingNodeFields(),
		),
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) updateNetworkingElements(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// 1. check if function has VPC config (SGs and subnets)
	// 2. If it does, check if there is a VPC endpoint that covers the lambda API or if the subnet has access to the public internet (NAT Gateway or NACL for public subnet + SG allows it)
	// 3. If there is no VPC endpoint and no internet access, add a VPC endpoint for the Lambda API and ensure that the endpoint allows traffic from the function's security group
	// 4. If requested (through annotations), add a NAT Gateway and SG rule to allow access to the Lambda API
	return output, nil
}

func invokeLambdaFunctionEnvVarName(
	userDefinedEnvVarName string,
	resourceInfo *provider.ResourceInfo,
) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}

	return fmt.Sprintf(
		"INVOKE_LAMBDA_FUNCTION_%s",
		resourceInfo.ResourceName,
	)
}

func createLambdaFunctionInvokeSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"InvokeLambdaFunction%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

type lambdaFunctionLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
}

func getLambdaFunctionLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *lambdaFunctionLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key: fmt.Sprintf(
				"aws.lambda.function.%s.populateEnvVars",
				otherResourceInfo.ResourceName,
			),
			FallbackKey: "aws.lambda.function.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf(
				"aws.lambda.function.%s.envVarName",
				otherResourceInfo.ResourceName,
			),
		},
	)

	return &lambdaFunctionLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
	}
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sExecutionRole",
		resourceInfo.ResourceName,
	)
}
