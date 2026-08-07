//go:build integration

package e2e

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Assertions over the individual rules of a security group, shared by the tests that
// check what a link opened between two groups.

func egressRules(rules []ec2types.SecurityGroupRule) []ec2types.SecurityGroupRule {
	return rulesByDirection(rules, true)
}

func ingressRules(rules []ec2types.SecurityGroupRule) []ec2types.SecurityGroupRule {
	return rulesByDirection(rules, false)
}

func rulesByDirection(
	rules []ec2types.SecurityGroupRule,
	egress bool,
) []ec2types.SecurityGroupRule {
	matched := []ec2types.SecurityGroupRule{}
	for _, rule := range rules {
		if aws.ToBool(rule.IsEgress) == egress {
			matched = append(matched, rule)
		}
	}

	return matched
}

func hasGroupEgress(rules []ec2types.SecurityGroupRule, groupID string, port int32) bool {
	return hasGroupRule(egressRules(rules), groupID, port)
}

func hasGroupIngress(rules []ec2types.SecurityGroupRule, groupID string, port int32) bool {
	return hasGroupRule(ingressRules(rules), groupID, port)
}

func hasGroupRule(rules []ec2types.SecurityGroupRule, groupID string, port int32) bool {
	for _, rule := range rules {
		if rule.ReferencedGroupInfo == nil {
			continue
		}
		if aws.ToString(rule.ReferencedGroupInfo.GroupId) == groupID &&
			aws.ToInt32(rule.FromPort) == port {
			return true
		}
	}

	return false
}

// The default egress rule AWS attaches to a new security group is all protocols
// to 0.0.0.0/0, which reports as protocol "-1".
func hasAllTrafficEgress(rules []ec2types.SecurityGroupRule) bool {
	for _, rule := range egressRules(rules) {
		if aws.ToString(rule.IpProtocol) == "-1" {
			return true
		}
	}

	return false
}
