package lambdadynamodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *dynamoDBTableLambdaFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// On destroy, we don't disable streams - that would be destructive
	// and the user may want streams for other purposes
	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	providerCtx := provider.NewProviderContextFromLinkContext(
		input.LinkContext,
		"aws",
	)

	// Check if streams are already enabled
	currentStreamEnabled := isStreamEnabled(input.ResourceInfo)
	if currentStreamEnabled {
		// Streams already enabled, nothing to do
		return &provider.LinkUpdateResourceOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	// Get the table name
	tableName, hasTableName := extractTableNameFromResourceInfo(input.ResourceInfo)
	if !hasTableName {
		return nil, fmt.Errorf(
			"table name could not be retrieved from the DynamoDB table %q",
			pluginutils.GetResourceName(input.ResourceInfo),
		)
	}

	// Get the desired stream view type from annotation
	streamViewType := getStreamViewTypeAnnotation(input.ResourceInfo)

	// Enable streams on the table
	dynamodbService, err := l.getDynamoDBService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	_, err = dynamodbService.UpdateTable(ctx, &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
		StreamSpecification: &dynamodbtypes.StreamSpecification{
			StreamEnabled:  aws.Bool(true),
			StreamViewType: dynamodbtypes.StreamViewType(streamViewType),
		},
	})
	if err != nil {
		return nil, fmt.Errorf(
			"failed to enable DynamoDB Streams on table %q: %w",
			tableName,
			err,
		)
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			"streamSpecification",
			core.MappingNodeFields(
				"streamEnabled",
				core.MappingNodeFromBool(true),
				"streamViewType",
				core.MappingNodeFromString(streamViewType),
			),
		),
	}, nil
}

func isStreamEnabled(resourceInfo *provider.ResourceInfo) bool {
	// The CFN-canonical DynamoDB table has no streamEnabled flag (CFN models
	// stream enablement by the presence of the streamSpecification block). The
	// computed streamArn is populated only once a stream is active, so its
	// presence is the reliable signal that streams are already enabled.
	streamARN, hasStreamARN := pluginutils.GetValueByPath(
		"$.streamArn",
		resourceInfo.CurrentResourceState.SpecData,
	)
	return hasStreamARN && core.StringValue(streamARN) != ""
}

func getStreamViewTypeAnnotation(resourceInfo *provider.ResourceInfo) string {
	viewType, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     "aws.dynamodb.stream.viewType",
			Default: "NEW_AND_OLD_IMAGES",
		},
	)
	return viewType
}

func (l *dynamoDBTableLambdaFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The Lambda function is not modified as part of this link.
	// The function's execution role should have stream read permissions,
	// typically configured via the access link (aws/lambda/function::aws/dynamodb/table).
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *dynamoDBTableLambdaFunctionLinkActions) UpdateIntermediaryResources(
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

	// Get the stream ARN from the DynamoDB table
	streamARN, hasStreamARN := extractStreamARNFromResourceInfo(input.ResourceAInfo)
	if !hasStreamARN {
		return nil, fmt.Errorf(
			"stream ARN could not be retrieved from the DynamoDB table %q; "+
				"ensure the table has DynamoDB Streams enabled",
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
	// the event source mapping for this specific function. A table can trigger multiple
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
	esmResourceID := tableFunctionESMResourceID(input.ResourceAInfo, input.ResourceBInfo)
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

func (l *dynamoDBTableLambdaFunctionLinkActions) createIntermediaryResources(
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
	esmOutput, err := l.createEventSourceMapping(ctx, streamARN, functionARN, annotations, lambdaService)
	if err != nil {
		return nil, err
	}

	esmIdentity := tableFunctionESMIdentity(input.ResourceAInfo, input.ResourceBInfo)
	return mergeIntermediaryOutputs(esmIdentity, esmOutput, permOutput), nil
}

func (l *dynamoDBTableLambdaFunctionLinkActions) updateIntermediaryResources(
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

	esmIdentity := tableFunctionESMIdentity(input.ResourceAInfo, input.ResourceBInfo)
	return mergeIntermediaryOutputs(esmIdentity, esmOutput, permOutput), nil
}

func (l *dynamoDBTableLambdaFunctionLinkActions) destroyIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	lambdaService lambdaservice.Service,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Delete the Event Source Mapping
	esmResourceID := tableFunctionESMResourceID(input.ResourceAInfo, input.ResourceBInfo)
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
	sid := createDynamoDBStreamSID(input.ResourceAInfo)
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

func (l *dynamoDBTableLambdaFunctionLinkActions) createEventSourceMapping(
	ctx context.Context,
	streamARN string,
	functionARN string,
	annotations *streamTriggerAnnotations,
	lambdaService lambdaservice.Service,
) (*esmLinkData, error) {
	createInput := &lambda.CreateEventSourceMappingInput{
		EventSourceArn:   aws.String(streamARN),
		FunctionName:     aws.String(functionARN),
		StartingPosition: types.EventSourcePosition(annotations.startingPosition),
		Enabled:          aws.Bool(annotations.enabled),
	}

	if annotations.batchSize > 0 {
		createInput.BatchSize = aws.Int32(int32(annotations.batchSize))
	}
	if annotations.batchWindow > 0 {
		createInput.MaximumBatchingWindowInSeconds = aws.Int32(int32(annotations.batchWindow))
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
	if annotations.filterCriteria != "" {
		filterCriteria, err := parseFilterCriteria(annotations.filterCriteria)
		if err != nil {
			return nil, fmt.Errorf("failed to parse filterCriteria: %w", err)
		}
		createInput.FilterCriteria = filterCriteria
	}

	// The execution role's stream-read grant is applied immediately before this call,
	// but IAM is eventually consistent: CreateEventSourceMapping can transiently fail
	// validation ("Cannot access stream ... ensure the role can perform ...") until the
	// policy propagates. Treat that as retryable so the link deploy retries it.
	createESM := pluginutils.RetryableReturnValue(
		func(ctx context.Context, in *lambda.CreateEventSourceMappingInput) (*lambda.CreateEventSourceMappingOutput, error) {
			return lambdaService.CreateEventSourceMapping(ctx, in)
		},
		isRoleNotYetPropagatedError,
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

// isRoleNotYetPropagatedError reports whether an error is the transient
// InvalidParameterValueException AWS returns when an execution role's freshly-granted
// stream permissions have not yet propagated (IAM eventual consistency).
func isRoleNotYetPropagatedError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidParameterValueException" {
		msg := apiErr.ErrorMessage()
		return strings.Contains(msg, "Cannot access stream") ||
			strings.Contains(msg, "ensure the role can perform")
	}
	return false
}

func (l *dynamoDBTableLambdaFunctionLinkActions) updateEventSourceMapping(
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
	if annotations.batchWindow >= 0 {
		updateInput.MaximumBatchingWindowInSeconds = aws.Int32(int32(annotations.batchWindow))
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

	if annotations.filterCriteria != "" {
		filterCriteria, err := parseFilterCriteria(annotations.filterCriteria)
		if err != nil {
			return nil, fmt.Errorf("failed to parse filterCriteria: %w", err)
		}
		updateInput.FilterCriteria = filterCriteria
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

// grantStreamPermissions packs the DynamoDB stream-read statement into the Lambda
// execution role's allocator-managed policy set (a single shared inline policy,
// with managed-policy overflow), replacing any prior statement with the same Sid.
func (l *dynamoDBTableLambdaFunctionLinkActions) grantStreamPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	streamARN string,
	setupCtx *linkutils.LambdaLinkSetupContext,
	iamService iamservice.Service,
) (*permissionLinkData, error) {
	sid := createDynamoDBStreamSID(input.ResourceAInfo)

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: dynamoDBStreamStatement(sid, streamARN),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q read access to the DynamoDB stream of table %q: %w",
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

// dynamoDBStreamStatement builds the IAM policy statement (canonical PascalCase, as
// the IAM API expects) granting read access to the table's stream.
func dynamoDBStreamStatement(sid, streamARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   dynamoDBStreamActions(),
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

// mergeIntermediaryOutputs combines the event source mapping link data with the
// stream-read grant's link data and (for inline placements) drift mappings, so the
// ESM outputs and the role-policy outputs are both preserved.
func mergeIntermediaryOutputs(
	esmIdentity linkutils.IntermediaryIdentity,
	esmData *esmLinkData,
	permData *permissionLinkData,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		camelStreamStatementNode(permData.sid, permData.streamARN),
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it:
	// inline placements map the statement by Sid; managed (overflow) placements map
	// the attached managed policy ARN.
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(
		mappings, roleLinkData, permData.roleResourceName, permData.executionRoleName, permData.sid, permData.result,
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

// camelStreamStatementNode builds the statement in the camelCase spec form the
// role's external state uses (after Cloud Control name translation), so the drift
// comparison against link data matches.
func camelStreamStatementNode(sid, streamARN string) *core.MappingNode {
	actions := dynamoDBStreamActions()
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

func createDynamoDBStreamSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"DynamoDBStreamRead%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

func createStreamLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sExecutionRole",
		resourceInfo.ResourceName,
	)
}

func dynamoDBStreamActions() []string {
	return []string{
		"dynamodb:DescribeStream",
		"dynamodb:GetRecords",
		"dynamodb:GetShardIterator",
		"dynamodb:ListStreams",
	}
}

type streamTriggerAnnotations struct {
	batchSize                  int
	batchWindow                int
	startingPosition           string
	parallelizationFactor      int
	maximumRetryAttempts       int
	maximumRecordAgeInSeconds  int
	bisectBatchOnFunctionError bool
	filterCriteria             string
	enabled                    bool
}

func getStreamTriggerAnnotations(
	resourceInfo *provider.ResourceInfo,
) *streamTriggerAnnotations {
	// Relationship annotations use aws.dynamodb.lambda.stream.* pattern to indicate
	// these configure the DynamoDB→Lambda stream relationship
	batchSize, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.dynamodb.lambda.stream.batchSize",
			Default: 100,
		},
	)

	batchWindow, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.dynamodb.lambda.stream.batchWindow",
			Default: 0,
		},
	)

	startingPosition, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     "aws.dynamodb.lambda.stream.startingPosition",
			Default: "LATEST",
		},
	)

	parallelizationFactor, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.dynamodb.lambda.stream.parallelizationFactor",
			Default: 1,
		},
	)

	maximumRetryAttempts, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.dynamodb.lambda.stream.maximumRetryAttempts",
			Default: -1,
		},
	)

	maximumRecordAgeInSeconds, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.dynamodb.lambda.stream.maximumRecordAgeInSeconds",
			Default: -1,
		},
	)

	bisectBatchOnFunctionError, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     "aws.dynamodb.lambda.stream.bisectBatchOnFunctionError",
			Default: false,
		},
	)

	filterCriteria, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: "aws.dynamodb.lambda.stream.filterCriteria",
		},
	)

	enabled, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     "aws.dynamodb.lambda.stream.enabled",
			Default: true,
		},
	)

	return &streamTriggerAnnotations{
		batchSize:                  batchSize,
		batchWindow:                batchWindow,
		startingPosition:           startingPosition,
		parallelizationFactor:      parallelizationFactor,
		maximumRetryAttempts:       maximumRetryAttempts,
		maximumRecordAgeInSeconds:  maximumRecordAgeInSeconds,
		bisectBatchOnFunctionError: bisectBatchOnFunctionError,
		filterCriteria:             filterCriteria,
		enabled:                    enabled,
	}
}

func extractStreamARNFromResourceInfo(resourceInfo *provider.ResourceInfo) (string, bool) {
	// The CFN-canonical DynamoDB table exposes the stream ARN as the computed
	// streamArn field (the native resource called it latestStreamArn).
	streamARN, hasStreamARN := pluginutils.GetValueByPath(
		"$.streamArn",
		resourceInfo.CurrentResourceState.SpecData,
	)
	if hasStreamARN && streamARN != nil {
		return core.StringValue(streamARN), true
	}

	return "", false
}

// tableFunctionESMIdentity is the single source of truth for the link-owned event source
// mapping's identity, used by UpdateIntermediaryResources (to persist it), StageChanges (to
// project it into link changes) and the UUID lookup. The event source mapping is created via
// the Lambda SDK rather than as a managed Cloud Control resource, but it is projected into
// link data the same way as other link-owned intermediaries.
func tableFunctionESMIdentity(tableInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/lambda/eventSourceMapping",
		ResourceID:   tableFunctionESMResourceID(tableInfo, functionInfo),
		ResourceName: tableFunctionESMResourceName(tableInfo, functionInfo),
	}
}

func tableFunctionESMResourceID(tableInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%s__%s__event-source-mapping",
		pluginutils.GetResourceName(tableInfo),
		pluginutils.GetResourceName(functionInfo),
	)
}

func tableFunctionESMResourceName(tableInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sStream%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(tableInfo)),
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

func parseFilterCriteria(filterCriteriaJSON string) (*types.FilterCriteria, error) {
	if filterCriteriaJSON == "" {
		return nil, nil
	}

	var filterCriteria struct {
		Filters []struct {
			Pattern string `json:"pattern"`
		} `json:"filters"`
	}

	if err := json.Unmarshal([]byte(filterCriteriaJSON), &filterCriteria); err != nil {
		return nil, err
	}

	var filters []types.Filter
	for _, f := range filterCriteria.Filters {
		filters = append(filters, types.Filter{
			Pattern: aws.String(f.Pattern),
		})
	}

	return &types.FilterCriteria{
		Filters: filters,
	}, nil
}
