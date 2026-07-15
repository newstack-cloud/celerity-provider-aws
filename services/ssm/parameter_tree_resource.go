package ssm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// ParameterTreeResource returns the resource implementation for an AWS Systems Manager
// (SSM) parameter tree: a prefix-scoped set of parameters managed as a single store with
// write-only blob semantics for the stored values. One parameter is created per entry in
// "values" (String) and "secureValues" (SecureString) at "<path>/<key>".
//
// Values are applied on create and on explicit blueprint change, but are never read back
// or reported as drift, so out-of-band writes (for example runtime configuration updates
// via a CLI or the console) survive redeploys. This mirrors how Secrets Manager's
// secretString is handled and deliberately diverges from aws/ssm/parameter, which keeps
// its value drift-tracked for hand-authored declarative parameters.
func ParameterTreeResource(
	ssmServiceFactory pluginutils.ServiceFactory[*aws.Config, ssmservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	actions := &parameterTreeResourceActions{
		ssmServiceFactory: ssmServiceFactory,
		awsConfigStore:    awsConfigStore,
	}

	basicExample, _ := examples.ReadFile("examples/resources/parameter_tree_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/parameter_tree_complete.md")

	return &providerv1.ResourceDefinition{
		Type:  "aws/ssm/parameterTree",
		Label: "AWS SSM Parameter Tree",
		PlainTextSummary: "A resource managing a prefix-scoped tree of AWS Systems Manager parameters " +
			"as a single store, with stored values treated as an opaque blob for drift purposes.",
		FormattedDescription: "The resource type for an AWS Systems Manager (SSM) parameter tree. It " +
			"manages every parameter beneath a path prefix (one per entry in `values`/`secureValues`) " +
			"as a single store. Stored values are treated like a Secrets Manager secret blob: they are " +
			"applied on create and on explicit blueprint change, but are never read back, reported as " +
			"drift, or reverted, so out-of-band configuration writes survive redeploys. Use " +
			"`aws/ssm/parameter` instead for individual parameters whose values should be drift-tracked.",
		Schema:         parameterTreeResourceSchema(),
		IDField:        "path",
		TaggingSupport: provider.TaggingSupportFull,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		// A parameter tree is a leaf configuration store that other resources link to; it
		// does not link out to other resources.
		CommonTerminal:       true,
		GetExternalStateFunc: actions.GetExternalState,
		CreateFunc:           actions.Create,
		UpdateFunc:           actions.Update,
		DestroyFunc:          actions.Destroy,
		StabilisedFunc:       actions.Stabilised,
	}
}

type parameterTreeResourceActions struct {
	ssmServiceFactory pluginutils.ServiceFactory[*aws.Config, ssmservice.Service]
	awsConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
}

func (a *parameterTreeResourceActions) getSSMServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
	meta map[string]*core.MappingNode,
) (ssmservice.Service, string, error) {
	awsConfig, err := a.awsConfigStore.FromProviderContext(ctx, providerContext, meta)
	if err != nil {
		return nil, "", err
	}

	return a.ssmServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}

type parameterTreeEntry struct {
	value  string
	secure bool
}

func parameterTreePath(specData *core.MappingNode) (string, error) {
	pathNode, hasPath := pluginutils.GetValueByPath("$.path", specData)
	if !hasPath {
		return "", errors.New("path is required for an SSM parameter tree")
	}
	return core.StringValue(pathNode), nil
}

// Merges the "values" and "secureValues" maps from the given spec
// (desired or prior state) into a single key -> entry map.
func parameterTreeEntries(specData *core.MappingNode) map[string]parameterTreeEntry {
	entries := map[string]parameterTreeEntry{}
	if valuesNode, ok := pluginutils.GetValueByPath("$.values", specData); ok && valuesNode != nil {
		for key, value := range valuesNode.Fields {
			entries[key] = parameterTreeEntry{
				value: core.StringValue(value),
			}
		}
	}
	if secureNode, ok := pluginutils.GetValueByPath("$.secureValues", specData); ok && secureNode != nil {
		for key, value := range secureNode.Fields {
			entries[key] = parameterTreeEntry{
				value:  core.StringValue(value),
				secure: true,
			}
		}
	}
	return entries
}

func sortedParameterTreeKeys[Value any](entries map[string]Value) []string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func parameterTreeFullName(path string, key string) string {
	return path + "/" + key
}

func parameterTreeValueHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func parameterTreePutInput(
	path string,
	key string,
	entry parameterTreeEntry,
	specData *core.MappingNode,
) *ssm.PutParameterInput {
	putInput := &ssm.PutParameterInput{
		Name:      aws.String(parameterTreeFullName(path, key)),
		Value:     aws.String(entry.value),
		Type:      ssmtypes.ParameterTypeString,
		Overwrite: aws.Bool(false),
	}

	if entry.secure {
		putInput.Type = ssmtypes.ParameterTypeSecureString
		if keyID, ok := pluginutils.GetValueByPath("$.keyId", specData); ok {
			putInput.KeyId = aws.String(core.StringValue(keyID))
		}
	}
	if tier, ok := pluginutils.GetValueByPath("$.tier", specData); ok {
		putInput.Tier = ssmtypes.ParameterTier(core.StringValue(tier))
	}
	if description, ok := pluginutils.GetValueByPath("$.description", specData); ok {
		putInput.Description = aws.String(core.StringValue(description))
	}

	return putInput
}

// Returns structural metadata (no values) for the tree's
// managed parameters, keyed by entry key. It pages through DescribeParameters with a
// recursive path filter and drops any parameter beneath the prefix that is not in
// managedKeys, so foreign parameters sharing the prefix are never surfaced.
func describeManagedTreeParameters[Value any](
	ctx context.Context,
	service ssmservice.Service,
	path string,
	managedEntries map[string]Value,
) (map[string]ssmtypes.ParameterMetadata, error) {
	metadataByKey := map[string]ssmtypes.ParameterMetadata{}
	prefix := path + "/"

	var nextToken *string
	for {
		output, err := service.DescribeParameters(ctx, &ssm.DescribeParametersInput{
			ParameterFilters: []ssmtypes.ParameterStringFilter{
				{
					Key:    aws.String("Path"),
					Option: aws.String("Recursive"),
					Values: []string{path},
				},
			},
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, metadata := range output.Parameters {
			key := strings.TrimPrefix(aws.ToString(metadata.Name), prefix)
			if _, managed := managedEntries[key]; managed {
				metadataByKey[key] = metadata
			}
		}

		if output.NextToken == nil || aws.ToString(output.NextToken) == "" {
			return metadataByKey, nil
		}
		nextToken = output.NextToken
	}
}

func parameterTreeComputedParameters(
	metadataByKey map[string]ssmtypes.ParameterMetadata,
	desiredEntries map[string]parameterTreeEntry,
) *core.MappingNode {
	parameters := &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	for key, entry := range desiredEntries {
		fields := map[string]*core.MappingNode{
			"valueHash": core.MappingNodeFromString(parameterTreeValueHash(entry.value)),
		}
		if metadata, ok := metadataByKey[key]; ok {
			fields["arn"] = core.MappingNodeFromString(aws.ToString(metadata.ARN))
			fields["type"] = core.MappingNodeFromString(string(metadata.Type))
		}
		parameters.Fields[key] = &core.MappingNode{Fields: fields}
	}
	return parameters
}

func parameterTreeStoredHashes(stateSpecData *core.MappingNode) map[string]string {
	hashes := map[string]string{}
	parametersNode, ok := pluginutils.GetValueByPath("$.parameters", stateSpecData)
	if !ok || parametersNode == nil {
		return hashes
	}
	for key, entryNode := range parametersNode.Fields {
		hashNode, hasHash := pluginutils.GetValueByPath("$.valueHash", entryNode)
		if hasHash {
			hashes[key] = core.StringValue(hashNode)
		}
	}
	return hashes
}
