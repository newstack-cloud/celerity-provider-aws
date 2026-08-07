package linkutils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink-provider-aws/ec2util"
	"github.com/newstack-cloud/bluelink-provider-aws/flex"
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
// reach a link target. When the caller is not VPC-attached ReconcileLinkNetworking is a
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

	// TargetSecurityGroupIDs and TargetPort, when set, pair the caller's security group
	// with an in-VPC resource's security group (RDS proxy/instance, ElastiCache).
	//
	// Every group the target carries is passed, not one chosen by the link. Exactly one
	// of them should be a group the flex VPC prepared from its securityGroups list, and
	// that is the one paired; the rest belong to the author and are left alone. Picking
	// here would mean guessing, and pairing against an author's own group could open a
	// path to every other resource sharing it.
	TargetSecurityGroupIDs []string
	TargetPort             int32
}

// ReconcileLinkNetworking brings the networking a VPC-attached caller needs to reach its
// link target into line with the link's current state, in either direction: opening it
// on create and update, and taking it away on destroy. It is a no-op when the caller is
// not attached to a VPC, which reaches AWS services over the public internet instead.
//
// Removal is reference counted. An endpoint or security group is shared by every link
// that needs the same service in the same VPC, so a destroy drops this link's tag and
// removes the resource only once no other link holds it.
//
// The caller must be connected to a flex VPC (often in reference mode) in the same
// blueprint.
func ReconcileLinkNetworking(
	ctx context.Context,
	ec2Service ec2service.Service,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	activation NetworkingActivation,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	if !activation.Caller.attached() {
		// On destroy the caller's attachment may already be gone, nothing orders this
		// link against the placement link that clears the function's vpcConfig. Giving
		// up here leaves the endpoint and its security group behind, and that group's
		// ingress rule then blocks the workload group, and with it the whole VPC, from
		// ever being deleted. What this link created is recoverable from its own
		// recorded state, which does not depend on the caller still being attached.
		if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
			return removeRecordedServiceEndpoint(ctx, ec2Service, input, output)
		}
		return output, nil
	}

	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	instanceID := pluginutils.GetInstanceID(input.ResourceAInfo)

	// The flex VPC resource's IDField is its name, not the AWS VPC ID, so the caller's
	// vpcConfig cannot be used to look it up directly. The name is recovered from the
	// VPC's own tag.
	flexVPCName, err := flexVPCNameForVPC(ctx, ec2Service, activation.Caller.VPCID)
	if err != nil {
		return nil, err
	}

	flexVPCResourceState, err := input.ResourceService.LookupResourceInState(
		ctx,
		&provider.ResourceLookupInput{
			InstanceID:      instanceID,
			ResourceType:    "aws/flex/vpc",
			ExternalID:      flexVPCName,
			ProviderContext: providerCtx,
		},
	)
	if err != nil {
		return nil, err
	}
	if flexVPCResourceState == nil {
		return nil, fmt.Errorf(
			"no aws/flex/vpc resource named %q was found in the blueprint instance, "+
				"so the networking for a caller placed in VPC %s cannot be activated",
			flexVPCName,
			activation.Caller.VPCID,
		)
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
			return reconcileGatewayEndpoint(
				ctx,
				ec2Service,
				input,
				activation,
				flexVPCResourceState,
				output,
			)
		}
		return reconcileServiceEndpoint(
			ctx,
			ec2Service,
			input,
			activation,
			flexVPCResourceState,
			output,
		)
	}

	if len(activation.TargetSecurityGroupIDs) > 0 {
		return reconcileSecurityGroupPair(
			ctx,
			ec2Service,
			input,
			activation,
			flexVPCResourceState,
			output,
		)
	}

	return output, nil
}

func reconcileServiceEndpoint(
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

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		// The endpoint may already be gone while its security group is not: deleting the
		// endpoint returns before AWS releases its network interfaces, so the group
		// delete can yield and be retried once the endpoint no longer exists. Skipping
		// the group here left it behind, and its ingress rule then blocked the caller's
		// group and the whole VPC.
		//
		// Falling through is not an option either: the create path below would provision
		// an endpoint during a teardown.
		if len(vpcEndpointsOutput.VpcEndpoints) == 0 {
			return removeLinkEndpointSecurityGroup(ctx, ec2Service, info, input, output)
		}
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

	err = pairSecurityGroups(
		ctx,
		ec2Service,
		info.callerSecurityGroup,
		aws.ToString(securityGroupOutput.GroupId),
		vpcEndpointPort,
		input.LinkID,
	)
	if err != nil {
		return nil, wrapRuleBudgetError(err, input)
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
				utils.CreateTagFilterBluelinkService(info.serviceName),
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
	err := pairSecurityGroups(
		ctx,
		ec2Service,
		info.callerSecurityGroup,
		linkSecurityGroup.securityGroupID,
		vpcEndpointPort,
		input.LinkID,
	)
	if err != nil {
		err = wrapRuleBudgetError(err, input)
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

	return removeLinkEndpointSecurityGroup(ctx, ec2Service, info, input, output)
}

// Removes the security group this link created for a service endpoint, or drops just
// this link's tag when other links still reference it.
//
// Separate from the endpoint removal because the two do not always happen together: the
// endpoint can be gone while its group is not, and a teardown that only cleaned up when
// it still found an endpoint left the group behind for good.
func removeLinkEndpointSecurityGroup(
	ctx context.Context,
	ec2Service ec2service.Service,
	info *vpcEndpointInfo,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	securityGroupsOutput, err := ec2Service.DescribeSecurityGroups(
		ctx,
		&ec2.DescribeSecurityGroupsInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{info.vpcID}},
				utils.CreateTagFilterBluelinkService(info.serviceName),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if securityGroupsOutput == nil {
		return output, nil
	}

	// The caller's egress rule points at the endpoint's group, and AWS will not delete a
	// group that another group's rules reference. Revoking it here is what makes the
	// delete below possible.
	//
	// This used to be left to the placement link, which deletes the caller's whole
	// group and took the rule with it. That only worked while nothing ordered the two:
	// once access links are torn down before the placement link that they require a
	// network attachment from, the caller's group is still standing, the delete fails
	// with DependencyViolation, and because that error is treated as transient it is
	// retried until the deployment times out.
	//
	// Only rules carrying this link's ID are revoked, so a group shared with other
	// links keeps theirs. This is what the gateway endpoint and security group pair
	// teardowns have always done.
	if info.callerSecurityGroup != "" {
		if err := revokeLinkRules(ctx, ec2Service, info.callerSecurityGroup, input.LinkID); err != nil {
			return nil, err
		}
	}

	for i := range securityGroupsOutput.SecurityGroups {
		securityGroup := securityGroupsOutput.SecurityGroups[i]
		if !utils.HasSecurityGroupTagForLink(&securityGroup, input.LinkID) {
			continue
		}
		// Remove the security group when the current link is the last one holding a
		// reference, otherwise just drop this link's tag.
		otherLinkTags := utils.GetOtherLinkTagsFromSecurityGroup(&securityGroup, input.LinkID)
		if len(otherLinkTags) == 0 {
			if err := revokeLinkRules(
				ctx,
				ec2Service,
				aws.ToString(securityGroup.GroupId),
				input.LinkID,
			); err != nil {
				return nil, err
			}

			// Deleting the endpoint above returns before AWS has released the interfaces
			// it created, so deleting its group in the same pass races them and fails
			// with DependencyViolation.
			err := ec2util.DeleteSecurityGroupWhenUnused(
				ctx,
				ec2Service,
				aws.ToString(securityGroup.GroupId),
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
	callerSecurityGroupID string,
) (*linkSecurityGroupInfo, error) {
	if len(securityGroupsOutput.SecurityGroups) == 0 {
		return nil, errors.New("no security groups found for the service VPC endpoint")
	}

	linkTag := utils.CreateTagBlueprintLinkID(linkID)
	for i := range securityGroupsOutput.SecurityGroups {
		securityGroup := securityGroupsOutput.SecurityGroups[i]
		for _, tag := range securityGroup.Tags {
			if utils.MatchesEC2Tag(tag, linkTag) &&
				utils.HasIngressFromSecurityGroupID(&securityGroup, callerSecurityGroupID) {
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

// Interface VPC endpoints terminate HTTPS, so this is the only port a caller needs to
// reach one. Opening all ports would grant reach the endpoint cannot serve anyway.
const vpcEndpointPort int32 = 443

// Recovers the flex VPC's Bluelink name from the AWS VPC the caller is attached to.
//
// A caller's vpcConfig carries the AWS VPC ID, but the flex VPC resource is identified
// in blueprint state by its name (its IDField), so the two cannot be matched without
// this step. The name is held on the VPC as a tag by the flex VPC resource itself.
func flexVPCNameForVPC(
	ctx context.Context,
	ec2Service ec2service.Service,
	vpcID string,
) (string, error) {
	output, err := ec2Service.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcID},
	})
	if err != nil {
		return "", err
	}

	if output == nil || len(output.Vpcs) == 0 {
		return "", fmt.Errorf("VPC %s could not be found", vpcID)
	}

	for _, tag := range output.Vpcs[0].Tags {
		if aws.ToString(tag.Key) == flex.TagFlexVPCName {
			return aws.ToString(tag.Value), nil
		}
	}

	return "", fmt.Errorf(
		"VPC %s carries no %s tag, so it is not a Bluelink-managed flex VPC",
		vpcID,
		flex.TagFlexVPCName,
	)
}

// Removes the VPC endpoint this link recorded when it was deployed, for the case where
// the caller has already been detached and the live configuration can no longer say
// which VPC the link was operating in.
//
// The endpoint's own description supplies the VPC and service name that the normal
// teardown path reads from the caller, so the ref-counted removal is unchanged.
func removeRecordedServiceEndpoint(
	ctx context.Context,
	ec2Service ec2service.Service,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	endpointID := recordedVPCEndpointID(input)
	if endpointID == "" {
		return output, nil
	}

	endpointsOutput, err := ec2Service.DescribeVpcEndpoints(
		ctx,
		&ec2.DescribeVpcEndpointsInput{VpcEndpointIds: []string{endpointID}},
	)
	if err != nil {
		// Already gone is the expected outcome of a retried teardown.
		if isVPCEndpointNotFoundError(err) {
			return output, nil
		}
		return nil, err
	}
	if endpointsOutput == nil || len(endpointsOutput.VpcEndpoints) == 0 {
		return output, nil
	}

	endpoint := endpointsOutput.VpcEndpoints[0]
	info := &vpcEndpointInfo{
		vpcID:       aws.ToString(endpoint.VpcId),
		serviceName: aws.ToString(endpoint.ServiceName),
	}

	return removeServiceEndpoint(ctx, ec2Service, info, &endpoint, input, output)
}

// The endpoint ID this link recorded under "<callerName>VPCEndpoint" when it deployed.
func recordedVPCEndpointID(input *provider.LinkUpdateIntermediaryResourcesInput) string {
	if input.CurrentLinkState == nil {
		return ""
	}

	suffix := "VPCEndpoint"
	for name, data := range input.CurrentLinkState.Data {
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		if id, ok := pluginutils.GetValueByPath("$.id", data); ok {
			if endpointID := core.StringValue(id); endpointID != "" {
				return endpointID
			}
		}
	}

	return ""
}

func isVPCEndpointNotFoundError(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode() == "InvalidVpcEndpointId.NotFound"
	}
	return false
}
