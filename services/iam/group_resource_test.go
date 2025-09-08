//go:build unit

package iam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractGroupNameFromARN(t *testing.T) {
	tests := []struct {
		name     string
		arn      string
		expected string
		hasError bool
	}{
		{
			name:     "valid group ARN",
			arn:      "arn:aws:iam::123456789012:group/test-group",
			expected: "test-group",
			hasError: false,
		},
		{
			name:     "valid group ARN with path",
			arn:      "arn:aws:iam::123456789012:group/path/to/test-group",
			expected: "path/to/test-group",
			hasError: false,
		},
		{
			name:     "invalid ARN format",
			arn:      "invalid-arn",
			expected: "",
			hasError: true,
		},
		{
			name:     "not a group ARN",
			arn:      "arn:aws:iam::123456789012:user/test-user",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractGroupNameFromARN(tt.arn)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
