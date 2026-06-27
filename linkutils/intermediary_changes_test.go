//go:build unit

package linkutils

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type IntermediaryChangesSuite struct {
	suite.Suite
}

const (
	icResourceID = "ruleA__fnB__eventbridge-invoke-permission"
	icRuleARN    = "arn:aws:events:us-west-2:123456789012:rule/a"
	icFnARN      = "arn:aws:lambda:us-west-2:123456789012:function:b"
)

func icIdentity() IntermediaryIdentity {
	return IntermediaryIdentity{
		ResourceType: "aws/lambda/permission",
		ResourceID:   icResourceID,
		ResourceName: "ruleAInvokeFnB",
	}
}

func icLeaf(leaf string) string {
	return "[\"intermediaries\"][\"" + icResourceID + "\"][\"" + leaf + "\"]"
}

func icResolvedChanges(arn string) *provider.Changes {
	return &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(arn),
				}},
			},
		},
	}
}

func icKnownOnDeployChanges() *provider.Changes {
	return &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
			},
		},
		FieldChangesKnownOnDeploy: []string{"spec.arn"},
	}
}

func (s *IntermediaryChangesSuite) stage(currentLinkData *core.MappingNode, sourceChanges *provider.Changes) *provider.LinkChanges {
	out := &provider.LinkChanges{}
	err := CollectIntermediaryChanges(currentLinkData, out, StageIntermediary{
		Identity: icIdentity(),
		DerivedLeaves: []DerivedLeaf{
			{Leaf: "sourceArn", ResourceChanges: sourceChanges, ResourceSpecPath: "$.spec.arn"},
		},
	})
	s.Require().NoError(err)
	return out
}

func (s *IntermediaryChangesSuite) Test_collect_create() {
	out := s.stage(&core.MappingNode{}, icResolvedChanges(icRuleARN))
	s.Require().Len(out.NewFields, 2)
	s.Equal(icLeaf("resourceType"), out.NewFields[0].FieldPath)
	s.Equal("aws/lambda/permission", core.StringValue(out.NewFields[0].NewValue))
	s.Equal(icLeaf("sourceArn"), out.NewFields[1].FieldPath)
	s.Equal(icRuleARN, core.StringValue(out.NewFields[1].NewValue))
}

func (s *IntermediaryChangesSuite) Test_collect_update() {
	prior := &core.MappingNode{Fields: map[string]*core.MappingNode{
		"intermediaries": {Fields: map[string]*core.MappingNode{
			icResourceID: {Fields: map[string]*core.MappingNode{
				"resourceType": core.MappingNodeFromString("aws/lambda/permission"),
				"sourceArn":    core.MappingNodeFromString("arn:aws:events:us-west-2:123456789012:rule/old"),
			}},
		}},
	}}
	out := s.stage(prior, icResolvedChanges(icRuleARN))
	s.Require().Len(out.ModifiedFields, 1)
	s.Equal(icLeaf("sourceArn"), out.ModifiedFields[0].FieldPath)
	s.Equal(icRuleARN, core.StringValue(out.ModifiedFields[0].NewValue))
	// resourceType is unchanged.
	s.Equal([]string{icLeaf("resourceType")}, out.UnchangedFields)
}

func (s *IntermediaryChangesSuite) Test_collect_known_on_deploy() {
	out := s.stage(&core.MappingNode{}, icKnownOnDeployChanges())
	// resourceType is always resolvable; the derived ARN is known on deploy.
	s.Require().Len(out.NewFields, 1)
	s.Equal(icLeaf("resourceType"), out.NewFields[0].FieldPath)
	s.Equal([]string{icLeaf("sourceArn")}, out.FieldChangesKnownOnDeploy)
}

func (s *IntermediaryChangesSuite) Test_intermediary_link_data() {
	node := IntermediaryLinkData(DeployedIntermediary{
		Identity: icIdentity(),
		Leaves: map[string]*core.MappingNode{
			"sourceArn":    core.MappingNodeFromString(icRuleARN),
			"functionName": core.MappingNodeFromString(icFnARN),
		},
	})
	entry := node.Fields["intermediaries"].Fields[icResourceID]
	s.Require().NotNil(entry)
	s.Equal("aws/lambda/permission", core.StringValue(entry.Fields["resourceType"]))
	s.Equal(icRuleARN, core.StringValue(entry.Fields["sourceArn"]))
	s.Equal(icFnARN, core.StringValue(entry.Fields["functionName"]))
}

func (s *IntermediaryChangesSuite) Test_intermediary_leaf_path() {
	s.Equal(
		"$[\"intermediaries\"][\""+icResourceID+"\"][\"sourceArn\"]",
		IntermediaryLeafPath(icResourceID, "sourceArn"),
	)
}

func TestIntermediaryChangesSuite(t *testing.T) {
	suite.Run(t, new(IntermediaryChangesSuite))
}
