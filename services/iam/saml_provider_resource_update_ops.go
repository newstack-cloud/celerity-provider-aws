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

type samlProviderMetadataUpdate struct {
	arn                  string
	samlMetadataDocument string
}

func (s *samlProviderMetadataUpdate) Name() string {
	return "update SAML provider metadata"
}

func (s *samlProviderMetadataUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the SAML provider ARN from the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	arn, err := utils.ExtractARNFromCurrentState(
		currentStateSpecData,
		"SAML provider metadata update",
	)
	if err != nil {
		return false, saveOpCtx, err
	}
	s.arn = arn

	// Check if samlMetadataDocument was modified
	metadataModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.samlMetadataDocument" {
			metadataModified = true
			break
		}
	}

	if !metadataModified {
		return false, saveOpCtx, nil
	}

	// Get new metadata document
	samlMetadataDocument, hasMetadata := pluginutils.GetValueByPath("$.samlMetadataDocument", specData)
	if !hasMetadata || core.StringValue(samlMetadataDocument) == "" {
		return false, saveOpCtx, fmt.Errorf("SAML metadata document must be provided and non-empty")
	}
	s.samlMetadataDocument = core.StringValue(samlMetadataDocument)

	return true, saveOpCtx, nil
}

func (s *samlProviderMetadataUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := iamService.UpdateSAMLProvider(ctx, &iam.UpdateSAMLProviderInput{
		SAMLMetadataDocument: aws.String(s.samlMetadataDocument),
		SAMLProviderArn:      aws.String(s.arn),
	})
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to update SAML provider metadata: %w", err)
	}

	saveOpCtx.Data["arn"] = s.arn

	return saveOpCtx, nil
}

type samlProviderTagsUpdate struct {
	arn          string
	tagsToAdd    []types.Tag
	tagsToRemove []string
}

func (s *samlProviderTagsUpdate) Name() string {
	return "update SAML provider tags"
}

func (s *samlProviderTagsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the SAML provider ARN from the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	arn, err := utils.ExtractARNFromCurrentState(
		currentStateSpecData,
		"SAML provider tags update",
	)
	if err != nil {
		return false, saveOpCtx, err
	}
	s.arn = arn

	deployInput, _ := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	diffResult := utils.DiffTagsWithBluelink(
		changes,
		deployInput,
		"$.tags",
		toIAMTag,
	)
	s.tagsToAdd = diffResult.ToSet
	s.tagsToRemove = diffResult.ToRemove

	return len(s.tagsToAdd) > 0 || len(s.tagsToRemove) > 0, saveOpCtx, nil
}

func (s *samlProviderTagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove tags
	if len(s.tagsToRemove) > 0 {
		_, err := iamService.UntagSAMLProvider(ctx, &iam.UntagSAMLProviderInput{
			SAMLProviderArn: aws.String(s.arn),
			TagKeys:         s.tagsToRemove,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove tags: %w", err)
		}
	}

	// Add tags
	if len(s.tagsToAdd) > 0 {
		_, err := iamService.TagSAMLProvider(ctx, &iam.TagSAMLProviderInput{
			SAMLProviderArn: aws.String(s.arn),
			Tags:            sortTagsByKey(s.tagsToAdd),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add tags: %w", err)
		}
	}

	return saveOpCtx, nil
}
