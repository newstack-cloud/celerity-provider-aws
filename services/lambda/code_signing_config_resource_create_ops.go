package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type codeSigningConfigCreate struct {
	input *lambda.CreateCodeSigningConfigInput
}

func (u *codeSigningConfigCreate) Name() string {
	return "create code signing config"
}

func (u *codeSigningConfigCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToCreateCodeSigningConfigInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}

	// Merge Bluelink system tags with user-defined tags
	// Note: CodeSigningConfig doesn't support tags at creation time,
	// tags are applied via TagResource after creation by the tagsUpdate operation
	u.input = input
	return hasValues, saveOpCtx, nil
}

func (u *codeSigningConfigCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createCodeSigningConfigOutput, err := lambdaService.CreateCodeSigningConfig(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(createCodeSigningConfigOutput.CodeSigningConfig.CodeSigningConfigArn)
	newSaveOpCtx.Data["createCodeSigningConfigOutput"] = createCodeSigningConfigOutput
	newSaveOpCtx.Data["codeSigningConfigArn"] = aws.ToString(createCodeSigningConfigOutput.CodeSigningConfig.CodeSigningConfigArn)

	return newSaveOpCtx, nil
}

func changesToCreateCodeSigningConfigInput(
	specData *core.MappingNode,
) (*lambda.CreateCodeSigningConfigInput, bool, error) {
	input := &lambda.CreateCodeSigningConfigInput{}

	valueSetters := []*pluginutils.ValueSetter[*lambda.CreateCodeSigningConfigInput]{
		pluginutils.NewValueSetter(
			"$.description",
			func(value *core.MappingNode, input *lambda.CreateCodeSigningConfigInput) {
				input.Description = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.allowedPublishers",
			func(value *core.MappingNode, input *lambda.CreateCodeSigningConfigInput) {
				allowedPublishers := &types.AllowedPublishers{}
				if arns, exists := pluginutils.GetValueByPath("$.signingProfileVersionArns", value); exists {
					var arnList []string
					for _, arnNode := range arns.Items {
						arnList = append(arnList, core.StringValue(arnNode))
					}
					allowedPublishers.SigningProfileVersionArns = arnList
				}
				input.AllowedPublishers = allowedPublishers
			},
		),
		pluginutils.NewValueSetter(
			"$.codeSigningPolicies",
			func(value *core.MappingNode, input *lambda.CreateCodeSigningConfigInput) {
				codeSigningPolicies := &types.CodeSigningPolicies{}
				if policy, exists := pluginutils.GetValueByPath("$.untrustedArtifactOnDeployment", value); exists {
					policyStr := core.StringValue(policy)
					codeSigningPolicies.UntrustedArtifactOnDeployment = types.CodeSigningPolicy(policyStr)
				}
				input.CodeSigningPolicies = codeSigningPolicies
			},
		),
	}

	hasValuesToSave := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasValuesToSave = hasValuesToSave || valueSetter.DidSet()
	}

	return input, hasValuesToSave, nil
}
