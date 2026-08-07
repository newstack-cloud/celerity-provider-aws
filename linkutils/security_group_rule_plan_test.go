//go:build unit

package linkutils

import (
	"errors"
	"fmt"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/require"
)

func linkInputForResources(
	resourceAName, resourceBName string,
) *provider.LinkUpdateIntermediaryResourcesInput {
	return &provider.LinkUpdateIntermediaryResourcesInput{
		ResourceAInfo: &provider.ResourceInfo{ResourceName: resourceAName},
		ResourceBInfo: &provider.ResourceInfo{ResourceName: resourceBName},
	}
}

func ruleRefs(count int, port int32) []SecurityGroupRuleRef {
	rules := make([]SecurityGroupRuleRef, count)
	for i := range rules {
		rules[i] = SecurityGroupRuleRef{
			PairedSecurityGroupID: fmt.Sprintf("sg-%d", i),
			Port:                  port,
		}
	}
	return rules
}

func TestPlanSecurityGroupRuleAddsAnAbsentRule(t *testing.T) {
	outcome, err := planSecurityGroupRule(
		ruleRefs(3, 443),
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-endpoint", Port: 443},
		DefaultSecurityGroupRuleLimits(),
	)

	require.NoError(t, err)
	require.Equal(t, SecurityGroupRuleFits, outcome)
}

// A rule authorised by an earlier deployment carries a different rule ID, so matching on
// what the rule permits rather than its identity is what stops the allocator adding a
// second copy of it on every deploy and burning the budget down to nothing.
func TestPlanSecurityGroupRuleRecognisesAnExistingRule(t *testing.T) {
	existing := append(
		ruleRefs(3, 443),
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-endpoint", Port: 443},
	)

	outcome, err := planSecurityGroupRule(
		existing,
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-endpoint", Port: 443},
		DefaultSecurityGroupRuleLimits(),
	)

	require.NoError(t, err)
	require.Equal(t, SecurityGroupRuleAlreadyPresent, outcome)
}

// The same paired group on a different port is a different rule. Treating the group
// alone as the identity would silently drop the second port.
func TestPlanSecurityGroupRuleTreatsADifferentPortAsADifferentRule(t *testing.T) {
	existing := []SecurityGroupRuleRef{
		{PairedSecurityGroupID: "sg-db", Port: 5432},
	}

	outcome, err := planSecurityGroupRule(
		existing,
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-db", Port: 6379},
		DefaultSecurityGroupRuleLimits(),
	)

	require.NoError(t, err)
	require.Equal(t, SecurityGroupRuleFits, outcome)
}

func TestPlanSecurityGroupRuleExhaustsAtTheBudget(t *testing.T) {
	limits := DefaultSecurityGroupRuleLimits()

	_, err := planSecurityGroupRule(
		ruleRefs(limits.MaxRulesPerDirection, 443),
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-one-too-many", Port: 443},
		limits,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrSecurityGroupRuleBudgetExhausted)
	require.Contains(t, err.Error(), "55")
}

// A full group must still accept a rule it already holds, or a redeployment of a
// blueprint that fills the group would fail on rules that are already in place.
func TestPlanSecurityGroupRuleAllowsAnExistingRuleInAFullGroup(t *testing.T) {
	limits := DefaultSecurityGroupRuleLimits()
	existing := ruleRefs(limits.MaxRulesPerDirection, 443)

	outcome, err := planSecurityGroupRule(existing, existing[0], limits)

	require.NoError(t, err)
	require.Equal(t, SecurityGroupRuleAlreadyPresent, outcome)
}

// The budget has to sit under the AWS quota rather than on it, so rules an author adds
// outside Bluelink do not push the group over the limit at the moment a link runs.
func TestDefaultSecurityGroupRuleLimitsKeepHeadroomUnderTheQuota(t *testing.T) {
	const awsRulesPerDirectionQuota = 60

	limits := DefaultSecurityGroupRuleLimits()

	require.Less(t, limits.MaxRulesPerDirection, awsRulesPerDirectionQuota)
	require.Positive(t, limits.MaxRulesPerDirection)
}

// The allocator reports a group ID at best; the link is the only place that knows which
// two resources the author linked to reach the limit.
func TestWrapRuleBudgetErrorNamesBothResources(t *testing.T) {
	wrapped := wrapRuleBudgetError(
		fmt.Errorf("%w: full", ErrSecurityGroupRuleBudgetExhausted),
		linkInputForResources("getOrderFunction", "ordersQueue"),
	)

	require.ErrorIs(t, wrapped, ErrSecurityGroupRuleBudgetExhausted)
	require.Contains(t, wrapped.Error(), "getOrderFunction")
	require.Contains(t, wrapped.Error(), "ordersQueue")
}

func TestWrapRuleBudgetErrorLeavesOtherErrorsAlone(t *testing.T) {
	original := errors.New("the security group could not be read")

	wrapped := wrapRuleBudgetError(
		original,
		linkInputForResources("getOrderFunction", "ordersQueue"),
	)

	require.Equal(t, original, wrapped)
}
