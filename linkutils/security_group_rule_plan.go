package linkutils

import (
	"errors"
	"fmt"
	"slices"
)

// This file holds the pure planning core of the security group rule allocator, the
// networking analogue of the role-access allocator in role_access_policy_plan.go: given
// the rules a group already holds and one rule a link wants, it decides whether the rule
// is already satisfied, fits, or exhausts the group. It performs no AWS I/O.
//
// The allocation strategy involves a single group per direction, with a hard ceiling:
//
// The IAM allocator spills overflow into additional managed policies, and the obvious
// analogue would be to attach another security group to the same network interface. That
// is not available to a link on the caller side. A Lambda function's groups come from
// its vpcConfig.SecurityGroupIds, which the placement link rewrites in full on every
// deployment, and the deployer runs placement before the access links. A group attached
// by an access link would therefore be dropped on the next deploy, taking its rules with
// it, and the function would lose that access until the links ran again.
//
// So the budget is a ceiling rather than a trigger to spill. Exhausting it is a typed
// error naming both resources. A workload with more than this many distinct linked
// targets is far outside the shape this is built for, and silently under-provisioning it
// would be worse than telling the author.

// ErrSecurityGroupRuleBudgetExhausted is returned (wrapped) when a rule cannot fit in
// its security group. Callers detect it with errors.Is and enrich it with the workload
// and target identity so the message names the two resources the author linked rather
// than a group ID they never wrote.
var ErrSecurityGroupRuleBudgetExhausted = errors.New("security group rule budget exhausted")

// The AWS default quota is 60 rules per security group in each direction. The budget
// keeps headroom under it for rules an author adds outside Bluelink, in the same spirit
// as the IAM allocator's byte budgets.
const defaultMaxRulesPerDirection = 55

// SecurityGroupRuleLimits is the budget the allocator packs against.
type SecurityGroupRuleLimits struct {
	// MaxRulesPerDirection caps the rules the allocator will add to one group in one
	// direction (headroom under the documented per-direction quota).
	MaxRulesPerDirection int
}

// DefaultSecurityGroupRuleLimits returns the conservative budget.
func DefaultSecurityGroupRuleLimits() SecurityGroupRuleLimits {
	return SecurityGroupRuleLimits{MaxRulesPerDirection: defaultMaxRulesPerDirection}
}

// SecurityGroupRuleRef identifies a rule by what it permits, which is what makes two
// rules the same rule as far as the allocator is concerned. Rules are matched on the
// paired group and port rather than on rule ID, because a rule authorised by an earlier
// deployment carries a different ID and must not be counted or added twice.
type SecurityGroupRuleRef struct {
	// PairedSecurityGroupID is the group on the other end of the rule.
	PairedSecurityGroupID string
	// Port is the single TCP port the rule permits.
	Port int32
}

// SecurityGroupRuleOutcome is what the allocator decided about a rule.
type SecurityGroupRuleOutcome int

const (
	// SecurityGroupRuleAlreadyPresent means the group already permits this, so
	// authorising it again would be a duplicate.
	SecurityGroupRuleAlreadyPresent SecurityGroupRuleOutcome = iota
	// SecurityGroupRuleFits means the rule is absent and within budget.
	SecurityGroupRuleFits
)

// planSecurityGroupRule decides whether a rule needs to be authorised, is already there,
// or cannot fit. existing is every rule the group currently holds in the direction the
// desired rule belongs to.
func planSecurityGroupRule(
	existing []SecurityGroupRuleRef,
	desired SecurityGroupRuleRef,
	limits SecurityGroupRuleLimits,
) (SecurityGroupRuleOutcome, error) {
	if slices.Contains(existing, desired) {
		return SecurityGroupRuleAlreadyPresent, nil
	}

	// Checked against the rules already in place rather than against a count the caller
	// tracks, so a group filled by an earlier deployment is measured as it actually is.
	if len(existing) >= limits.MaxRulesPerDirection {
		return 0, fmt.Errorf(
			"%w: the security group already holds %d of a maximum %d rules",
			ErrSecurityGroupRuleBudgetExhausted,
			len(existing),
			limits.MaxRulesPerDirection,
		)
	}

	return SecurityGroupRuleFits, nil
}
