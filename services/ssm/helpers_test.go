//go:build unit

package ssm

import "errors"

var (
	errTestPutParameter    = errors.New("failed to put parameter")
	errTestGetParameter    = errors.New("failed to get parameter")
	errTestDeleteParameter = errors.New("failed to delete parameter")
)
