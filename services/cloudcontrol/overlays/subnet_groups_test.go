//go:build unit

package overlays

import (
	"context"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/substitutions"
	"github.com/stretchr/testify/suite"
)

type SubnetGroupsOverlaySuite struct {
	suite.Suite
}

func subnetIDsSpec(field string, ids ...string) *core.MappingNode {
	items := []*core.MappingNode{}
	for _, id := range ids {
		items = append(items, core.MappingNodeFromString(id))
	}
	return core.MappingNodeFields(field, core.MappingNodeItems(items...))
}

func (s *SubnetGroupsOverlaySuite) Test_deploy_time_validation_enforces_minimums() {
	testCases := []struct {
		name         string
		resourceType string
		spec         *core.MappingNode
		expectError  bool
	}{
		{
			name:         "rds dbSubnetGroup rejects a single subnet",
			resourceType: "aws/rds/dbSubnetGroup",
			spec:         subnetIDsSpec("subnetIds", "subnet-1"),
			expectError:  true,
		},
		{
			name:         "rds dbSubnetGroup rejects an empty list",
			resourceType: "aws/rds/dbSubnetGroup",
			spec:         subnetIDsSpec("subnetIds"),
			expectError:  true,
		},
		{
			name:         "rds dbSubnetGroup accepts two subnets",
			resourceType: "aws/rds/dbSubnetGroup",
			spec:         subnetIDsSpec("subnetIds", "subnet-1", "subnet-2"),
			expectError:  false,
		},
		{
			name:         "rds dbProxy rejects a single subnet",
			resourceType: "aws/rds/dbProxy",
			spec:         subnetIDsSpec("vpcSubnetIds", "subnet-1"),
			expectError:  true,
		},
		{
			name:         "rds dbProxy accepts two subnets",
			resourceType: "aws/rds/dbProxy",
			spec:         subnetIDsSpec("vpcSubnetIds", "subnet-1", "subnet-2"),
			expectError:  false,
		},
		{
			name:         "elasticache subnetGroup rejects an empty list",
			resourceType: "aws/elasticache/subnetGroup",
			spec:         subnetIDsSpec("subnetIds"),
			expectError:  true,
		},
		{
			name:         "elasticache subnetGroup accepts a single subnet",
			resourceType: "aws/elasticache/subnetGroup",
			spec:         subnetIDsSpec("subnetIds", "subnet-1"),
			expectError:  false,
		},
		{
			name:         "missing field is left to required-field validation",
			resourceType: "aws/rds/dbSubnetGroup",
			spec:         core.MappingNodeFields(),
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			behaviour := BehaviourFor(tc.resourceType)
			s.Require().NotNil(behaviour)
			s.Require().NotNil(behaviour.ValidateResolvedSpec)

			err := behaviour.ValidateResolvedSpec(tc.spec)
			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.resourceType)
				s.Contains(err.Error(), "aws/flex/vpc")
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *SubnetGroupsOverlaySuite) Test_plan_time_validation_counts_literal_lists() {
	behaviour := BehaviourFor("aws/rds/dbSubnetGroup")
	s.Require().NotNil(behaviour)
	s.Require().NotNil(behaviour.CustomValidate)

	output, err := behaviour.CustomValidate(context.Background(), &provider.ResourceValidateInput{
		SchemaResource: &schema.Resource{
			Spec: subnetIDsSpec("subnetIds", "subnet-1"),
		},
	})
	s.Require().NoError(err)
	s.Require().Len(output.Diagnostics, 1)
	s.Equal(core.DiagnosticLevelError, output.Diagnostics[0].Level)
	s.Contains(output.Diagnostics[0].Message, "at least 2 subnets")
}

func (s *SubnetGroupsOverlaySuite) Test_plan_time_validation_skips_reference_wired_lists() {
	behaviour := BehaviourFor("aws/rds/dbSubnetGroup")
	s.Require().NotNil(behaviour)

	// A field wired from a reference (e.g. a flex VPC's privateSubnetIds) only
	// resolves at deploy time, so plan-time validation must not flag it.
	output, err := behaviour.CustomValidate(context.Background(), &provider.ResourceValidateInput{
		SchemaResource: &schema.Resource{
			Spec: core.MappingNodeFields(
				"subnetIds",
				&core.MappingNode{
					StringWithSubstitutions: &substitutions.StringOrSubstitutions{},
				},
			),
		},
	})
	s.Require().NoError(err)
	s.Empty(output.Diagnostics)
}

func (s *SubnetGroupsOverlaySuite) Test_plan_time_validation_skips_empty_lists() {
	// The blueprint framework rebuilds reference-valued lists (e.g.
	// ${resources.appVpc.spec.privateSubnetIds}) as empty non-nil Items slices,
	// so an empty list at plan time may still resolve to enough subnets and
	// must be left to the deploy-time hook.
	for _, resourceType := range []string{"aws/rds/dbSubnetGroup", "aws/elasticache/subnetGroup"} {
		s.Run(resourceType, func() {
			behaviour := BehaviourFor(resourceType)
			s.Require().NotNil(behaviour)

			output, err := behaviour.CustomValidate(context.Background(), &provider.ResourceValidateInput{
				SchemaResource: &schema.Resource{
					Spec: subnetIDsSpec("subnetIds"),
				},
			})
			s.Require().NoError(err)
			s.Empty(output.Diagnostics)
		})
	}
}

func TestSubnetGroupsOverlaySuite(t *testing.T) {
	suite.Run(t, new(SubnetGroupsOverlaySuite))
}
