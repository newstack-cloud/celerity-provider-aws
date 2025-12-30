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

type oidcProviderClientIdsUpdate struct {
	arn      string
	toAdd    []string
	toRemove []string
}

func (o *oidcProviderClientIdsUpdate) Name() string {
	return "update OIDC provider client IDs"
}

func (o *oidcProviderClientIdsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the OIDC provider ARN from the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	if currentStateSpecData == nil {
		return false, saveOpCtx, fmt.Errorf("current state spec data is required for OIDC provider update")
	}
	arn, hasArn := pluginutils.GetValueByPath("$.arn", currentStateSpecData)
	if !hasArn {
		return false, saveOpCtx, fmt.Errorf("OIDC provider ARN is required for update")
	}
	o.arn = core.StringValue(arn)

	// Check if clientIdList was modified
	clientIdListModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.clientIdList" {
			clientIdListModified = true
			break
		}
	}

	if !clientIdListModified {
		return false, saveOpCtx, nil
	}

	// Get current client IDs
	currentClientIds := []string{}
	if currentClientIdList, exists := pluginutils.GetValueByPath("$.clientIdList", currentStateSpecData); exists && currentClientIdList != nil {
		for _, item := range currentClientIdList.Items {
			currentClientIds = append(currentClientIds, core.StringValue(item))
		}
	}

	// Get desired client IDs
	desiredClientIds := []string{}
	if desiredClientIdList, exists := pluginutils.GetValueByPath("$.clientIdList", specData); exists && desiredClientIdList != nil {
		for _, item := range desiredClientIdList.Items {
			desiredClientIds = append(desiredClientIds, core.StringValue(item))
		}
	}

	// Calculate additions and removals
	o.toAdd = diffStringSlices(desiredClientIds, currentClientIds)
	o.toRemove = diffStringSlices(currentClientIds, desiredClientIds)

	return len(o.toAdd) > 0 || len(o.toRemove) > 0, saveOpCtx, nil
}

func (o *oidcProviderClientIdsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove client IDs
	for _, clientId := range o.toRemove {
		_, err := iamService.RemoveClientIDFromOpenIDConnectProvider(ctx, &iam.RemoveClientIDFromOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(o.arn),
			ClientID:                 aws.String(clientId),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove client ID %s: %w", clientId, err)
		}
	}

	// Add client IDs
	for _, clientId := range o.toAdd {
		_, err := iamService.AddClientIDToOpenIDConnectProvider(ctx, &iam.AddClientIDToOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(o.arn),
			ClientID:                 aws.String(clientId),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add client ID %s: %w", clientId, err)
		}
	}

	return saveOpCtx, nil
}

type oidcProviderThumbprintsUpdate struct {
	arn            string
	thumbprintList []string
}

func (o *oidcProviderThumbprintsUpdate) Name() string {
	return "update OIDC provider thumbprints"
}

func (o *oidcProviderThumbprintsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the OIDC provider ARN from the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	arn, err := utils.ExtractARNFromCurrentState(
		currentStateSpecData,
		"OIDC provider thumbprint update",
	)
	if err != nil {
		return false, saveOpCtx, err
	}
	o.arn = arn

	// Check if thumbprintList was modified
	thumbprintListModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.thumbprintList" {
			thumbprintListModified = true
			break
		}
	}

	if !thumbprintListModified {
		return false, saveOpCtx, nil
	}

	// Get desired thumbprints
	o.thumbprintList = []string{}
	if desiredThumbprintList, exists := pluginutils.GetValueByPath("$.thumbprintList", specData); exists && desiredThumbprintList != nil {
		for _, item := range desiredThumbprintList.Items {
			o.thumbprintList = append(o.thumbprintList, core.StringValue(item))
		}
	}

	return true, saveOpCtx, nil
}

func (o *oidcProviderThumbprintsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// AWS replaces the entire thumbprint list
	_, err := iamService.UpdateOpenIDConnectProviderThumbprint(ctx, &iam.UpdateOpenIDConnectProviderThumbprintInput{
		OpenIDConnectProviderArn: aws.String(o.arn),
		ThumbprintList:           o.thumbprintList,
	})
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to update thumbprints: %w", err)
	}

	return saveOpCtx, nil
}

type oidcProviderTagsUpdate struct {
	arn      string
	toAdd    []types.Tag
	toRemove []string
}

func (o *oidcProviderTagsUpdate) Name() string {
	return "update OIDC provider tags"
}

func (o *oidcProviderTagsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the OIDC provider ARN from the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	arn, err := utils.ExtractARNFromCurrentState(
		currentStateSpecData,
		"OIDC provider tags update",
	)
	if err != nil {
		return false, saveOpCtx, err
	}
	o.arn = arn

	deployInput, _ := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	diffResult := utils.DiffTagsWithBluelink(
		changes,
		deployInput,
		"$.tags",
		toIAMTag,
	)

	o.toAdd = diffResult.ToSet
	o.toRemove = diffResult.ToRemove

	return len(o.toAdd) > 0 || len(o.toRemove) > 0, saveOpCtx, nil
}

func (o *oidcProviderTagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove tags
	if len(o.toRemove) > 0 {
		_, err := iamService.UntagOpenIDConnectProvider(ctx, &iam.UntagOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(o.arn),
			TagKeys:                  o.toRemove,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove tags: %w", err)
		}
	}

	// Add/update tags
	if len(o.toAdd) > 0 {
		_, err := iamService.TagOpenIDConnectProvider(ctx, &iam.TagOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(o.arn),
			Tags:                     o.toAdd,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add tags: %w", err)
		}
	}

	return saveOpCtx, nil
}

// diffStringSlices returns elements in a that are not in b.
func diffStringSlices(a, b []string) []string {
	bMap := make(map[string]bool)
	for _, v := range b {
		bMap[v] = true
	}

	var diff []string
	for _, v := range a {
		if !bMap[v] {
			diff = append(diff, v)
		}
	}

	return diff
}
