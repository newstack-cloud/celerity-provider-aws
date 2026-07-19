package sqslambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
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

func (l *sqsQueueLambdaFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// SQS has no source-side stream to enable (unlike DynamoDB), so the queue is not
	// modified by this link. The queue's ARN is read directly from its state when the
	// event source mapping is created.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *sqsQueueLambdaFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The Lambda function is not modified as part of this link.
	// The function's execution role is granted queue consume permissions as a
	// link-owned change in UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *sqsQueueLambdaFunctionLinkActions) UpdateIntermediaryResources(
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

	// Get the queue ARN from the SQS queue
	queueARN, hasQueueARN := extractQueueARNFromResourceInfo(input.ResourceAInfo)
	if !hasQueueARN {
		return nil, fmt.Errorf(
			"queue ARN could not be retrieved from the SQS queue %q",
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}

	// Get the function ARN from the Lambda function
	functionARN, hasFunctionARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the Lambda function %q",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	// Queue trigger annotations are on resourceB (Lambda function) because they configure
	// the event source mapping for this specific function. A queue can trigger multiple
	// functions with different configurations.
	annotations := getQueueTriggerAnnotations(input.ResourceBInfo)

	// Set up Lambda link context to get role info for queue consume permissions
	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceBInfo,
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
			InstanceID:      pluginutils.GetInstanceID(input.ResourceBInfo),
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

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.destroyIntermediaryResources(
			ctx,
			input,
			setupCtx,
			lambdaService,
			iamService,
		)
	}

	// Check if we have an existing event source mapping UUID in the link state
	esmResourceID := queueFunctionESMResourceID(input.ResourceAInfo, input.ResourceBInfo)
	existingUUID := getExistingEventSourceMappingUUID(input.CurrentLinkState, esmResourceID)

	if existingUUID != "" {
		return l.updateIntermediaryResources(
			ctx,
			input,
			existingUUID,
			queueARN,
			functionARN,
			annotations,
			setupCtx,
			lambdaService,
			iamService,
		)
	}

	return l.createIntermediaryResources(
		ctx,
		input,
		queueARN,
		functionARN,
		annotations,
		setupCtx,
		lambdaService,
		iamService,
	)
}

func (l *sqsQueueLambdaFunctionLinkActions) createIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	queueARN string,
	functionARN string,
	annotations *queueTriggerAnnotations,
	setupCtx *linkutils.LambdaLinkSetupContext,
	lambdaService lambdaservice.Service,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Grant queue consume permissions to the Lambda execution role before creating the
	// event source mapping: CreateEventSourceMapping validates that the role can consume
	// from the queue, so the grant must be in place first.
	permOutput, err := l.grantConsumePermissions(
		ctx,
		input,
		queueARN,
		setupCtx,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	// Create the Event Source Mapping.
	esmOutput, err := l.createEventSourceMapping(ctx, queueARN, functionARN, annotations, lambdaService)
	if err != nil {
		return nil, err
	}

	esmIdentity := queueFunctionESMIdentity(input.ResourceAInfo, input.ResourceBInfo)
	return mergeIntermediaryOutputs(esmIdentity, esmOutput, permOutput), nil
}

func (l *sqsQueueLambdaFunctionLinkActions) updateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	existingUUID string,
	queueARN string,
	functionARN string,
	annotations *queueTriggerAnnotations,
	setupCtx *linkutils.LambdaLinkSetupContext,
	lambdaService lambdaservice.Service,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	esmOutput, err := l.updateEventSourceMapping(
		ctx,
		existingUUID,
		queueARN,
		functionARN,
		annotations,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	permOutput, err := l.grantConsumePermissions(
		ctx,
		input,
		queueARN,
		setupCtx,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	esmIdentity := queueFunctionESMIdentity(input.ResourceAInfo, input.ResourceBInfo)
	return mergeIntermediaryOutputs(esmIdentity, esmOutput, permOutput), nil
}

func (l *sqsQueueLambdaFunctionLinkActions) destroyIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	lambdaService lambdaservice.Service,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	esmResourceID := queueFunctionESMResourceID(input.ResourceAInfo, input.ResourceBInfo)
	uuid := getExistingEventSourceMappingUUID(input.CurrentLinkState, esmResourceID)
	if uuid != "" {
		// ESM deletion transiently fails with ResourceInUseException while the
		// mapping is processing or mid-update. Retry in-call with backoff
		// (~2 minutes, mirroring CloudFormation), then fall back to the
		// engine's link-update retries.
		deleteESM := linkutils.RetryOnESMInUse(
			func(ctx context.Context, in *lambda.DeleteEventSourceMappingInput) (*lambda.DeleteEventSourceMappingOutput, error) {
				return lambdaService.DeleteEventSourceMapping(ctx, in)
			},
		)
		_, err := deleteESM(ctx, &lambda.DeleteEventSourceMappingInput{
			UUID: aws.String(uuid),
		})
		if err != nil {
			return nil, fmt.Errorf(
				"failed to delete event source mapping %q: %w",
				uuid,
				err,
			)
		}
	}

	// Remove the queue-consume grant from the role's allocator-managed policy set.
	sid := createSQSQueueSID(input.ResourceAInfo)
	_, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName: setupCtx.RoleName,
		SID:      sid,
	})
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData:                   core.MappingNodeFields(),
	}, nil
}

func (l *sqsQueueLambdaFunctionLinkActions) createEventSourceMapping(
	ctx context.Context,
	queueARN string,
	functionARN string,
	annotations *queueTriggerAnnotations,
	lambdaService lambdaservice.Service,
) (*esmLinkData, error) {
	createInput := &lambda.CreateEventSourceMappingInput{
		EventSourceArn: aws.String(queueARN),
		FunctionName:   aws.String(functionARN),
		Enabled:        aws.Bool(annotations.enabled),
	}

	if annotations.batchSize > 0 {
		createInput.BatchSize = aws.Int32(int32(annotations.batchSize))
	}
	if annotations.maximumBatchingWindowInSeconds > 0 {
		createInput.MaximumBatchingWindowInSeconds = aws.Int32(int32(annotations.maximumBatchingWindowInSeconds))
	}
	if annotations.maximumConcurrency >= 2 {
		createInput.ScalingConfig = &types.ScalingConfig{
			MaximumConcurrency: aws.Int32(int32(annotations.maximumConcurrency)),
		}
	}
	if annotations.reportBatchItemFailures {
		createInput.FunctionResponseTypes = []types.FunctionResponseType{
			types.FunctionResponseTypeReportBatchItemFailures,
		}
	}
	if len(annotations.filterPatterns) > 0 {
		createInput.FilterCriteria = buildFilterCriteria(annotations.filterPatterns)
	}

	// The execution role's queue-consume grant is applied immediately before this call,
	// but IAM is eventually consistent: CreateEventSourceMapping can transiently fail
	// validation until the policy propagates. Retry in-call with backoff (~2 minutes,
	// mirroring CloudFormation), then fall back to the engine's link-update retries.
	createESM := linkutils.RetryOnIAMPropagation(
		func(ctx context.Context, in *lambda.CreateEventSourceMappingInput) (*lambda.CreateEventSourceMappingOutput, error) {
			return lambdaService.CreateEventSourceMapping(ctx, in)
		},
	)

	output, err := createESM(ctx, createInput)
	if err != nil {
		return nil, err
	}

	return &esmLinkData{
		uuid:           aws.ToString(output.UUID),
		arn:            aws.ToString(output.EventSourceMappingArn),
		eventSourceArn: queueARN,
		functionArn:    functionARN,
	}, nil
}

func (l *sqsQueueLambdaFunctionLinkActions) updateEventSourceMapping(
	ctx context.Context,
	uuid string,
	queueARN string,
	functionARN string,
	annotations *queueTriggerAnnotations,
	lambdaService lambdaservice.Service,
) (*esmLinkData, error) {
	// EventSourceArn is create-only and is therefore not set on update.
	updateInput := &lambda.UpdateEventSourceMappingInput{
		UUID:         aws.String(uuid),
		FunctionName: aws.String(functionARN),
		Enabled:      aws.Bool(annotations.enabled),
	}

	if annotations.batchSize > 0 {
		updateInput.BatchSize = aws.Int32(int32(annotations.batchSize))
	}
	if annotations.maximumBatchingWindowInSeconds >= 0 {
		updateInput.MaximumBatchingWindowInSeconds = aws.Int32(int32(annotations.maximumBatchingWindowInSeconds))
	}
	if annotations.maximumConcurrency >= 2 {
		updateInput.ScalingConfig = &types.ScalingConfig{
			MaximumConcurrency: aws.Int32(int32(annotations.maximumConcurrency)),
		}
	}
	if annotations.reportBatchItemFailures {
		updateInput.FunctionResponseTypes = []types.FunctionResponseType{
			types.FunctionResponseTypeReportBatchItemFailures,
		}
	}
	if len(annotations.filterPatterns) > 0 {
		updateInput.FilterCriteria = buildFilterCriteria(annotations.filterPatterns)
	}

	output, err := lambdaService.UpdateEventSourceMapping(ctx, updateInput)
	if err != nil {
		return nil, err
	}

	return &esmLinkData{
		uuid:           uuid,
		arn:            aws.ToString(output.EventSourceMappingArn),
		eventSourceArn: queueARN,
		functionArn:    functionARN,
	}, nil
}

type esmLinkData struct {
	uuid           string
	arn            string
	eventSourceArn string
	functionArn    string
}

func (l *sqsQueueLambdaFunctionLinkActions) grantConsumePermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	queueARN string,
	setupCtx *linkutils.LambdaLinkSetupContext,
	iamService iamservice.Service,
) (*permissionLinkData, error) {
	sid := createSQSQueueSID(input.ResourceAInfo)

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: sqsConsumeStatement(sid, queueARN),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q consume access to the SQS queue %q: %w",
				pluginutils.GetResourceName(input.ResourceBInfo),
				pluginutils.GetResourceName(input.ResourceAInfo),
				err,
			)
		}
		return nil, err
	}

	return &permissionLinkData{
		executionRoleName: createQueueLinkDataExecutionRoleName(input.ResourceBInfo),
		roleResourceName:  setupCtx.RoleResourceName,
		sid:               sid,
		queueARN:          queueARN,
		result:            result,
	}, nil
}

func sqsConsumeStatement(sid, queueARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   sqsConsumeActions(),
		"Resource": queueARN,
	}
}

type permissionLinkData struct {
	executionRoleName string
	roleResourceName  string
	sid               string
	queueARN          string
	result            linkutils.RoleAccessResult
}

func mergeIntermediaryOutputs(
	esmIdentity linkutils.IntermediaryIdentity,
	esmData *esmLinkData,
	permData *permissionLinkData,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specSQSConsumeStatementNode(permData.sid, permData.queueARN),
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it:
	// inline placements map the statement by Sid; managed (overflow) placements map
	// the attached managed policy ARN.
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(
		mappings,
		roleLinkData,
		permData.roleResourceName,
		permData.executionRoleName,
		permData.sid,
		permData.result,
	)

	// Project the event source mapping into the "intermediaries" link-data map (alongside
	// the role-policy outputs) so its create/update/destroy is surfaced as link changes
	// the same way as other link-owned intermediaries. The runtime uuid/arn are carried as
	// extra leaves used for reconciliation.
	linkData := linkutils.IntermediaryLinkData(linkutils.DeployedIntermediary{
		Identity: esmIdentity,
		Leaves: map[string]*core.MappingNode{
			"uuid":           core.MappingNodeFromString(esmData.uuid),
			"arn":            core.MappingNodeFromString(esmData.arn),
			"eventSourceArn": core.MappingNodeFromString(esmData.eventSourceArn),
			"functionArn":    core.MappingNodeFromString(esmData.functionArn),
		},
	})
	linkData.Fields[permData.executionRoleName] = roleLinkData

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData:                   linkData,
		ResourceDataMappings:       mappings,
	}
}

func specSQSConsumeStatementNode(sid, queueARN string) *core.MappingNode {
	actions := sqsConsumeActions()
	actionItems := make([]*core.MappingNode, len(actions))
	for i, action := range actions {
		actionItems[i] = core.MappingNodeFromString(action)
	}
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{Items: actionItems},
		"resource", core.MappingNodeFromString(queueARN),
	)
}

func createSQSQueueSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"SQSQueueConsume%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

func createQueueLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sExecutionRole",
		resourceInfo.ResourceName,
	)
}

func sqsConsumeActions() []string {
	return []string{
		"sqs:ReceiveMessage",
		"sqs:DeleteMessage",
		"sqs:GetQueueAttributes",
	}
}

type queueTriggerAnnotations struct {
	batchSize                      int
	maximumBatchingWindowInSeconds int
	maximumConcurrency             int
	reportBatchItemFailures        bool
	filterPatterns                 []string
	enabled                        bool
}

func getQueueTriggerAnnotations(
	resourceInfo *provider.ResourceInfo,
) *queueTriggerAnnotations {
	// Relationship annotations use aws.sqs.lambda.* pattern to indicate
	// these configure the SQS→Lambda trigger relationship
	batchSize, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.sqs.lambda.batchSize",
			Default: 10,
		},
	)

	maximumBatchingWindowInSeconds, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.sqs.lambda.maximumBatchingWindowInSeconds",
			Default: 0,
		},
	)

	// maximumConcurrency has no default; a value below 2 (including the zero default)
	// leaves the scaling config unset.
	maximumConcurrency, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.sqs.lambda.maximumConcurrency",
			Default: 0,
		},
	)

	reportBatchItemFailures, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     "aws.sqs.lambda.reportBatchItemFailures",
			Default: false,
		},
	)

	filterPatterns := getQueueFilterPatterns(resourceInfo)

	enabled, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     "aws.sqs.lambda.enabled",
			Default: true,
		},
	)

	return &queueTriggerAnnotations{
		batchSize:                      batchSize,
		maximumBatchingWindowInSeconds: maximumBatchingWindowInSeconds,
		maximumConcurrency:             maximumConcurrency,
		reportBatchItemFailures:        reportBatchItemFailures,
		filterPatterns:                 filterPatterns,
		enabled:                        enabled,
	}
}

func extractQueueARNFromResourceInfo(resourceInfo *provider.ResourceInfo) (string, bool) {
	// The CFN-canonical SQS queue exposes the queue ARN as the computed arn field.
	queueARN, hasQueueARN := pluginutils.GetValueByPath(
		"$.arn",
		resourceInfo.CurrentResourceState.SpecData,
	)
	if hasQueueARN && queueARN != nil {
		return core.StringValue(queueARN), true
	}

	return "", false
}

func queueFunctionESMIdentity(queueInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/lambda/eventSourceMapping",
		ResourceID:   queueFunctionESMResourceID(queueInfo, functionInfo),
		ResourceName: queueFunctionESMResourceName(queueInfo, functionInfo),
	}
}

func queueFunctionESMResourceID(queueInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%s__%s__event-source-mapping",
		pluginutils.GetResourceName(queueInfo),
		pluginutils.GetResourceName(functionInfo),
	)
}

func queueFunctionESMResourceName(queueInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sTrigger%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(queueInfo)),
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(functionInfo)),
	)
}

func getExistingEventSourceMappingUUID(linkState *state.LinkState, resourceID string) string {
	if linkState == nil || linkState.Data == nil {
		return ""
	}

	linkData := &core.MappingNode{
		Fields: linkState.Data,
	}
	uuid, hasUUID := pluginutils.GetValueByPath(
		linkutils.IntermediaryLeafPath(resourceID, "uuid"),
		linkData,
	)
	if !hasUUID || uuid == nil {
		return ""
	}

	return core.StringValue(uuid)
}

// Reads the indexed filter-pattern annotations (aws.sqs.lambda.filter.<index>) in
// order, stopping at the first absent index. Each value is a single AWS event-filtering
// pattern string; the set forms the event source mapping's filter criteria.
func getQueueFilterPatterns(resourceInfo *provider.ResourceInfo) []string {
	var patterns []string
	for i := 0; ; i++ {
		pattern, found := pluginutils.GetStringAnnotation(
			resourceInfo,
			&pluginutils.AnnotationQuery[string]{
				Key: fmt.Sprintf("aws.sqs.lambda.filter.%d", i),
			},
		)
		if !found || pattern == "" {
			break
		}
		patterns = append(patterns, pattern)
	}
	return patterns
}

func buildFilterCriteria(patterns []string) *types.FilterCriteria {
	if len(patterns) == 0 {
		return nil
	}

	filters := make([]types.Filter, 0, len(patterns))
	for _, pattern := range patterns {
		filters = append(filters, types.Filter{
			Pattern: aws.String(pattern),
		})
	}

	return &types.FilterCriteria{
		Filters: filters,
	}
}
