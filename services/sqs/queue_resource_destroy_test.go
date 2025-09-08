//go:build unit

package sqs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SQSQueueResourceDestroySuite struct {
	suite.Suite
}

func (s *SQSQueueResourceDestroySuite) Test_destroy_sqs_queue() {
	// TODO: Implement SQS queue destruction test logic
}

func TestSQSQueueResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(SQSQueueResourceDestroySuite))
}
