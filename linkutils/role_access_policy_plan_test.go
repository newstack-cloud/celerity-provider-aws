//go:build unit

package linkutils

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RoleAccessPolicyPlanSuite struct {
	suite.Suite
}

func planStmt(sid string, padBytes int) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(
		`{"Sid":%q,"Effect":"Allow","Action":"s3:GetObject","Resource":%q}`,
		sid, "arn:"+strings.Repeat("x", padBytes),
	))
}

func planSmallLimits() AccessPolicyLimits {
	// Small budgets so a handful of statements force overflow.
	return AccessPolicyLimits{MaxInlineBytes: 250, MaxManagedBytes: 250, MaxManagedSlots: 2}
}

func (s *RoleAccessPolicyPlanSuite) Test_places_first_grant_inline() {
	plan, err := planRoleAccess(nil, "sid-1", planStmt("sid-1", 0), false, DefaultAccessPolicyLimits())
	s.Require().NoError(err)

	s.Equal(inlineAccessPolicyName, plan.PlacedSlot)
	s.Require().Len(plan.Upserts, 1)
	s.Equal(inlineAccessPolicyName, plan.Upserts[0].Name)
	s.Equal(AccessPolicyInline, plan.Upserts[0].Kind)
	s.Contains(plan.Upserts[0].Statements, "sid-1")
}

func (s *RoleAccessPolicyPlanSuite) Test_upsert_is_idempotent_on_same_slot() {
	current := []AccessPolicySlot{{
		Kind: AccessPolicyInline,
		Name: inlineAccessPolicyName,
		Statements: map[string]json.RawMessage{
			"sid-1": planStmt("sid-1", 0),
		},
	}}
	// Re-applying the same sid replaces in place; still inline, one upsert.
	plan, err := planRoleAccess(current, "sid-1", planStmt("sid-1", 5), false, DefaultAccessPolicyLimits())
	s.Require().NoError(err)

	s.Equal(inlineAccessPolicyName, plan.PlacedSlot)
	s.Empty(plan.Deletes)
	s.Require().Len(plan.Upserts, 1)
	s.Len(plan.Upserts[0].Statements, 1)
}

func (s *RoleAccessPolicyPlanSuite) Test_overflows_to_managed() {
	limits := planSmallLimits()
	var current []AccessPolicySlot

	// With these budgets each slot holds one padded statement: 1 inline + 2 managed
	// slots = capacity 3, so the run exercises inline then overflow to managed.
	placements := map[string]int{}
	for i := range 3 {
		sid := fmt.Sprintf("sid-%d", i)
		plan, err := planRoleAccess(current, sid, planStmt(sid, 60), false, limits)
		s.Require().NoErrorf(err, "grant %d", i)
		current = applyPlanToSlots(current, plan)
		placements[plan.PlacedSlot]++
	}

	s.NotZero(placements[inlineAccessPolicyName], "expected some grants placed inline")
	managed := 0
	for name, n := range placements {
		if strings.HasPrefix(name, managedAccessPolicyPrefix) {
			managed += n
		}
	}
	s.NotZerof(managed, "expected overflow to managed slots, placements: %v", placements)
}

func (s *RoleAccessPolicyPlanSuite) Test_errors_when_budget_exhausted() {
	limits := AccessPolicyLimits{MaxInlineBytes: 200, MaxManagedBytes: 200, MaxManagedSlots: 1}
	var current []AccessPolicySlot

	var lastErr error
	for i := range 20 {
		sid := fmt.Sprintf("sid-%d", i)
		plan, err := planRoleAccess(current, sid, planStmt(sid, 80), false, limits)
		if err != nil {
			lastErr = err
			break
		}
		current = applyPlanToSlots(current, plan)
	}
	s.Require().Error(lastErr, "expected a capacity-exhausted error once inline + 1 managed slot are full")
	s.Contains(lastErr.Error(), "permission budget exhausted")
}

func (s *RoleAccessPolicyPlanSuite) Test_remove_deletes_empty_slot() {
	current := []AccessPolicySlot{{
		Kind: AccessPolicyInline,
		Name: inlineAccessPolicyName,
		Statements: map[string]json.RawMessage{
			"sid-1": planStmt("sid-1", 0),
		},
	}}
	plan, err := planRoleAccess(current, "sid-1", nil, true, DefaultAccessPolicyLimits())
	s.Require().NoError(err)

	s.Empty(plan.PlacedSlot)
	s.Require().Len(plan.Deletes, 1)
	s.Equal(inlineAccessPolicyName, plan.Deletes[0].Name)
	s.Empty(plan.Upserts)
}

func (s *RoleAccessPolicyPlanSuite) Test_remove_one_of_many_rewrites_slot() {
	current := []AccessPolicySlot{{
		Kind: AccessPolicyInline,
		Name: inlineAccessPolicyName,
		Statements: map[string]json.RawMessage{
			"sid-1": planStmt("sid-1", 0),
			"sid-2": planStmt("sid-2", 0),
		},
	}}
	plan, err := planRoleAccess(current, "sid-1", nil, true, DefaultAccessPolicyLimits())
	s.Require().NoError(err)

	s.Empty(plan.Deletes, "slot still has sid-2, should not be deleted")
	s.Require().Len(plan.Upserts, 1)
	s.NotContains(plan.Upserts[0].Statements, "sid-1")
	s.Contains(plan.Upserts[0].Statements, "sid-2")
}

func TestRoleAccessPolicyPlanSuite(t *testing.T) {
	suite.Run(t, new(RoleAccessPolicyPlanSuite))
}

// Dolds a plan back into a slot set so a sequence of grants can
// be simulated in tests.
func applyPlanToSlots(current []AccessPolicySlot, plan AccessPolicyPlan) []AccessPolicySlot {
	byName := map[string]AccessPolicySlot{}
	for _, slot := range current {
		byName[slot.Name] = slot
	}

	for _, slot := range plan.Upserts {
		byName[slot.Name] = slot
	}

	for _, slot := range plan.Deletes {
		delete(byName, slot.Name)
	}

	out := make([]AccessPolicySlot, 0, len(byName))
	for _, slot := range byName {
		out = append(out, slot)
	}

	return out
}
