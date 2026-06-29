package lambdadynamodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
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

// UpdateIntermediaryResources grants (or revokes) the Lambda execution role access
// to the DynamoDB table by packing a single IAM statement into the role's
// allocator-managed policies. The role is an existing intermediary in the
// blueprint, so the role lock is held while its shared policy set is read-modified.
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

	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	sid := createDynamoDBAccessSID(input.ResourceBInfo)

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

	tableARN, hasTableARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasTableARN {
		return nil, fmt.Errorf(
			"table ARN could not be retrieved from the linked to %q DynamoDB table resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	annotations := getDynamoDBTableLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: dynamoDBAccessStatement(sid, tableARN, annotations.accessLevel),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q access to DynamoDB table %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	return accessLinkOutput(input, setupCtx.RoleResourceName, sid, tableARN, annotations.accessLevel, result), nil
}

// dynamoDBAccessStatement builds the IAM policy statement (canonical PascalCase, as
// the IAM API expects) granting the access level to the table.
func dynamoDBAccessStatement(sid, tableARN, accessLevel string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   dynamoDBActionsForAccessLevel(accessLevel),
		"Resource": tableARN,
	}
}

// accessLinkOutput records the granted statement in link data and, for inline
// placements, maps it onto the role's spec so the framework attributes the
// statement to this link and does not treat it as drift / strip it on redeploy.
func accessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, tableARN, accessLevel string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specAccessStatementNode(sid, tableARN, accessLevel),
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

// specAccessStatementNode builds the statement in the camelCase spec form the
// role's external state uses (after Cloud Control name translation), so the drift
// comparison against link data matches.
func specAccessStatementNode(sid, tableARN, accessLevel string) *core.MappingNode {
	actions := dynamoDBActionsForAccessLevel(accessLevel)
	actionItems := make([]*core.MappingNode, len(actions))
	for i, action := range actions {
		actionItems[i] = core.MappingNodeFromString(action)
	}
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{Items: actionItems},
		"resource", core.MappingNodeFromString(tableARN),
	)
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
