package lambdadynamodb

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

func (l *lambdaFunctionDynamoDBTableLinkActions) UpdateResourceA(
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

	annotations := getDynamoDBTableLinkAnnotations(
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
		nil,
		provider.NewProviderContextFromLinkContext(
			input.LinkContext,
			"aws",
		),
	)
	if err != nil {
		return nil, err
	}

	tableName, hasTableName := extractTableNameFromResourceInfo(input.OtherResourceInfo)
	if !hasTableName {
		return nil, fmt.Errorf(
			"table name could not be retrieved from the linked to %q DynamoDB table resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeFunctionEnvVars(
			ctx,
			input,
			setupCtx.FunctionARN,
			annotations.envVarName,
			setupCtx.LambdaOutput,
			lambdaService,
		)
	}

	return l.addFunctionEnvVars(
		ctx,
		input,
		setupCtx.FunctionARN,
		tableName,
		annotations.envVarName,
		setupCtx.LambdaOutput,
		lambdaService,
	)
}

func (l *lambdaFunctionDynamoDBTableLinkActions) addFunctionEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	tableName string,
	envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := dynamoDBTableEnvVarName(
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
			finalEnvVarName: tableName,
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
					finalEnvVarName,
					core.MappingNodeFromString(tableName),
				),
			),
		),
		ResourceDataMappings: map[string]string{
			dataMappingKey: linkDataFieldPath,
		},
	}, nil
}

func (l *lambdaFunctionDynamoDBTableLinkActions) removeFunctionEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := dynamoDBTableEnvVarName(
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

func (l *lambdaFunctionDynamoDBTableLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The DynamoDB table is not updated as a part of the link update,
	// the link only requires updates to the linked from lambda function
	// and its execution role that will allow it to access the DynamoDB table.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *lambdaFunctionDynamoDBTableLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(
		input.LinkContext,
		"aws",
	)
	lambdaService, err := l.getLambdaService(ctx, providerCtx)
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

	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	rolePolicies, err := iamService.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: aws.String(setupCtx.RoleName),
	})
	if err != nil {
		return nil, err
	}

	tableARN, hasTableARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasTableARN {
		return nil, fmt.Errorf(
			"table ARN could not be retrieved from the linked to %q DynamoDB table resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	annotations := getDynamoDBTableLinkAnnotations(
		input.ResourceAInfo,
		input.ResourceBInfo,
	)
	sid := createDynamoDBAccessSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeRolePolicyPermissions(
			ctx,
			input,
			setupCtx,
			rolePolicies,
			iamService,
		)
	}

	return l.setRolePolicyPermissions(
		ctx,
		input,
		tableARN,
		sid,
		annotations.accessLevel,
		setupCtx,
		rolePolicies,
		iamService,
	)
}

func (l *lambdaFunctionDynamoDBTableLinkActions) setRolePolicyPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	tableARN string,
	sid string,
	accessLevel string,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	actions := dynamoDBActionsForAccessLevel(accessLevel)
	permission := map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   actions,
		"Resource": tableARN,
	}

	if len(rolePolicies.PolicyNames) > 0 {
		return l.updateExistingRolePolicy(
			ctx,
			input,
			&dynamoDBRoleInfo{
				roleName:          setupCtx.RoleName,
				roleResourceState: setupCtx.RoleResourceState,
				rolePolicies:      rolePolicies,
				permission:        permission,
				sid:               sid,
			},
			iamService,
		)
	}

	return l.addNewRolePolicy(
		ctx,
		input,
		&dynamoDBRoleInfo{
			roleName:          setupCtx.RoleName,
			roleResourceState: setupCtx.RoleResourceState,
			rolePolicies:      rolePolicies,
			permission:        permission,
			sid:               sid,
		},
		iamService,
	)
}

func (l *lambdaFunctionDynamoDBTableLinkActions) addNewRolePolicy(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleData *dynamoDBRoleInfo,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	policyName, err := linkutils.CreateNewRolePolicy(
		ctx,
		iamService,
		input.InstanceName,
		roleData.roleName,
		[]map[string]any{roleData.permission},
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

	permissionNode, err := pluginutils.AnyToMappingNode(roleData.permission)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: core.MappingNodeFields(
			linkDataExecRoleName,
			core.MappingNodeFields(
				linkutils.PermissionFieldName,
				permissionNode,
				linkutils.PolicyNameFieldName,
				policyName,
			),
		),
		ResourceDataMappings: map[string]string{
			roleResourcePath: linkDataFieldPath,
		},
	}, nil
}

type dynamoDBRoleInfo struct {
	roleName          string
	roleResourceState *state.ResourceState
	rolePolicies      *iam.ListRolePoliciesOutput
	permission        map[string]any
	sid               string
}

func (l *lambdaFunctionDynamoDBTableLinkActions) updateExistingRolePolicy(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleData *dynamoDBRoleInfo,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	executionRoleName := createLinkDataExecutionRoleName(input.ResourceAInfo)
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
		[]map[string]any{roleData.permission},
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

	permissionNode, err := pluginutils.AnyToMappingNode(roleData.permission)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: core.MappingNodeFields(
			executionRoleName,
			core.MappingNodeFields(
				linkutils.PermissionFieldName,
				permissionNode,
				linkutils.PolicyNameFieldName,
				policyName,
			),
		),
		ResourceDataMappings: map[string]string{
			roleResourcePath: linkDataFieldPath,
		},
	}, nil
}

func (l *lambdaFunctionDynamoDBTableLinkActions) removeRolePolicyPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	executionRoleName := createLinkDataExecutionRoleName(input.ResourceAInfo)

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

func dynamoDBTableEnvVarName(
	userDefinedEnvVarName string,
	resourceInfo *provider.ResourceInfo,
) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}

	return fmt.Sprintf(
		"DYNAMODB_TABLE_%s",
		resourceInfo.ResourceName,
	)
}

func createDynamoDBAccessSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"DynamoDBAccess%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

type dynamoDBTableLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
	accessLevel     string
}

func getDynamoDBTableLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *dynamoDBTableLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key: fmt.Sprintf(
				"aws.lambda.dynamodb.%s.populateEnvVars",
				otherResourceInfo.ResourceName,
			),
			FallbackKey: "aws.lambda.dynamodb.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf(
				"aws.lambda.dynamodb.%s.envVarName",
				otherResourceInfo.ResourceName,
			),
		},
	)

	accessLevel, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf(
				"aws.lambda.dynamodb.%s.accessLevel",
				otherResourceInfo.ResourceName,
			),
			Default: "readwrite",
		},
	)

	return &dynamoDBTableLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
		accessLevel:     accessLevel,
	}
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sExecutionRole",
		resourceInfo.ResourceName,
	)
}

func extractTableNameFromResourceInfo(resourceInfo *provider.ResourceInfo) (string, bool) {
	tableName, hasTableName := pluginutils.GetValueByPath(
		"$.tableName",
		resourceInfo.CurrentResourceState.SpecData,
	)
	if !hasTableName {
		return "", false
	}
	return core.StringValue(tableName), true
}

func dynamoDBActionsForAccessLevel(accessLevel string) []string {
	switch accessLevel {
	case "read":
		return []string{
			"dynamodb:GetItem",
			"dynamodb:Query",
			"dynamodb:Scan",
			"dynamodb:BatchGetItem",
		}
	case "write":
		return []string{
			"dynamodb:PutItem",
			"dynamodb:UpdateItem",
			"dynamodb:DeleteItem",
			"dynamodb:BatchWriteItem",
		}
	case "readwrite":
		fallthrough
	default:
		return []string{
			"dynamodb:GetItem",
			"dynamodb:Query",
			"dynamodb:Scan",
			"dynamodb:BatchGetItem",
			"dynamodb:PutItem",
			"dynamodb:UpdateItem",
			"dynamodb:DeleteItem",
			"dynamodb:BatchWriteItem",
		}
	}
}
