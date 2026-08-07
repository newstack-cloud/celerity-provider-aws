package linkutils

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
)

// This file is the AWS I/O layer of the security group rule allocator. It reads the
// rules a group currently holds, asks the pure planner ([planSecurityGroupRule]) what to
// do, and authorises the rule if it is needed, all under a lock on the group so the
// read-modify-write is serialised the way the role allocator serialises per role.

// Guards the discover-plan-apply cycle per security group.
//
// This is what keeps two links off the same group. Links made ready together are
// deployed concurrently, so two of them reaching the same group would otherwise each
// read the rules, plan against what they saw, and authorise on top of a group the other
// had already changed. Concurrent deployments of different blueprint instances do not
// contend here because each workload gets a group of its own.
var securityGroupRuleLocks sync.Map

func lockSecurityGroup(securityGroupID string) func() {
	value, _ := securityGroupRuleLocks.LoadOrStore(securityGroupID, &sync.Mutex{})
	mutex, ok := value.(*sync.Mutex)
	if !ok {
		// Nothing else writes this map, so this is unreachable. Returning an
		// unlocked no-op rather than panicking keeps a deployment alive if it ever
		// stops being true, at the cost of the exclusion this provides.
		return func() {}
	}
	mutex.Lock()

	return mutex.Unlock
}

// RuleDirection is which side of a group a rule sits on.
type RuleDirection int

const (
	RuleIngress RuleDirection = iota
	RuleEgress
)

func (d RuleDirection) isEgress() bool {
	return d == RuleEgress
}

// AuthorizeRuleWithinBudget authorises one rule on a group if the group does not already
// permit it and has room, and reports ErrSecurityGroupRuleBudgetExhausted if it does not.
//
// This replaces authorising blind and swallowing the duplicate error: the group has to be
// read to count its rules anyway, so the same read settles whether the rule is already
// there. The duplicate error is still tolerated, because between the read and the write
// another writer may have added it.
func AuthorizeRuleWithinBudget(
	ctx context.Context,
	ec2Service ec2service.Service,
	securityGroupID string,
	direction RuleDirection,
	rule SecurityGroupRuleRef,
	linkID string,
) error {
	release := lockSecurityGroup(securityGroupID)
	defer release()

	existing, err := securityGroupRulesInDirection(ctx, ec2Service, securityGroupID, direction)
	if err != nil {
		return err
	}

	outcome, err := planSecurityGroupRule(existing, rule, DefaultSecurityGroupRuleLimits())
	if err != nil {
		return err
	}
	if outcome == SecurityGroupRuleAlreadyPresent {
		return nil
	}

	permissions := securityGroupPairPermissions(rule.PairedSecurityGroupID, rule.Port)
	tags := securityGroupRuleTagSpecifications(linkID)

	if direction.isEgress() {
		_, err = ec2Service.AuthorizeSecurityGroupEgress(
			ctx,
			&ec2.AuthorizeSecurityGroupEgressInput{
				GroupId:           aws.String(securityGroupID),
				IpPermissions:     permissions,
				TagSpecifications: tags,
			},
		)
	} else {
		_, err = ec2Service.AuthorizeSecurityGroupIngress(
			ctx,
			&ec2.AuthorizeSecurityGroupIngressInput{
				GroupId:           aws.String(securityGroupID),
				IpPermissions:     permissions,
				TagSpecifications: tags,
			},
		)
	}

	return ignoreDuplicateRuleError(err)
}

// The rules a group holds in one direction, reduced to what the allocator matches on.
//
// Only rules that reference another security group are counted. A rule permitting a CIDR
// is a different shape that the pairing path never creates, but it still occupies the
// group's budget, so it is counted without being matchable.
func securityGroupRulesInDirection(
	ctx context.Context,
	ec2Service ec2service.Service,
	securityGroupID string,
	direction RuleDirection,
) ([]SecurityGroupRuleRef, error) {
	rules := []SecurityGroupRuleRef{}

	var nextToken *string
	for {
		output, err := ec2Service.DescribeSecurityGroupRules(
			ctx,
			&ec2.DescribeSecurityGroupRulesInput{
				Filters: []ec2types.Filter{
					{
						Name:   aws.String("group-id"),
						Values: []string{securityGroupID},
					},
				},
				NextToken: nextToken,
			},
		)
		if err != nil {
			return nil, err
		}
		if output == nil {
			return rules, nil
		}

		for _, rule := range output.SecurityGroupRules {
			if aws.ToBool(rule.IsEgress) != direction.isEgress() {
				continue
			}

			// Absent on a rule that permits a CIDR rather than another group. Such a
			// rule still occupies the group's budget, so it is counted, but it can
			// never match a pairing rule and is left with an empty paired group.
			pairedGroupID := ""
			if rule.ReferencedGroupInfo != nil {
				pairedGroupID = aws.ToString(rule.ReferencedGroupInfo.GroupId)
			}

			rules = append(rules, SecurityGroupRuleRef{
				PairedSecurityGroupID: pairedGroupID,
				Port:                  aws.ToInt32(rule.ToPort),
			})
		}

		if output.NextToken == nil || aws.ToString(output.NextToken) == "" {
			return rules, nil
		}
		nextToken = output.NextToken
	}
}
