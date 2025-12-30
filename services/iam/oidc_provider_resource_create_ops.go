package iam

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type oidcProviderCreate struct {
	url                            string
	clientIdList                   []string
	thumbprintList                 []string
	tags                           []types.Tag
	uniqueOIDCProviderUrlGenerator utils.UniqueNameGenerator
}

func (o *oidcProviderCreate) Name() string {
	return "create OIDC provider"
}

func (o *oidcProviderCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	deployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	if !ok || deployInput == nil {
		return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in SaveOperationContext.Data")
	}

	// Extract URL from spec data
	url, hasUrl := pluginutils.GetValueByPath("$.url", specData)
	if !hasUrl || core.StringValue(url) == "" {
		return false, saveOpCtx, fmt.Errorf("OIDC provider URL must be provided and non-empty")
	}
	o.url = core.StringValue(url)

	// Extract client IDs
	clientIdList, hasClientIds := pluginutils.GetValueByPath("$.clientIdList", specData)
	if hasClientIds && clientIdList != nil && len(clientIdList.Items) > 0 {
		clientIdItems := clientIdList.Items
		o.clientIdList = make([]string, len(clientIdItems))
		for i, item := range clientIdItems {
			o.clientIdList[i] = core.StringValue(item)
		}
	}

	// Extract thumbprints
	thumbprintList, hasThumbprints := pluginutils.GetValueByPath("$.thumbprintList", specData)
	if hasThumbprints && thumbprintList != nil && len(thumbprintList.Items) > 0 {
		thumbprintItems := thumbprintList.Items
		o.thumbprintList = make([]string, len(thumbprintItems))
		for i, item := range thumbprintItems {
			o.thumbprintList[i] = core.StringValue(item)
		}
	}

	// Extract user tags and merge with Bluelink system tags
	userTags, err := iamTagsFromSpecData(specData)
	if err != nil {
		return false, saveOpCtx, err
	}
	o.tags = mergeBluelinkTagsWithIAMTags(deployInput, userTags)

	return true, saveOpCtx, nil
}

func (o *oidcProviderCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	input := &iam.CreateOpenIDConnectProviderInput{
		Url:            aws.String(o.url),
		ClientIDList:   o.clientIdList,
		ThumbprintList: o.thumbprintList,
		Tags:           sortTagsByKeyForOIDC(o.tags),
	}

	output, err := iamService.CreateOpenIDConnectProvider(ctx, input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	newSaveOpCtx.Data["createOIDCProviderOutput"] = output
	newSaveOpCtx.Data["oidcProviderArn"] = output.OpenIDConnectProviderArn
	return newSaveOpCtx, nil
}

func newOIDCProviderCreate(generator utils.UniqueNameGenerator) *oidcProviderCreate {
	return &oidcProviderCreate{
		uniqueOIDCProviderUrlGenerator: generator,
	}
}

// sortTagsByKeyForOIDC sorts a slice of types.Tag by their Key field.
func sortTagsByKeyForOIDC(tags []types.Tag) []types.Tag {
	sorted := make([]types.Tag, len(tags))
	copy(sorted, tags)
	sort.Slice(sorted, func(i, j int) bool {
		return aws.ToString(sorted[i].Key) < aws.ToString(sorted[j].Key)
	})
	return sorted
}
