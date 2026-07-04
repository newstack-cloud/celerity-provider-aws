//go:build unit

package s3sns

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

type BucketTopicLinkUpdateSuite struct {
	suite.Suite
}

const (
	btBucketName = "orders"
	btBucketARN  = "arn:aws:s3:::orders"
	btTopicARN   = "arn:aws:sns:us-west-2:123456789012:order-events"
	btInstanceID = "instance-1"
	btPerEventID = "bluelink-ordersBucket-orderEventsTopic-s3ObjectCreated"
	btResourceID = "ordersBucket__orderEventsTopic__s3-publish-policy"
)

func btBucketInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersBucket",
		InstanceID:   btInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("bucketName", core.MappingNodeFromString(btBucketName)),
		},
	}
}

func btTopicInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "orderEventsTopic",
		InstanceID:   btInstanceID,
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: core.MappingNodeFields(
					"aws.s3.sns.event.0", core.MappingNodeFromString("s3:ObjectCreated:*"),
					"aws.s3.sns.filterPrefix", core.MappingNodeFromString("incoming/"),
				),
			},
		},
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("topicArn", core.MappingNodeFromString(btTopicARN)),
		},
	}
}

func (s *BucketTopicLinkUpdateSuite) Test_create_grants_policy_and_writes_notification() {
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
			ResourceAInfo:   btBucketInfo(),
			ResourceBInfo:   btTopicInfo(),
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
		},
	)
	s.Require().NoError(err)

	// The topic publish policy is deployed, scoped to the bucket ARN.
	s.Require().Len(rs.DeployCalls, 1)
	call := rs.DeployCalls[0]
	s.Equal("aws/sns/topicInlinePolicy", call.ResourceType)
	s.Equal(btResourceID, call.Input.DeployInput.ResourceID)
	spec := call.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(btTopicARN, core.StringValue(spec.Fields["topicArn"]))
	statements := spec.Fields["policyDocument"].Fields["statement"].Items
	s.Require().Len(statements, 1)
	s.Equal(
		"s3.amazonaws.com",
		core.StringValue(statements[0].Fields["principal"].Fields["service"]),
	)
	s.Equal("sns:Publish", core.StringValue(statements[0].Fields["action"]))
	s.Equal(btTopicARN, core.StringValue(statements[0].Fields["resource"]))
	s.Equal(
		btBucketARN,
		core.StringValue(statements[0].Fields["condition"].Fields["ArnLike"].Fields["aws:SourceArn"]),
	)

	// The topic notification entry is written into the bucket's notification config.
	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*s3.PutBucketNotificationConfigurationInput)
		if !ok || in.NotificationConfiguration == nil || len(in.NotificationConfiguration.TopicConfigurations) != 1 {
			return false
		}
		entry := in.NotificationConfiguration.TopicConfigurations[0]
		return aws.ToString(in.Bucket) == btBucketName &&
			aws.ToString(entry.Id) == btPerEventID &&
			aws.ToString(entry.TopicArn) == btTopicARN &&
			len(entry.Events) == 1 && entry.Events[0] == s3types.EventS3ObjectCreated &&
			entry.Filter != nil
	})

	s.Require().Len(out.IntermediaryResourceStates, 1)

	// The entry is claimed on the bucket via a compound (topic ARN AND event) mapping
	// so the bucket's drift/update does not strip it.
	mappingKey := "ordersBucket::spec.notificationConfiguration.topicConfigurations" +
		"[@.topic = \"" + btTopicARN + "\" && @.event = \"s3:ObjectCreated:*\"]"
	s.Contains(out.ResourceDataMappings, mappingKey)
}

// A read-modify-write preserves other links' notification entries and replaces this
// link's own entry (by id) rather than duplicating it.
func (s *BucketTopicLinkUpdateSuite) Test_create_preserves_foreign_entries() {
	loader := &testutils.MockAWSConfigLoader{}
	s3Svc := s3mock.CreateS3ServiceMock(
		s3mock.WithGetBucketNotificationConfigurationOutput(&s3.GetBucketNotificationConfigurationOutput{
			TopicConfigurations: []s3types.TopicConfiguration{
				{Id: aws.String("other-link"), TopicArn: aws.String("arn:aws:sns:us-west-2:123456789012:other")},
				{Id: aws.String(btPerEventID), TopicArn: aws.String("arn:aws:sns:us-west-2:123456789012:stale")},
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
			ResourceAInfo:   btBucketInfo(),
			ResourceBInfo:   btTopicInfo(),
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
		},
	)
	s.Require().NoError(err)

	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in := arg.(*s3.PutBucketNotificationConfigurationInput)
		entries := in.NotificationConfiguration.TopicConfigurations
		if len(entries) != 2 {
			return false
		}
		var foreign, own *s3types.TopicConfiguration
		for i := range entries {
			switch aws.ToString(entries[i].Id) {
			case "other-link":
				foreign = &entries[i]
			case btPerEventID:
				own = &entries[i]
			}
		}
		// The foreign entry is preserved untouched; the link's own entry is refreshed.
		return foreign != nil && own != nil &&
			aws.ToString(own.TopicArn) == btTopicARN
	})
}

func (s *BucketTopicLinkUpdateSuite) Test_destroy_removes_notification_and_policy() {
	loader := &testutils.MockAWSConfigLoader{}
	s3Svc := s3mock.CreateS3ServiceMock(
		s3mock.WithGetBucketNotificationConfigurationOutput(&s3.GetBucketNotificationConfigurationOutput{
			TopicConfigurations: []s3types.TopicConfiguration{
				{Id: aws.String(btPerEventID), TopicArn: aws.String(btTopicARN)},
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
			ResourceAInfo:   btBucketInfo(),
			ResourceBInfo:   btTopicInfo(),
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
			CurrentLinkState: &state.LinkState{
				IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
					{ResourceID: btResourceID, ResourceType: "aws/sns/topicInlinePolicy", InstanceID: btInstanceID},
				},
			},
		},
	)
	s.Require().NoError(err)

	s.Require().Len(rs.DestroyCalls, 1)
	s.Equal(btResourceID, rs.DestroyCalls[0].Input.ResourceID)
	// The link's notification entry is removed (written back without it).
	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in := arg.(*s3.PutBucketNotificationConfigurationInput)
		return len(in.NotificationConfiguration.TopicConfigurations) == 0
	})
}

func TestBucketTopicLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(BucketTopicLinkUpdateSuite))
}
