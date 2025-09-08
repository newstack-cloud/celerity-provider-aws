//go:build unit

package lambdalinks

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type FunctionFunctionLinkUpdateSuite struct {
	suite.Suite
}

func (s *FunctionFunctionLinkUpdateSuite) Test_link_update_linked_resources() {
	// loader := &testutils.MockAWSConfigLoader{}
	// linkCtx := plugintestutils.NewTestLinkContext(
	// 	map[string]map[string]*core.ScalarValue{
	// 		"aws": {
	// 			"region": core.ScalarFromString("us-west-2"),
	// 		},
	// 	},
	// 	map[string]*core.ScalarValue{
	// 		"session_id": core.ScalarFromString("test-session-id"),
	// 	},
	// )

	// testCases := []plugintestutils.LinkUpdateResourceTestCase[
	// 	*aws.Config,
	// 	lambdaservice.Service,
	// 	*aws.Config,
	// 	lambdaservice.Service,
	// ]{
	// 	s.createCallerFunctionUpdateTestCase(linkCtx, loader),
	// 	s.createCalleeFunctionUpdateTestCase(linkCtx, loader),
	// 	s.createDisableEnvVarPopulationTestCase(linkCtx, loader),
	// 	s.createRemoveEnvVarsForLinkDestroyTestCase(linkCtx, loader),
	// 	s.createPointEnvVarsToNewCalleeFunctionTestCase(linkCtx, loader),
	// 	s.createErrorCallerFuncMissingARNTestCase(linkCtx, loader),
	// 	s.createErrorCaleeFuncMissingARNTestCase(linkCtx, loader),
	// 	s.createErrorUpdateServiceErrorTestCase(linkCtx, loader),
	// 	s.createErrorRemoveServiceErrorTestCase(linkCtx, loader),
	// }

	// plugintestutils.RunLinkUpdateResourceTestCases(
	// 	testCases,
	// 	FunctionCodeSigningConfigLink,
	// 	&s.Suite,
	// )
}

func (s *FunctionFunctionLinkUpdateSuite) Test_link_update_intermediary_resources() {
	// loader := &testutils.MockAWSConfigLoader{}
	// linkCtx := plugintestutils.NewTestLinkContext(
	// 	map[string]map[string]*core.ScalarValue{
	// 		"aws": {
	// 			"region": core.ScalarFromString("us-west-2"),
	// 		},
	// 	},
	// 	map[string]*core.ScalarValue{
	// 		"session_id": core.ScalarFromString("test-session-id"),
	// 	},
	// )

	// testCases := []plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	// 	*aws.Config,
	// 	lambdaservice.Service,
	// 	*aws.Config,
	// 	lambdaservice.Service,
	// ]{
	// 	s.createExistingRolePolicyUpdateTestCase(linkCtx, loader),
	// 	s.createNewRolePolicyUpdateTestCase(linkCtx, loader),
	// 	s.createRolePolicyPermRemovalTestCase(linkCtx, loader),
	// 	s.createRolePolicyRemovalTestCase(linkCtx, loader),
	// 	s.createPermUpdateForRecreatedCallerFunctionTestCase(linkCtx, loader),
	// 	s.createPermUpdateForRecreatedCalleeFunctionTestCase(linkCtx, loader),
	// 	s.createErrorMissingRoleErrorTestCase(linkCtx, loader),
	// }

	// plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
	// 	testCases,
	// 	FunctionFunctionLink(iamServiceFactory),
	// 	&s.Suite,
	// )
}

func TestFunctionFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionFunctionLinkUpdateSuite))
}
