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

type serverCertificateUpdate struct {
	input *iam.UpdateServerCertificateInput
}

func (s *serverCertificateUpdate) Name() string {
	return "update server certificate"
}

func (s *serverCertificateUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	currentSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentServerCertificateName, hasCurrentServerCertificateName := pluginutils.GetValueByPath(
		"$.serverCertificateName",
		currentSpecData,
	)
	if !hasCurrentServerCertificateName {
		return false, saveOpCtx, fmt.Errorf("server certificate name is required for update operation")
	}

	input := &iam.UpdateServerCertificateInput{
		ServerCertificateName: aws.String(core.StringValue(currentServerCertificateName)),
	}
	finalServerCertificateName := core.StringValue(currentServerCertificateName)
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.serverCertificateName" {
			newCertificateName := core.StringValue(fieldChange.NewValue)
			input.NewServerCertificateName = aws.String(newCertificateName)
			finalServerCertificateName = newCertificateName
		}
		if fieldChange.FieldPath == "spec.path" {
			input.NewPath = aws.String(core.StringValue(fieldChange.NewValue))
		}
	}

	s.input = input

	saveOpCtx.Data["finalServerCertificateName"] = finalServerCertificateName

	hasChanges := input.NewPath != nil || input.NewServerCertificateName != nil
	return hasChanges, saveOpCtx, nil
}

func (s *serverCertificateUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := iamService.UpdateServerCertificate(ctx, s.input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to update server certificate: %w", err)
	}

	return saveOpCtx, nil
}

type serverCertificateTagsUpdate struct {
	serverCertificateName string
	tagsToAdd             []types.Tag
	tagsToRemove          []string
}

func (s *serverCertificateTagsUpdate) Name() string {
	return "update server certificate tags"
}

func (s *serverCertificateTagsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	serverCertificateName, hasServerCertificateName := pluginutils.GetValueByPath(
		"$.serverCertificateName",
		currentStateSpecData,
	)
	if !hasServerCertificateName {
		return false, saveOpCtx, fmt.Errorf("server certificate name is required for update operation")
	}

	s.serverCertificateName = core.StringValue(serverCertificateName)

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

func (s *serverCertificateTagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove tags
	if len(s.tagsToRemove) > 0 {
		_, err := iamService.UntagServerCertificate(ctx, &iam.UntagServerCertificateInput{
			ServerCertificateName: aws.String(s.serverCertificateName),
			TagKeys:               s.tagsToRemove,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove tags: %w", err)
		}
	}

	// Add tags
	if len(s.tagsToAdd) > 0 {
		_, err := iamService.TagServerCertificate(ctx, &iam.TagServerCertificateInput{
			ServerCertificateName: aws.String(s.serverCertificateName),
			Tags:                  sortTagsByKey(s.tagsToAdd),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add tags: %w", err)
		}
	}

	return saveOpCtx, nil
}
