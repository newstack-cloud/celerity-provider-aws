//go:build unit

package utils

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/stretchr/testify/assert"
)

func TestDefaultUniqueNameGenerator(t *testing.T) {
	tests := []struct {
		name           string
		maxLength      int
		instanceName   string
		resourceName   string
		expectedPrefix string
		expectedLength int
		hasInstanceID  bool
	}{
		{
			name:           "normal case with instance name",
			maxLength:      64,
			instanceName:   "production-env",
			resourceName:   "TestRole",
			expectedPrefix: "production-env-TestRole-",
			expectedLength: 64,
			hasInstanceID:  true,
		},
		{
			name:           "new deployment without instance name",
			maxLength:      64,
			instanceName:   "",
			resourceName:   "TestRole",
			expectedPrefix: "TestRole-",
			expectedLength: 64,
			hasInstanceID:  false,
		},
		{
			name:           "short limit with instance name",
			maxLength:      20,
			instanceName:   "very-long-instance-name",
			resourceName:   "VeryLongResourceName",
			expectedPrefix: "",
			expectedLength: 20,
			hasInstanceID:  true,
		},
		{
			name:           "short limit without instance name",
			maxLength:      20,
			instanceName:   "",
			resourceName:   "VeryLongResourceName",
			expectedPrefix: "",
			expectedLength: 20,
			hasInstanceID:  false,
		},
		{
			name:           "exact fit with instance name",
			maxLength:      30,
			instanceName:   "short",
			resourceName:   "role",
			expectedPrefix: "short-role-",
			expectedLength: 30,
			hasInstanceID:  true,
		},
		{
			name:           "exact fit without instance name",
			maxLength:      30,
			instanceName:   "",
			resourceName:   "role",
			expectedPrefix: "role-",
			expectedLength: 30,
			hasInstanceID:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := DefaultUniqueNameGenerator(tt.maxLength)

			input := &provider.ResourceDeployInput{
				InstanceID:   "uuid-instance-id",
				InstanceName: tt.instanceName,
				ResourceID:   "test-resource-id",
				Changes: &provider.Changes{
					AppliedResourceInfo: provider.ResourceInfo{
						ResourceID:   "test-resource-id",
						ResourceName: tt.resourceName,
						InstanceID:   "uuid-instance-id",
						ResourceWithResolvedSubs: &provider.ResolvedResource{
							Type: &schema.ResourceTypeWrapper{
								Value: "aws/iam/role",
							},
							Spec: &core.MappingNode{},
						},
					},
				},
			}

			result, err := generator(input)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedLength)

			if tt.expectedPrefix != "" {
				assert.Contains(t, result, tt.expectedPrefix)
			}

			// Verify the result contains parts of the instance name and resource name
			// For short limits, we check for truncated versions
			if tt.maxLength >= 30 {
				// For longer limits, we should see the full names
				if tt.hasInstanceID && len(tt.instanceName) <= 20 {
					assert.Contains(t, result, tt.instanceName)
				}
				if len(tt.resourceName) <= 20 {
					assert.Contains(t, result, tt.resourceName)
				}
			} else {
				// For short limits, check for truncated versions
				if tt.hasInstanceID && len(tt.instanceName) > 0 {
					assert.Contains(t, result, tt.instanceName[:3])
				}
				if len(tt.resourceName) > 0 {
					assert.Contains(t, result, tt.resourceName[:3])
				}
			}
		})
	}
}

func TestPredefinedGenerators(t *testing.T) {
	tests := []struct {
		name          string
		generator     UniqueNameGenerator
		maxLength     int
		description   string
		hasInstanceID bool
	}{
		{
			name:          "IAM Role Generator with instance name",
			generator:     IAMRoleNameGenerator,
			maxLength:     64,
			description:   "IAM roles have 64 character limit",
			hasInstanceID: true,
		},
		{
			name:          "IAM Role Generator without instance name",
			generator:     IAMRoleNameGenerator,
			maxLength:     64,
			description:   "IAM roles have 64 character limit (new deployment)",
			hasInstanceID: false,
		},
		{
			name:          "Lambda Function Generator",
			generator:     LambdaFunctionNameGenerator,
			maxLength:     64,
			description:   "Lambda functions have 64 character limit",
			hasInstanceID: true,
		},
		{
			name:          "S3 Bucket Generator",
			generator:     S3BucketNameGenerator,
			maxLength:     63,
			description:   "S3 buckets have 63 character limit",
			hasInstanceID: true,
		},
		{
			name:          "EC2 Instance Generator",
			generator:     EC2InstanceNameGenerator,
			maxLength:     255,
			description:   "EC2 instances have 255 character limit",
			hasInstanceID: true,
		},
		{
			name:          "DynamoDB Table Generator",
			generator:     DynamoDBTableNameGenerator,
			maxLength:     255,
			description:   "DynamoDB tables have 255 character limit",
			hasInstanceID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instanceName := "production-env"
			if !tt.hasInstanceID {
				instanceName = ""
			}

			input := &provider.ResourceDeployInput{
				InstanceID:   "uuid-instance-id",
				InstanceName: instanceName,
				ResourceID:   "test-resource-id",
				Changes: &provider.Changes{
					AppliedResourceInfo: provider.ResourceInfo{
						ResourceID:   "test-resource-id",
						ResourceName: "TestResource",
						InstanceID:   "uuid-instance-id",
						ResourceWithResolvedSubs: &provider.ResolvedResource{
							Type: &schema.ResourceTypeWrapper{
								Value: "aws/iam/role",
							},
							Spec: &core.MappingNode{},
						},
					},
				},
			}

			result, err := tt.generator(input)

			assert.NoError(t, err)
			assert.LessOrEqual(t, len(result), tt.maxLength, tt.description)
			assert.Greater(t, len(result), 0)

			// Verify it contains the expected components
			if tt.hasInstanceID {
				assert.Contains(t, result, "production-env")
			}
			assert.Contains(t, result, "TestResource")
		})
	}
}

func TestGeneratorWithEmptyInputs(t *testing.T) {
	input := &provider.ResourceDeployInput{
		InstanceID:   "uuid-instance-id",
		InstanceName: "",
		ResourceID:   "test-resource-id",
		Changes: &provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceID:   "test-resource-id",
				ResourceName: "",
				InstanceID:   "uuid-instance-id",
				ResourceWithResolvedSubs: &provider.ResolvedResource{
					Type: &schema.ResourceTypeWrapper{
						Value: "aws/iam/role",
					},
					Spec: &core.MappingNode{},
				},
			},
		},
	}

	generator := DefaultUniqueNameGenerator(64)
	result, err := generator(input)

	assert.NoError(t, err)
	assert.Len(t, result, 64)
	assert.Contains(t, result, "--") // Should contain two hyphens for empty instance and resource names
}

func TestGeneratorWithVeryLongInputs(t *testing.T) {
	tests := []struct {
		name          string
		hasInstanceID bool
	}{
		{
			name:          "with instance name",
			hasInstanceID: true,
		},
		{
			name:          "without instance name",
			hasInstanceID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instanceName := "very-long-instance-name-that-exceeds-normal-limits-and-should-be-truncated"
			if !tt.hasInstanceID {
				instanceName = ""
			}

			input := &provider.ResourceDeployInput{
				InstanceID:   "uuid-instance-id",
				InstanceName: instanceName,
				ResourceID:   "test-resource-id",
				Changes: &provider.Changes{
					AppliedResourceInfo: provider.ResourceInfo{
						ResourceID:   "test-resource-id",
						ResourceName: "very-long-resource-name-that-exceeds-normal-limits-and-should-be-truncated",
						InstanceID:   "uuid-instance-id",
						ResourceWithResolvedSubs: &provider.ResolvedResource{
							Type: &schema.ResourceTypeWrapper{
								Value: "aws/iam/role",
							},
							Spec: &core.MappingNode{},
						},
					},
				},
			}

			generator := DefaultUniqueNameGenerator(30)
			result, err := generator(input)

			assert.NoError(t, err)
			assert.Len(t, result, 30)

			// Should still contain parts of the original names
			// For 30 char limit, the algorithm allocates:
			// - 8 chars for nanoid
			// - 2 chars for separators
			// - 20 chars to distribute between instance and resource names
			// - 3/5 to instance name = 12 chars
			// - 2/5 to resource name = 8 chars
			if tt.hasInstanceID {
				// Should contain truncated instance name (first 12 chars)
				assert.Contains(t, result, "very-long-in")
			}
			// Should contain truncated resource name (first 8 chars)
			assert.Contains(t, result, "very-long")
		})
	}
}
