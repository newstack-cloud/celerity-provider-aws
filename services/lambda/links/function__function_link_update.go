package lambdalinks

import (
	"context"
	"errors"
	"fmt"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	lambdaService, _, err := l.getLambdaServiceWithRegion(
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

// UpdateIntermediaryResources grants (or revokes) the caller function's
// execution role permission to invoke the target function by packing a single IAM
// statement into the role's allocator-managed policies. The role is an existing
// intermediary in the blueprint, so the role lock is held while its shared policy
// set is read-modified.
func (l *lambdaFunctionFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(
		input.LinkContext,
		"aws",
	)
	lambdaService, region, err := l.getLambdaServiceWithRegion(
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

	// The execution role is shared by every link that grants it access, so lock it
	// for the read-modify-write of its policy set.
	err = input.ResourceService.AcquireResourceLock(
		ctx,
		&provider.AcquireResourceLockInput{
			InstanceID:      pluginutils.GetInstanceID(input.ResourceAInfo),
			ResourceName:    setupCtx.RoleResourceName,
			ProviderContext: providerCtx,
			AcquiredBy:      input.LinkID,
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

	sid := createLambdaFunctionInvokeSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		_, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setupCtx.RoleName,
			SID:      sid,
		})
		if err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
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

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: lambdaInvokeAccessStatement(sid, otherFunctionARN),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q permission to invoke Lambda %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	output := invokeAccessLinkOutput(input, setupCtx.RoleResourceName, sid, otherFunctionARN, result)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-attached caller reaching another Lambda needs an interface VPC endpoint for
	// the Lambda service; ActivateLinkNetworking is a no-op for non-VPC functions.
	return linkutils.ActivateLinkNetworking(
		ctx,
		ec2Service,
		input,
		linkutils.NetworkingActivation{
			Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
			Region:       region,
			AWSService:   "lambda",
			EndpointType: ec2types.VpcEndpointTypeInterface,
		},
		output,
	)
}

// lambdaInvokeAccessStatement builds the IAM policy statement (canonical
// PascalCase, as the IAM API expects) granting permission to invoke the target
// function.
func lambdaInvokeAccessStatement(sid, otherFunctionARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   "lambda:InvokeFunction",
		"Resource": otherFunctionARN,
	}
}

// invokeAccessLinkOutput records the granted statement in link data and, for
// inline placements, maps it onto the role's spec so the framework attributes the
// statement to this link and does not treat it as drift / strip it on redeploy.
func invokeAccessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, otherFunctionARN string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specInvokeStatementNode(sid, otherFunctionARN),
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it:
	// inline placements map the statement by Sid; managed (overflow) placements map
	// the attached managed policy ARN.
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

// specInvokeStatementNode builds the statement in the camelCase spec form the
// role's external state uses (after Cloud Control name translation), so the drift
// comparison against link data matches.
func specInvokeStatementNode(sid, otherFunctionARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", core.MappingNodeFromString("lambda:InvokeFunction"),
		"resource", core.MappingNodeFromString(otherFunctionARN),
	)
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
	// Intra-service link annotations use aws.lambda.invoke.* pattern
	// where "invoke" is the feature being configured
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key: fmt.Sprintf(
				"aws.lambda.invoke.%s.populateEnvVars",
				otherResourceInfo.ResourceName,
			),
			FallbackKey: "aws.lambda.invoke.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf(
				"aws.lambda.invoke.%s.envVarName",
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
