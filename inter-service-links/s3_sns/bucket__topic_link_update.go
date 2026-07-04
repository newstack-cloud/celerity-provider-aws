package s3sns

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

const snsAnnotationPrefix = "aws.s3.sns"

func (l *bucketTopicLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The bucket's notification configuration is written in the intermediary phase, the
	// only phase with access to the resource service (for the bucket lock and the managed
	// topic policy), and after the topic policy is set.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *bucketTopicLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The topic is not modified by this link; the publish permission is a separate,
	// link-owned aws/sns/topicInlinePolicy intermediary handled in
	// UpdateIntermediaryResources.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

// UpdateIntermediaryResources grants S3 permission to publish to the topic and merges this
// link's topic entry into the bucket's notification configuration. The policy is set before
// the notification entry is written, since S3 validates the destination permission when the
// configuration is applied.
func (l *bucketTopicLinkActions) UpdateIntermediaryResources(
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

	identity := topicPolicyIntermediaryIdentity(input.ResourceAInfo, input.ResourceBInfo)
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

	topicARN, hasTopicARN := topicARN(input.ResourceBInfo)
	if !hasTopicARN {
		return nil, fmt.Errorf(
			"topic ARN could not be retrieved from the SNS topic %q",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	sid := topicStatementID(input.ResourceAInfo)
	policyDocument := buildTopicPolicyDocument(sid, topicARN, bucketARN)

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
				"topicArn", core.MappingNodeFromString(topicARN),
				"policyDocument", policyDocument,
			),
		},
	)
	if err != nil {
		return nil, err
	}

	events := linkutils.S3NotificationEvents(input.ResourceBInfo, snsAnnotationPrefix)
	prefix, suffix := linkutils.S3KeyFilterParts(input.ResourceBInfo, snsAnnotationPrefix)
	if err := linkutils.PutS3Notification(
		ctx, s3Service, linkutils.S3TopicTarget, bucketName, entryID, topicARN, events, prefix, suffix,
	); err != nil {
		return nil, err
	}

	// Claim the link's notification entries on the bucket so its drift/update does not
	// strip them. Each per-event entry is mapped by its compound (topic ARN AND event)
	// selector onto the bucket's notification configuration.
	specEntries := linkutils.S3NotificationSpecEntries(linkutils.S3TopicTarget, topicARN, events, prefix, suffix)
	notificationLinkData, mappings := linkutils.S3NotificationLinkData(
		input.ResourceAInfo.ResourceName, linkutils.S3TopicTarget, specEntries,
	)

	intermediaryLinkData := linkutils.IntermediaryLinkData(linkutils.DeployedIntermediary{
		Identity: identity,
		Leaves: map[string]*core.MappingNode{
			"sourceArn": core.MappingNodeFromString(bucketARN),
			"topicArn":  core.MappingNodeFromString(topicARN),
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

func topicARN(topicInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(topicInfo)
	arnNode, has := pluginutils.GetValueByPath("$.topicArn", spec)
	if !has {
		return "", false
	}
	return core.StringValue(arnNode), true
}

func notificationEntryID(bucketInfo, topicInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"bluelink-%s-%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(bucketInfo)),
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(topicInfo)),
	)
}

func buildTopicPolicyDocument(sid, topicARN, bucketARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"version", core.MappingNodeFromString("2012-10-17"),
		"statement", core.MappingNodeItems(
			core.MappingNodeFields(
				"sid", core.MappingNodeFromString(sid),
				"effect", core.MappingNodeFromString("Allow"),
				"principal", core.MappingNodeFields(
					"service", core.MappingNodeFromString("s3.amazonaws.com"),
				),
				"action", core.MappingNodeFromString("sns:Publish"),
				"resource", core.MappingNodeFromString(topicARN),
				"condition", core.MappingNodeFields(
					"ArnLike", core.MappingNodeFields(
						"aws:SourceArn", core.MappingNodeFromString(bucketARN),
					),
				),
			),
		),
	)
}

func topicStatementID(bucketInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"S3%s",
		pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(bucketInfo)),
	)
}

func topicPolicyIntermediaryIdentity(bucketInfo, topicInfo *provider.ResourceInfo) linkutils.IntermediaryIdentity {
	return linkutils.IntermediaryIdentity{
		ResourceType: "aws/sns/topicInlinePolicy",
		ResourceID: fmt.Sprintf(
			"%s__%s__s3-publish-policy",
			pluginutils.GetResourceName(bucketInfo),
			pluginutils.GetResourceName(topicInfo),
		),
		ResourceName: fmt.Sprintf(
			"%sPublish%s",
			pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(bucketInfo)),
			pluginutils.StripNonAlphaNumericChars(pluginutils.GetResourceName(topicInfo)),
		),
	}
}
