package linkutils

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// CallerNetworking is the VPC attachment of a link's caller, which is the compute running the
// workload, independent of the compute platform. A Lambda function's VpcConfig, an ECS
// task's awsvpc configuration and an EKS pod's networking all reduce to a VPC, the
// subnets the workload's network interfaces live in and the security groups attached to
// them. Keeping the caller platform-agnostic lets the same networking activation serve
// FaaS and containerised compute (ECS/EKS).
type CallerNetworking struct {
	VPCID            string
	SubnetIDs        []string
	SecurityGroupIDs []string
}

func (c CallerNetworking) attached() bool {
	return c.VPCID != "" && len(c.SubnetIDs) > 0 && len(c.SecurityGroupIDs) > 0
}

// NetworkingActivation describes the networking a VPC-attached caller needs opened to
// reach a link target. When the caller is not VPC-attached ActivateLinkNetworking is a
// no-op.
type NetworkingActivation struct {
	// Caller is the VPC attachment of the link's caller (resource A), read from its
	// computed state and reduced to a platform-agnostic form.
	Caller CallerNetworking
	// Region is the AWS region the caller and its target are deployed in.
	Region string

	// AWSService, when set, provisions an interface VPC endpoint so a VPC-isolated
	// caller can reach an AWS managed service. It is the short service segment (e.g.
	// "lambda", "sns", "sqs", "secretsmanager", "ssm", "kms"); the full endpoint
	// service name is com.amazonaws.<region>.<AWSService>.
	AWSService string
	// EndpointType selects the VPC endpoint type. Interface (the default) provisions an
	// interface endpoint with a security group; Gateway provisions a gateway endpoint
	// attached to the caller's route tables (used for S3 and DynamoDB).
	EndpointType ec2types.VpcEndpointType

	// TargetSecurityGroupID and TargetPort, when set, pair the caller's security group
	// with an in-VPC resource's security group (RDS proxy/instance, ElastiCache).
	TargetSecurityGroupID string
	TargetPort            int32
}

// ActivateLinkNetworking opens the networking a VPC-attached caller needs to reach its
// link target, and ref-count-aware removes it on destroy. It is a no-op when the caller
// is not attached to a VPC.
//
// The caller must be connected to a flex VPC (often in reference mode) in the same
// blueprint.
func ActivateLinkNetworking(
	ctx context.Context,
	ec2Service ec2service.Service,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	activation NetworkingActivation,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	if !activation.Caller.attached() {
		return output, nil
	}

	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	instanceID := pluginutils.GetInstanceID(input.ResourceAInfo)

	flexVPCResourceState, err := input.ResourceService.LookupResourceInState(
		ctx,
		&provider.ResourceLookupInput{
			InstanceID:      instanceID,
			ResourceType:    "aws/flex/vpc",
			ExternalID:      activation.Caller.VPCID,
			ProviderContext: providerCtx,
		},
	)
	if err != nil {
		return nil, err
	}

	// Lock the flex VPC to prevent concurrent updates to the same flex VPC from
	// multiple links in the same blueprint.
	err = input.ResourceService.AcquireResourceLock(
		ctx,
		&provider.AcquireResourceLockInput{
			InstanceID:      instanceID,
			ResourceName:    flexVPCResourceState.Name,
			ProviderContext: providerCtx,
		},
	)
	if err != nil {
		return nil, err
	}

	if activation.AWSService != "" {
		if activation.EndpointType == ec2types.VpcEndpointTypeGateway {
			return activateGatewayEndpoint(
				ctx,
				ec2Service,
				input,
				activation,
				flexVPCResourceState,
				output,
			)
		}
		return activateServiceEndpoint(
			ctx,
			ec2Service,
			input,
			activation,
			flexVPCResourceState,
			output,
		)
	}

	if activation.TargetSecurityGroupID != "" {
		return activateSecurityGroupPair(ctx, ec2Service, input, activation, output)
	}

	return output, nil
}

func activateServiceEndpoint(
	ctx context.Context,
	ec2Service ec2service.Service,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	activation NetworkingActivation,
	flexVPCResourceState *state.ResourceState,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	flexVPCNameValue, hasFlexVPCName := pluginutils.GetValueByPath(
		"$.name",
		flexVPCResourceState.SpecData,
	)
	if !hasFlexVPCName {
		return nil, fmt.Errorf("flex VPC name could not be retrieved from the flex VPC resource")
	}
	flexVPCName := core.StringValue(flexVPCNameValue)

	serviceName := serviceEndpointName(activation.Region, activation.AWSService)
	vpcID := activation.Caller.VPCID

	vpcEndpointsOutput, err := ec2Service.DescribeVpcEndpoints(
		ctx,
		&ec2.DescribeVpcEndpointsInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{vpcID}},
				{Name: aws.String("service-name"), Values: []string{serviceName}},
				utils.CreateTagFilterFlexVPCNameForLink(flexVPCName),
			},
		},
	)
	if err != nil {
		return nil, err
	}

	info := &vpcEndpointInfo{
		flexVPCName:         flexVPCName,
		vpcID:               vpcID,
		serviceName:         serviceName,
		privateDNS:          flexVPCHasDNSSupportWithHostnames(flexVPCResourceState),
		callerSubnetIDs:     activation.Caller.SubnetIDs,
		callerSecurityGroup: activation.Caller.SecurityGroupIDs[0],
		region:              activation.Region,
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy &&
		len(vpcEndpointsOutput.VpcEndpoints) > 0 {
		return removeServiceEndpoint(ctx, ec2Service, info, &vpcEndpointsOutput.VpcEndpoints[0], input, output)
	}

	if len(vpcEndpointsOutput.VpcEndpoints) == 0 {
		return createServiceEndpoint(ctx, ec2Service, info, input, output)
	}

	return updateServiceEndpoint(ctx, ec2Service, info, &vpcEndpointsOutput.VpcEndpoints[0], input, output)
}

func createServiceEndpoint(
	ctx context.Context,
	ec2Service ec2service.Service,
	info *vpcEndpointInfo,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	securityGroupOutput, err := ec2Service.CreateSecurityGroup(
		ctx,
		&ec2.CreateSecurityGroupInput{
			GroupName:   aws.String(fmt.Sprintf("%s-vpc-endpoint-access", info.serviceName)),
			VpcId:       aws.String(info.vpcID),
			Description: aws.String(fmt.Sprintf("Security group for %s service access", info.serviceName)),
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeSecurityGroup,
					Tags: []ec2types.Tag{
						utils.CreateTagLinkSecurityGroup(),
						utils.CreateTagBlueprintInstanceName(input.InstanceName),
						utils.CreateTagBlueprintLinkID(input.LinkID),
						utils.CreateTagFlexVPCNameForLink(info.flexVPCName),
						utils.CreateTagBluelinkService(info.serviceName),
					},
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	_, err = ec2Service.AuthorizeSecurityGroupIngress(
		ctx,
		&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:                 securityGroupOutput.GroupId,
			SourceSecurityGroupName: aws.String(info.callerSecurityGroup),
		},
	)
	if err != nil {
		return nil, err
	}

	vpcEndpointOutput, err := ec2Service.CreateVpcEndpoint(
		ctx,
		&ec2.CreateVpcEndpointInput{
			VpcId:             aws.String(info.vpcID),
			PrivateDnsEnabled: aws.Bool(info.privateDNS),
			VpcEndpointType:   ec2types.VpcEndpointTypeInterface,
			ServiceName:       aws.String(info.serviceName),
			SubnetIds:         info.callerSubnetIDs,
			SecurityGroupIds:  []string{aws.ToString(securityGroupOutput.GroupId)},
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeVpcEndpoint,
					Tags: []ec2types.Tag{
						utils.CreateTagLinkVPCEndpoint(),
						utils.CreateTagBlueprintInstanceName(input.InstanceName),
						utils.CreateTagBlueprintLinkID(input.LinkID),
						utils.CreateTagFlexVPCNameForLink(info.flexVPCName),
					},
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	vpcEndpointName := fmt.Sprintf("%sVPCEndpoint", input.ResourceAInfo.ResourceName)
	vpcEndpointLinkData := core.MappingNodeFields(
		vpcEndpointName,
		core.MappingNodeFields(
			"id",
			core.MappingNodeFromString(aws.ToString(vpcEndpointOutput.VpcEndpoint.VpcEndpointId)),
		),
	)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: output.IntermediaryResourceStates,
		LinkData:                   core.MergeMaps(output.LinkData, vpcEndpointLinkData),
		ResourceDataMappings:       output.ResourceDataMappings,
	}, nil
}

func updateServiceEndpoint(
	ctx context.Context,
	ec2Service ec2service.Service,
	info *vpcEndpointInfo,
	endpoint *ec2types.VpcEndpoint,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Tag the VPC endpoint with the current link, for reference counting so it is only
	// removed when the last link that references it is destroyed.
	if !utils.HasVPCEndpointTagForLink(endpoint, input.LinkID) {
		_, err := ec2Service.CreateTags(
			ctx,
			&ec2.CreateTagsInput{
				Resources: []string{aws.ToString(endpoint.VpcEndpointId)},
				Tags:      []ec2types.Tag{utils.CreateTagBlueprintLinkID(input.LinkID)},
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Make sure the VPC endpoint provides a network interface in the caller's subnets.
	if !utils.VPCEndpointInSubnets(endpoint, info.callerSubnetIDs) {
		_, err := ec2Service.ModifyVpcEndpoint(
			ctx,
			&ec2.ModifyVpcEndpointInput{
				VpcEndpointId: endpoint.VpcEndpointId,
				AddSubnetIds:  info.callerSubnetIDs,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Make sure the link is attached to a security group that allows access to the
	// VPC endpoint.
	securityGroupsOutput, err := ec2Service.DescribeSecurityGroups(
		ctx,
		&ec2.DescribeSecurityGroupsInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{info.vpcID}},
				{Name: aws.String(utils.TagBluelinkService), Values: []string{info.serviceName}},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	linkSecurityGroup, err := linkSecurityGroupForEndpoint(
		securityGroupsOutput,
		input.LinkID,
		info.callerSecurityGroup,
	)
	if err != nil {
		return nil, err
	}

	if !linkSecurityGroup.attached {
		return attachLinkToSecurityGroup(ctx, ec2Service, info, linkSecurityGroup, input, output)
	}

	return output, nil
}

func attachLinkToSecurityGroup(
	ctx context.Context,
	ec2Service ec2service.Service,
	info *vpcEndpointInfo,
	linkSecurityGroup *linkSecurityGroupInfo,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	_, err := ec2Service.AuthorizeSecurityGroupIngress(
		ctx,
		&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:                 aws.String(linkSecurityGroup.securityGroupID),
			SourceSecurityGroupName: aws.String(info.callerSecurityGroup),
		},
	)
	if err != nil {
		return nil, err
	}

	_, err = ec2Service.CreateTags(
		ctx,
		&ec2.CreateTagsInput{
			Resources: []string{linkSecurityGroup.securityGroupID},
			Tags:      []ec2types.Tag{utils.CreateTagBlueprintLinkID(input.LinkID)},
		},
	)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func removeServiceEndpoint(
	ctx context.Context,
	ec2Service ec2service.Service,
	info *vpcEndpointInfo,
	endpoint *ec2types.VpcEndpoint,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Remove the VPC endpoint when the current link is the last one holding a reference, otherwise
	// just drop this link's tag.
	if utils.HasVPCEndpointTagForLink(endpoint, input.LinkID) {
		otherLinkTags := utils.GetOtherLinkTagsFromVPCEndpoint(endpoint, input.LinkID)
		if len(otherLinkTags) == 0 {
			_, err := ec2Service.DeleteVpcEndpoints(
				ctx,
				&ec2.DeleteVpcEndpointsInput{
					VpcEndpointIds: []string{aws.ToString(endpoint.VpcEndpointId)},
				},
			)
			if err != nil {
				return nil, err
			}
		} else {
			_, err := ec2Service.DeleteTags(
				ctx,
				&ec2.DeleteTagsInput{
					Resources: []string{aws.ToString(endpoint.VpcEndpointId)},
					Tags:      []ec2types.Tag{utils.CreateTagBlueprintLinkID(input.LinkID)},
				},
			)
			if err != nil {
				return nil, err
			}
		}
	}

	securityGroupsOutput, err := ec2Service.DescribeSecurityGroups(
		ctx,
		&ec2.DescribeSecurityGroupsInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{info.vpcID}},
				{Name: aws.String(utils.TagBluelinkService), Values: []string{info.serviceName}},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	for i := range securityGroupsOutput.SecurityGroups {
		securityGroup := securityGroupsOutput.SecurityGroups[i]
		if !utils.HasSecurityGroupTagForLink(&securityGroup, input.LinkID) {
			continue
		}
		// Remove the security group when the current link is the last one holding a reference,
		// otherwise just drop this link's tag. We do not remove the ingress rule, since
		// the same rule may be shared between links (multiple functions often share a
		// security group).
		otherLinkTags := utils.GetOtherLinkTagsFromSecurityGroup(&securityGroup, input.LinkID)
		if len(otherLinkTags) == 0 {
			_, err := ec2Service.DeleteSecurityGroup(
				ctx,
				&ec2.DeleteSecurityGroupInput{GroupId: securityGroup.GroupId},
			)
			if err != nil {
				return nil, err
			}
		} else {
			_, err := ec2Service.DeleteTags(
				ctx,
				&ec2.DeleteTagsInput{
					Resources: []string{aws.ToString(securityGroup.GroupId)},
					Tags:      []ec2types.Tag{utils.CreateTagBlueprintLinkID(input.LinkID)},
				},
			)
			if err != nil {
				return nil, err
			}
		}
	}

	return output, nil
}

type vpcEndpointInfo struct {
	flexVPCName         string
	vpcID               string
	serviceName         string
	privateDNS          bool
	region              string
	callerSubnetIDs     []string
	callerSecurityGroup string
}

type linkSecurityGroupInfo struct {
	securityGroupID string
	attached        bool
}

func linkSecurityGroupForEndpoint(
	securityGroupsOutput *ec2.DescribeSecurityGroupsOutput,
	linkID string,
	callerSecurityGroup string,
) (*linkSecurityGroupInfo, error) {
	if len(securityGroupsOutput.SecurityGroups) == 0 {
		return nil, errors.New("no security groups found for the service VPC endpoint")
	}

	linkTag := utils.CreateTagBlueprintLinkID(linkID)
	for i := range securityGroupsOutput.SecurityGroups {
		securityGroup := securityGroupsOutput.SecurityGroups[i]
		for _, tag := range securityGroup.Tags {
			if utils.MatchesEC2Tag(tag, linkTag) &&
				utils.HasIngressWithSourceSecurityGroupName(&securityGroup, callerSecurityGroup) {
				return &linkSecurityGroupInfo{
					securityGroupID: aws.ToString(securityGroup.GroupId),
					attached:        true,
				}, nil
			}
		}
	}

	return &linkSecurityGroupInfo{
		securityGroupID: aws.ToString(securityGroupsOutput.SecurityGroups[0].GroupId),
		attached:        false,
	}, nil
}

func flexVPCHasDNSSupportWithHostnames(flexVPCResourceState *state.ResourceState) bool {
	enableDNSSupportValue, _ := pluginutils.GetValueByPath(
		"$.enableDNSSupport",
		flexVPCResourceState.SpecData,
	)
	enableDNSHostnamesValue, _ := pluginutils.GetValueByPath(
		"$.enableDNSHostnames",
		flexVPCResourceState.SpecData,
	)
	return core.BoolValue(enableDNSSupportValue) && core.BoolValue(enableDNSHostnamesValue)
}

func serviceEndpointName(region, service string) string {
	return fmt.Sprintf("com.amazonaws.%s.%s", region, service)
}
