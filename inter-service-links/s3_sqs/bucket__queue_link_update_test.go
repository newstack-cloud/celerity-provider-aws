//go:build unit

package s3sqs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	s3mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/s3_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type BucketQueueLinkUpdateSuite struct {
	suite.Suite
}

const (
	bqBucketName = "orders"
	bqBucketARN  = "arn:aws:s3:::orders"
	bqQueueARN   = "arn:aws:sqs:us-west-2:123456789012:order-events"
	bqQueueURL   = "https://sqs.us-west-2.amazonaws.com/123456789012/order-events"
	bqInstanceID = "instance-1"
	bqPerEventID = "bluelink-ordersBucket-orderEventsQueue-s3ObjectCreated"
	bqResourceID = "ordersBucket__orderEventsQueue__s3-send-policy"
)

func bqBucketInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersBucket",
		InstanceID:   bqInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("bucketName", core.MappingNodeFromString(bqBucketName)),
		},
	}
}

func bqQueueInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "orderEventsQueue",
		InstanceID:   bqInstanceID,
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: core.MappingNodeFields(
					"aws.s3.sqs.event.0", core.MappingNodeFromString("s3:ObjectCreated:*"),
					"aws.s3.sqs.filterPrefix", core.MappingNodeFromString("incoming/"),
				),
			},
		},
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(bqQueueARN),
				"queueUrl", core.MappingNodeFromString(bqQueueURL),
			),
		},
	}
}

func (s *BucketQueueLinkUpdateSuite) Test_create_grants_permission_and_writes_notification() {
	loader := &testutils.MockAWSConfigLoader{}
	s3Svc := s3mock.CreateS3ServiceMock(
		s3mock.WithGetBucketNotificationConfigurationOutput(&s3.GetBucketNotificationConfigurationOutput{}),
		s3mock.WithPutBucketNotificationConfigurationOutput(&s3.PutBucketNotificationConfigurationOutput{}),
	)
	rs := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{}))

	out, err := testActions(loader, s3Svc).UpdateIntermediaryResources(
		context.Background(),
		&provider.LinkUpdateIntermediaryResourcesInput{
			LinkID:          "link-1",
			LinkUpdateType:  provider.LinkUpdateTypeCreate,
			ResourceAInfo:   bqBucketInfo(),
			ResourceBInfo:   bqQueueInfo(),
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
		},
	)
	s.Require().NoError(err)

	// The queue policy is deployed, scoped to the bucket ARN.
	s.Require().Len(rs.DeployCalls, 1)
	call := rs.DeployCalls[0]
	s.Equal("aws/sqs/queueInlinePolicy", call.ResourceType)
	s.Equal(bqResourceID, call.Input.DeployInput.ResourceID)
	spec := call.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(bqQueueURL, core.StringValue(spec.Fields["queue"]))
	statement := spec.Fields["policyDocument"].Fields["statement"].Items[0]
	s.Equal("s3.amazonaws.com", core.StringValue(statement.Fields["principal"].Fields["service"]))
	s.Equal("sqs:SendMessage", core.StringValue(statement.Fields["action"]))
	s.Equal(bqQueueARN, core.StringValue(statement.Fields["resource"]))
	sourceArn := statement.Fields["condition"].Fields["ArnLike"].Fields["aws:SourceArn"]
	s.Equal(bqBucketARN, core.StringValue(sourceArn))

	// The queue notification entry is written into the bucket's notification config.
	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*s3.PutBucketNotificationConfigurationInput)
		if !ok || in.NotificationConfiguration == nil || len(in.NotificationConfiguration.QueueConfigurations) != 1 {
			return false
		}
		entry := in.NotificationConfiguration.QueueConfigurations[0]
		return aws.ToString(in.Bucket) == bqBucketName &&
			aws.ToString(entry.Id) == bqPerEventID &&
			aws.ToString(entry.QueueArn) == bqQueueARN &&
			len(entry.Events) == 1 && entry.Events[0] == s3types.EventS3ObjectCreated &&
			entry.Filter != nil
	})

	s.Require().Len(out.IntermediaryResourceStates, 1)

	// The entry is claimed on the bucket via a compound (queue ARN AND event) mapping
	// so the bucket's drift/update does not strip it.
	mappingKey := "ordersBucket::spec.notificationConfiguration.queueConfigurations" +
		"[@.queue = \"" + bqQueueARN + "\" && @.event = \"s3:ObjectCreated:*\"]"
	s.Contains(out.ResourceDataMappings, mappingKey)
}

// A read-modify-write preserves other links' notification entries and replaces this
// link's own entry (by id) rather than duplicating it.
func (s *BucketQueueLinkUpdateSuite) Test_create_preserves_foreign_entries() {
	loader := &testutils.MockAWSConfigLoader{}
	s3Svc := s3mock.CreateS3ServiceMock(
		s3mock.WithGetBucketNotificationConfigurationOutput(&s3.GetBucketNotificationConfigurationOutput{
			QueueConfigurations: []s3types.QueueConfiguration{
				{Id: aws.String("other-link"), QueueArn: aws.String("arn:aws:sqs:us-west-2:123456789012:other")},
				{Id: aws.String(bqPerEventID), QueueArn: aws.String("arn:aws:sqs:us-west-2:123456789012:stale")},
			},
		}),
		s3mock.WithPutBucketNotificationConfigurationOutput(&s3.PutBucketNotificationConfigurationOutput{}),
	)
	rs := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{}))

	_, err := testActions(loader, s3Svc).UpdateIntermediaryResources(
		context.Background(),
		&provider.LinkUpdateIntermediaryResourcesInput{
			LinkID:          "link-1",
			LinkUpdateType:  provider.LinkUpdateTypeCreate,
			ResourceAInfo:   bqBucketInfo(),
			ResourceBInfo:   bqQueueInfo(),
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
		},
	)
	s.Require().NoError(err)

	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in := arg.(*s3.PutBucketNotificationConfigurationInput)
		entries := in.NotificationConfiguration.QueueConfigurations
		if len(entries) != 2 {
			return false
		}
		var foreign, own *s3types.QueueConfiguration
		for i := range entries {
			switch aws.ToString(entries[i].Id) {
			case "other-link":
				foreign = &entries[i]
			case bqPerEventID:
				own = &entries[i]
			}
		}
		// The foreign entry is preserved untouched; the link's own entry is refreshed.
		return foreign != nil && own != nil &&
			aws.ToString(own.QueueArn) == bqQueueARN
	})
}

func (s *BucketQueueLinkUpdateSuite) Test_destroy_removes_notification_and_policy() {
	loader := &testutils.MockAWSConfigLoader{}
	s3Svc := s3mock.CreateS3ServiceMock(
		s3mock.WithGetBucketNotificationConfigurationOutput(&s3.GetBucketNotificationConfigurationOutput{
			QueueConfigurations: []s3types.QueueConfiguration{
				{Id: aws.String(bqPerEventID), QueueArn: aws.String(bqQueueARN)},
			},
		}),
		s3mock.WithPutBucketNotificationConfigurationOutput(&s3.PutBucketNotificationConfigurationOutput{}),
	)
	rs := resourceservicemock.Create()

	_, err := testActions(loader, s3Svc).UpdateIntermediaryResources(
		context.Background(),
		&provider.LinkUpdateIntermediaryResourcesInput{
			LinkID:          "link-1",
			LinkUpdateType:  provider.LinkUpdateTypeDestroy,
			ResourceAInfo:   bqBucketInfo(),
			ResourceBInfo:   bqQueueInfo(),
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
			CurrentLinkState: &state.LinkState{
				IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
					{ResourceID: bqResourceID, ResourceType: "aws/sqs/queueInlinePolicy", InstanceID: bqInstanceID},
				},
			},
		},
	)
	s.Require().NoError(err)

	s.Require().Len(rs.DestroyCalls, 1)
	s.Equal(bqResourceID, rs.DestroyCalls[0].Input.ResourceID)
	// The link's notification entry is removed (written back without it).
	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in := arg.(*s3.PutBucketNotificationConfigurationInput)
		return len(in.NotificationConfiguration.QueueConfigurations) == 0
	})
}

func TestBucketQueueLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(BucketQueueLinkUpdateSuite))
}
