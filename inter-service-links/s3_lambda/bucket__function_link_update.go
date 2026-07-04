package s3lambda

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

const lambdaAnnotationPrefix = "aws.s3.lambda"

func (l *bucketFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The bucket's notification configuration is written in the intermediary phase, the
	// only phase with access to the resource service (for the bucket lock and the managed
	// permission), and after the permission is granted.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *bucketFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The function is not modified by this link; the invoke permission is a separate,
	// link-owned aws/lambda/permission intermediary handled in UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

// UpdateIntermediaryResources grants S3 permission to invoke the function and merges this
// link's lambda entry into the bucket's notification configuration. The permission is
// granted before the notification entry is written, since S3 validates the destination
// permission when the configuration is applied.
func (l *bucketFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	instanceID := pluginutils.GetInstanceID(input.ResourceAInfo)

	bucketName, hasBucketName := bucketName(input.ResourceAInfo)
	if !hasBucketName {
		return nil, fmt.Errorf(
			"bucket name could not be retrieved from the S3 bucket %q",
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}
	bucketARN := fmt.Sprintf("arn:aws:s3:::%s", bucketName)
	entryID := notificationEntryID(input.ResourceAInfo, input.ResourceBInfo)

	s3Service, err := l.getS3Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// Serialise notification read-modify-writes on the bucket across links.
	if err := input.ResourceService.AcquireResourceLock(
		ctx,
		&provider.AcquireResourceLockInput{
			InstanceID:      instanceID,
			ResourceName:    input.ResourceAInfo.ResourceName,
			ProviderContext: providerCtx,
			AcquiredBy:      input.LinkID,
		},
	); err != nil {
		return nil, err
	}

	identity := permissionIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)
	priorState := linkutils.FindIntermediaryState(input.CurrentLinkState, identity.ResourceID)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if err := linkutils.RemoveS3Notification(ctx, s3Service, bucketName, entryID); err != nil {
			return nil, err
		}

		if err := linkutils.DestroyManagedIntermediary(
			ctx, input.ResourceService, instanceID, providerCtx, priorState,
		); err != nil {
			return nil, err
		}

		return &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	functionARN, hasFunctionARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the Lambda function %q",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	intermediaryState, err := linkutils.DeployManagedIntermediary(
		ctx,
		input.ResourceService,
		instanceID,
		input.InstanceName,
		providerCtx,
		priorState,
		linkutils.ManagedIntermediary{
			ResourceType: identity.ResourceType,
			ResourceID:   identity.ResourceID,
			ResourceName: identity.ResourceName,
			Spec: core.MappingNodeFields(
				"functionName", core.MappingNodeFromString(functionARN),
				"action", core.MappingNodeFromString("lambda:InvokeFunction"),
				"principal", core.MappingNodeFromString("s3.amazonaws.com"),
				"sourceArn", core.MappingNodeFromString(bucketARN),
			),
		},
	)
	if err != nil {
		return nil, err
	}

	events := linkutils.S3NotificationEvents(input.ResourceBInfo, lambdaAnnotationPrefix)
	prefix, suffix := linkutils.S3KeyFilterParts(input.ResourceBInfo, lambdaAnnotationPrefix)
	if err := linkutils.PutS3Notification(
		ctx, s3Service, linkutils.S3LambdaTarget, bucketName, entryID, functionARN, events, prefix, suffix,
	); err != nil {
		return nil, err
	}

	// Claim the link's notification entries on the bucket so its drift/update does not
	// strip them. Each per-event entry is mapped by its compound (function ARN AND event)
	// selector onto the bucket's notification configuration.
	specEntries := linkutils.S3NotificationSpecEntries(linkutils.S3LambdaTarget, functionARN, events, prefix, suffix)
	notificationLinkData, mappings := linkutils.S3NotificationLinkData(
		input.ResourceAInfo.ResourceName, linkutils.S3LambdaTarget, specEntries,
	)

	intermediaryLinkData := linkutils.IntermediaryLinkData(linkutils.DeployedIntermediary{
		Identity: identity,
		Leaves: map[string]*core.MappingNode{
			"sourceArn":   core.MappingNodeFromString(bucketARN),
			"functionArn": core.MappingNodeFromString(functionARN),
		},
	})

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:                   core.MergeMaps(intermediaryLinkData, notificationLinkData),
		ResourceDataMappings:       mappings,
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{intermediaryState},
	}, nil
}

func bucketName(bucketInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(bucketInfo)
	nameNode, has := pluginutils.GetValueByPath("$.bucketName", spec)
	if !has {
		return "", false
	}
	return core.StringValue(nameNode), true
}

func notificationEntryID(bucketInfo, functionInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"bluelink-%s-%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(bucketInfo)),
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(functionInfo)),
	)
}

func permissionIntermediaryIdentity(bucketInfo, functionInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/lambda/permission",
		ResourceID: fmt.Sprintf(
			"%s__%s__s3-invoke-permission",
			pluginutils.GetResourceName(bucketInfo),
			pluginutils.GetResourceName(functionInfo),
		),
		ResourceName: fmt.Sprintf(
			"%sInvoke%s",
			pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(bucketInfo)),
			pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(functionInfo)),
		),
	}
}
