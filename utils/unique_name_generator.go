package utils

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// UniqueNameGenerator is a function type for generating unique names.
type UniqueNameGenerator func(input *provider.ResourceDeployInput) (string, error)

// DefaultUniqueNameGenerator creates a unique name using blueprint instance name, resource name, and a nanoid
// with a configurable character limit. Uses input.InstanceName which provides meaningful, human-readable names.
func DefaultUniqueNameGenerator(maxLength int) UniqueNameGenerator {
	return func(input *provider.ResourceDeployInput) (string, error) {
		instanceName := input.InstanceName
		resourceName := input.Changes.AppliedResourceInfo.ResourceName

		// Generate a short nanoid (8 chars)
		nid, err := gonanoid.New(8)
		if err != nil {
			return "", err
		}

		var base string
		if instanceName != "" {
			// Use full format: instance-resource-nanoid
			base = fmt.Sprintf("%s-%s-%s", instanceName, resourceName, nid)
		} else {
			// Use simplified format: resource-nanoid (for new deployments)
			base = fmt.Sprintf("%s-%s", resourceName, nid)
		}

		// Ensure the name is exactly the specified length
		if len(base) > maxLength {
			if instanceName != "" {
				// Calculate how much space we need for the nanoid and separators
				nanoidLength := 8
				separatorLength := 2 // two hyphens
				availableSpace := maxLength - nanoidLength - separatorLength

				if availableSpace <= 0 {
					// If we can't fit the required parts, just use the nanoid
					return nid[:maxLength], nil
				}

				// Distribute available space between instance name and resource name
				// Give slightly more space to instance name as it's typically more important
				maxInstance := availableSpace * 3 / 5
				maxResource := availableSpace * 2 / 5

				if len(instanceName) > maxInstance {
					instanceName = instanceName[:maxInstance]
				}
				if len(resourceName) > maxResource {
					resourceName = resourceName[:maxResource]
				}

				base = fmt.Sprintf("%s-%s-%s", instanceName, resourceName, nid)
			} else {
				// For simplified format, just truncate resource name
				nanoidLength := 8
				separatorLength := 1 // one hyphen
				availableSpace := maxLength - nanoidLength - separatorLength

				if availableSpace <= 0 {
					// If we can't fit the required parts, just use the nanoid
					return nid[:maxLength], nil
				}

				if len(resourceName) > availableSpace {
					resourceName = resourceName[:availableSpace]
				}

				base = fmt.Sprintf("%s-%s", resourceName, nid)
			}
		}

		// Pad or truncate to exact length
		if len(base) < maxLength {
			// Pad with hyphens if too short
			base = base + "-"
			for len(base) < maxLength {
				base = base + "-"
			}
		} else if len(base) > maxLength {
			// Truncate if too long
			base = base[:maxLength]
		}

		return base, nil
	}
}

// Common name generators for different AWS services.
var (
	// IAMRoleNameGenerator generates names for IAM roles (64 char limit).
	IAMRoleNameGenerator = DefaultUniqueNameGenerator(64)

	// IAMUserNameGenerator generates names for IAM users (64 char limit).
	IAMUserNameGenerator = DefaultUniqueNameGenerator(64)

	// IAMGroupNameGenerator generates names for IAM groups (128 char limit).
	IAMGroupNameGenerator = DefaultUniqueNameGenerator(128)

	// IAMInstanceProfileNameGenerator generates names for IAM instance profiles (128 char limit).
	IAMInstanceProfileNameGenerator = DefaultUniqueNameGenerator(128)

	// LambdaFunctionNameGenerator generates names for Lambda functions (64 char limit).
	LambdaFunctionNameGenerator = DefaultUniqueNameGenerator(64)

	// S3BucketNameGenerator generates names for S3 buckets (63 char limit).
	S3BucketNameGenerator = DefaultUniqueNameGenerator(63)

	// EC2InstanceNameGenerator generates names for EC2 instances (255 char limit).
	EC2InstanceNameGenerator = DefaultUniqueNameGenerator(255)

	// DynamoDBTableNameGenerator generates names for DynamoDB tables (255 char limit).
	DynamoDBTableNameGenerator = DefaultUniqueNameGenerator(255)

	// IAMPolicyNameGenerator generates names for IAM policies (128 char limit).
	IAMPolicyNameGenerator = DefaultUniqueNameGenerator(128)

	// IAMOIDCProviderUrlGenerator generates names for IAM OIDC providers (255 char limit).
	IAMOIDCProviderUrlGenerator = DefaultUniqueNameGenerator(255)

	// IAMSAMLProviderNameGenerator generates names for IAM SAML providers (128 char limit).
	IAMSAMLProviderNameGenerator = DefaultUniqueNameGenerator(128)

	// IAMServerCertificateNameGenerator generates names for IAM server certificates (128 char limit).
	IAMServerCertificateNameGenerator = DefaultUniqueNameGenerator(128)
)

// SQSQueueNameGenerator generates names for SQS queues (80 char limit) and accounts
// for FIFO queues, ensuring the name ends with '.fifo'.
func SQSQueueNameGenerator(fifoQueue bool) UniqueNameGenerator {
	return func(input *provider.ResourceDeployInput) (string, error) {
		if fifoQueue {
			generatedName, err := DefaultUniqueNameGenerator(75)(input)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%s.fifo", generatedName), nil
		}

		return DefaultUniqueNameGenerator(80)(input)
	}
}
