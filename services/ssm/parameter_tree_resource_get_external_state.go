package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// GetExternalState reports only structural metadata for the tree: which managed
// parameters exist, their ARNs and types, and their tags. Stored values are NEVER read
// back. No GetParameter call is made and nothing is decrypted, mirroring how write-only
// fields such as Secrets Manager's secretString are withheld from external state so that
// out-of-band value writes can never surface as drift or be reverted.
func (a *parameterTreeResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	service, region, err := a.getSSMServiceWithRegion(
		ctx,
		input.ProviderContext,
		parameterRegionMeta(input.CurrentResourceSpec),
	)
	if err != nil {
		return nil, err
	}

	path, err := parameterTreePath(input.CurrentResourceSpec)
	if err != nil {
		return nil, err
	}

	managed := parameterTreeEntries(input.CurrentResourceSpec)
	metadataByKey, err := describeManagedTreeParameters(ctx, service, path, managed)
	if err != nil {
		return nil, err
	}
	if len(metadataByKey) == 0 {
		return emptyExternalState(), nil
	}

	fields := map[string]*core.MappingNode{
		"path":       core.MappingNodeFromString(path),
		"parameters": externalTreeParameters(metadataByKey),
	}
	if region != "" {
		fields["region"] = core.MappingNodeFromString(region)
	}

	// All parameters in the tree share one tag set by construction, so the first managed
	// parameter (in key order) is representative; reading it keeps tag drift detectable.
	firstKey := sortedParameterTreeKeys(metadataByKey)[0]
	tags, err := getParameterTags(ctx, service, parameterTreeFullName(path, firstKey))
	if err != nil {
		return nil, err
	}
	if len(tags.Fields) > 0 {
		fields["tags"] = tags
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{Fields: fields},
	}, nil
}

func externalTreeParameters(
	metadataByKey map[string]ssmtypes.ParameterMetadata,
) *core.MappingNode {
	parameters := &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	for key, metadata := range metadataByKey {
		parameters.Fields[key] = core.MappingNodeFields(
			"arn", core.MappingNodeFromString(aws.ToString(metadata.ARN)),
			"type", core.MappingNodeFromString(string(metadata.Type)),
		)
	}
	return parameters
}
