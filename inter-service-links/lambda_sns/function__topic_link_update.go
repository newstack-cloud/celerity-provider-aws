package lambdasns

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

func (l *functionTopicLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, _, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getTopicLinkAnnotations(input.ResourceInfo, input.OtherResourceInfo)
	if !annotations.populateEnvVars {
		return &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields(),
			ResourceDataMappings: map[string]string{},
		}, nil
	}

	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{LambdaFuncResourceInfo: input.ResourceInfo, LoadRoleInfo: false},
		lambdaService,
		nil,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	topicARN, hasTopicARN := extractTopicARN(input.OtherResourceInfo)
	if !hasTopicARN {
		return nil, fmt.Errorf(
			"topic ARN could not be retrieved from the linked to %q topic resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeTopicEnvVars(ctx, input, setupCtx.FunctionARN, annotations.envVarName, setupCtx.LambdaOutput, lambdaService)
	}

	return l.addTopicEnvVars(ctx, input, setupCtx.FunctionARN, topicARN, annotations.envVarName, setupCtx.LambdaOutput, lambdaService)
}

func (l *functionTopicLinkActions) addTopicEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, topicARN, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := topicEnvVarName(envVarName, input.OtherResourceInfo)
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
		map[string]string{finalEnvVarName: topicARN},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			pluginutils.GetResourceName(input.ResourceInfo),
			core.MappingNodeFields(
				"environmentVariables",
				core.MappingNodeFields(finalEnvVarName, core.MappingNodeFromString(topicARN)),
			),
		),
		ResourceDataMappings: map[string]string{dataMappingKey: linkDataFieldPath},
	}, nil
}

func (l *functionTopicLinkActions) removeTopicEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := topicEnvVarName(envVarName, input.OtherResourceInfo)
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

func (l *functionTopicLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The topic is not modified by this link; only the Lambda function and its execution
	// role are updated to allow publishing to the topic.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) the Lambda execution role permission to
// publish to the topic, then activates the SNS VPC endpoint when the function is
// VPC-isolated.
func (l *functionTopicLinkActions) UpdateIntermediaryResources(
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

	sid := createPublishSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if _, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setupCtx.RoleName,
			SID:      sid,
		}); err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}, nil
	}

	topicARN, hasTopicARN := extractTopicARN(input.ResourceBInfo)
	if !hasTopicARN {
		return nil, fmt.Errorf(
			"topic ARN could not be retrieved from the linked to %q topic resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: snsPublishStatement(sid, topicARN),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q permission to publish to SNS topic %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	output := publishAccessLinkOutput(input, setupCtx.RoleResourceName, sid, topicARN, result)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-isolated caller needs an interface VPC endpoint for SNS; this is a no-op for
	// non-VPC functions.
	return linkutils.ActivateLinkNetworking(
		ctx,
		ec2Service,
		input,
		linkutils.NetworkingActivation{
			Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
			Region:       region,
			AWSService:   "sns",
			EndpointType: ec2types.VpcEndpointTypeInterface,
		},
		output,
	)
}

func snsPublishStatement(sid, topicARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   "sns:Publish",
		"Resource": topicARN,
	}
}

func publishAccessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, topicARN string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specPublishStatementNode(sid, topicARN),
	)

	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

func specPublishStatementNode(sid, topicARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", core.MappingNodeFromString("sns:Publish"),
		"resource", core.MappingNodeFromString(topicARN),
	)
}

func extractTopicARN(topicInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(topicInfo)
	topicARNNode, has := pluginutils.GetValueByPath("$.topicArn", spec)
	if !has {
		return "", false
	}
	return core.StringValue(topicARNNode), true
}

func topicEnvVarName(userDefinedEnvVarName string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	return fmt.Sprintf("SNS_TOPIC_%s", resourceInfo.ResourceName)
}

func createPublishSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("SNSPublish%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

type topicLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
}

func getTopicLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *topicLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.sns.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.sns.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.sns.%s.envVarName", otherResourceInfo.ResourceName),
		},
	)

	return &topicLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
	}
}
