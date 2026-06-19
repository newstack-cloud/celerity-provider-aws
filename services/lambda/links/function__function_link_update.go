package lambdalinks

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	lambdaService, _, err := l.getLambdaServiceWithRegion(
		ctx,
		provider.NewProviderContextFromLinkContext(
			input.LinkContext,
			"aws",
		),
	)
	if err != nil {
		return nil, err
	}

	annotations := getLambdaFunctionLinkAnnotations(
		input.ResourceInfo,
		input.OtherResourceInfo,
	)
	if !annotations.populateEnvVars {
		return &provider.LinkUpdateResourceOutput{
			LinkData:             core.MappingNodeFields(),
			ResourceDataMappings: map[string]string{},
		}, nil
	}

	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceInfo,
			LoadRoleInfo:           false,
		},
		lambdaService,
		// Resource service is only needed if we need to load the role info.
		/* resourceService */
		nil,
		provider.NewProviderContextFromLinkContext(
			input.LinkContext,
			"aws",
		),
	)
	if err != nil {
		return nil, err
	}

	otherFunctionARN, hasOtherFunctionARN := utils.ExtractARNFromResourceInfo(
		input.OtherResourceInfo,
	)
	if !hasOtherFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked to %q function resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeCallerFunctionEnvVars(
			ctx,
			input,
			setupCtx.FunctionARN,
			annotations.envVarName,
			setupCtx.LambdaOutput,
			lambdaService,
		)
	}

	return l.addCallerFunctionEnvVars(
		ctx,
		input,
		setupCtx.FunctionARN,
		otherFunctionARN,
		annotations.envVarName,
		setupCtx.LambdaOutput,
		lambdaService,
	)
}

func (l *lambdaFunctionFunctionLinkActions) addCallerFunctionEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	otherFunctionARN string,
	envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := invokeLambdaFunctionEnvVarName(
		envVarName,
		input.OtherResourceInfo,
	)
	dataMappingKey := fmt.Sprintf(
		"%s::spec.environment.variables[\"%s\"]",
		input.ResourceInfo.ResourceName,
		finalEnvVarName,
	)
	linkDataFieldPath := fmt.Sprintf(
		"%s.environmentVariables[\"%s\"]",
		input.ResourceInfo.ResourceName,
		finalEnvVarName,
	)

	err := linkutils.UpdateLambdaEnvironmentVariables(
		ctx,
		lambdaService,
		functionARN,
		currentFunctionConfig,
		map[string]string{
			finalEnvVarName: otherFunctionARN,
		},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			pluginutils.GetResourceName(input.ResourceInfo),
			core.MappingNodeFields(
				"environmentVariables",
				core.MappingNodeFields(
					envVarName,
					core.MappingNodeFromString(otherFunctionARN),
				),
			),
		),
		ResourceDataMappings: map[string]string{
			dataMappingKey: linkDataFieldPath,
		},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) removeCallerFunctionEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	envVarName string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	finalEnvVarName := invokeLambdaFunctionEnvVarName(
		envVarName,
		input.OtherResourceInfo,
	)

	err := linkutils.RemoveLambdaEnvironmentVariables(
		ctx,
		lambdaService,
		functionARN,
		currentFunctionConfig,
		[]string{finalEnvVarName},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			pluginutils.GetResourceName(input.ResourceInfo),
			core.MappingNodeFields(),
		),
		ResourceDataMappings: map[string]string{},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The target function is not updated as a part of the link update,
	// the link only requires updates to the linked from function
	// and its execution role that will allow it to invoke the target function.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(
		input.LinkContext,
		"aws",
	)
	lambdaService, region, err := l.getLambdaServiceWithRegion(
		ctx,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceAInfo,
			LoadRoleInfo:           true,
		},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	// Acquire a lock on the role resource to prevent concurrent updates
	// to the same role from multiple links.
	err = input.ResourceService.AcquireResourceLock(
		ctx,
		&provider.AcquireResourceLockInput{
			InstanceID:      pluginutils.GetInstanceID(input.ResourceAInfo),
			ResourceName:    setupCtx.RoleResourceName,
			ProviderContext: providerCtx,
		},
	)
	if err != nil {
		return nil, err
	}

	iamService, err := l.getIamService(
		ctx,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	rolePolicies, err := iamService.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: aws.String(setupCtx.RoleName),
	})
	if err != nil {
		return nil, err
	}

	otherFunctionARN, hasOtherFunctionARN := utils.ExtractARNFromResourceInfo(
		input.ResourceBInfo,
	)
	if !hasOtherFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked to %q function resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	sid := createLambdaFunctionInvokeSID(input.ResourceBInfo)
	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeRolePolicyPermissions(
			ctx,
			input,
			setupCtx,
			rolePolicies,
			iamService,
		)
	}

	output, err := l.setRolePolicyPermissions(
		ctx,
		input,
		otherFunctionARN,
		sid,
		setupCtx,
		rolePolicies,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	return l.updateNetworkingElements(
		ctx,
		setupCtx,
		input,
		output,
		region,
	)
}

func (l *lambdaFunctionFunctionLinkActions) setRolePolicyPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	otherFunctionARN string,
	sid string,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	invokePermission := map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   "lambda:InvokeFunction",
		"Resource": otherFunctionARN,
	}

	if len(rolePolicies.PolicyNames) > 0 {
		return l.updateExistingRolePolicy(
			ctx,
			input,
			&roleInfo{
				roleName:          setupCtx.RoleName,
				roleResourceState: setupCtx.RoleResourceState,
				rolePolicies:      rolePolicies,
				invokePermission:  invokePermission,
				sid:               sid,
			},
			iamService,
		)
	}

	return l.addNewRolePolicy(
		ctx,
		input,
		&roleInfo{
			roleName:          setupCtx.RoleName,
			roleResourceState: setupCtx.RoleResourceState,
			rolePolicies:      rolePolicies,
			invokePermission:  invokePermission,
			sid:               sid,
		},
		iamService,
	)
}

func (l *lambdaFunctionFunctionLinkActions) addNewRolePolicy(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleData *roleInfo,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	policyName, err := linkutils.CreateNewRolePolicy(
		ctx,
		iamService,
		input.InstanceName,
		roleData.roleName,
		[]map[string]any{roleData.invokePermission},
	)
	if err != nil {
		return nil, err
	}

	roleResourcePath := fmt.Sprintf(
		"%s::spec.policies[@.name=\"%q\"].Statement[@.Sid=\"%q\"]",
		roleData.roleResourceState.Name,
		policyName,
		roleData.sid,
	)
	linkDataExecRoleName := createLinkDataExecutionRoleName(input.ResourceAInfo)
	linkDataFieldPath := linkutils.PermissionFieldPath(linkDataExecRoleName)

	invokePermissionNode, err := pluginutils.AnyToMappingNode(roleData.invokePermission)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: core.MappingNodeFields(
			linkDataExecRoleName,
			core.MappingNodeFields(
				linkutils.PermissionFieldName,
				invokePermissionNode,
				linkutils.PolicyNameFieldName,
				core.MappingNodeFromString(policyName),
			),
		),
		ResourceDataMappings: map[string]string{
			roleResourcePath: linkDataFieldPath,
		},
	}, nil
}

type roleInfo struct {
	roleName          string
	roleResourceState *state.ResourceState
	rolePolicies      *iam.ListRolePoliciesOutput
	invokePermission  map[string]any
	sid               string
}

func (l *lambdaFunctionFunctionLinkActions) updateExistingRolePolicy(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleData *roleInfo,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	executionRoleName := fmt.Sprintf(
		"%sExecutionRole",
		input.ResourceAInfo.ResourceName,
	)
	linkDataFieldPath := linkutils.PermissionFieldPath(executionRoleName)

	policyName, existingPermSID := linkutils.ExtractPolicyNameAndCurrentPermissionSID(
		input.CurrentLinkState,
		executionRoleName,
		roleData.rolePolicies,
	)

	err := linkutils.UpdateExistingRolePolicy(
		ctx,
		iamService,
		roleData.roleName,
		policyName,
		[]map[string]any{roleData.invokePermission},
		[]string{existingPermSID},
	)
	if err != nil {
		return nil, err
	}

	roleResourcePath := fmt.Sprintf(
		"%s::spec.policies[@.name=\"%q\"].Statement[@.Sid=\"%q\"]",
		roleData.roleResourceState.Name,
		policyName,
		roleData.sid,
	)

	invokePermissionNode, err := pluginutils.AnyToMappingNode(roleData.invokePermission)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: core.MappingNodeFields(
			executionRoleName,
			core.MappingNodeFields(
				linkutils.PermissionFieldName,
				invokePermissionNode,
				linkutils.PolicyNameFieldName,
				core.MappingNodeFromString(policyName),
			),
		),
		ResourceDataMappings: map[string]string{
			roleResourcePath: linkDataFieldPath,
		},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) removeRolePolicyPermissions(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	rolePolicies *iam.ListRolePoliciesOutput,
	iamService iamservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	executionRoleName := fmt.Sprintf(
		"%sExecutionRole",
		input.ResourceAInfo.ResourceName,
	)

	policyName, existingPermSID := linkutils.ExtractPolicyNameAndCurrentPermissionSID(
		input.CurrentLinkState,
		executionRoleName,
		rolePolicies,
	)

	err := linkutils.RemoveExistingRolePolicyPermissions(
		ctx,
		iamService,
		setupCtx.RoleName,
		policyName,
		[]string{existingPermSID},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: core.MappingNodeFields(
			executionRoleName,
			core.MappingNodeFields(),
		),
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) updateNetworkingElements(
	ctx context.Context,
	setupCtx *linkutils.LambdaLinkSetupContext,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
	region string,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {

	if !hasVPCConfig(setupCtx.LambdaOutput.VpcConfig) {
		return output, nil
	}

	providerCtx := provider.NewProviderContextFromLinkContext(
		input.LinkContext,
		"aws",
	)
	instanceID := pluginutils.GetInstanceID(input.ResourceAInfo)

	// In order to update networking elements, the function must be connected to a flex VPC
	// (often in reference mode) in the same blueprint.
	flexVPCResourceState, err := input.ResourceService.LookupResourceInState(
		ctx,
		&provider.ResourceLookupInput{
			InstanceID:      instanceID,
			ResourceType:    "aws/flex/vpc",
			ExternalID:      aws.ToString(setupCtx.LambdaOutput.VpcConfig.VpcId),
			ProviderContext: providerCtx,
		},
	)
	if err != nil {
		return nil, err
	}

	// Lock the flex VPC resource to prevent concurrent updates to the same flex VPC
	// from multiple links in the same blueprint.
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

	ec2Service, err := l.getEC2Service(
		ctx,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	flexVPCName, hasFlexVPCName := pluginutils.GetValueByPath(
		"$.name",
		flexVPCResourceState.SpecData,
	)
	if !hasFlexVPCName {
		return nil, fmt.Errorf(
			"flex VPC name could not be retrieved from the flex VPC resource",
		)
	}

	vpcEndpointsOutput, err := ec2Service.DescribeVpcEndpoints(
		ctx,
		&ec2.DescribeVpcEndpointsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{aws.ToString(setupCtx.LambdaOutput.VpcConfig.VpcId)},
				},
				{
					Name: aws.String("service-name"),
					Values: []string{
						lambdaServiceName(region),
					},
				},
				utils.CreateTagFilterFlexVPCNameForLink(
					core.StringValue(flexVPCName),
				),
			},
		},
	)
	if err != nil {
		return nil, err
	}

	endpointInfo := &vpcEndpointInfo{
		flexVPCName:                 core.StringValue(flexVPCName),
		vpcID:                       aws.ToString(setupCtx.LambdaOutput.VpcConfig.VpcId),
		privateDNS:                  flexVPCHasDNSSupportWithHostnames(flexVPCResourceState),
		lambdaFunctionSubnetIDs:     setupCtx.LambdaOutput.VpcConfig.SubnetIds,
		lambdaFunctionSecurityGroup: setupCtx.LambdaOutput.VpcConfig.SecurityGroupIds[0],
		region:                      region,
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy &&
		len(vpcEndpointsOutput.VpcEndpoints) > 0 {
		return l.removeNetworkingElements(
			ctx,
			setupCtx,
			&vpcEndpointsOutput.VpcEndpoints[0],
			input,
			output,
			region,
			ec2Service,
		)
	}

	if len(vpcEndpointsOutput.VpcEndpoints) == 0 {
		return l.createVPCEndpoint(
			ctx,
			endpointInfo,
			input,
			output,
			ec2Service,
		)
	}

	return l.updateVPCEndpoint(
		ctx,
		endpointInfo,
		&vpcEndpointsOutput.VpcEndpoints[0],
		input,
		output,
		ec2Service,
	)
}

func (l *lambdaFunctionFunctionLinkActions) updateVPCEndpoint(
	ctx context.Context,
	info *vpcEndpointInfo,
	endpoint *ec2types.VpcEndpoint,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
	ec2Service ec2service.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// Make sure the VPC endpoint has a tag that associates it with the current link
	// as a part of a reference counting mechanism to ensure the VPC endpoint is only removed
	// when the last link that references it is destroyed.
	if !utils.HasVPCEndpointTagForLink(endpoint, input.LinkID) {
		_, err := ec2Service.CreateTags(
			ctx,
			&ec2.CreateTagsInput{
				Resources: []string{aws.ToString(endpoint.VpcEndpointId)},
				Tags: []ec2types.Tag{
					utils.CreateTagBlueprintLinkID(input.LinkID),
				},
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Make sure the VPC endpoint provides a network interface in the caller lambda function subnet.
	if !utils.VPCEndpointInSubnets(endpoint, info.lambdaFunctionSubnetIDs) {
		_, err := ec2Service.ModifyVpcEndpoint(
			ctx,
			&ec2.ModifyVpcEndpointInput{
				VpcEndpointId: endpoint.VpcEndpointId,
				AddSubnetIds:  info.lambdaFunctionSubnetIDs,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Make sure that the link is attached to a security group that allows access to the
	// VPC endpoint.
	securityGroupsOutput, err := ec2Service.DescribeSecurityGroups(
		ctx,
		&ec2.DescribeSecurityGroupsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{info.vpcID},
				},
				{
					Name:   aws.String(utils.TagBluelinkService),
					Values: []string{lambdaServiceName(info.region)},
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	linkSecurityGroupInfo, err := isLinkAttachedToSecurityGroup(
		securityGroupsOutput,
		input.LinkID,
		info.lambdaFunctionSecurityGroup,
	)
	if err != nil {
		return nil, err
	}

	if !linkSecurityGroupInfo.attached {
		return l.attackLinkToSecurityGroup(
			ctx,
			info,
			linkSecurityGroupInfo,
			input,
			output,
			ec2Service,
		)
	}

	return output, nil
}

func (l *lambdaFunctionFunctionLinkActions) attackLinkToSecurityGroup(
	ctx context.Context,
	info *vpcEndpointInfo,
	linkSecurityGroupInfo *linkSecurityGroupInfo,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
	ec2Service ec2service.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	_, err := ec2Service.AuthorizeSecurityGroupIngress(
		ctx,
		&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:                 aws.String(linkSecurityGroupInfo.securityGroupID),
			SourceSecurityGroupName: aws.String(info.lambdaFunctionSecurityGroup),
		},
	)
	if err != nil {
		return nil, err
	}

	_, err = ec2Service.CreateTags(
		ctx,
		&ec2.CreateTagsInput{
			Resources: []string{linkSecurityGroupInfo.securityGroupID},
			Tags: []ec2types.Tag{
				utils.CreateTagBlueprintLinkID(input.LinkID),
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (l *lambdaFunctionFunctionLinkActions) removeNetworkingElements(
	ctx context.Context,
	setupCtx *linkutils.LambdaLinkSetupContext,
	endpoint *ec2types.VpcEndpoint,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
	region string,
	ec2Service ec2service.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// If VPC endpoint exists and is only associated with the current link, then remove it.
	// If VPC endpoint exists and is associated with multiple links, then remove the tag
	// that associates it with the current link.

	// If VPC endpoint does not exist, then do nothing.

	// If security group exists for the current link, and is not associated with any other links,
	// then remove the security group.

	// If security group exists for the current link, and is associated with other links,
	// then remove the tag that associates it with the current link.
	// A trade-off here is rule accumulation for lots of security groups associated with
	// a single VPC endpoint.

	// If security group does not exist, then do nothing.

	// Handle VPC endpoint cleanup
	if utils.HasVPCEndpointTagForLink(endpoint, input.LinkID) {
		// Check if this is the only link using this VPC endpoint
		otherLinkTags := utils.GetOtherLinkTagsFromVPCEndpoint(endpoint, input.LinkID)
		if len(otherLinkTags) == 0 {
			// No other links use this VPC endpoint, so we can delete it
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
			// Other links use this VPC endpoint, so just remove the current link's tag.
			_, err := ec2Service.DeleteTags(
				ctx,
				&ec2.DeleteTagsInput{
					Resources: []string{aws.ToString(endpoint.VpcEndpointId)},
					Tags: []ec2types.Tag{
						utils.CreateTagBlueprintLinkID(input.LinkID),
					},
				},
			)
			if err != nil {
				return nil, err
			}
		}
	}

	// Find security groups associated with this link
	securityGroupsOutput, err := ec2Service.DescribeSecurityGroups(
		ctx,
		&ec2.DescribeSecurityGroupsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{aws.ToString(setupCtx.LambdaOutput.VpcConfig.VpcId)},
				},
				{
					Name:   aws.String(utils.TagBluelinkService),
					Values: []string{lambdaServiceName(region)},
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	// Handle security group cleanup
	for _, securityGroup := range securityGroupsOutput.SecurityGroups {
		if utils.HasSecurityGroupTagForLink(&securityGroup, input.LinkID) {
			// Check if this is the only link using this security group
			otherLinkTags := utils.GetOtherLinkTagsFromSecurityGroup(&securityGroup, input.LinkID)
			if len(otherLinkTags) == 0 {
				// No other links use this security group, so we can delete it
				_, err := ec2Service.DeleteSecurityGroup(
					ctx,
					&ec2.DeleteSecurityGroupInput{
						GroupId: securityGroup.GroupId,
					},
				)
				if err != nil {
					return nil, err
				}
			} else {
				// Other links use this security group, so just remove our tag
				// Note: We don't remove the ingress rule because the same rule might be shared between links,
				// as multiple lambda functions will often share the same security group.
				_, err := ec2Service.DeleteTags(
					ctx,
					&ec2.DeleteTagsInput{
						Resources: []string{aws.ToString(securityGroup.GroupId)},
						Tags: []ec2types.Tag{
							utils.CreateTagBlueprintLinkID(input.LinkID),
						},
					},
				)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return output, nil
}

type vpcEndpointInfo struct {
	flexVPCName                 string
	vpcID                       string
	privateDNS                  bool
	region                      string
	lambdaFunctionSubnetIDs     []string
	lambdaFunctionSecurityGroup string
}

func (l *lambdaFunctionFunctionLinkActions) createVPCEndpoint(
	ctx context.Context,
	info *vpcEndpointInfo,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	output *provider.LinkUpdateIntermediaryResourcesOutput,
	ec2Service ec2service.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	securityGroupOutput, err := ec2Service.CreateSecurityGroup(
		ctx,
		&ec2.CreateSecurityGroupInput{
			GroupName:   aws.String("lambda-vpc-endpoint-access"),
			VpcId:       aws.String(info.vpcID),
			Description: aws.String("Security group for lambda service access"),
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeSecurityGroup,
					Tags: []ec2types.Tag{
						utils.CreateTagLinkSecurityGroup(),
						utils.CreateTagBlueprintInstanceName(input.InstanceName),
						utils.CreateTagBlueprintLinkID(input.LinkID),
						utils.CreateTagFlexVPCNameForLink(
							info.flexVPCName,
						),
						utils.CreateTagBluelinkService(lambdaServiceName(info.region)),
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
			SourceSecurityGroupName: aws.String(info.lambdaFunctionSecurityGroup),
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
			ServiceName:       aws.String(lambdaServiceName(info.region)),
			SubnetIds:         info.lambdaFunctionSubnetIDs,
			SecurityGroupIds:  []string{aws.ToString(securityGroupOutput.GroupId)},
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeVpcEndpoint,
					Tags: []ec2types.Tag{
						utils.CreateTagLinkVPCEndpoint(),
						utils.CreateTagBlueprintInstanceName(input.InstanceName),
						utils.CreateTagBlueprintLinkID(input.LinkID),
						utils.CreateTagFlexVPCNameForLink(
							info.flexVPCName,
						),
					},
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	vpcEndpointName := fmt.Sprintf(
		"%sVPCEndpoint",
		input.ResourceAInfo.ResourceName,
	)
	vpcEndpointLinkData := core.MappingNodeFields(
		vpcEndpointName,
		core.MappingNodeFields(
			"id",
			aws.ToString(vpcEndpointOutput.VpcEndpoint.VpcEndpointId),
		),
	)
	linkData := core.MergeMaps(
		output.LinkData,
		vpcEndpointLinkData,
	)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: output.IntermediaryResourceStates,
		LinkData:                   linkData,
		ResourceDataMappings:       output.ResourceDataMappings,
	}, nil
}

func invokeLambdaFunctionEnvVarName(
	userDefinedEnvVarName string,
	resourceInfo *provider.ResourceInfo,
) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}

	return fmt.Sprintf(
		"INVOKE_LAMBDA_FUNCTION_%s",
		resourceInfo.ResourceName,
	)
}

func createLambdaFunctionInvokeSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"InvokeLambdaFunction%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

type lambdaFunctionLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
}

func getLambdaFunctionLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *lambdaFunctionLinkAnnotations {
	// Intra-service link annotations use aws.lambda.invoke.* pattern
	// where "invoke" is the feature being configured
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key: fmt.Sprintf(
				"aws.lambda.invoke.%s.populateEnvVars",
				otherResourceInfo.ResourceName,
			),
			FallbackKey: "aws.lambda.invoke.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf(
				"aws.lambda.invoke.%s.envVarName",
				otherResourceInfo.ResourceName,
			),
		},
	)

	return &lambdaFunctionLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
	}
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sExecutionRole",
		resourceInfo.ResourceName,
	)
}

func hasVPCConfig(vpcConfig *types.VpcConfigResponse) bool {
	return vpcConfig != nil &&
		vpcConfig.VpcId != nil &&
		len(vpcConfig.SubnetIds) > 0 &&
		len(vpcConfig.SecurityGroupIds) > 0
}

func flexVPCHasDNSSupportWithHostnames(flexVPCResourceState *state.ResourceState) bool {
	enableDNSSupportValue, _ := pluginutils.GetValueByPath(
		"$.enableDNSSupport",
		flexVPCResourceState.SpecData,
	)
	enableDNSSupport := core.BoolValue(enableDNSSupportValue)

	enableDNSHostnamesValue, _ := pluginutils.GetValueByPath(
		"$.enableDNSHostnames",
		flexVPCResourceState.SpecData,
	)
	enableDNSHostnames := core.BoolValue(enableDNSHostnamesValue)

	return enableDNSSupport && enableDNSHostnames
}

func lambdaServiceName(region string) string {
	return fmt.Sprintf("com.amazonaws.%s.lambda", region)
}

type linkSecurityGroupInfo struct {
	securityGroupID string
	attached        bool
}

func isLinkAttachedToSecurityGroup(
	securityGroupsOutput *ec2.DescribeSecurityGroupsOutput,
	linkID string,
	lambdaFunctionSecurityGroup string,
) (*linkSecurityGroupInfo, error) {
	if len(securityGroupsOutput.SecurityGroups) == 0 {
		return nil, errors.New("no security groups found for lambda service endpoint")
	}

	linkTag := utils.CreateTagBlueprintLinkID(linkID)
	for _, securityGroup := range securityGroupsOutput.SecurityGroups {
		for _, tag := range securityGroup.Tags {
			if utils.MatchesEC2Tag(tag, linkTag) &&
				utils.HasIngressWithSourceSecurityGroupName(&securityGroup, lambdaFunctionSecurityGroup) {
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
