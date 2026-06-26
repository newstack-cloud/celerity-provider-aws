//go:build unit

package overlays

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_stabilise_required_registry(t *testing.T) {
	// aws/rds/dbInstance is registered by this package's rds_db_instance.go init().
	assert.True(t, IsStabiliseRequired("aws/rds/dbInstance"))
	assert.False(t, IsStabiliseRequired("aws/sqs/queue"))

	types := StabiliseRequiredTypes()
	assert.True(t, slices.Contains(types, "aws/rds/dbInstance"))
	// Returned sorted for deterministic output.
	assert.True(t, slices.IsSorted(types))
}

func Test_register_stabilise_required_is_idempotent(t *testing.T) {
	RegisterStabiliseRequired("aws/test/slowExample")
	RegisterStabiliseRequired("aws/test/slowExample")

	assert.True(t, IsStabiliseRequired("aws/test/slowExample"))
	assert.Equal(t, 1, countOccurrences(StabiliseRequiredTypes(), "aws/test/slowExample"))
}

func countOccurrences(values []string, target string) int {
	count := 0
	for _, v := range values {
		if v == target {
			count++
		}
	}
	return count
}
