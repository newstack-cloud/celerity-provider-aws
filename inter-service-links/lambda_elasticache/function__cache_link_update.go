package lambdaelasticache

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionCacheLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, _, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getCacheLinkAnnotations(input.ResourceInfo, input.OtherResourceInfo)
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

	prefix := cacheEnvVarPrefix(annotations.envVarPrefix, input.OtherResourceInfo)
	envVarNames := cacheEnvVarNames(prefix)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeCacheEnvVars(ctx, input, setupCtx.FunctionARN, envVarNames, setupCtx.LambdaOutput, lambdaService)
	}

	endpoint, hasEndpoint := extractCacheEndpoint(input.OtherResourceInfo)
	if !hasEndpoint {
		return nil, fmt.Errorf(
			"endpoint could not be retrieved from the linked to %q ElastiCache replication group resource",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	envVars := map[string]string{
		envVarNames.host: endpoint,
		envVarNames.port: strconv.Itoa(int(annotations.port)),
	}

	return l.addCacheEnvVars(ctx, input, setupCtx.FunctionARN, envVars, setupCtx.LambdaOutput, lambdaService)
}

func (l *functionCacheLinkActions) addCacheEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	envVars map[string]string,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	if err := linkutils.UpdateLambdaEnvironmentVariables(
		ctx, lambdaService, functionARN, currentFunctionConfig, envVars,
	); err != nil {
		return nil, err
	}

	functionName := pluginutils.GetResourceName(input.ResourceInfo)
	envVarNodes := core.MappingNodeFields()
	mappings := map[string]string{}
	for _, name := range sortedKeys(envVars) {
		envVarNodes.Fields[name] = core.MappingNodeFromString(envVars[name])
		mappings[fmt.Sprintf("%s::spec.environment.variables[%q]", functionName, name)] =
			fmt.Sprintf("%s.environmentVariables[%q]", functionName, name)
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			functionName,
			core.MappingNodeFields("environmentVariables", envVarNodes),
		),
		ResourceDataMappings: mappings,
	}, nil
}

func (l *functionCacheLinkActions) removeCacheEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	envVarNames cacheEnvVarNameSet,
	currentFunctionConfig *types.FunctionConfiguration,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	if err := linkutils.RemoveLambdaEnvironmentVariables(
		ctx, lambdaService, functionARN, currentFunctionConfig, envVarNames.all(),
	); err != nil {
		return nil, err
	}
	return &provider.LinkUpdateResourceOutput{
		LinkData:             core.MappingNodeFields(pluginutils.GetResourceName(input.ResourceInfo), core.MappingNodeFields()),
		ResourceDataMappings: map[string]string{},
	}, nil
}

func (l *functionCacheLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The ElastiCache replication group is not modified by this link; only the Lambda function and
	// the security-group rules are updated so it can reach the cache.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

// UpdateIntermediaryResources opens the security-group rule so the function can reach the cache on
// the cache port. (for IAM authentication, the elasticache:Connect grant is added in a follow-up
// slice; on the password path there is no execution-role change.)
func (l *functionCacheLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, region, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getCacheLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)

	// The execution role is only needed for the IAM auth grant; skip the role lookup for password
	// auth (which grants no IAM permission).
	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceAInfo,
			LoadRoleInfo:           annotations.authMode == "iam",
		},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}

	// IAM elasticache:Connect grant (only for token-based auth).
	if annotations.authMode == "iam" {
		output, err = l.reconcileConnectGrant(ctx, providerCtx, input, setupCtx, annotations, region)
		if err != nil {
			return nil, err
		}
	}

	// Open the security-group path from the function to the cache on the cache port.
	targetSecurityGroupID, hasTargetSG := extractCacheSecurityGroupID(input.ResourceBInfo)
	if !hasTargetSG {
		return nil, fmt.Errorf(
			"security group could not be retrieved from the linked to %q ElastiCache replication group resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	return linkutils.ActivateLinkNetworking(
		ctx,
		ec2Service,
		input,
		linkutils.NetworkingActivation{
			Caller:                linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
			Region:                region,
			TargetSecurityGroupID: targetSecurityGroupID,
			TargetPort:            annotations.port,
		},
		output,
	)
}

func extractCacheEndpoint(cacheInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(cacheInfo)
	for _, path := range []string{"$.configurationEndPoint.address", "$.primaryEndPoint.address"} {
		if node, has := pluginutils.GetValueByPath(path, spec); has && node != nil && core.StringValue(node) != "" {
			return core.StringValue(node), true
		}
	}
	return "", false
}

// extractCacheSecurityGroupID reads the cache's security group id. securityGroupIds is write-only
// on the ElastiCache replication group (not returned in external state), so it is read from the
// resolved blueprint spec, where it is set as a reference to the flex VPC security group.
func extractCacheSecurityGroupID(cacheInfo *provider.ResourceInfo) (string, bool) {
	specs := []*core.MappingNode{
		resolvedSpecData(cacheInfo),
		pluginutils.GetCurrentStateSpecDataFromResourceInfo(cacheInfo),
	}
	for _, spec := range specs {
		node, has := pluginutils.GetValueByPath("$.securityGroupIds", spec)
		if has && node != nil && len(node.Items) > 0 {
			return core.StringValue(node.Items[0]), true
		}
	}
	return "", false
}

func resolvedSpecData(resourceInfo *provider.ResourceInfo) *core.MappingNode {
	if resourceInfo == nil || resourceInfo.ResourceWithResolvedSubs == nil {
		return &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	}
	return resourceInfo.ResourceWithResolvedSubs.Spec
}

type cacheEnvVarNameSet struct {
	host string
	port string
}

func (s cacheEnvVarNameSet) all() []string {
	return []string{s.host, s.port}
}

func cacheEnvVarNames(prefix string) cacheEnvVarNameSet {
	return cacheEnvVarNameSet{
		host: prefix + "_HOST",
		port: prefix + "_PORT",
	}
}

func cacheEnvVarPrefix(userDefinedPrefix string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedPrefix != "" {
		return userDefinedPrefix
	}
	return "CACHE_" + strings.ToUpper(pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type cacheLinkAnnotations struct {
	populateEnvVars bool
	envVarPrefix    string
	authMode        string
	userId          string
	port            int32
}

func getCacheLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *cacheLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.elasticache.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.elasticache.populateEnvVars",
			Default:     true,
		},
	)
	envVarPrefix, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.elasticache.%s.envVarPrefix", otherResourceInfo.ResourceName),
		},
	)
	authMode, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.elasticache.%s.authMode", otherResourceInfo.ResourceName),
			Default: "password",
		},
	)
	userId, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.elasticache.%s.userId", otherResourceInfo.ResourceName),
			Default: "*",
		},
	)
	port, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     fmt.Sprintf("aws.lambda.elasticache.%s.port", otherResourceInfo.ResourceName),
			Default: 6379,
		},
	)

	return &cacheLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarPrefix:    envVarPrefix,
		authMode:        authMode,
		userId:          userId,
		port:            int32(port),
	}
}
