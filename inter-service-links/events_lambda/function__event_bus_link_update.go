package eventslambda

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

func (l *functionEventBusLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	lambdaService, err := l.getLambdaService(
		ctx,
		provider.NewProviderContextFromLinkContext(input.LinkContext, "aws"),
	)
	if err != nil {
		return nil, err
	}

	annotations := getEventBusLinkAnnotations(input.ResourceInfo, input.OtherResourceInfo)
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
		provider.NewProviderContextFromLinkContext(input.LinkContext, "aws"),
	)
	if err != nil {
		return nil, err
	}

	eventBusName, hasEventBusName := extractEventBusNameFromResourceInfo(input.OtherResourceInfo)
	if !hasEventBusName {
		return nil, fmt.Errorf(
			"event bus name could not be retrieved from the linked to %q event bus resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeFunctionEnvVars(
			ctx, input, setupCtx.FunctionARN, annotations.envVarName, setupCtx.LambdaOutput, lambdaService,
		)
	}

	return l.addFunctionEnvVars(
		ctx, input, setupCtx.FunctionARN, eventBusName, annotations.envVarName, setupCtx.LambdaOutput, lambdaService,
	)
}

func (l *functionEventBusLinkActions) addFunctionEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	eventBusName string,
	envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := eventBusEnvVarName(envVarName, input.OtherResourceInfo)
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
		map[string]string{finalEnvVarName: eventBusName},
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
					core.MappingNodeFromString(eventBusName),
				),
			),
		),
		ResourceDataMappings: map[string]string{
			dataMappingKey: linkDataFieldPath,
		},
	}, nil
}

func (l *functionEventBusLinkActions) removeFunctionEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := eventBusEnvVarName(envVarName, input.OtherResourceInfo)

	err := linkutils.RemoveLambdaEnvironmentVariables(
		ctx, lambdaService, functionARN, currentFunctionConfig, []string{finalEnvVarName},
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

func (l *functionEventBusLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The event bus is not modified by the link; only the Lambda function and its
	// execution role are updated to allow publishing events to the bus.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) the Lambda execution role
// permission to publish events to the event bus by packing a single IAM statement
// into the role's allocator-managed policies. The role is an existing intermediary
// in the blueprint, so the role lock is held while its shared policy set is
// read-modified.
func (l *functionEventBusLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
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

	sid := createPutEventsSID(input.ResourceBInfo)

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

	eventBusARN, hasEventBusARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasEventBusARN {
		return nil, fmt.Errorf(
			"event bus ARN could not be retrieved from the linked to %q event bus resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: putEventsStatement(sid, eventBusARN),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q permission to publish events to event bus %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	return putEventsLinkOutput(input, setupCtx.RoleResourceName, sid, eventBusARN, result), nil
}

func putEventsStatement(sid, eventBusARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   []any{"events:PutEvents"},
		"Resource": eventBusARN,
	}
}

// Records the granted statement in link data and, for inline
// placements, maps it onto the role's spec so the framework attributes the
// statement to this link and does not treat it as drift / strip it on redeploy.
func putEventsLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, eventBusARN string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specPutEventsStatementNode(sid, eventBusARN),
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it:
	// inline placements map the statement by Sid. Managed (overflow) placements map
	// the attached managed policy ARN.
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

func specPutEventsStatementNode(sid, eventBusARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{Items: []*core.MappingNode{
			core.MappingNodeFromString("events:PutEvents"),
		}},
		"resource", core.MappingNodeFromString(eventBusARN),
	)
}

func eventBusEnvVarName(
	userDefinedEnvVarName string,
	resourceInfo *provider.ResourceInfo,
) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	return fmt.Sprintf("EVENT_BUS_%s", resourceInfo.ResourceName)
}

func createPutEventsSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"EventBridgePutEvents%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

type eventBusLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
}

func getEventBusLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *eventBusLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key: fmt.Sprintf(
				"aws.lambda.events.%s.populateEnvVars",
				otherResourceInfo.ResourceName,
			),
			FallbackKey: "aws.lambda.events.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf(
				"aws.lambda.events.%s.envVarName",
				otherResourceInfo.ResourceName,
			),
		},
	)

	return &eventBusLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
	}
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

func extractEventBusNameFromResourceInfo(resourceInfo *provider.ResourceInfo) (string, bool) {
	eventBusName, hasEventBusName := pluginutils.GetValueByPath(
		"$.name",
		resourceInfo.CurrentResourceState.SpecData,
	)
	if !hasEventBusName {
		return "", false
	}
	return core.StringValue(eventBusName), true
}
