package lambdakms

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionKeyLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, _, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getKeyLinkAnnotations(input.ResourceInfo, input.OtherResourceInfo)
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
		nil,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	keyARN, hasKeyARN := extractKeyARN(input.OtherResourceInfo)
	if !hasKeyARN {
		return nil, fmt.Errorf(
			"key ARN could not be retrieved from the linked to %q KMS key resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeKeyEnvVars(
			ctx, input, setupCtx.FunctionARN, annotations.envVarName,
			setupCtx.LambdaOutput, lambdaService,
		)
	}

	return l.addKeyEnvVars(
		ctx, input, setupCtx.FunctionARN, keyARN, annotations.envVarName,
		setupCtx.LambdaOutput, lambdaService,
	)
}

func (l *functionKeyLinkActions) addKeyEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, keyARN, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := keyEnvVarName(envVarName, input.OtherResourceInfo)
	dataMappingKey := fmt.Sprintf(
		"%s::spec.environment.variables[\"%s\"]",
		input.ResourceInfo.ResourceName, finalEnvVarName,
	)
	linkDataFieldPath := fmt.Sprintf(
		"%s.environmentVariables[\"%s\"]",
		input.ResourceInfo.ResourceName, finalEnvVarName,
	)

	err := linkutils.UpdateLambdaEnvironmentVariables(
		ctx, lambdaService, functionARN, currentFunctionConfig,
		map[string]string{finalEnvVarName: keyARN},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			pluginutils.GetResourceName(input.ResourceInfo),
			core.MappingNodeFields(
				"environmentVariables",
				core.MappingNodeFields(finalEnvVarName, core.MappingNodeFromString(keyARN)),
			),
		),
		ResourceDataMappings: map[string]string{dataMappingKey: linkDataFieldPath},
	}, nil
}

func (l *functionKeyLinkActions) removeKeyEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := keyEnvVarName(envVarName, input.OtherResourceInfo)
	err := linkutils.RemoveLambdaEnvironmentVariables(
		ctx, lambdaService, functionARN, currentFunctionConfig, []string{finalEnvVarName},
	)
	if err != nil {
		return nil, err
	}
	return &provider.LinkUpdateResourceOutput{
		LinkData:             core.MappingNodeFields(pluginutils.GetResourceName(input.ResourceInfo), core.MappingNodeFields()),
		ResourceDataMappings: map[string]string{},
	}, nil
}

func (l *functionKeyLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The KMS key is not modified by this link; only the Lambda function and its execution
	// role are updated to allow it to use the key.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) the Lambda execution role use
// of the KMS key, then activates the KMS interface VPC endpoint when the function is
// VPC-isolated.
func (l *functionKeyLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, region, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
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

	// The execution role is shared by every link that grants it access, so lock it for the
	// read-modify-write of its policy set.
	err = input.ResourceService.AcquireResourceLock(ctx, &provider.AcquireResourceLockInput{
		InstanceID:      pluginutils.GetInstanceID(input.ResourceAInfo),
		ResourceName:    setupCtx.RoleResourceName,
		ProviderContext: providerCtx,
		AcquiredBy:      input.LinkID,
	})
	if err != nil {
		return nil, err
	}

	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	sid := createKeyAccessSID(input.ResourceBInfo)
	annotations := getKeyLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
	granteeRoleARN := aws.ToString(setupCtx.LambdaOutput.Role)
	grantName := keyGrantName(input.ResourceAInfo, input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if _, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setupCtx.RoleName,
			SID:      sid,
		}); err != nil {
			return nil, err
		}
		// Revoke any KMS grant this link created for the execution role.
		if keyARN, ok := extractKeyARN(input.ResourceBInfo); ok {
			if err := l.reconcileKeyGrant(ctx, providerCtx, keyGrantReconcile{
				granteeRoleARN: granteeRoleARN,
				keyID:          keyARN,
				grantName:      grantName,
				accessLevel:    annotations.accessLevel,
				manageGrant:    false,
				isDestroy:      true,
			}); err != nil {
				return nil, err
			}
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	keyARN, hasKeyARN := extractKeyARN(input.ResourceBInfo)
	if !hasKeyARN {
		return nil, fmt.Errorf(
			"key ARN could not be retrieved from the linked to %q KMS key resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: keyAccessStatement(sid, keyARN, annotations.accessLevel),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q access to KMS key %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	// Ensure the key-side authorisation matches the manageKeyGrant annotation: create/update
	// the KMS grant when enabled, or revoke a previously-created one when disabled.
	if err := l.reconcileKeyGrant(ctx, providerCtx, keyGrantReconcile{
		granteeRoleARN: granteeRoleARN,
		keyID:          keyARN,
		grantName:      grantName,
		accessLevel:    annotations.accessLevel,
		manageGrant:    annotations.manageKeyGrant,
		isDestroy:      false,
	}); err != nil {
		return nil, err
	}

	output := accessLinkOutput(input, setupCtx.RoleResourceName, sid, keyARN, annotations.accessLevel, result)

	// Record the managed KMS grant in link data so it is tracked in state and surfaced in
	// staged changes. When unmanaged, the field is absent, which reflects removal on toggle-off.
	if annotations.manageKeyGrant {
		output.LinkData.Fields[keyGrantLinkDataField] = keyGrantLinkDataNode(
			input.ResourceAInfo, input.ResourceBInfo, annotations.accessLevel,
		)
	}

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-isolated caller reaches KMS through an interface VPC endpoint; this is a no-op
	// for non-VPC functions.
	return linkutils.ActivateLinkNetworking(
		ctx,
		ec2Service,
		input,
		linkutils.NetworkingActivation{
			Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
			Region:       region,
			AWSService:   "kms",
			EndpointType: ec2types.VpcEndpointTypeInterface,
		},
		output,
	)
}

func keyAccessStatement(sid, keyARN, accessLevel string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   keyActionsForAccessLevel(accessLevel),
		"Resource": []string{keyARN},
	}
}

func accessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, keyARN, accessLevel string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specKeyAccessStatementNode(sid, keyARN, accessLevel),
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it: inline
	// placements map the statement by Sid; managed (overflow) placements map the attached
	// managed policy ARN.
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(
		mappings,
		roleLinkData,
		roleResourceName,
		linkDataKey,
		sid,
		result,
	)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

func specKeyAccessStatementNode(sid, keyARN, accessLevel string) *core.MappingNode {
	actions := keyActionsForAccessLevel(accessLevel)
	actionItems := make([]*core.MappingNode, len(actions))
	for i, action := range actions {
		actionItems[i] = core.MappingNodeFromString(action)
	}
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{Items: actionItems},
		"resource", &core.MappingNode{Items: []*core.MappingNode{
			core.MappingNodeFromString(keyARN),
		}},
	)
}

func extractKeyARN(keyInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(keyInfo)
	arnNode, has := pluginutils.GetValueByPath("$.arn", spec)
	if !has {
		return "", false
	}
	return core.StringValue(arnNode), true
}

func keyEnvVarName(userDefinedEnvVarName string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	return fmt.Sprintf("KMS_KEY_%s", resourceInfo.ResourceName)
}

func createKeyAccessSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("KMSAccess%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

func keyActionsForAccessLevel(accessLevel string) []string {
	switch accessLevel {
	case "encryptDecrypt":
		return []string{
			"kms:Decrypt",
			"kms:DescribeKey",
			"kms:Encrypt",
			"kms:GenerateDataKey",
			"kms:GenerateDataKeyWithoutPlaintext",
		}
	case "decrypt":
		fallthrough
	default:
		return []string{"kms:Decrypt", "kms:DescribeKey"}
	}
}

type keyLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
	accessLevel     string
	manageKeyGrant  bool
}

func getKeyLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *keyLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.kms.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.kms.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.kms.%s.envVarName", otherResourceInfo.ResourceName),
		},
	)

	accessLevel, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.kms.%s.accessLevel", otherResourceInfo.ResourceName),
			Default: "decrypt",
		},
	)

	manageKeyGrant, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     fmt.Sprintf("aws.lambda.kms.%s.manageKeyGrant", otherResourceInfo.ResourceName),
			Default: false,
		},
	)

	return &keyLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
		accessLevel:     accessLevel,
		manageKeyGrant:  manageKeyGrant,
	}
}
