//go:build unit

package eventslinks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type RuleAPIDestinationLinkStageChangesSuite struct {
	suite.Suite
}

func (s *RuleAPIDestinationLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		eventsservice.Service,
		*aws.Config,
		eventsservice.Service,
	]{
		{
			Name: "stages no field changes for existing resources (role policy only)",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{
					AppliedResourceInfo: provider.ResourceInfo{ResourceName: "orderWebhookRule"},
				},
				ResourceBChanges: &provider.Changes{
					AppliedResourceInfo: provider.ResourceInfo{ResourceName: "orderWebhookDestination"},
				},
				CurrentLinkState: &state.LinkState{LinkID: "test-link"},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{},
			},
		},
		{
			Name: "stages the role permission as known on deploy when the rule is new",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{
					AppliedResourceInfo: provider.ResourceInfo{ResourceName: "orderWebhookRule"},
					NewFields: []provider.FieldChange{
						{FieldPath: "spec.arn"},
					},
				},
				ResourceBChanges: &provider.Changes{
					AppliedResourceInfo: provider.ResourceInfo{ResourceName: "orderWebhookDestination"},
				},
				CurrentLinkState: &state.LinkState{LinkID: "test-link"},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					FieldChangesKnownOnDeploy: []string{"orderWebhookRuleRole"},
				},
			},
		},
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		ruleAPIDestinationLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

func TestRuleAPIDestinationLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(RuleAPIDestinationLinkStageChangesSuite))
}
