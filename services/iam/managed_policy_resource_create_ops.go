package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type managedPolicyCreate struct {
	input                     *iam.CreatePolicyInput
	uniquePolicyNameGenerator utils.UniqueNameGenerator
}

func (m *managedPolicyCreate) Name() string {
	return "create managed policy"
}

func (m *managedPolicyCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	createInput, hasValues, err := changesToCreatePolicyInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}

	deployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	if !ok || deployInput == nil {
		return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in SaveOperationContext.Data")
	}

	// Generate unique policy name if not provided
	if createInput.PolicyName == nil || aws.ToString(createInput.PolicyName) == "" {
		// Use the injected generator or default to the IAM policy generator
		generator := m.uniquePolicyNameGenerator
		if generator == nil {
			generator = utils.IAMPolicyNameGenerator
		}

		uniquePolicyName, err := generator(deployInput)
		if err != nil {
			return false, saveOpCtx, err
		}

		createInput.PolicyName = aws.String(uniquePolicyName)
		hasValues = true
	}

	// Merge Bluelink system tags with user-defined tags
	createInput.Tags = mergeBluelinkTagsWithIAMTags(deployInput, createInput.Tags)

	m.input = createInput
	return hasValues, saveOpCtx, nil
}

func (m *managedPolicyCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createPolicyOutput, err := iamService.CreatePolicy(ctx, m.input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create IAM managed policy: %w", err)
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(createPolicyOutput.Policy.Arn)
	newSaveOpCtx.Data["createPolicyOutput"] = createPolicyOutput
	newSaveOpCtx.Data["policyArn"] = aws.ToString(createPolicyOutput.Policy.Arn)

	return newSaveOpCtx, nil
}

func newManagedPolicyCreate(generator utils.UniqueNameGenerator) *managedPolicyCreate {
	return &managedPolicyCreate{
		uniquePolicyNameGenerator: generator,
	}
}
