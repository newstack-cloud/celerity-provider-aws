package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type samlProviderCreate struct {
	name                            string
	samlMetadataDocument            string
	tags                            []types.Tag
	uniqueSAMLProviderNameGenerator utils.UniqueNameGenerator
}

func (s *samlProviderCreate) Name() string {
	return "create SAML provider"
}

func (s *samlProviderCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	deployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	if !ok || deployInput == nil {
		return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in SaveOperationContext.Data")
	}

	// Extract name from spec data or generate one
	name, hasName := pluginutils.GetValueByPath("$.name", specData)
	if hasName && core.StringValue(name) != "" {
		s.name = core.StringValue(name)
	} else {
		// Generate a unique name if not provided
		generatedName, err := s.uniqueSAMLProviderNameGenerator(deployInput)
		if err != nil {
			return false, saveOpCtx, fmt.Errorf("failed to generate unique name: %w", err)
		}
		s.name = generatedName
	}

	// Extract SAML metadata document
	samlMetadataDocument, hasMetadata := pluginutils.GetValueByPath("$.samlMetadataDocument", specData)
	if !hasMetadata || core.StringValue(samlMetadataDocument) == "" {
		return false, saveOpCtx, fmt.Errorf("SAML metadata document must be provided and non-empty")
	}
	s.samlMetadataDocument = core.StringValue(samlMetadataDocument)

	// Extract user tags and merge with Bluelink system tags
	userTags, err := iamTagsFromSpecData(specData)
	if err != nil {
		return false, saveOpCtx, err
	}
	s.tags = mergeBluelinkTagsWithIAMTags(deployInput, userTags)

	return true, saveOpCtx, nil
}

func (s *samlProviderCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	input := &iam.CreateSAMLProviderInput{
		Name:                 aws.String(s.name),
		SAMLMetadataDocument: aws.String(s.samlMetadataDocument),
		Tags:                 sortTagsByKey(s.tags),
	}
	if input.Tags == nil {
		input.Tags = []types.Tag{}
	}

	output, err := iamService.CreateSAMLProvider(ctx, input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create SAML provider: %w", err)
	}

	newSaveOpCtx.Data["createSAMLProviderOutput"] = output
	return newSaveOpCtx, nil
}

func newSAMLProviderCreate(generator utils.UniqueNameGenerator) *samlProviderCreate {
	return &samlProviderCreate{
		uniqueSAMLProviderNameGenerator: generator,
	}
}
