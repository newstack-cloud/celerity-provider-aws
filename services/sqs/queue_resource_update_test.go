//go:build unit

package sqs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SQSQueueResourceUpdateSuite struct {
	suite.Suite
}

func (s *SQSQueueResourceUpdateSuite) Test_update_sqs_queue() {
	// TODO: Implement SQS queue update test logic
}

func TestSQSQueueResourceUpdateSuite(t *testing.T) {
	suite.Run(t, new(SQSQueueResourceUpdateSuite))
}
