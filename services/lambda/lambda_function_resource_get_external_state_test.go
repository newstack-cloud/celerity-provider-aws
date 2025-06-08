package lambda

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LambdaFunctionResourceGetExternalStateSuite struct {
	suite.Suite
}

func TestLambdaFunctionResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceGetExternalStateSuite))
}
