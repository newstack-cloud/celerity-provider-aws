package lambdadynamodb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
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
	streamEnabled, hasStreamEnabled := pluginutils.GetValueByPath(
		"$.streamSpecification.streamEnabled",
		resourceInfo.CurrentResourceState.SpecData,
	)
	if !hasStreamEnabled || streamEnabled == nil {
		return false
	}
	return core.BoolValue(streamEnabled)
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

	// Acquire a lock on the role resource to prevent concurrent updates
	err = input.ResourceService.AcquireResourceLock(
		ctx,
		&provider.AcquireResourceLockInput{
			InstanceID:      pluginutils.GetInstanceID(input.ResourceBInfo),
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

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.destroyIntermediaryResources(
			ctx,
			input,
			setupCtx,
			rolePolicies,
			lambdaService,
			iamService,
		)
	}

	// Check if we have an existing event source mapping UUID in the link state
	existingUUID := getExistingEventSourceMappingUUID(input.CurrentLinkState)

	if existingUUID != "" {
		return l.updateIntermediaryResources(
			ctx,
			input,
			existingUUID,
			streamARN,
			functionARN,
			annotations,
			setupCtx,
			rolePolicies,
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
		rolePolicies,
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
	rolePolicies *iam.ListRolePoliciesOutput,
	lambdaService lambdaservice.Service,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Create the Event Source Mapping
	esmOutput, err := l.createEventSourceMapping(ctx, streamARN, functionARN, annotations, lambdaService)
	if err != nil {
		return nil, err
	}

	// Add stream read permissions to the Lambda execution role
	permOutput, err := l.addStreamPermissions(
		ctx,
		input,
		streamARN,
		setupCtx,
		rolePolicies,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	return mergeIntermediaryOutputs(esmOutput, permOutput), nil
}

func (l *dynamoDBTableLambdaFunctionLinkActions) updateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	existingUUID string,
	streamARN string,
	functionARN string,
	annotations *streamTriggerAnnotations,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
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

	// Update stream read permissions on the Lambda execution role
	permOutput, err := l.updateStreamPermissions(
		ctx,
		input,
		streamARN,
		setupCtx,
		rolePolicies,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	return mergeIntermediaryOutputs(esmOutput, permOutput), nil
}

func (l *dynamoDBTableLambdaFunctionLinkActions) destroyIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	lambdaService lambdaservice.Service,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Delete the Event Source Mapping
	uuid := getExistingEventSourceMappingUUID(input.CurrentLinkState)
	if uuid != "" {
		_, err := lambdaService.DeleteEventSourceMapping(ctx, &lambda.DeleteEventSourceMappingInput{
			UUID: aws.String(uuid),
		})
		if err != nil {
			return nil, err
		}
	}

	// Remove stream read permissions from the Lambda execution role
	err := l.removeStreamPermissions(
		ctx,
		input,
		setupCtx,
		rolePolicies,
		iamService,
	)
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

	output, err := lambdaService.CreateEventSourceMapping(ctx, createInput)
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

func (l *dynamoDBTableLambdaFunctionLinkActions) addStreamPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	streamARN string,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	iamService iamservice.Service,
) (*permissionLinkData, error) {
	sid := createDynamoDBStreamSID(input.ResourceAInfo)
	permission := map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   dynamoDBStreamActions(),
		"Resource": streamARN,
	}

	if len(rolePolicies.PolicyNames) > 0 {
		return l.updateExistingRolePolicyWithStreamPerms(
			ctx,
			input,
			setupCtx,
			rolePolicies,
			permission,
			sid,
			iamService,
		)
	}

	return l.addNewRolePolicyWithStreamPerms(
		ctx,
		input,
		setupCtx,
		permission,
		sid,
		iamService,
	)
}

func (l *dynamoDBTableLambdaFunctionLinkActions) updateStreamPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	streamARN string,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	iamService iamservice.Service,
) (*permissionLinkData, error) {
	// For updates, we use the same logic as add - it will update or create as needed
	return l.addStreamPermissions(ctx, input, streamARN, setupCtx, rolePolicies, iamService)
}

func (l *dynamoDBTableLambdaFunctionLinkActions) removeStreamPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	iamService iamservice.Service,
) error {
	executionRoleName := createStreamLinkDataExecutionRoleName(input.ResourceBInfo)

	policyName, existingPermSID := linkutils.ExtractPolicyNameAndCurrentPermissionSID(
		input.CurrentLinkState,
		executionRoleName,
		rolePolicies,
	)

	return linkutils.RemoveExistingRolePolicyPermissions(
		ctx,
		iamService,
		setupCtx.RoleName,
		policyName,
		[]string{existingPermSID},
	)
}

func (l *dynamoDBTableLambdaFunctionLinkActions) addNewRolePolicyWithStreamPerms(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	permission map[string]any,
	sid string,
	iamService iamservice.Service,
) (*permissionLinkData, error) {
	policyName, err := linkutils.CreateNewRolePolicy(
		ctx,
		iamService,
		input.InstanceName,
		setupCtx.RoleName,
		[]map[string]any{permission},
	)
	if err != nil {
		return nil, err
	}

	return &permissionLinkData{
		executionRoleName: createStreamLinkDataExecutionRoleName(input.ResourceBInfo),
		roleResourceState: setupCtx.RoleResourceState,
		policyName:        policyName,
		permission:        permission,
		sid:               sid,
	}, nil
}

func (l *dynamoDBTableLambdaFunctionLinkActions) updateExistingRolePolicyWithStreamPerms(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	permission map[string]any,
	sid string,
	iamService iamservice.Service,
) (*permissionLinkData, error) {
	executionRoleName := createStreamLinkDataExecutionRoleName(input.ResourceBInfo)

	policyName, existingPermSID := linkutils.ExtractPolicyNameAndCurrentPermissionSID(
		input.CurrentLinkState,
		executionRoleName,
		rolePolicies,
	)

	err := linkutils.UpdateExistingRolePolicy(
		ctx,
		iamService,
		setupCtx.RoleName,
		policyName,
		[]map[string]any{permission},
		[]string{existingPermSID},
	)
	if err != nil {
		return nil, err
	}

	return &permissionLinkData{
		executionRoleName: executionRoleName,
		roleResourceState: setupCtx.RoleResourceState,
		policyName:        policyName,
		permission:        permission,
		sid:               sid,
	}, nil
}

type permissionLinkData struct {
	executionRoleName string
	roleResourceState *state.ResourceState
	policyName        string
	permission        map[string]any
	sid               string
}

func mergeIntermediaryOutputs(
	esmData *esmLinkData,
	permData *permissionLinkData,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	permissionNode, _ := pluginutils.AnyToMappingNode(permData.permission)

	roleResourcePath := fmt.Sprintf(
		"%s::spec.policies[@.name=\"%q\"].Statement[@.Sid=\"%q\"]",
		permData.roleResourceState.Name,
		permData.policyName,
		permData.sid,
	)
	linkDataFieldPath := linkutils.PermissionFieldPath(permData.executionRoleName)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: core.MappingNodeFields(
			"eventSourceMapping",
			core.MappingNodeFields(
				"uuid",
				core.MappingNodeFromString(esmData.uuid),
				"arn",
				core.MappingNodeFromString(esmData.arn),
				"eventSourceArn",
				core.MappingNodeFromString(esmData.eventSourceArn),
				"functionArn",
				core.MappingNodeFromString(esmData.functionArn),
			),
			permData.executionRoleName,
			core.MappingNodeFields(
				linkutils.PermissionFieldName,
				permissionNode,
				linkutils.PolicyNameFieldName,
				core.MappingNodeFromString(permData.policyName),
			),
		),
		ResourceDataMappings: map[string]string{
			roleResourcePath: linkDataFieldPath,
		},
	}
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
	// Look for latestStreamArn in the resource state (stored in SpecData)
	streamARN, hasStreamARN := pluginutils.GetValueByPath(
		"$.latestStreamArn",
		resourceInfo.CurrentResourceState.SpecData,
	)
	if hasStreamARN && streamARN != nil {
		return core.StringValue(streamARN), true
	}

	return "", false
}

func getExistingEventSourceMappingUUID(linkState *state.LinkState) string {
	if linkState == nil || linkState.Data == nil {
		return ""
	}

	linkData := &core.MappingNode{
		Fields: linkState.Data,
	}
	uuid, hasUUID := pluginutils.GetValueByPath(
		"$.eventSourceMapping.uuid",
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
