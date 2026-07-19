package s3sqs

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

const sqsAnnotationPrefix = "aws.s3.sqs"

func (l *bucketQueueLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The bucket's notification configuration is written in the intermediary phase, the
	// only phase with access to the resource service (for the bucket lock and the managed
	// queue policy), and after the queue policy is set.
	return &provider.LinkUpdateResourceOutput{LinkData: core.MappingNodeFields()}, nil
}

func (l *bucketQueueLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The queue is not modified by this link; the send-message permission is a separate,
	// link-owned aws/sqs/queueInlinePolicy intermediary handled in
	// UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{LinkData: core.MappingNodeFields()}, nil
}

// UpdateIntermediaryResources grants S3 permission to send messages to the queue and
// merges this link's queue entry into the bucket's notification configuration. The queue
// policy is set before the notification entry is written, since S3 validates the
// destination permission when the configuration is applied.
func (l *bucketQueueLinkActions) UpdateIntermediaryResources(
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

	identity := queuePolicyIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)
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
		return &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}, nil
	}

	queueARN, hasQueueARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasQueueARN {
		return nil, fmt.Errorf(
			"queue ARN could not be retrieved from the SQS queue %q",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	queueSpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(input.ResourceBInfo)
	queueURLNode, hasQueueURL := pluginutils.GetValueByPath("$.queueUrl", queueSpec)
	if !hasQueueURL {
		return nil, fmt.Errorf(
			"queue URL could not be retrieved from the SQS queue %q",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}
	queueURL := core.StringValue(queueURLNode)

	sid := s3StatementID(input.ResourceAInfo)
	policyDocument := buildQueuePolicyDocument(sid, queueARN, bucketARN)

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
				"queue", core.MappingNodeFromString(queueURL),
				"policyDocument", policyDocument,
			),
		},
	)
	if err != nil {
		return nil, err
	}

	events := linkutils.S3NotificationEvents(input.ResourceBInfo, sqsAnnotationPrefix)
	prefix, suffix := linkutils.S3KeyFilterParts(input.ResourceBInfo, sqsAnnotationPrefix)
	if err := linkutils.PutS3Notification(
		ctx, s3Service, linkutils.S3QueueTarget, bucketName, entryID, queueARN, events, prefix, suffix,
	); err != nil {
		return nil, err
	}

	// Claim the link's notification entries on the bucket so its drift/update does not
	// strip them. Each per-event entry is mapped by its compound (queue ARN AND event)
	// selector onto the bucket's notification configuration.
	specEntries := linkutils.S3NotificationSpecEntries(linkutils.S3QueueTarget, queueARN, events, prefix, suffix)
	notificationLinkData, mappings := linkutils.S3NotificationLinkData(
		input.ResourceAInfo.ResourceName, linkutils.S3QueueTarget, specEntries,
	)

	intermediaryLinkData := linkutils.IntermediaryLinkData(linkutils.DeployedIntermediary{
		Identity: identity,
		Leaves: map[string]*core.MappingNode{
			"sourceArn": core.MappingNodeFromString(bucketARN),
			"queueArn":  core.MappingNodeFromString(queueARN),
		},
	})

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:                   core.MergeMaps(intermediaryLinkData, notificationLinkData),
		ResourceDataMappings:       mappings,
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{intermediaryState},
	}, nil
}

// Falls back to deriving the name from the bucket ARN for name-less
// (auto-named) buckets whose assigned name is not in state at link-update time.
func bucketName(bucketInfo *provider.ResourceInfo) (string, bool) {
	return linkutils.PhysicalResourceName(bucketInfo, "bucketName")
}

func notificationEntryID(bucketInfo, queueInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"bluelink-%s-%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(bucketInfo)),
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(queueInfo)),
	)
}

func buildQueuePolicyDocument(sid, queueARN, bucketARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"version", core.MappingNodeFromString("2012-10-17"),
		"statement", core.MappingNodeItems(
			core.MappingNodeFields(
				"sid", core.MappingNodeFromString(sid),
				"effect", core.MappingNodeFromString("Allow"),
				"principal", core.MappingNodeFields(
					"service", core.MappingNodeFromString("s3.amazonaws.com"),
				),
				"action", core.MappingNodeFromString("sqs:SendMessage"),
				"resource", core.MappingNodeFromString(queueARN),
				"condition", core.MappingNodeFields(
					"ArnLike", core.MappingNodeFields(
						"aws:SourceArn", core.MappingNodeFromString(bucketARN),
					),
				),
			),
		),
	)
}

func s3StatementID(bucketInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"S3%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(bucketInfo)),
	)
}

func queuePolicyIntermediaryIdentity(bucketInfo, queueInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/sqs/queueInlinePolicy",
		ResourceID: fmt.Sprintf(
			"%s__%s__s3-send-policy",
			pluginutils.GetResourceName(bucketInfo),
			pluginutils.GetResourceName(queueInfo),
		),
		ResourceName: fmt.Sprintf(
			"%sSend%s",
			pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(bucketInfo)),
			pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(queueInfo)),
		),
	}
}
