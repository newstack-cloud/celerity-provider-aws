//go:build unit

package s3lambda

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	s3mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/s3_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type BucketFunctionLinkUpdateSuite struct {
	suite.Suite
}

const (
	bfBucketName  = "orders"
	bfBucketARN   = "arn:aws:s3:::orders"
	bfFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:process-order"
	bfInstanceID  = "instance-1"
	bfEntryID     = "bluelink-ordersBucket-processOrderFunction"
	bfPerEventID  = "bluelink-ordersBucket-processOrderFunction-s3ObjectCreated"
	bfResourceID  = "ordersBucket__processOrderFunction__s3-invoke-permission"
)

func bfBucketInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersBucket",
		InstanceID:   bfInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("bucketName", core.MappingNodeFromString(bfBucketName)),
		},
	}
}

func bfFunctionInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "processOrderFunction",
		InstanceID:   bfInstanceID,
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: core.MappingNodeFields(
					"aws.s3.lambda.event.0", core.MappingNodeFromString("s3:ObjectCreated:*"),
					"aws.s3.lambda.filterPrefix", core.MappingNodeFromString("incoming/"),
				),
			},
		},
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("arn", core.MappingNodeFromString(bfFunctionARN)),
		},
	}
}

func (s *BucketFunctionLinkUpdateSuite) Test_create_grants_permission_and_writes_notification() {
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
			ResourceAInfo:   bfBucketInfo(),
			ResourceBInfo:   bfFunctionInfo(),
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
		},
	)
	s.Require().NoError(err)

	// The Lambda invoke permission is deployed, scoped to the bucket ARN.
	s.Require().Len(rs.DeployCalls, 1)
	call := rs.DeployCalls[0]
	s.Equal("aws/lambda/permission", call.ResourceType)
	s.Equal(bfResourceID, call.Input.DeployInput.ResourceID)
	spec := call.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(bfFunctionARN, core.StringValue(spec.Fields["functionName"]))
	s.Equal("s3.amazonaws.com", core.StringValue(spec.Fields["principal"]))
	s.Equal(bfBucketARN, core.StringValue(spec.Fields["sourceArn"]))

	// The lambda notification entry is written into the bucket's notification config.
	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*s3.PutBucketNotificationConfigurationInput)
		if !ok || in.NotificationConfiguration == nil || len(in.NotificationConfiguration.LambdaFunctionConfigurations) != 1 {
			return false
		}
		entry := in.NotificationConfiguration.LambdaFunctionConfigurations[0]
		return aws.ToString(in.Bucket) == bfBucketName &&
			aws.ToString(entry.Id) == bfPerEventID &&
			aws.ToString(entry.LambdaFunctionArn) == bfFunctionARN &&
			len(entry.Events) == 1 && entry.Events[0] == s3types.EventS3ObjectCreated &&
			entry.Filter != nil
	})

	s.Require().Len(out.IntermediaryResourceStates, 1)

	// The entry is claimed on the bucket via a compound (function ARN AND event) mapping
	// so the bucket's drift/update does not strip it.
	mappingKey := "ordersBucket::spec.notificationConfiguration.lambdaConfigurations" +
		"[@.function = \"" + bfFunctionARN + "\" && @.event = \"s3:ObjectCreated:*\"]"
	s.Contains(out.ResourceDataMappings, mappingKey)
}

// A read-modify-write preserves other links' notification entries and replaces this
// link's own entry (by id) rather than duplicating it.
func (s *BucketFunctionLinkUpdateSuite) Test_create_preserves_foreign_entries() {
	loader := &testutils.MockAWSConfigLoader{}
	s3Svc := s3mock.CreateS3ServiceMock(
		s3mock.WithGetBucketNotificationConfigurationOutput(&s3.GetBucketNotificationConfigurationOutput{
			LambdaFunctionConfigurations: []s3types.LambdaFunctionConfiguration{
				{Id: aws.String("other-link"), LambdaFunctionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:function:other")},
				{Id: aws.String(bfPerEventID), LambdaFunctionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:function:stale")},
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
			ResourceAInfo:   bfBucketInfo(),
			ResourceBInfo:   bfFunctionInfo(),
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
		},
	)
	s.Require().NoError(err)

	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in := arg.(*s3.PutBucketNotificationConfigurationInput)
		entries := in.NotificationConfiguration.LambdaFunctionConfigurations
		if len(entries) != 2 {
			return false
		}
		var foreign, own *s3types.LambdaFunctionConfiguration
		for i := range entries {
			switch aws.ToString(entries[i].Id) {
			case "other-link":
				foreign = &entries[i]
			case bfPerEventID:
				own = &entries[i]
			}
		}
		// The foreign entry is preserved untouched; the link's own entry is refreshed.
		return foreign != nil && own != nil &&
			aws.ToString(own.LambdaFunctionArn) == bfFunctionARN
	})
}

func (s *BucketFunctionLinkUpdateSuite) Test_destroy_removes_notification_and_permission() {
	loader := &testutils.MockAWSConfigLoader{}
	s3Svc := s3mock.CreateS3ServiceMock(
		s3mock.WithGetBucketNotificationConfigurationOutput(&s3.GetBucketNotificationConfigurationOutput{
			LambdaFunctionConfigurations: []s3types.LambdaFunctionConfiguration{
				{Id: aws.String(bfPerEventID), LambdaFunctionArn: aws.String(bfFunctionARN)},
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
			ResourceAInfo:   bfBucketInfo(),
			ResourceBInfo:   bfFunctionInfo(),
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
			CurrentLinkState: &state.LinkState{
				IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
					{ResourceID: bfResourceID, ResourceType: "aws/lambda/permission", InstanceID: bfInstanceID},
				},
			},
		},
	)
	s.Require().NoError(err)

	s.Require().Len(rs.DestroyCalls, 1)
	s.Equal(bfResourceID, rs.DestroyCalls[0].Input.ResourceID)
	// The link's notification entry is removed (written back without it).
	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in := arg.(*s3.PutBucketNotificationConfigurationInput)
		return len(in.NotificationConfiguration.LambdaFunctionConfigurations) == 0
	})
}

// Multiple events configured on the link produce one notification entry (and one
// compound drift mapping) per event, each uniquely identified by (function ARN, event).
func (s *BucketFunctionLinkUpdateSuite) Test_create_writes_one_entry_per_event() {
	loader := &testutils.MockAWSConfigLoader{}
	s3Svc := s3mock.CreateS3ServiceMock(
		s3mock.WithGetBucketNotificationConfigurationOutput(&s3.GetBucketNotificationConfigurationOutput{}),
		s3mock.WithPutBucketNotificationConfigurationOutput(&s3.PutBucketNotificationConfigurationOutput{}),
	)
	rs := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{}))

	functionInfo := bfFunctionInfo()
	functionInfo.ResourceWithResolvedSubs.Metadata.Annotations = core.MappingNodeFields(
		"aws.s3.lambda.event.0", core.MappingNodeFromString("s3:ObjectCreated:*"),
		"aws.s3.lambda.event.1", core.MappingNodeFromString("s3:ObjectRemoved:*"),
	)

	out, err := testActions(loader, s3Svc).UpdateIntermediaryResources(
		context.Background(),
		&provider.LinkUpdateIntermediaryResourcesInput{
			LinkID:          "link-1",
			LinkUpdateType:  provider.LinkUpdateTypeCreate,
			ResourceAInfo:   bfBucketInfo(),
			ResourceBInfo:   functionInfo,
			ResourceService: rs,
			LinkContext:     testLinkContext(),
			InstanceName:    "instance",
		},
	)
	s.Require().NoError(err)

	// Two notification entries are written, one per event, with distinct per-event ids.
	s3Svc.AssertCalledWith(&s.Suite, "PutBucketNotificationConfiguration", 0, plugintestutils.Any, func(arg any) bool {
		in := arg.(*s3.PutBucketNotificationConfigurationInput)
		entries := in.NotificationConfiguration.LambdaFunctionConfigurations
		if len(entries) != 2 {
			return false
		}
		ids := map[string]string{} // id -> event
		for _, e := range entries {
			if len(e.Events) != 1 || aws.ToString(e.LambdaFunctionArn) != bfFunctionARN {
				return false
			}
			ids[aws.ToString(e.Id)] = string(e.Events[0])
		}
		return ids[bfEntryID+"-s3ObjectCreated"] == "s3:ObjectCreated:*" &&
			ids[bfEntryID+"-s3ObjectRemoved"] == "s3:ObjectRemoved:*"
	})

	// One compound (function ARN AND event) drift mapping is emitted per event.
	s.Contains(out.ResourceDataMappings,
		"ordersBucket::spec.notificationConfiguration.lambdaConfigurations"+
			"[@.function = \""+bfFunctionARN+"\" && @.event = \"s3:ObjectCreated:*\"]")
	s.Contains(out.ResourceDataMappings,
		"ordersBucket::spec.notificationConfiguration.lambdaConfigurations"+
			"[@.function = \""+bfFunctionARN+"\" && @.event = \"s3:ObjectRemoved:*\"]")
}

func TestBucketFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(BucketFunctionLinkUpdateSuite))
}
