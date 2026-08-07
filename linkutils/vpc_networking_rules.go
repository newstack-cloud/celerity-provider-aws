package linkutils

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
)

// Tags a security group rule with the link that created it.
//
// This is the networking analogue of the SID an IAM grant carries: a group is written
// to by several parties (the placement link, each access link, and the VPC resource),
// and the tag is what lets each of them find and remove exactly its own rules while
// leaving everyone else's alone. A rule with no link tag belongs to another writer and
// is never touched.
func securityGroupRuleTagSpecifications(linkID string) []ec2types.TagSpecification {
	return []ec2types.TagSpecification{
		{
			ResourceType: ec2types.ResourceTypeSecurityGroupRule,
			Tags:         []ec2types.Tag{utils.CreateTagBlueprintLinkID(linkID)},
		},
	}
}

// Removes every rule a link created on a security group.
//
// Rules used to be left in place on destroy, which was safe only while every workload
// shared the flex VPC's one security group: a dangling caller-to-target rule granted
// nothing that some other caller did not already have. Now that a placed workload has
// its own group, a rule outliving its link is a grant that workload should no longer
// hold, so it has to go.
func revokeLinkRules(
	ctx context.Context,
	ec2Service ec2service.Service,
	securityGroupID string,
	linkID string,
) error {
	ingressRuleIDs, egressRuleIDs, err := linkRuleIDs(ctx, ec2Service, securityGroupID, linkID)
	if err != nil {
		return err
	}

	if len(ingressRuleIDs) > 0 {
		_, err = ec2Service.RevokeSecurityGroupIngress(
			ctx,
			&ec2.RevokeSecurityGroupIngressInput{
				GroupId:              aws.String(securityGroupID),
				SecurityGroupRuleIds: ingressRuleIDs,
			},
		)
		if err := ignoreMissingRuleError(err); err != nil {
			return err
		}
	}

	if len(egressRuleIDs) > 0 {
		_, err = ec2Service.RevokeSecurityGroupEgress(
			ctx,
			&ec2.RevokeSecurityGroupEgressInput{
				GroupId:              aws.String(securityGroupID),
				SecurityGroupRuleIds: egressRuleIDs,
			},
		)
		if err := ignoreMissingRuleError(err); err != nil {
			return err
		}
	}

	return nil
}

// Rules are revoked by ID rather than by matching on their contents, because two links
// can legitimately produce identical permissions on the same group. Revoking by shape
// would take away a rule another link still depends on.
func linkRuleIDs(
	ctx context.Context,
	ec2Service ec2service.Service,
	securityGroupID string,
	linkID string,
) (ingressRuleIDs []string, egressRuleIDs []string, err error) {
	output, err := ec2Service.DescribeSecurityGroupRules(
		ctx,
		&ec2.DescribeSecurityGroupRulesInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("group-id"),
					Values: []string{securityGroupID},
				},
				{
					Name: aws.String("tag-key"),
					Values: []string{
						fmt.Sprintf("%s%s", utils.TagBlueprintLinkIDPrefix, linkID),
					},
				},
			},
		},
	)
	if err != nil {
		return nil, nil, err
	}
	if output == nil {
		return nil, nil, nil
	}

	for _, rule := range output.SecurityGroupRules {
		ruleID := aws.ToString(rule.SecurityGroupRuleId)
		if ruleID == "" {
			continue
		}
		if aws.ToBool(rule.IsEgress) {
			egressRuleIDs = append(egressRuleIDs, ruleID)
			continue
		}
		ingressRuleIDs = append(ingressRuleIDs, ruleID)
	}

	return ingressRuleIDs, egressRuleIDs, nil
}

// A rule that is already gone is the expected outcome of a retried teardown, not a
// failure.
func ignoreMissingRuleError(err error) error {
	if err == nil {
		return nil
	}

	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		switch apiErr.ErrorCode() {
		case "InvalidSecurityGroupRuleId.NotFound",
			"InvalidPermission.NotFound",
			"InvalidGroup.NotFound":
			return nil
		}
	}

	return err
}
