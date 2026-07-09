//go:build unit

package apigatewayv2lambda

import (
	"context"
	"testing"

	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/suite"
)

type APIFunctionLinkStageChangesSuite struct {
	suite.Suite
}

func (s *APIFunctionLinkStageChangesSuite) Test_projects_intermediaries_as_new_fields() {
	loader := &testutils.MockAWSConfigLoader{}

	out, err := testActions(loader).StageChanges(context.Background(), &provider.LinkStageChangesInput{
		ResourceAChanges: &provider.Changes{AppliedResourceInfo: *afAPIInfo("HTTP")},
		ResourceBChanges: &provider.Changes{AppliedResourceInfo: *afFunctionInfo()},
		CurrentLinkState: &state.LinkState{LinkID: "link-1"},
	})
	s.Require().NoError(err)

	newFieldPaths := map[string]bool{}
	for _, change := range out.Changes.NewFields {
		newFieldPaths[change.FieldPath] = true
	}
	anyChange := map[string]bool{}
	for _, change := range out.Changes.NewFields {
		anyChange[change.FieldPath] = true
	}
	for _, path := range out.Changes.FieldChangesKnownOnDeploy {
		anyChange[path] = true
	}

	// Each intermediary is surfaced via its (static) resourceType leaf as a new field.
	s.True(newFieldPaths[intermediaryLeaf(afIntegrationResourceID, "resourceType")])
	s.True(newFieldPaths[intermediaryLeaf(afRouteResourceID, "resourceType")])
	s.True(newFieldPaths[intermediaryLeaf(afPermissionResourceID, "resourceType")])
	// Derived leaves come from the linked resources (new or known-on-deploy).
	s.True(anyChange[intermediaryLeaf(afIntegrationResourceID, "integrationUri")])
	s.True(anyChange[intermediaryLeaf(afPermissionResourceID, "functionArn")])
}

func intermediaryLeaf(resourceID, leaf string) string {
	return "[\"intermediaries\"][\"" + resourceID + "\"][\"" + leaf + "\"]"
}

func TestAPIFunctionLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(APIFunctionLinkStageChangesSuite))
}
