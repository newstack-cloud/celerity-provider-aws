package kinesislambda

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *kinesisStreamLambdaFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// A Kinesis data stream needs no enabling to act as an event source, so there
	// is nothing to do on the stream itself for this link.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *kinesisStreamLambdaFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The Lambda function is not modified as part of this link.
	// The function's execution role is granted stream read permissions via the
	// intermediary resource update.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *kinesisStreamLambdaFunctionLinkActions) UpdateIntermediaryResources(
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

	// Get the stream ARN from the Kinesis stream (the stream is the event source).
	streamARN, hasStreamARN := extractStreamARNFromResourceInfo(input.ResourceAInfo)
	if !hasStreamARN {
		return nil, fmt.Errorf(
			"stream ARN could not be retrieved from the Kinesis stream %q",
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

	// Stream trigger annotations are on resourceB (Lambda function) because they configure
	// the event source mapping for this specific function. A stream can trigger multiple
	// functions with different configurations.
	annotations := getStreamTriggerAnnotations(input.ResourceBInfo)

	// Set up Lambda link context to get role info for stream permissions
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
	esmResourceID := streamFunctionESMResourceID(input.ResourceAInfo, input.ResourceBInfo)
	existingUUID := getExistingEventSourceMappingUUID(input.CurrentLinkState, esmResourceID)

	if existingUUID != "" {
		return l.updateIntermediaryResources(
			ctx,
			input,
			existingUUID,
			streamARN,
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
		streamARN,
		functionARN,
		annotations,
		setupCtx,
		lambdaService,
		iamService,
	)
}

func (l *kinesisStreamLambdaFunctionLinkActions) createIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	streamARN string,
	functionARN string,
	annotations *streamTriggerAnnotations,
	setupCtx *linkutils.LambdaLinkSetupContext,
	lambdaService lambdaservice.Service,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Grant stream read permissions to the Lambda execution role before creating the
	// event source mapping: CreateEventSourceMapping validates that the role can read
	// the stream, so the grant must be in place first.
	permOutput, err := l.grantStreamPermissions(
		ctx,
		input,
		streamARN,
		setupCtx,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	// Create the Event Source Mapping.
	esmOutput, err := l.createEventSourceMapping(
		ctx, streamARN, functionARN, annotations, lambdaService,
	)
	if err != nil {
		return nil, err
	}

	esmIdentity := streamFunctionESMIdentity(input.ResourceAInfo, input.ResourceBInfo)
	return mergeIntermediaryOutputs(esmIdentity, esmOutput, permOutput), nil
}

func (l *kinesisStreamLambdaFunctionLinkActions) updateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	existingUUID string,
	streamARN string,
	functionARN string,
	annotations *streamTriggerAnnotations,
	setupCtx *linkutils.LambdaLinkSetupContext,
	lambdaService lambdaservice.Service,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Update the Event Source Mapping
	esmOutput, err := l.updateEventSourceMapping(
		ctx,
		existingUUID,
		streamARN,
		functionARN,
		annotations,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	// Grant stream read permissions to the Lambda execution role
	// (the allocator replaces the statement in place on update).
	permOutput, err := l.grantStreamPermissions(
		ctx,
		input,
		streamARN,
		setupCtx,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	esmIdentity := streamFunctionESMIdentity(input.ResourceAInfo, input.ResourceBInfo)
	return mergeIntermediaryOutputs(esmIdentity, esmOutput, permOutput), nil
}

func (l *kinesisStreamLambdaFunctionLinkActions) destroyIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	lambdaService lambdaservice.Service,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Delete the Event Source Mapping
	esmResourceID := streamFunctionESMResourceID(input.ResourceAInfo, input.ResourceBInfo)
	uuid := getExistingEventSourceMappingUUID(input.CurrentLinkState, esmResourceID)
	if uuid != "" {
		_, err := lambdaService.DeleteEventSourceMapping(ctx, &lambda.DeleteEventSourceMappingInput{
			UUID: aws.String(uuid),
		})
		if err != nil {
			return nil, err
		}
	}

	// Remove the stream-read grant from the role's allocator-managed policy set.
	sid := createKinesisStreamSID(input.ResourceAInfo)
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

func (l *kinesisStreamLambdaFunctionLinkActions) createEventSourceMapping(
	ctx context.Context,
	streamARN string,
	functionARN string,
	annotations *streamTriggerAnnotations,
	lambdaService lambdaservice.Service,
) (*esmLinkData, error) {
	createInput := &lambda.CreateEventSourceMappingInput{
		EventSourceArn:   aws.String(streamARN),
		FunctionName:     aws.String(functionARN),
		StartingPosition: lambdatypes.EventSourcePosition(annotations.startingPosition),
		Enabled:          aws.Bool(annotations.enabled),
	}

	if annotations.startingPosition == "AT_TIMESTAMP" && annotations.startingPositionTimestamp != "" {
		ts, err := parseStartingPositionTimestamp(annotations.startingPositionTimestamp)
		if err != nil {
			return nil, err
		}
		createInput.StartingPositionTimestamp = aws.Time(ts)
	}

	if annotations.batchSize > 0 {
		createInput.BatchSize = aws.Int32(int32(annotations.batchSize))
	}
	if annotations.maximumBatchingWindowInSeconds > 0 {
		createInput.MaximumBatchingWindowInSeconds = aws.Int32(int32(annotations.maximumBatchingWindowInSeconds))
	}
	if annotations.parallelizationFactor > 0 {
		createInput.ParallelizationFactor = aws.Int32(int32(annotations.parallelizationFactor))
	}
	if annotations.maximumRetryAttempts != 0 {
		createInput.MaximumRetryAttempts = aws.Int32(int32(annotations.maximumRetryAttempts))
	}
	if annotations.maximumRecordAgeInSeconds != 0 {
		createInput.MaximumRecordAgeInSeconds = aws.Int32(int32(annotations.maximumRecordAgeInSeconds))
	}
	if annotations.bisectBatchOnFunctionError {
		createInput.BisectBatchOnFunctionError = aws.Bool(true)
	}
	if annotations.tumblingWindowInSeconds > 0 {
		createInput.TumblingWindowInSeconds = aws.Int32(int32(annotations.tumblingWindowInSeconds))
	}
	if annotations.reportBatchItemFailures {
		createInput.FunctionResponseTypes = []lambdatypes.FunctionResponseType{
			lambdatypes.FunctionResponseTypeReportBatchItemFailures,
		}
	}
	if len(annotations.filterPatterns) > 0 {
		createInput.FilterCriteria = buildFilterCriteria(annotations.filterPatterns)
	}

	// The execution role's stream-read grant is applied immediately before this call,
	// but IAM is eventually consistent: CreateEventSourceMapping can transiently fail
	// validation ("Cannot access stream ... ensure the role can perform ...") until the
	// policy propagates. Treat that as retryable so the link deploy retries it.
	createESM := pluginutils.RetryableReturnValue(
		func(ctx context.Context, in *lambda.CreateEventSourceMappingInput) (*lambda.CreateEventSourceMappingOutput, error) {
			return lambdaService.CreateEventSourceMapping(ctx, in)
		},
		linkutils.IsRoleNotYetPropagatedError,
	)

	output, err := createESM(ctx, createInput)
	if err != nil {
		return nil, err
	}

	return &esmLinkData{
		uuid:           aws.ToString(output.UUID),
		arn:            aws.ToString(output.EventSourceMappingArn),
		eventSourceArn: streamARN,
		functionArn:    functionARN,
	}, nil
}

// parseStartingPositionTimestamp parses an AT_TIMESTAMP starting-position annotation
// (Unix epoch seconds) into a time, tolerating a value expressed as a float.
func parseStartingPositionTimestamp(value string) (time.Time, error) {
	secs, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return time.Unix(secs, 0), nil
	}

	if floatSecs, floatErr := strconv.ParseFloat(value, 64); floatErr == nil {
		return time.Unix(int64(floatSecs), 0), nil
	}

	return time.Time{}, fmt.Errorf(
		"invalid aws.kinesis.lambda.startingPositionTimestamp %q: expected Unix epoch seconds: %w",
		value,
		err,
	)
}

func (l *kinesisStreamLambdaFunctionLinkActions) updateEventSourceMapping(
	ctx context.Context,
	uuid string,
	streamARN string,
	functionARN string,
	annotations *streamTriggerAnnotations,
	lambdaService lambdaservice.Service,
) (*esmLinkData, error) {
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
	if annotations.parallelizationFactor > 0 {
		updateInput.ParallelizationFactor = aws.Int32(int32(annotations.parallelizationFactor))
	}
	if annotations.maximumRetryAttempts != 0 {
		updateInput.MaximumRetryAttempts = aws.Int32(int32(annotations.maximumRetryAttempts))
	}
	if annotations.maximumRecordAgeInSeconds != 0 {
		updateInput.MaximumRecordAgeInSeconds = aws.Int32(int32(annotations.maximumRecordAgeInSeconds))
	}
	updateInput.BisectBatchOnFunctionError = aws.Bool(annotations.bisectBatchOnFunctionError)
	if annotations.tumblingWindowInSeconds > 0 {
		updateInput.TumblingWindowInSeconds = aws.Int32(int32(annotations.tumblingWindowInSeconds))
	}
	if annotations.reportBatchItemFailures {
		updateInput.FunctionResponseTypes = []lambdatypes.FunctionResponseType{
			lambdatypes.FunctionResponseTypeReportBatchItemFailures,
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
		eventSourceArn: streamARN,
		functionArn:    functionARN,
	}, nil
}

type esmLinkData struct {
	uuid           string
	arn            string
	eventSourceArn string
	functionArn    string
}

// Packs the Kinesis stream-read statement into the Lambda
// execution role's allocator-managed policy set (a single shared inline policy,
// with managed-policy overflow), replacing any prior statement with the same Sid.
func (l *kinesisStreamLambdaFunctionLinkActions) grantStreamPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	streamARN string,
	setupCtx *linkutils.LambdaLinkSetupContext,
	iamService iamservice.Service,
) (*permissionLinkData, error) {
	sid := createKinesisStreamSID(input.ResourceAInfo)

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: kinesisStreamStatement(sid, streamARN),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q read access to the Kinesis stream %q: %w",
				pluginutils.GetResourceName(input.ResourceBInfo),
				pluginutils.GetResourceName(input.ResourceAInfo),
				err,
			)
		}
		return nil, err
	}

	return &permissionLinkData{
		executionRoleName: createStreamLinkDataExecutionRoleName(input.ResourceBInfo),
		roleResourceName:  setupCtx.RoleResourceName,
		sid:               sid,
		streamARN:         streamARN,
		result:            result,
	}, nil
}

// Builds the IAM policy statement (canonical PascalCase, as
// the IAM API expects) granting read access to the Kinesis stream.
func kinesisStreamStatement(sid, streamARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   kinesisStreamActions(),
		"Resource": streamARN,
	}
}

type permissionLinkData struct {
	executionRoleName string
	roleResourceName  string
	sid               string
	streamARN         string
	result            linkutils.RoleAccessResult
}

// Combines the event source mapping link data with the
// stream-read grant's link data and (for inline placements) drift mappings, so the
// ESM outputs and the role-policy outputs are both preserved.
func mergeIntermediaryOutputs(
	esmIdentity linkutils.IntermediaryIdentity,
	esmData *esmLinkData,
	permData *permissionLinkData,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specStreamStatementNode(permData.sid, permData.streamARN),
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

func specStreamStatementNode(sid, streamARN string) *core.MappingNode {
	actions := kinesisStreamActions()
	actionItems := make([]*core.MappingNode, len(actions))
	for i, action := range actions {
		actionItems[i] = core.MappingNodeFromString(action)
	}
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{Items: actionItems},
		"resource", core.MappingNodeFromString(streamARN),
	)
}

func createKinesisStreamSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"KinesisStreamRead%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

func createStreamLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sExecutionRole",
		resourceInfo.ResourceName,
	)
}

func kinesisStreamActions() []string {
	return []string{
		"kinesis:DescribeStream",
		"kinesis:DescribeStreamSummary",
		"kinesis:GetRecords",
		"kinesis:GetShardIterator",
		"kinesis:ListShards",
		"kinesis:ListStreams",
	}
}

type streamTriggerAnnotations struct {
	batchSize                      int
	maximumBatchingWindowInSeconds int
	startingPosition               string
	startingPositionTimestamp      string
	parallelizationFactor          int
	maximumRetryAttempts           int
	maximumRecordAgeInSeconds      int
	bisectBatchOnFunctionError     bool
	tumblingWindowInSeconds        int
	reportBatchItemFailures        bool
	filterPatterns                 []string
	enabled                        bool
}

func getStreamTriggerAnnotations(
	resourceInfo *provider.ResourceInfo,
) *streamTriggerAnnotations {
	// Relationship annotations use aws.kinesis.lambda.* pattern to indicate
	// these configure the Kinesis→Lambda stream relationship.
	batchSize, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.kinesis.lambda.batchSize",
			Default: 100,
		},
	)

	maximumBatchingWindowInSeconds, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.kinesis.lambda.maximumBatchingWindowInSeconds",
			Default: 0,
		},
	)

	startingPosition, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     "aws.kinesis.lambda.startingPosition",
			Default: "LATEST",
		},
	)

	startingPositionTimestamp, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: "aws.kinesis.lambda.startingPositionTimestamp",
		},
	)

	parallelizationFactor, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.kinesis.lambda.parallelizationFactor",
			Default: 1,
		},
	)

	maximumRetryAttempts, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.kinesis.lambda.maximumRetryAttempts",
			Default: -1,
		},
	)

	maximumRecordAgeInSeconds, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.kinesis.lambda.maximumRecordAgeInSeconds",
			Default: -1,
		},
	)

	bisectBatchOnFunctionError, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     "aws.kinesis.lambda.bisectBatchOnFunctionError",
			Default: false,
		},
	)

	tumblingWindowInSeconds, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.kinesis.lambda.tumblingWindowInSeconds",
			Default: 0,
		},
	)

	reportBatchItemFailures, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     "aws.kinesis.lambda.reportBatchItemFailures",
			Default: false,
		},
	)

	filterPatterns := getStreamFilterPatterns(resourceInfo)

	enabled, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     "aws.kinesis.lambda.enabled",
			Default: true,
		},
	)

	return &streamTriggerAnnotations{
		batchSize:                      batchSize,
		maximumBatchingWindowInSeconds: maximumBatchingWindowInSeconds,
		startingPosition:               startingPosition,
		startingPositionTimestamp:      startingPositionTimestamp,
		parallelizationFactor:          parallelizationFactor,
		maximumRetryAttempts:           maximumRetryAttempts,
		maximumRecordAgeInSeconds:      maximumRecordAgeInSeconds,
		bisectBatchOnFunctionError:     bisectBatchOnFunctionError,
		tumblingWindowInSeconds:        tumblingWindowInSeconds,
		reportBatchItemFailures:        reportBatchItemFailures,
		filterPatterns:                 filterPatterns,
		enabled:                        enabled,
	}
}

func extractStreamARNFromResourceInfo(resourceInfo *provider.ResourceInfo) (string, bool) {
	// The Kinesis stream is the event source; its ARN is the computed arn field.
	streamARN, hasStreamARN := pluginutils.GetValueByPath(
		"$.arn",
		resourceInfo.CurrentResourceState.SpecData,
	)
	if hasStreamARN && streamARN != nil {
		return core.StringValue(streamARN), true
	}

	return "", false
}

func streamFunctionESMIdentity(streamInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/lambda/eventSourceMapping",
		ResourceID:   streamFunctionESMResourceID(streamInfo, functionInfo),
		ResourceName: streamFunctionESMResourceName(streamInfo, functionInfo),
	}
}

func streamFunctionESMResourceID(streamInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%s__%s__event-source-mapping",
		pluginutils.GetResourceName(streamInfo),
		pluginutils.GetResourceName(functionInfo),
	)
}

func streamFunctionESMResourceName(streamInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sStream%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(streamInfo)),
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

// Reads the indexed filter-pattern annotations
// (aws.kinesis.lambda.filter.<index>) in order, stopping at the first absent index.
// Each value is a single AWS event-filtering pattern string; the set forms the event
// source mapping's filter criteria.
func getStreamFilterPatterns(resourceInfo *provider.ResourceInfo) []string {
	var patterns []string
	for i := 0; ; i++ {
		pattern, found := pluginutils.GetStringAnnotation(
			resourceInfo,
			&pluginutils.AnnotationQuery[string]{
				Key: fmt.Sprintf("aws.kinesis.lambda.filter.%d", i),
			},
		)
		if !found || pattern == "" {
			break
		}
		patterns = append(patterns, pattern)
	}
	return patterns
}

func buildFilterCriteria(patterns []string) *lambdatypes.FilterCriteria {
	if len(patterns) == 0 {
		return nil
	}

	filters := make([]lambdatypes.Filter, 0, len(patterns))
	for _, pattern := range patterns {
		filters = append(filters, lambdatypes.Filter{
			Pattern: aws.String(pattern),
		})
	}

	return &lambdatypes.FilterCriteria{
		Filters: filters,
	}
}
