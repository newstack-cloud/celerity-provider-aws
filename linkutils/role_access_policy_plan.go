package linkutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"sort"
	"strconv"
	"strings"
)

// ErrAccessPolicyBudgetExhausted is returned (wrapped) when a grant cannot fit in
// the role's inline policy or any managed policy slot. Callers detect it with
// errors.Is and enrich it with the role and link identity for a clear message.
var ErrAccessPolicyBudgetExhausted = errors.New("execution role permission budget exhausted")

// This file holds the pure planning core of the role-access permission allocator:
// given the role's current Bluelink-managed policy slots and a single grant
// (a statement keyed by Sid), it decides where the statement lives and produces a
// mutation plan. It performs no AWS I/O, the
// IAM read/discovery and write/apply live alongside it and call into here.
//
// Allocation strategy — inline-first with managed overflow, respecting AWS limits:
//   - One inline policy on the role carries statements up to the inline budget
//     (the role's aggregate inline limit is 10,240 chars; we keep headroom).
//   - Overflow spills into managed policies (each capped at 6,144 chars; we keep
//     headroom), up to a bounded number of attachments so practitioner-attached
//     managed policies still have room under the per-role attachment limit.
// Placement is deterministic and a grant only ever moves slots when its current
// slot can no longer hold it, which keeps the steady-state IAM footprint flat.

const (
	// inlineAccessPolicyName is the single Bluelink-managed inline policy on a role.
	inlineAccessPolicyName = "bluelink-link-access"
	// managedAccessPolicyPrefix prefixes the Bluelink-managed overflow policies,
	// e.g. "bluelink-link-access-1".
	managedAccessPolicyPrefix = "bluelink-link-access-"
)

// AccessPolicyLimits are the size/count budgets the allocator packs against. The
// defaults keep headroom under the documented IAM quotas and leave attachment
// budget for practitioner-managed policies.
type AccessPolicyLimits struct {
	// MaxInlineBytes caps the inline policy document (headroom under the 10,240
	// aggregate inline-policy limit per role).
	MaxInlineBytes int
	// MaxManagedBytes caps each managed policy document (headroom under 6,144).
	MaxManagedBytes int
	// MaxManagedSlots caps how many managed policies the allocator will attach,
	// leaving room under the per-role managed-policy attachment limit (10 default).
	MaxManagedSlots int
}

// DefaultAccessPolicyLimits returns conservative budgets with headroom under the
// IAM quotas.
func DefaultAccessPolicyLimits() AccessPolicyLimits {
	return AccessPolicyLimits{
		MaxInlineBytes:  9000,
		MaxManagedBytes: 5500,
		MaxManagedSlots: 5,
	}
}

// AccessPolicySlotKind distinguishes an inline role policy from a managed policy.
type AccessPolicySlotKind int

const (
	AccessPolicyInline AccessPolicySlotKind = iota
	AccessPolicyManaged
)

// AccessPolicySlot is one Bluelink-managed policy resource and the statements it
// holds, keyed by Sid. ARN is only set for managed slots that already exist.
type AccessPolicySlot struct {
	Kind       AccessPolicySlotKind
	Name       string
	ARN        string
	Statements map[string]json.RawMessage
}

// AccessPolicyPlan is the set of mutations to bring the role's policies in line
// with the desired statement set, plus the slot the grant now lives in.
type AccessPolicyPlan struct {
	// Upserts are slots whose document must be written (inline put, or managed
	// create / new version).
	Upserts []AccessPolicySlot
	// Deletes are slots that are now empty and must be removed (inline delete, or
	// managed detach + delete).
	Deletes []AccessPolicySlot
	// PlacedSlot is the name of the slot the grant was placed in, recorded in link
	// data so later reconciles target it directly. Empty when the grant was removed.
	PlacedSlot string
}

// planRoleAccess computes the mutation plan for upserting or removing a single
// statement (by Sid) across the role's current policy slots.
func planRoleAccess(
	current []AccessPolicySlot,
	sid string,
	statement json.RawMessage,
	removeOnly bool,
	limits AccessPolicyLimits,
) (AccessPolicyPlan, error) {
	working, originalNames := cloneSlots(current)

	changed := map[string]bool{}
	removeStatement(working, sid, changed)

	placedSlot := ""
	if !removeOnly {
		target, err := selectSlot(working, statement, limits)
		if err != nil {
			return AccessPolicyPlan{}, err
		}
		target.Statements[sid] = statement
		changed[target.Name] = true
		placedSlot = target.Name
	}

	return buildPlan(working, originalNames, changed, placedSlot), nil
}

func cloneSlots(current []AccessPolicySlot) (map[string]*AccessPolicySlot, map[string]bool) {
	working := make(map[string]*AccessPolicySlot, len(current))
	originalNames := make(map[string]bool, len(current))
	for _, slot := range current {
		statements := make(map[string]json.RawMessage, len(slot.Statements))
		maps.Copy(statements, slot.Statements)
		working[slot.Name] = &AccessPolicySlot{
			Kind:       slot.Kind,
			Name:       slot.Name,
			ARN:        slot.ARN,
			Statements: statements,
		}
		originalNames[slot.Name] = true
	}
	return working, originalNames
}

func removeStatement(working map[string]*AccessPolicySlot, sid string, changed map[string]bool) {
	for name, slot := range working {
		if _, ok := slot.Statements[sid]; ok {
			delete(slot.Statements, sid)
			changed[name] = true
		}
	}
}

// Returns the slot to place the statement in, materialising the inline
// slot or a new managed slot as needed. It prefers the inline slot, then existing
// managed slots in order, then a new managed slot.
func selectSlot(
	working map[string]*AccessPolicySlot,
	statement json.RawMessage,
	limits AccessPolicyLimits,
) (*AccessPolicySlot, error) {
	inline := working[inlineAccessPolicyName]
	if inline == nil {
		inline = &AccessPolicySlot{
			Kind:       AccessPolicyInline,
			Name:       inlineAccessPolicyName,
			Statements: map[string]json.RawMessage{},
		}
	}
	if fitsWith(inline.Statements, statement, limits.MaxInlineBytes) {
		working[inlineAccessPolicyName] = inline
		return inline, nil
	}

	for _, name := range sortedManagedSlotNames(working) {
		slot := working[name]
		if fitsWith(slot.Statements, statement, limits.MaxManagedBytes) {
			return slot, nil
		}
	}

	if managedSlotCount(working) < limits.MaxManagedSlots {
		slot := &AccessPolicySlot{
			Kind:       AccessPolicyManaged,
			Name:       nextManagedSlotName(working),
			Statements: map[string]json.RawMessage{},
		}
		if fitsWith(slot.Statements, statement, limits.MaxManagedBytes) {
			working[slot.Name] = slot
			return slot, nil
		}
	}

	return nil, fmt.Errorf(
		"%w: the statement does not fit in the inline policy or any of the %d managed "+
			"policy slots; reduce the number of access links on this role or request an "+
			"IAM quota increase",
		ErrAccessPolicyBudgetExhausted,
		limits.MaxManagedSlots,
	)
}

func buildPlan(
	working map[string]*AccessPolicySlot,
	originalNames map[string]bool,
	changed map[string]bool,
	placedSlot string,
) AccessPolicyPlan {
	plan := AccessPolicyPlan{PlacedSlot: placedSlot}
	for _, name := range sortedSlotNames(working) {
		slot := working[name]
		switch {
		case len(slot.Statements) == 0:
			if originalNames[name] {
				plan.Deletes = append(plan.Deletes, *slot)
			}
		case changed[name]:
			plan.Upserts = append(plan.Upserts, *slot)
		}
	}
	return plan
}

func fitsWith(statements map[string]json.RawMessage, candidate json.RawMessage, maxBytes int) bool {
	size := policyDocumentSize(statements) + len(candidate) + 1
	return size <= maxBytes
}

func policyDocumentSize(statements map[string]json.RawMessage) int {
	const wrapper = len(`{"Version":"2012-10-17","Statement":[]}`)
	size := wrapper
	for _, statement := range statements {
		size += len(statement) + 1
	}

	return size
}

func managedSlotCount(working map[string]*AccessPolicySlot) int {
	count := 0
	for name := range working {
		if strings.HasPrefix(name, managedAccessPolicyPrefix) {
			count += 1
		}
	}

	return count
}

func sortedManagedSlotNames(working map[string]*AccessPolicySlot) []string {
	var names []string
	for name := range working {
		if strings.HasPrefix(name, managedAccessPolicyPrefix) {
			names = append(names, name)
		}
	}
	sort.Slice(names, func(i, j int) bool {
		return managedSlotIndex(names[i]) < managedSlotIndex(names[j])
	})
	return names
}

func sortedSlotNames(working map[string]*AccessPolicySlot) []string {
	names := make([]string, 0, len(working))
	for name := range working {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Returns the lowest unused managed slot name so deletes free
// up indices for reuse.
func nextManagedSlotName(working map[string]*AccessPolicySlot) string {
	used := map[int]bool{}
	for name := range working {
		if strings.HasPrefix(name, managedAccessPolicyPrefix) {
			used[managedSlotIndex(name)] = true
		}
	}
	for index := 1; ; index++ {
		if !used[index] {
			return managedAccessPolicyPrefix + strconv.Itoa(index)
		}
	}
}

func managedSlotIndex(name string) int {
	index, err := strconv.Atoi(strings.TrimPrefix(name, managedAccessPolicyPrefix))
	if err != nil {
		return 0
	}
	return index
}
