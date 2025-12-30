package dynamodb

import (
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// mergeBluelinkTagsWithDynamoDBTags merges Bluelink system tags with user-defined DynamoDB tags.
// User tags take precedence on key conflicts.
// Returns nil if there are no tags to set (preserves original behavior when tagging is not enabled).
func mergeBluelinkTagsWithDynamoDBTags(
	deployInput *provider.ResourceDeployInput,
	userTags map[string]string,
) map[string]string {
	if userTags == nil {
		userTags = make(map[string]string)
	}
	result := utils.MergeBluelinkTagsWithUserTags(deployInput, userTags)
	if len(result) == 0 {
		return nil
	}
	return result
}
