package utils

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// MatchesEC2Tag checks if the current tag matches the expected tag
// for EC2 resources.
func MatchesEC2Tag(
	expectedTag ec2types.Tag,
	currentTag ec2types.Tag,
) bool {
	return currentTag.Key != nil &&
		currentTag.Value != nil &&
		aws.ToString(currentTag.Key) == aws.ToString(expectedTag.Key) &&
		aws.ToString(currentTag.Value) == aws.ToString(expectedTag.Value)
}
