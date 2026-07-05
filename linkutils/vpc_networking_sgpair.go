package linkutils

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Opens connectivity between a VPC-attached caller's security group
// and an in-VPC target resource's security group (an RDS proxy/instance, an ElastiCache node)
// on a specific port. It authorises egress on the caller's security group to the target and
// ingress on the target from the caller. When the caller and target share the flex VPC's
// managed security group, both are self-referencing rules on that one group.
//
// The rules are shared across links (multiple callers commonly share a security group), so on
// destroy they are left in place rather than ref-counted per rule. A dangling caller↔target
// rule grants no external access. Creation is idempotent, an already-present rule
// (InvalidPermission.Duplicate) is treated as success.
func activateSecurityGroupPair(
	ctx context.Context,
	ec2Service ec2service.Service,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	activation NetworkingActivation,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return output, nil
	}

	callerSecurityGroupID := activation.Caller.SecurityGroupIDs[0]
	targetSecurityGroupID := activation.TargetSecurityGroupID
	port := activation.TargetPort

	// For ingress on the target's security group,
	// allow the caller's security group to reach it on the target port.
	if err := authorizeIngressIgnoreDuplicate(
		ctx, ec2Service, targetSecurityGroupID, callerSecurityGroupID, port,
	); err != nil {
		return nil, err
	}

	// For egress on the caller's security group,
	// allow it to reach the target's security group on the target port.
	// (Flex VPC security groups have their default allow-all egress revoked.)
	if err := authorizeEgressIgnoreDuplicate(
		ctx, ec2Service, callerSecurityGroupID, targetSecurityGroupID, port,
	); err != nil {
		return nil, err
	}

	return output, nil
}

func authorizeIngressIgnoreDuplicate(
	ctx context.Context,
	ec2Service ec2service.Service,
	securityGroupID, sourceSecurityGroupID string,
	port int32,
) error {
	_, err := ec2Service.AuthorizeSecurityGroupIngress(
		ctx,
		&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       aws.String(securityGroupID),
			IpPermissions: securityGroupPairPermissions(sourceSecurityGroupID, port),
		},
	)
	return ignoreDuplicateRuleError(err)
}

func authorizeEgressIgnoreDuplicate(
	ctx context.Context,
	ec2Service ec2service.Service,
	securityGroupID, destinationSecurityGroupID string,
	port int32,
) error {
	_, err := ec2Service.AuthorizeSecurityGroupEgress(
		ctx,
		&ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       aws.String(securityGroupID),
			IpPermissions: securityGroupPairPermissions(destinationSecurityGroupID, port),
		},
	)
	return ignoreDuplicateRuleError(err)
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
