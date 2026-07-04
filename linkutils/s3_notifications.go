package linkutils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	s3service "github.com/newstack-cloud/bluelink-provider-aws/services/s3/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// S3NotificationEvents reads the indexed event annotations (aws.{prefix}.event.<index>)
// from the target resource in order, stopping at the first absent index. When none are
// set it defaults to s3:ObjectCreated:*.
func S3NotificationEvents(resourceInfo *provider.ResourceInfo, annotationPrefix string) []string {
	var events []string
	for i := 0; ; i++ {
		value, found := pluginutils.GetStringAnnotation(
			resourceInfo,
			&pluginutils.AnnotationQuery[string]{
				Key: fmt.Sprintf("%s.event.%d", annotationPrefix, i),
			},
		)
		if !found || value == "" {
			break
		}
		events = append(events, value)
	}
	if len(events) == 0 {
		events = []string{string(s3types.EventS3ObjectCreated)}
	}
	return events
}

// S3KeyFilterParts reads the prefix/suffix filter annotations (aws.{prefix}.filterPrefix
// and aws.{prefix}.filterSuffix).
func S3KeyFilterParts(resourceInfo *provider.ResourceInfo, annotationPrefix string) (prefix, suffix string) {
	prefix, _ = pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("%s.filterPrefix", annotationPrefix),
		},
	)
	suffix, _ = pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("%s.filterSuffix", annotationPrefix),
		},
	)
	return prefix, suffix
}

// S3NotificationTarget identifies the destination type of a bucket notification and the
// spec/SDK field that holds the destination ARN.
type S3NotificationTarget struct {
	// ConfigField is the notification-configuration array on the bucket spec, e.g.
	// "lambdaConfigurations", "queueConfigurations" or "topicConfigurations".
	ConfigField string
	// ARNField is the spec attribute holding the destination ARN within an entry, e.g.
	// "function", "queue" or "topic".
	ARNField string
}

var (
	// S3LambdaTarget describes a Lambda function notification destination.
	S3LambdaTarget = S3NotificationTarget{
		ConfigField: "lambdaConfigurations",
		ARNField:    "function",
	}
	// S3QueueTarget describes an SQS queue notification destination.
	S3QueueTarget = S3NotificationTarget{
		ConfigField: "queueConfigurations",
		ARNField:    "queue",
	}
	// S3TopicTarget describes an SNS topic notification destination.
	S3TopicTarget = S3NotificationTarget{
		ConfigField: "topicConfigurations",
		ARNField:    "topic",
	}
)

// S3NotificationSpecEntry is one per-event notification entry in the bucket-spec
// shape, together with the compound selector that uniquely identifies it
// within the bucket's notification array (by destination ARN AND event).
type S3NotificationSpecEntry struct {
	Event    string
	SpecNode *core.MappingNode
	// Selector is the array-item selector (e.g.
	// `[@.function = "<arn>" && @.event = "<event>"]`) that targets this entry.
	Selector string
}

// S3NotificationSpecEntries builds the per-event spec-shaped entries for a destination,
// used both to record the link's contribution in link data and to map it onto the
// bucket's notification configuration for drift suppression.
func S3NotificationSpecEntries(
	target S3NotificationTarget,
	targetARN string,
	events []string,
	prefix, suffix string,
) []S3NotificationSpecEntry {
	filterNode := specFilterNode(prefix, suffix)
	entries := make([]S3NotificationSpecEntry, 0, len(events))
	for _, event := range events {
		fields := []any{
			"event", core.MappingNodeFromString(event),
			target.ARNField, core.MappingNodeFromString(targetARN),
		}
		if filterNode != nil {
			fields = append(fields, "filter", filterNode)
		}
		entries = append(entries, S3NotificationSpecEntry{
			Event:    event,
			SpecNode: core.MappingNodeFields(fields...),
			Selector: fmt.Sprintf(
				"[@.%s = %q && @.event = %q]",
				target.ARNField, targetARN, event,
			),
		})
	}
	return entries
}

func specFilterNode(prefix, suffix string) *core.MappingNode {
	var rules []*core.MappingNode
	if prefix != "" {
		rules = append(rules, core.MappingNodeFields(
			"name", core.MappingNodeFromString("prefix"),
			"value", core.MappingNodeFromString(prefix),
		))
	}
	if suffix != "" {
		rules = append(rules, core.MappingNodeFields(
			"name", core.MappingNodeFromString("suffix"),
			"value", core.MappingNodeFromString(suffix),
		))
	}

	if len(rules) == 0 {
		return nil
	}

	return core.MappingNodeFields(
		"s3Key", core.MappingNodeFields("rules", &core.MappingNode{Items: rules}),
	)
}

// S3NotificationLinkData builds the link-data subtree and the resource-data mappings that
// claim the link's notification entries on the bucket, so the bucket's drift/update does
// not strip them. Each entry is mapped by its compound (destination ARN AND event)
// selector onto the bucket's notification-configuration array.
func S3NotificationLinkData(
	bucketResourceName string,
	target S3NotificationTarget,
	entries []S3NotificationSpecEntry,
) (*core.MappingNode, map[string]string) {
	items := make([]*core.MappingNode, len(entries))
	for i, entry := range entries {
		items[i] = entry.SpecNode
	}

	linkData := core.MappingNodeFields(
		bucketResourceName, core.MappingNodeFields(
			"notificationConfiguration", core.MappingNodeFields(
				target.ConfigField, &core.MappingNode{Items: items},
			),
		),
	)

	mappings := map[string]string{}
	for _, entry := range entries {
		specPath := fmt.Sprintf(
			"%s::spec.notificationConfiguration.%s%s",
			bucketResourceName, target.ConfigField, entry.Selector,
		)
		linkPath := fmt.Sprintf(
			"%s.notificationConfiguration.%s%s",
			bucketResourceName, target.ConfigField, entry.Selector,
		)
		mappings[specPath] = linkPath
	}
	return linkData, mappings
}

// CollectS3NotificationChanges projects the link's bucket notification contribution as a
// known-on-deploy change when either linked resource is new (the entries embed the
// destination ARN, which is resolved on deploy), so the bucket notification change is
// surfaced in the staged plan alongside the permission intermediary.
func CollectS3NotificationChanges(
	changes *provider.LinkChanges,
	bucketResourceName string,
	target S3NotificationTarget,
	resourceAChanges, resourceBChanges *provider.Changes,
) {
	if pluginutils.IsResourceNew(resourceAChanges) || pluginutils.IsResourceNew(resourceBChanges) {
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			fmt.Sprintf("%s.notificationConfiguration.%s", bucketResourceName, target.ConfigField),
		)
	}
}

// PutS3Notification merges this link's notification entries (one per event, identified by
// "<baseID>-<eventSlug>") into the bucket's notification configuration for the given
// destination type, preserving every other entry, and writes it back. The caller must
// hold the bucket lock.
func PutS3Notification(
	ctx context.Context,
	s3Service s3service.Service,
	target S3NotificationTarget,
	bucket, baseID, targetARN string,
	events []string,
	prefix, suffix string,
) error {
	cfg, err := readBucketNotification(ctx, s3Service, bucket)
	if err != nil {
		return err
	}

	filter := sdkFilter(prefix, suffix)
	link := buildSDKEntries(baseID, events, filter)
	applyTargetEntries(cfg, target, baseID, targetARN, link)

	return writeBucketNotification(ctx, s3Service, bucket, cfg)
}

// RemoveS3Notification drops this link's notification entries (those whose id begins with
// "<baseID>-") from the bucket's notification configuration across all destination types,
// preserving every other entry. The caller must hold the bucket resource lock.
func RemoveS3Notification(
	ctx context.Context,
	s3Service s3service.Service,
	bucket, baseID string,
) error {
	cfg, err := readBucketNotification(ctx, s3Service, bucket)
	if err != nil {
		return err
	}
	prefix := baseID + "-"

	cfg.LambdaFunctionConfigurations = keepLambda(cfg.LambdaFunctionConfigurations, func(id string) bool {
		return !strings.HasPrefix(id, prefix)
	})
	cfg.QueueConfigurations = keepQueue(cfg.QueueConfigurations, func(id string) bool {
		return !strings.HasPrefix(id, prefix)
	})
	cfg.TopicConfigurations = keepTopic(cfg.TopicConfigurations, func(id string) bool {
		return !strings.HasPrefix(id, prefix)
	})

	return writeBucketNotification(ctx, s3Service, bucket, cfg)
}

// S3NotificationEntryID is the stable per-event S3 notification id this link writes,
// "<baseID>-<eventSlug>".
func S3NotificationEntryID(baseID, event string) string {
	return fmt.Sprintf("%s-%s", baseID, pluginutils.StripNonAlphaNumericChars(event))
}

type sdkNotificationEntry struct {
	id     string
	event  string
	filter *s3types.NotificationConfigurationFilter
}

func buildSDKEntries(
	baseID string,
	events []string,
	filter *s3types.NotificationConfigurationFilter,
) []sdkNotificationEntry {
	entries := make([]sdkNotificationEntry, 0, len(events))
	for _, event := range events {
		entries = append(entries, sdkNotificationEntry{
			id:     S3NotificationEntryID(baseID, event),
			event:  event,
			filter: filter,
		})
	}
	return entries
}

func applyTargetEntries(
	cfg *s3types.NotificationConfiguration,
	target S3NotificationTarget,
	baseID, targetARN string,
	entries []sdkNotificationEntry,
) {
	prefix := baseID + "-"
	keep := func(id string) bool { return !strings.HasPrefix(id, prefix) }

	switch target.ConfigField {
	case S3LambdaTarget.ConfigField:
		kept := keepLambda(cfg.LambdaFunctionConfigurations, keep)
		for _, e := range entries {
			kept = append(kept, s3types.LambdaFunctionConfiguration{
				Id:                aws.String(e.id),
				LambdaFunctionArn: aws.String(targetARN),
				Events:            []s3types.Event{s3types.Event(e.event)},
				Filter:            e.filter,
			})
		}
		cfg.LambdaFunctionConfigurations = kept
	case S3QueueTarget.ConfigField:
		kept := keepQueue(cfg.QueueConfigurations, keep)
		for _, e := range entries {
			kept = append(kept, s3types.QueueConfiguration{
				Id:       aws.String(e.id),
				QueueArn: aws.String(targetARN),
				Events:   []s3types.Event{s3types.Event(e.event)},
				Filter:   e.filter,
			})
		}
		cfg.QueueConfigurations = kept
	case S3TopicTarget.ConfigField:
		kept := keepTopic(cfg.TopicConfigurations, keep)
		for _, e := range entries {
			kept = append(kept, s3types.TopicConfiguration{
				Id:       aws.String(e.id),
				TopicArn: aws.String(targetARN),
				Events:   []s3types.Event{s3types.Event(e.event)},
				Filter:   e.filter,
			})
		}
		cfg.TopicConfigurations = kept
	}
}

func keepLambda(in []s3types.LambdaFunctionConfiguration, keep func(string) bool) []s3types.LambdaFunctionConfiguration {
	out := in[:0]
	for _, c := range in {
		if keep(aws.ToString(c.Id)) {
			out = append(out, c)
		}
	}
	return out
}

func keepQueue(in []s3types.QueueConfiguration, keep func(string) bool) []s3types.QueueConfiguration {
	out := in[:0]
	for _, c := range in {
		if keep(aws.ToString(c.Id)) {
			out = append(out, c)
		}
	}
	return out
}

func keepTopic(in []s3types.TopicConfiguration, keep func(string) bool) []s3types.TopicConfiguration {
	out := in[:0]
	for _, c := range in {
		if keep(aws.ToString(c.Id)) {
			out = append(out, c)
		}
	}
	return out
}

func sdkFilter(prefix, suffix string) *s3types.NotificationConfigurationFilter {
	var rules []s3types.FilterRule
	if prefix != "" {
		rules = append(rules, s3types.FilterRule{
			Name:  s3types.FilterRuleNamePrefix,
			Value: aws.String(prefix),
		})
	}
	if suffix != "" {
		rules = append(rules, s3types.FilterRule{
			Name:  s3types.FilterRuleNameSuffix,
			Value: aws.String(suffix),
		})
	}
	if len(rules) == 0 {
		return nil
	}
	return &s3types.NotificationConfigurationFilter{
		Key: &s3types.S3KeyFilter{FilterRules: rules},
	}
}

func readBucketNotification(
	ctx context.Context,
	s3Service s3service.Service,
	bucket string,
) (*s3types.NotificationConfiguration, error) {
	out, err := s3Service.GetBucketNotificationConfiguration(
		ctx,
		&s3.GetBucketNotificationConfigurationInput{
			Bucket: aws.String(bucket),
		},
	)
	if err != nil {
		return nil, err
	}

	return &s3types.NotificationConfiguration{
		LambdaFunctionConfigurations: out.LambdaFunctionConfigurations,
		QueueConfigurations:          out.QueueConfigurations,
		TopicConfigurations:          out.TopicConfigurations,
		EventBridgeConfiguration:     out.EventBridgeConfiguration,
	}, nil
}

func writeBucketNotification(
	ctx context.Context,
	s3Service s3service.Service,
	bucket string,
	cfg *s3types.NotificationConfiguration,
) error {
	// S3 validates the destination permission when the notification configuration is
	// applied; the permission is granted moments earlier in the same link update, so a
	// transient validation failure is retried.
	put := pluginutils.Retryable(
		func(ctx context.Context, in *s3.PutBucketNotificationConfigurationInput) error {
			_, err := s3Service.PutBucketNotificationConfiguration(ctx, in)
			return err
		},
		isS3DestinationNotReadyError,
	)
	return put(ctx, &s3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(bucket),
		NotificationConfiguration: cfg,
	})
}

// isS3DestinationNotReadyError reports whether an error is the transient validation
// failure S3 returns when a notification destination's permission has not yet propagated.
func isS3DestinationNotReadyError(err error) bool {
	if err == nil {
		return false
	}
	// S3 returns a 400 InvalidArgument when it cannot validate a notification destination
	// (e.g. the destination permission has not yet propagated); the message disambiguates
	// this transient case from other InvalidArgument failures. This is an eventual-
	// consistency retry the AWS SDK's default retryer will not perform.
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == "InvalidArgument" {
		msg := apiErr.ErrorMessage()
		return strings.Contains(msg, "Unable to validate the following destination configurations") ||
			strings.Contains(msg, "Permissions on the destination")
	}
	return false
}
