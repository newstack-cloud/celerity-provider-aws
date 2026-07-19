package lambdas3

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

func (l *functionBucketLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, _, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getBucketLinkAnnotations(input.ResourceInfo, input.OtherResourceInfo)
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

	bucketName, hasBucketName := extractBucketName(input.OtherResourceInfo)
	if !hasBucketName {
		return nil, fmt.Errorf(
			"bucket name could not be retrieved from the linked to %q S3 bucket resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeBucketEnvVars(ctx, input, setupCtx.FunctionARN, annotations.envVarName, setupCtx.LambdaOutput, lambdaService)
	}

	return l.addBucketEnvVars(ctx, input, setupCtx.FunctionARN, bucketName, annotations.envVarName, setupCtx.LambdaOutput, lambdaService)
}

func (l *functionBucketLinkActions) addBucketEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, bucketName, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := bucketEnvVarName(envVarName, input.OtherResourceInfo)
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
		map[string]string{finalEnvVarName: bucketName},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			pluginutils.GetResourceName(input.ResourceInfo),
			core.MappingNodeFields(
				"environmentVariables",
				core.MappingNodeFields(finalEnvVarName, core.MappingNodeFromString(bucketName)),
			),
		),
		ResourceDataMappings: map[string]string{dataMappingKey: linkDataFieldPath},
	}, nil
}

func (l *functionBucketLinkActions) removeBucketEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN, envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := bucketEnvVarName(envVarName, input.OtherResourceInfo)
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

func (l *functionBucketLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The S3 bucket is not modified by this link; only the Lambda function and its
	// execution role are updated to allow it to access the bucket.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) the Lambda execution role access to the
// S3 bucket, then activates the S3 gateway VPC endpoint when the function is VPC-isolated.
func (l *functionBucketLinkActions) UpdateIntermediaryResources(
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

	// The execution role is shared by every link that grants it access, so lock it for
	// the read-modify-write of its policy set.
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

	sid := createS3AccessSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if _, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setupCtx.RoleName,
			SID:      sid,
		}); err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}, nil
	}

	bucketName, hasBucketName := extractBucketName(input.ResourceBInfo)
	if !hasBucketName {
		return nil, fmt.Errorf(
			"bucket name could not be retrieved from the linked to %q S3 bucket resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	annotations := getBucketLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: s3AccessStatement(sid, bucketName, annotations.accessLevel),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q access to S3 bucket %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	output := accessLinkOutput(input, setupCtx.RoleResourceName, sid, bucketName, annotations.accessLevel, result)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-isolated caller reaches S3 through a gateway VPC endpoint; this is a no-op
	// for non-VPC functions.
	return linkutils.ActivateLinkNetworking(
		ctx,
		ec2Service,
		input,
		linkutils.NetworkingActivation{
			Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
			Region:       region,
			AWSService:   "s3",
			EndpointType: ec2types.VpcEndpointTypeGateway,
		},
		output,
	)
}

func bucketResourceARN(bucketName string) string {
	return fmt.Sprintf("arn:aws:s3:::%s", bucketName)
}

func bucketObjectsResourceARN(bucketName string) string {
	return fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)
}

func s3AccessStatement(sid, bucketName, accessLevel string) map[string]any {
	return map[string]any{
		"Sid":    sid,
		"Effect": "Allow",
		"Action": s3ActionsForAccessLevel(accessLevel),
		"Resource": []string{
			bucketResourceARN(bucketName),
			bucketObjectsResourceARN(bucketName),
		},
	}
}

func accessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, bucketName, accessLevel string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specS3AccessStatementNode(sid, bucketName, accessLevel),
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it:
	// inline placements map the statement by Sid; managed (overflow) placements map the
	// attached managed policy ARN.
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

func specS3AccessStatementNode(sid, bucketName, accessLevel string) *core.MappingNode {
	actions := s3ActionsForAccessLevel(accessLevel)
	actionItems := make([]*core.MappingNode, len(actions))
	for i, action := range actions {
		actionItems[i] = core.MappingNodeFromString(action)
	}
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{Items: actionItems},
		"resource", &core.MappingNode{Items: []*core.MappingNode{
			core.MappingNodeFromString(bucketResourceARN(bucketName)),
			core.MappingNodeFromString(bucketObjectsResourceARN(bucketName)),
		}},
	)
}

// Falls back to deriving the name from the bucket ARN for name-less
// (auto-named) buckets whose assigned name is not in state at link-update time.
func extractBucketName(bucketInfo *provider.ResourceInfo) (string, bool) {
	return linkutils.PhysicalResourceName(bucketInfo, "bucketName")
}

func bucketEnvVarName(userDefinedEnvVarName string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	return fmt.Sprintf("S3_BUCKET_%s", resourceInfo.ResourceName)
}

func createS3AccessSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("S3Access%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

func s3ActionsForAccessLevel(accessLevel string) []string {
	switch accessLevel {
	case "read":
		return []string{"s3:GetObject", "s3:ListBucket"}
	case "write":
		return []string{"s3:PutObject", "s3:DeleteObject"}
	case "readwrite":
		fallthrough
	default:
		return []string{"s3:GetObject", "s3:ListBucket", "s3:PutObject", "s3:DeleteObject"}
	}
}

type bucketLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
	accessLevel     string
}

func getBucketLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *bucketLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.s3.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.s3.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.s3.%s.envVarName", otherResourceInfo.ResourceName),
		},
	)

	accessLevel, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.s3.%s.accessLevel", otherResourceInfo.ResourceName),
			Default: "readwrite",
		},
	)

	return &bucketLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
		accessLevel:     accessLevel,
	}
}
