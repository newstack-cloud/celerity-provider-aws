package lambdassm

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionParameterPathLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, _, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getParameterLinkAnnotations(input.ResourceInfo, input.OtherResourceInfo)
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
		nil,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	finalEnvVarName := parameterPathEnvVarName(annotations.envVarName, input.OtherResourceInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeParameterEnvVars(
			ctx, input, setupCtx.FunctionARN, finalEnvVarName,
			setupCtx.LambdaOutput, lambdaService,
		)
	}

	// The environment variable holds the path prefix, since the runtime enumerates the
	// namespace via ssm:GetParametersByPath(path).
	path, hasPath := extractParameterPath(input.OtherResourceInfo)
	if !hasPath {
		return nil, fmt.Errorf(
			"parameter path could not be retrieved from the linked to %q SSM parameter path resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	return l.addParameterEnvVars(
		ctx, input, setupCtx.FunctionARN, path, finalEnvVarName,
		setupCtx.LambdaOutput, lambdaService,
	)
}

func (l *functionParameterPathLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The parameter path is not modified by this link; only the Lambda function and its
	// execution role are updated to allow it to access the namespace.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

// UpdateIntermediaryResources grants (or revokes) the Lambda execution role access to every
// SSM parameter beneath the path prefix, then activates the SSM interface VPC endpoint when
// the function is VPC-isolated.
func (l *functionParameterPathLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, region, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
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

	// The execution role is shared by every link that grants it access, so lock it for the
	// read-modify-write of its policy set.
	err = input.ResourceService.AcquireResourceLock(ctx, &provider.AcquireResourceLockInput{
		InstanceID:      pluginutils.GetInstanceID(input.ResourceAInfo),
		ResourceName:    setupCtx.RoleResourceName,
		ProviderContext: providerCtx,
		AcquiredBy:      input.LinkID,
	})
	if err != nil {
		return nil, err
	}

	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	sid := createParameterPathAccessSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if _, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setupCtx.RoleName,
			SID:      sid,
		}); err != nil {
			return nil, err
		}
		// The endpoint this link provisioned is removed here; returning early would
		// leave it, and its security group, behind.
		ec2Service, err := l.getEC2Service(ctx, providerCtx)
		if err != nil {
			return nil, err
		}
		return linkutils.ReconcileLinkNetworking(
			ctx,
			ec2Service,
			input,
			ssmParameterPathNetworkingActivation(setupCtx, region),
			&provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
		)
	}

	path, hasPath := extractParameterPath(input.ResourceBInfo)
	if !hasPath {
		return nil, fmt.Errorf(
			"parameter path could not be retrieved from the linked to %q SSM parameter path resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	pathARN, err := parameterPathARN(setupCtx.FunctionARN, region, path)
	if err != nil {
		return nil, err
	}

	annotations := getParameterLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
	statementResources := parameterPathStatementResources(pathARN)
	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: parameterPathAccessStatement(sid, statementResources, annotations.accessLevel),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q access to SSM parameter path %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	statementNode := specAccessStatementNode(
		sid,
		parameterActionsForAccessLevel(annotations.accessLevel),
		statementResources,
	)
	output := accessLinkOutput(input, setupCtx.RoleResourceName, sid, statementNode, result)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-isolated caller reaches SSM through an interface VPC endpoint; this is a no-op
	// for non-VPC functions.
	return linkutils.ReconcileLinkNetworking(
		ctx,
		ec2Service,
		input,
		ssmParameterPathNetworkingActivation(setupCtx, region),
		output,
	)
}

// The path ARN authorises ssm:GetParametersByPath over the hierarchy; the wildcard ARN
// authorises actions on the individual parameters beneath it (e.g. ssm:GetParameter).
func parameterPathStatementResources(pathARN string) []string {
	return []string{pathARN, pathARN + "/*"}
}

func parameterPathAccessStatement(sid string, resources []string, accessLevel string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   parameterActionsForAccessLevel(accessLevel),
		"Resource": resources,
	}
}

// Builds the ARN of the parameter hierarchy. The synthetic parameter path
// resource has no ARN of its own, so the partition and account ID are derived from the
// function's ARN (the parameters live in the same account as the function reading them).
func parameterPathARN(functionARN, region, path string) (string, error) {
	parts := strings.Split(functionARN, ":")
	if len(parts) < 5 || parts[1] == "" || parts[4] == "" {
		return "", fmt.Errorf(
			"could not derive the partition and account ID from function ARN %q",
			functionARN,
		)
	}
	return fmt.Sprintf("arn:%s:ssm:%s:%s:parameter%s", parts[1], region, parts[4], path), nil
}

func extractParameterPath(parameterPathInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(parameterPathInfo)
	pathNode, has := pluginutils.GetValueByPath("$.path", spec)
	if !has {
		return "", false
	}
	return core.StringValue(pathNode), true
}

var envVarUnsafeChars = regexp.MustCompile(`[^A-Za-z0-9_]`)

// Lambda environment variable keys must match [a-zA-Z][a-zA-Z0-9_]*, while blueprint
// resource names may contain characters like "-", so the default name sanitises the
// resource name rather than embedding it verbatim.
func parameterPathEnvVarName(userDefinedEnvVarName string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	safeName := envVarUnsafeChars.ReplaceAllString(resourceInfo.ResourceName, "_")
	return fmt.Sprintf("SSM_PARAMETER_PATH_%s", safeName)
}

func createParameterPathAccessSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("SSMPathAccess%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

// Shared by the create and destroy paths so a teardown removes exactly what the create
// path provisioned.
//
// Destroy used to return before reaching the activation, so the VPC endpoint and its
// security group were left behind on every teardown. That group's ingress rule
// references the caller's group, which then blocks the caller's group, and with it the
// whole VPC, from being deleted.
func ssmParameterPathNetworkingActivation(
	setupCtx *linkutils.LambdaLinkSetupContext,
	region string,
) linkutils.NetworkingActivation {
	return linkutils.NetworkingActivation{
		Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
		Region:       region,
		AWSService:   "ssm",
		EndpointType: ec2types.VpcEndpointTypeInterface,
	}
}
