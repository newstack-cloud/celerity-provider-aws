package utils

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Tag is an intermediary representation of a tag that is used across
// AWS services (The SDK provides a different type for each service).
// This is used to provide a consistent interface for the tag changes
// that should be converted to the upstream service's tag type.
type Tag struct {
	Key   string
	Value string
}

// TagsDiffResult is the result of the DiffTags function,
// it contains the tags that should be set and the tag keys that should be removed.
type TagsDiffResult[UpstreamTag any] struct {
	ToSet    []UpstreamTag
	ToRemove []string
}

// DiffTags provides a general purpose utility to derive the difference
// between two sets of tags stored in a resource spec for the purpose of making
// calls to the upstream service to apply tag changes.
// There is a limitation in the `Changes` data that the plugin receives
// in that when the tags are stored in a list, it does not provide sufficient information
// on the key of the tags to be removed as it just reports an updated or removed index
// in the list.
// For this reason, resource implementations need to use the actual current and upcoming
// resource spec data to derive the tag changes so that the correct tags are removed, added
// and replaced.
//
// tagsRootPath is the path to the tags field in the resource spec,
// the expected format is to use "$" to represent the root of the spec (e.g. "$.tags").
func DiffTags[UpstreamTag any](
	changes *provider.Changes,
	tagsRootPath string,
	transformTag func(tag *Tag) UpstreamTag,
) *TagsDiffResult[UpstreamTag] {
	result := &TagsDiffResult[UpstreamTag]{
		ToSet:    []UpstreamTag{},
		ToRemove: []string{},
	}

	currentSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentSpecTags, _ := pluginutils.GetValueByPath(tagsRootPath, currentSpecData)
	newSpecData := pluginutils.GetResolvedResourceSpecData(changes)
	newSpecTags, _ := pluginutils.GetValueByPath(tagsRootPath, newSpecData)

	currentTags := ToTagsMap(currentSpecTags)
	desiredTags := ToTagsMap(newSpecTags)

	// Calculate tags to add/update
	toSetIntermediary := []*Tag{}
	for key, value := range desiredTags {
		toSetIntermediary = append(toSetIntermediary, &Tag{
			Key:   key,
			Value: value,
		})
	}

	// Calculate tags to remove
	for key := range currentTags {
		if _, exists := desiredTags[key]; !exists {
			result.ToRemove = append(result.ToRemove, key)
		}
	}

	// Sort the tags to set and remove by key, this is mostly helpful
	// for deterministic comparison of output.
	slices.SortFunc(toSetIntermediary, func(i, j *Tag) int {
		return strings.Compare(i.Key, j.Key)
	})
	slices.Sort(result.ToRemove)

	for _, tag := range toSetIntermediary {
		result.ToSet = append(result.ToSet, transformTag(tag))
	}

	return result
}

// DiffTagsWithBluelink provides tag diffing that includes Bluelink system tags.
// It works like DiffTags but merges Bluelink tags with the desired user tags
// before computing the diff.
//
// tagsRootPath is the path to the tags field in the resource spec,
// the expected format is to use "$" to represent the root of the spec (e.g. "$.tags").
func DiffTagsWithBluelink[UpstreamTag any](
	changes *provider.Changes,
	deployInput *provider.ResourceDeployInput,
	tagsRootPath string,
	transformTag func(tag *Tag) UpstreamTag,
) *TagsDiffResult[UpstreamTag] {
	result := &TagsDiffResult[UpstreamTag]{
		ToSet:    []UpstreamTag{},
		ToRemove: []string{},
	}

	currentSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentSpecTags, _ := pluginutils.GetValueByPath(tagsRootPath, currentSpecData)
	newSpecData := pluginutils.GetResolvedResourceSpecData(changes)
	newSpecTags, _ := pluginutils.GetValueByPath(tagsRootPath, newSpecData)

	currentTags := ToTagsMap(currentSpecTags)
	userDesiredTags := ToTagsMap(newSpecTags)

	// Merge Bluelink system tags with user-defined tags
	desiredTags := MergeBluelinkTagsWithUserTags(deployInput, userDesiredTags)

	// Calculate tags to add/update
	toSetIntermediary := []*Tag{}
	for key, value := range desiredTags {
		toSetIntermediary = append(toSetIntermediary, &Tag{
			Key:   key,
			Value: value,
		})
	}

	// Calculate tags to remove
	for key := range currentTags {
		if _, exists := desiredTags[key]; !exists {
			result.ToRemove = append(result.ToRemove, key)
		}
	}

	// Sort the tags to set and remove by key, this is mostly helpful
	// for deterministic comparison of output.
	slices.SortFunc(toSetIntermediary, func(i, j *Tag) int {
		return strings.Compare(i.Key, j.Key)
	})
	slices.Sort(result.ToRemove)

	for _, tag := range toSetIntermediary {
		result.ToSet = append(result.ToSet, transformTag(tag))
	}

	return result
}

// ToTagsMap converts a MappingNode to a map of tags.
// The MappingNode is expected to be an array of objects with "key" and "value" fields.
func ToTagsMap(specTags *core.MappingNode) map[string]string {
	tagMap := make(map[string]string)
	if core.IsArrayMappingNode(specTags) {
		for _, item := range specTags.Items {
			key, _ := pluginutils.GetValueByPath("$.key", item)
			value, _ := pluginutils.GetValueByPath("$.value", item)
			tagMap[core.StringValue(key)] = core.StringValue(value)
		}
	} else if core.IsObjectMappingNode(specTags) {
		for key, value := range specTags.Fields {
			tagMap[key] = core.StringValue(value)
		}
	}
	return tagMap
}

const (
	// TagLinkSecurityGroup is a tag that is used to identify the security group
	// that is used to allow access to VPC endpoints.
	TagLinkSecurityGroup = "bluelink:link:security-group"
	// TagLinkVPCName is a tag that is used to identify the flex VPC
	// that is used to allow access to VPC endpoints.
	TagLinkVPCName = "bluelink:link:flex-vpc:name"
	// TagBlueprintInstanceName is a tag that is used to identify the blueprint instance
	// that is used to allow access to VPC endpoints.
	TagBlueprintInstanceName = "bluelink:blueprint-instance:name"
	// TagBlueprintLinkIDPrefix is a tag prefix that is used to identify the blueprint link
	// that is used for networking resources created as a part of a link implementation.
	// Each link will have its own key entry in the tag with the link ID as the suffix.
	TagBlueprintLinkIDPrefix = "bluelink:blueprint-link:id:"
	// TagLinkVPCEndpoint is a tag that is used to identify the VPC endpoint
	// that is used to allow access to VPC endpoints.
	TagLinkVPCEndpoint = "bluelink:link:vpc-endpoint"
	// TagBluelinkService is a tag that is used to identify the service
	// that a resource such as a security group is intended to provide access to.
	TagBluelinkService = "bluelink:service"
)

// CreateTagLinkSecurityGroup creates a tag that is used to identify the security group
// that is used to allow access to VPC endpoints.
func CreateTagLinkSecurityGroup() ec2types.Tag {
	return ec2types.Tag{
		Key:   aws.String(TagLinkSecurityGroup),
		Value: aws.String("true"),
	}
}

// CreateTagLinkVPCEndpoint creates a tag that is used to identify the VPC endpoint
// that is used to allow access to VPC endpoints.
func CreateTagLinkVPCEndpoint() ec2types.Tag {
	return ec2types.Tag{
		Key:   aws.String(TagLinkVPCEndpoint),
		Value: aws.String("true"),
	}
}

// CreateTagFlexVPCNameForLink creates a filter that is used to identify the flex VPC
// that is used to allow access to VPC endpoints.
// This is to be used for components created as a part of a link implementation
// and is different from the tag used for the core flex VPC resources.
func CreateTagFlexVPCNameForLink(flexVPCName string) ec2types.Tag {
	return ec2types.Tag{
		Key:   aws.String(TagLinkVPCName),
		Value: aws.String(flexVPCName),
	}
}

// CreateTagFilterFlexVPCNameForLink creates a filter that is used to identify the flex VPC
// that is used to allow access to VPC endpoints.
// This is to be used for components created as a part of a link implementation
// and is different from the tag used for the core flex VPC resources.
func CreateTagFilterFlexVPCNameForLink(flexVPCName string) ec2types.Filter {
	return ec2types.Filter{
		Name:   aws.String(fmt.Sprintf("tag:%s", TagLinkVPCName)),
		Values: []string{flexVPCName},
	}
}

// CreateTagBlueprintInstanceName creates a tag that is used to identify the blueprint instance
// associated with a networking resource created as a part of a link implementation.
func CreateTagBlueprintInstanceName(instanceName string) ec2types.Tag {
	return ec2types.Tag{
		Key:   aws.String(TagBlueprintInstanceName),
		Value: aws.String(instanceName),
	}
}

// CreateTagBlueprintLinkID creates a tag that is used to identify the blueprint link
// that is used to allow access to VPC endpoints.
func CreateTagBlueprintLinkID(linkID string) ec2types.Tag {
	return ec2types.Tag{
		Key:   aws.String(fmt.Sprintf("%s%s", TagBlueprintLinkIDPrefix, linkID)),
		Value: aws.String("true"),
	}
}

// CreateTagBluelinkService creates a tag that is used to identify the service
// that is used to allow access to VPC endpoints.
func CreateTagBluelinkService(serviceName string) ec2types.Tag {
	return ec2types.Tag{
		Key:   aws.String(TagBluelinkService),
		Value: aws.String(serviceName),
	}
}
