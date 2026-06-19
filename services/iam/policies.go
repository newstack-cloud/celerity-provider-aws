package iam

import (
	"encoding/json"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
)

type iamPoliciesDiffResult struct {
	toAdd    []*core.MappingNode
	toUpdate []*core.MappingNode
	toRemove []string
}

// Diff checks are carried out on IAM policies, as policies are in lists,
// using the position-based information in the change set is not enough to
// catch policies that have been removed as in the new changes, the policy
// in the same position might be completely different.
func diffIAMPolicies(currentPolicies *core.MappingNode, newPolicies *core.MappingNode) *iamPoliciesDiffResult {
	result := &iamPoliciesDiffResult{}

	// Create maps for easier comparison
	currentMap := make(map[string]*core.MappingNode)
	if currentPolicies != nil {
		for _, policy := range currentPolicies.Items {
			policyName := core.StringValue(policy.Fields["policyName"])
			currentMap[policyName] = policy
		}
	}
	newMap := make(map[string]*core.MappingNode)
	if newPolicies != nil {
		for _, policy := range newPolicies.Items {
			policyName := core.StringValue(policy.Fields["policyName"])
			newMap[policyName] = policy
		}
	}
	// Determine what needs to be added, updated, or removed
	for name, policy := range newMap {
		if currentPolicy, exists := currentMap[name]; !exists {
			result.toAdd = append(result.toAdd, policy)
		} else if !policiesEqual(currentPolicy, policy) {
			result.toUpdate = append(result.toUpdate, policy)
		}
	}
	for name := range currentMap {
		if _, exists := newMap[name]; !exists {
			result.toRemove = append(result.toRemove, name)
		}
	}

	return result
}

// A simple serialised JSON comparison is used to compare policy documents
// in the current implementation.
func policiesEqual(policy1, policy2 *core.MappingNode) bool {
	doc1JSON, err1 := json.Marshal(policy1.Fields["policyDocument"])
	doc2JSON, err2 := json.Marshal(policy2.Fields["policyDocument"])

	if err1 != nil || err2 != nil {
		return false
	}

	return string(doc1JSON) == string(doc2JSON)
}
