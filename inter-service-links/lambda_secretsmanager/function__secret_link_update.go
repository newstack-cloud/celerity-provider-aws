package lambdasecretsmanager

import (
	"context"
	"errors"
	"fmt"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionSecretLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, _, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getSecretLinkAnnotations(input.ResourceInfo, input.OtherResourceInfo)
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

	secretARN, hasSecretARN := extractSecretARN(input.OtherResourceInfo)
	if !hasSecretARN {
		return nil, fmt.Errorf(
			"secret ARN could not be retrieved from the linked to %q Secrets Manager secret resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeSecretEnvVars(
			ctx, input, setupCtx.FunctionARN,
			annotations.envVarName, setupCtx.LambdaOutput, lambdaService,
		)
	}

	return l.addSecretEnvVars(
		ctx, input, setupCtx.FunctionARN, secretARN, annotations.envVarName,
		setupCtx.LambdaOutput, lambdaService,
	)
}

func (l *functionSecretLinkActions) addSecretEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, secretARN, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := secretEnvVarName(envVarName, input.OtherResourceInfo)
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
		map[string]string{finalEnvVarName: secretARN},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			pluginutils.GetResourceName(input.ResourceInfo),
			core.MappingNodeFields(
				"environmentVariables",
				core.MappingNodeFields(finalEnvVarName, core.MappingNodeFromString(secretARN)),
			),
		),
		ResourceDataMappings: map[string]string{dataMappingKey: linkDataFieldPath},
	}, nil
}

func (l *functionSecretLinkActions) removeSecretEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := secretEnvVarName(envVarName, input.OtherResourceInfo)
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

func (l *functionSecretLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The secret is not modified by this link; only the Lambda function and its execution
	// role are updated to allow it to access the secret.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) the Lambda execution role access to the
// Secrets Manager secret, then activates the Secrets Manager interface VPC endpoint when the
// function is VPC-isolated.
func (l *functionSecretLinkActions) UpdateIntermediaryResources(
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

	sid := createSecretAccessSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if _, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setupCtx.RoleName,
			SID:      sid,
		}); err != nil {
			return nil, err
		}
		// The endpoint this link provisioned is removed here; returning early would
		// leave it, and its security group, behind.
		ec2Service, err := l.getEC2Service(ctx, providerCtx)
		if err != nil {
			return nil, err
		}
		return linkutils.ReconcileLinkNetworking(
			ctx,
			ec2Service,
			input,
			secretsManagerNetworkingActivation(setupCtx, region),
			&provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
		)
	}

	secretARN, hasSecretARN := extractSecretARN(input.ResourceBInfo)
	if !hasSecretARN {
		return nil, fmt.Errorf(
			"secret ARN could not be retrieved from the linked to %q Secrets Manager secret resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	annotations := getSecretLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: secretAccessStatement(sid, secretARN, annotations.accessLevel),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q access to Secrets Manager secret %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	output := accessLinkOutput(input, setupCtx.RoleResourceName, sid, secretARN, annotations.accessLevel, result)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-isolated caller reaches Secrets Manager through an interface VPC endpoint; this
	// is a no-op for non-VPC functions.
	return linkutils.ReconcileLinkNetworking(
		ctx,
		ec2Service,
		input,
		secretsManagerNetworkingActivation(setupCtx, region),
		output,
	)
}

func secretAccessStatement(sid, secretARN, accessLevel string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   secretActionsForAccessLevel(accessLevel),
		"Resource": []string{secretARN},
	}
}

func accessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, secretARN, accessLevel string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specSecretAccessStatementNode(sid, secretARN, accessLevel),
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

func specSecretAccessStatementNode(sid, secretARN, accessLevel string) *core.MappingNode {
	actions := secretActionsForAccessLevel(accessLevel)
	actionItems := make([]*core.MappingNode, len(actions))
	for i, action := range actions {
		actionItems[i] = core.MappingNodeFromString(action)
	}
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{Items: actionItems},
		"resource", &core.MappingNode{Items: []*core.MappingNode{
			core.MappingNodeFromString(secretARN),
		}},
	)
}

func extractSecretARN(secretInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(secretInfo)
	// The Secrets Manager secret's primary identifier (its "id" field) is the secret ARN.
	arnNode, has := pluginutils.GetValueByPath("$.id", spec)
	if !has {
		return "", false
	}
	return core.StringValue(arnNode), true
}

func secretEnvVarName(userDefinedEnvVarName string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	return fmt.Sprintf("SECRET_%s", resourceInfo.ResourceName)
}

func createSecretAccessSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("SecretsManagerAccess%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

func secretActionsForAccessLevel(accessLevel string) []string {
	switch accessLevel {
	case "write":
		return []string{"secretsmanager:PutSecretValue", "secretsmanager:UpdateSecret"}
	case "readwrite":
		return []string{
			"secretsmanager:GetSecretValue",
			"secretsmanager:DescribeSecret",
			"secretsmanager:PutSecretValue",
			"secretsmanager:UpdateSecret",
		}
	case "read":
		fallthrough
	default:
		return []string{"secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"}
	}
}

type secretLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
	accessLevel     string
}

func getSecretLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *secretLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.secretsmanager.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.secretsmanager.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.secretsmanager.%s.envVarName", otherResourceInfo.ResourceName),
		},
	)

	accessLevel, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.secretsmanager.%s.accessLevel", otherResourceInfo.ResourceName),
			Default: "read",
		},
	)

	return &secretLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
		accessLevel:     accessLevel,
	}
}

// Shared by the create and destroy paths so a teardown removes exactly what the create
// path provisioned.
//
// Destroy used to return before reaching the activation, so the VPC endpoint and its
// security group were left behind on every teardown. That group's ingress rule
// references the caller's group, which then blocks the caller's group, and with it the
// whole VPC, from being deleted.
func secretsManagerNetworkingActivation(
	setupCtx *linkutils.LambdaLinkSetupContext,
	region string,
) linkutils.NetworkingActivation {
	return linkutils.NetworkingActivation{
		Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
		Region:       region,
		AWSService:   "secretsmanager",
		EndpointType: ec2types.VpcEndpointTypeInterface,
	}
}
