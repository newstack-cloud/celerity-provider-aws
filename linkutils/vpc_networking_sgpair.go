package linkutils

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Opens connectivity between a VPC-attached caller's security group
// and an in-VPC target resource's security group (an RDS proxy/instance, an ElastiCache node)
// on a specific port. It authorises egress on the caller's security group to the target and
// ingress on the target from the caller.
//
// Every rule is tagged with the link that created it, and destroy revokes exactly those
// rules. Rules used to be left in place instead, which was safe only while every
// workload shared the flex VPC's one security group: a dangling caller-to-target rule
// granted nothing that some other caller did not already have. A placed workload now
// has its own group, so a rule outliving its link is a grant that workload should no
// longer hold.
//
// Creation is idempotent, an already-present rule (InvalidPermission.Duplicate) is
// treated as success.
func reconcileSecurityGroupPair(
	ctx context.Context,
	ec2Service ec2service.Service,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	activation NetworkingActivation,
	flexVPCResourceState *state.ResourceState,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	callerSecurityGroupID := activation.Caller.SecurityGroupIDs[0]
	targetSecurityGroupID, err := targetGroupMintedByVPC(
		activation.TargetSecurityGroupIDs,
		flexVPCResourceState,
		input,
	)
	if err != nil {
		return nil, err
	}
	port := activation.TargetPort

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		// Both ends, since this link wrote to both: egress on the caller and ingress
		// on the target.
		if err := revokeLinkRules(ctx, ec2Service, callerSecurityGroupID, input.LinkID); err != nil {
			return nil, err
		}
		if err := revokeLinkRules(ctx, ec2Service, targetSecurityGroupID, input.LinkID); err != nil {
			return nil, err
		}
		return output, nil
	}

	if err := pairSecurityGroups(
		ctx, ec2Service, callerSecurityGroupID, targetSecurityGroupID, port, input.LinkID,
	); err != nil {
		return nil, wrapRuleBudgetError(err, input)
	}

	return output, nil
}

// Picks the group to pair from those the target carries: the one the flex VPC minted from
// its securityGroups list.
//
// A target can carry groups the author manages themselves, and those must be left alone.
// Pairing against one would open a path to every other resource sharing it, which is the
// leak a group per target exists to close. Taking the first of the list would make which
// group that is a matter of ordering.
func targetGroupMintedByVPC(
	targetSecurityGroupIDs []string,
	flexVPCResourceState *state.ResourceState,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (string, error) {
	mintedByName := map[string]string{}
	if flexVPCResourceState != nil {
		node, has := pluginutils.GetValueByPath(
			"$.securityGroupIdsByName",
			flexVPCResourceState.SpecData,
		)
		if has && node != nil {
			for name, idNode := range node.Fields {
				mintedByName[core.StringValue(idNode)] = name
			}
		}
	}

	matched := []string{}
	for _, groupID := range targetSecurityGroupIDs {
		if _, minted := mintedByName[groupID]; minted {
			matched = append(matched, groupID)
		}
	}

	targetName := pluginutils.GetResourceName(input.ResourceBInfo)

	if len(matched) == 0 {
		return "", fmt.Errorf(
			"%q does not reference a security group created by the flex VPC it is in. "+
				"Add a name to the VPC's securityGroups and reference it from %q, so the "+
				"link has a group of its own to open a path to",
			targetName,
			targetName,
		)
	}
	if len(matched) > 1 {
		sort.Strings(matched)
		return "", fmt.Errorf(
			"%q references %d security groups created by the flex VPC it is in (%s), "+
				"and exactly one is required so the link knows which to open a path to",
			targetName,
			len(matched),
			strings.Join(matched, ", "),
		)
	}

	return matched[0], nil
}

// The allocator knows it ran out of room in a security group; only the link knows which
// two resources the author linked to get there. A budget error that names a group ID
// leaves the author to work out which of their links is the one to remove.
func wrapRuleBudgetError(
	err error,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) error {
	if !errors.Is(err, ErrSecurityGroupRuleBudgetExhausted) {
		return err
	}

	return fmt.Errorf(
		"cannot open a network path from %q to %q: %w",
		pluginutils.GetResourceName(input.ResourceAInfo),
		pluginutils.GetResourceName(input.ResourceBInfo),
		err,
	)
}

// Opens a one-way path from a caller's security group to a destination group on a port:
// ingress on the destination from the caller, and egress on the caller to the
// destination.
//
// Both halves are required and it is the same pair every time, whether the destination
// is a database, a cache or a VPC endpoint. Keeping them in one place is what stops the
// two sides drifting apart, which is how the endpoint path ended up authorising ingress
// that could never match and no egress at all.
func pairSecurityGroups(
	ctx context.Context,
	ec2Service ec2service.Service,
	callerSecurityGroupID string,
	destinationSecurityGroupID string,
	port int32,
	linkID string,
) error {
	if err := AuthorizeRuleWithinBudget(
		ctx,
		ec2Service,
		destinationSecurityGroupID,
		RuleIngress,
		SecurityGroupRuleRef{PairedSecurityGroupID: callerSecurityGroupID, Port: port},
		linkID,
	); err != nil {
		return err
	}

	// Flex VPC security groups have their default allow-all egress revoked, so without
	// this half the caller cannot reach the destination at all.
	return AuthorizeRuleWithinBudget(
		ctx,
		ec2Service,
		callerSecurityGroupID,
		RuleEgress,
		SecurityGroupRuleRef{PairedSecurityGroupID: destinationSecurityGroupID, Port: port},
		linkID,
	)
}

func securityGroupPairPermissions(pairedSecurityGroupID string, port int32) []ec2types.IpPermission {
	return []ec2types.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(port),
			ToPort:     aws.Int32(port),
			UserIdGroupPairs: []ec2types.UserIdGroupPair{
				{GroupId: aws.String(pairedSecurityGroupID)},
			},
		},
	}
}

func ignoreDuplicateRuleError(err error) error {
	if err == nil {
		return nil
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok &&
		apiErr.ErrorCode() == "InvalidPermission.Duplicate" {
		return nil
	}
	return err
}
