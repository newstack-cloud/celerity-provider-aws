package overlays

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// AWS enforces minimum subnet requirements on VPC-placed grouping resources that
// surface only as opaque API errors at apply time: an RDS DB subnet group and an
// RDS Proxy must span at least 2 availability zones, and an ElastiCache subnet
// group needs at least one subnet. These minimums are validated here with
// actionable diagnostics instead.
//
// Subnet IDs are typically wired from another resource's computed output (e.g.
// an aws/flex/vpc's privateSubnetIds), which only resolves at deploy time. For
// a flex VPC in reference mode the topology is not knowable any earlier. The
// deploy-time ValidateResolvedSpec hook is therefore the authoritative guard;
// the plan-time CustomValidate only fires for non-empty literal subnet ID
// lists, as anything else is not provably invalid before resolution.
//
// The check counts subnet IDs rather than distinct availability zones, as zone
// membership would need an EC2 lookup. A count below the minimum can never span
// enough zones, which catches every preset-shaped failure (flex VPC tiers hold
// one subnet per zone); same-zone duplicates are left to AWS's own error.
type minSubnetIDsConstraint struct {
	resourceType string
	field        string
	minimum      int
	requirement  string
}

func init() {
	constraints := []minSubnetIDsConstraint{
		{
			resourceType: "aws/rds/dbSubnetGroup",
			field:        "subnetIds",
			minimum:      2,
			requirement:  "at least 2 subnets in different availability zones",
		},
		{
			resourceType: "aws/rds/dbProxy",
			field:        "vpcSubnetIds",
			minimum:      2,
			requirement:  "at least 2 subnets in different availability zones",
		},
		{
			resourceType: "aws/elasticache/subnetGroup",
			field:        "subnetIds",
			minimum:      1,
			requirement:  "at least 1 subnet",
		},
	}

	for _, constraint := range constraints {
		RegisterBehaviour(constraint.resourceType, &Behaviour{
			CustomValidate:       constraint.customValidate(),
			ValidateResolvedSpec: constraint.validateResolvedSpec(),
		})
	}
}

func (c minSubnetIDsConstraint) validateResolvedSpec() SpecTransform {
	return func(spec *core.MappingNode) error {
		count, countable := c.countSubnetIDs(spec)
		if !countable || count >= c.minimum {
			return nil
		}
		return fmt.Errorf("%s", c.message(count))
	}
}

// Validates literal subnet ID lists at the earliest opportunity; a field wired
// from a reference is not countable until deploy time and is skipped here.
// The blueprint framework will in some cases rebuild a reference-valued list as an empty
// (non-nil) Items slice, so an empty list at plan time is indistinguishable
// from an unresolved reference and is left to the deploy-time hook too.
func (c minSubnetIDsConstraint) customValidate() func(
	context.Context,
	*provider.ResourceValidateInput,
) (*provider.ResourceValidateOutput, error) {
	return func(
		_ context.Context,
		input *provider.ResourceValidateInput,
	) (*provider.ResourceValidateOutput, error) {
		diagnostics := []*core.Diagnostic{}
		if input != nil && input.SchemaResource != nil {
			count, countable := c.countSubnetIDs(input.SchemaResource.Spec)
			if countable && count > 0 && count < c.minimum {
				diagnostics = append(diagnostics, &core.Diagnostic{
					Level:   core.DiagnosticLevelError,
					Message: c.message(count),
				})
			}
		}
		return &provider.ResourceValidateOutput{
			Diagnostics: diagnostics,
		}, nil
	}
}

// Returns countable = false when the field is absent or is a single
// reference rather than an array, in which case other validation
// (required fields, type checks, the deploy-time hook) is responsible.
func (c minSubnetIDsConstraint) countSubnetIDs(spec *core.MappingNode) (int, bool) {
	subnetIDs, hasField := pluginutils.GetValueByPath("$."+c.field, spec)
	if !hasField || subnetIDs == nil {
		return 0, false
	}
	if subnetIDs.Items == nil && subnetIDs.StringWithSubstitutions != nil {
		return 0, false
	}
	return len(subnetIDs.Items), true
}

func (c minSubnetIDsConstraint) message(count int) string {
	return fmt.Sprintf(
		"%s requires %s to contain %s, but %d %s provided. "+
			"When placing this resource in an aws/flex/vpc using its privateSubnetIds output, "+
			"use a preset that provides enough private subnets (\"standard\" and \"isolated\" "+
			"provide 3 across distinct availability zones); for a flex VPC in reference mode, "+
			"make sure the referenced VPC has enough private-tier subnets.",
		c.resourceType,
		c.field,
		c.requirement,
		count,
		pluralWas(count),
	)
}

func pluralWas(count int) string {
	if count == 1 {
		return "was"
	}
	return "were"
}
