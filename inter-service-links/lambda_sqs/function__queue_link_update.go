package lambdasqs

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

func (l *functionQueueLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, _, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getQueueLinkAnnotations(input.ResourceInfo, input.OtherResourceInfo)
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

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeQueueEnvVars(ctx, input, setupCtx.FunctionARN, annotations.envVarName, setupCtx.LambdaOutput, lambdaService)
	}

	queueURL, hasQueueURL := extractQueueURL(input.OtherResourceInfo)
	if !hasQueueURL {
		return nil, fmt.Errorf(
			"queue URL could not be retrieved from the linked to %q queue resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	return l.addQueueEnvVars(ctx, input, setupCtx.FunctionARN, queueURL, annotations.envVarName, setupCtx.LambdaOutput, lambdaService)
}

func (l *functionQueueLinkActions) addQueueEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, queueURL, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := queueEnvVarName(envVarName, input.OtherResourceInfo)
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
		map[string]string{finalEnvVarName: queueURL},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			pluginutils.GetResourceName(input.ResourceInfo),
			core.MappingNodeFields(
				"environmentVariables",
				core.MappingNodeFields(finalEnvVarName, core.MappingNodeFromString(queueURL)),
			),
		),
		ResourceDataMappings: map[string]string{dataMappingKey: linkDataFieldPath},
	}, nil
}

func (l *functionQueueLinkActions) removeQueueEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := queueEnvVarName(envVarName, input.OtherResourceInfo)
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

func (l *functionQueueLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The queue is not modified by this link; only the Lambda function and its execution role are
	// updated to allow sending to (and optionally receiving from) the queue.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) the Lambda execution role permission to access
// the queue, then activates the SQS VPC endpoint when the function is VPC-isolated.
func (l *functionQueueLinkActions) UpdateIntermediaryResources(
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
		&linkutils.LambdaLinkSetupData{LambdaFuncResourceInfo: input.ResourceAInfo, LoadRoleInfo: true},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

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

	sid := createAccessSID(input.ResourceBInfo)

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
			sqsNetworkingActivation(setupCtx, region),
			&provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
		)
	}

	queueARN, hasQueueARN := extractQueueARN(input.ResourceBInfo)
	if !hasQueueARN {
		return nil, fmt.Errorf(
			"queue ARN could not be retrieved from the linked to %q queue resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	annotations := getQueueLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
	actions := sqsActionsForAccessLevel(annotations.accessLevel)

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: sqsAccessStatement(sid, queueARN, actions),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q permission to access SQS queue %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	output := accessLinkOutput(input, setupCtx.RoleResourceName, sid, queueARN, actions, result)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-isolated caller needs an interface VPC endpoint for SQS; this is a no-op for
	// non-VPC functions.
	return linkutils.ReconcileLinkNetworking(
		ctx,
		ec2Service,
		input,
		sqsNetworkingActivation(setupCtx, region),
		output,
	)
}

// Returns the SQS actions granted for an access level. GetQueueUrl and
// GetQueueAttributes are always included so the SDK can resolve/inspect the queue.
func sqsActionsForAccessLevel(accessLevel string) []string {
	base := []string{"sqs:GetQueueUrl", "sqs:GetQueueAttributes"}
	switch accessLevel {
	case "receive":
		return append([]string{"sqs:ReceiveMessage", "sqs:DeleteMessage"}, base...)
	case "sendReceive":
		return append([]string{"sqs:SendMessage", "sqs:ReceiveMessage", "sqs:DeleteMessage"}, base...)
	default: // send
		return append([]string{"sqs:SendMessage"}, base...)
	}
}

func sqsAccessStatement(sid, queueARN string, actions []string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   actions,
		"Resource": queueARN,
	}
}

func accessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, queueARN string,
	actions []string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	actionItems := make([]*core.MappingNode, 0, len(actions))
	for _, action := range actions {
		actionItems = append(actionItems, core.MappingNodeFromString(action))
	}
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		core.MappingNodeFields(
			"sid", core.MappingNodeFromString(sid),
			"effect", core.MappingNodeFromString("Allow"),
			"action", &core.MappingNode{Items: actionItems},
			"resource", core.MappingNodeFromString(queueARN),
		),
	)

	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

func extractQueueURL(queueInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(queueInfo)
	node, has := pluginutils.GetValueByPath("$.queueUrl", spec)
	if !has || node == nil || core.StringValue(node) == "" {
		return "", false
	}
	return core.StringValue(node), true
}

func extractQueueARN(queueInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(queueInfo)
	node, has := pluginutils.GetValueByPath("$.arn", spec)
	if !has || node == nil || core.StringValue(node) == "" {
		return "", false
	}
	return core.StringValue(node), true
}

func queueEnvVarName(userDefinedEnvVarName string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	return fmt.Sprintf("SQS_QUEUE_%s", resourceInfo.ResourceName)
}

func createAccessSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("SQSAccess%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

type queueLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
	accessLevel     string
}

func getQueueLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *queueLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.sqs.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.sqs.populateEnvVars",
			Default:     true,
		},
	)
	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.sqs.%s.envVarName", otherResourceInfo.ResourceName),
		},
	)
	accessLevel, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.sqs.%s.accessLevel", otherResourceInfo.ResourceName),
			Default: "send",
		},
	)

	return &queueLinkAnnotations{
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
func sqsNetworkingActivation(
	setupCtx *linkutils.LambdaLinkSetupContext,
	region string,
) linkutils.NetworkingActivation {
	return linkutils.NetworkingActivation{
		Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
		Region:       region,
		AWSService:   "sqs",
		EndpointType: ec2types.VpcEndpointTypeInterface,
	}
}
