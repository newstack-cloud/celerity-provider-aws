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

type serverCertificateCreate struct {
	input                                *iam.UploadServerCertificateInput
	uniqueServerCertificateNameGenerator utils.UniqueNameGenerator
}

func (s *serverCertificateCreate) Name() string {
	return "create SAML provider"
}

func (s *serverCertificateCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	deployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	if !ok || deployInput == nil {
		return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in SaveOperationContext.Data")
	}

	name, err := s.getServerCertificateName(specData, changes)
	if err != nil {
		return false, saveOpCtx, err
	}

	userTags, err := iamTagsFromSpecData(specData)
	if err != nil {
		return false, saveOpCtx, err
	}

	// Merge Bluelink system tags with user-defined tags
	mergedTags := mergeBluelinkTagsWithIAMTags(deployInput, userTags)

	input := newSpecToUploadServerCertificateInput(specData, name, mergedTags)

	s.input = input

	return true, saveOpCtx, nil
}

func (s *serverCertificateCreate) getServerCertificateName(
	specData *core.MappingNode,
	changes *provider.Changes,
) (string, error) {
	name, hasName := pluginutils.GetValueByPath("$.serverCertificateName", specData)
	if hasName && core.StringValue(name) != "" {
		return core.StringValue(name), nil
	}

	generatedName, err := s.uniqueServerCertificateNameGenerator(&provider.ResourceDeployInput{
		Changes: changes,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate unique name: %w", err)
	}
	return generatedName, nil
}

func newSpecToUploadServerCertificateInput(
	specData *core.MappingNode,
	serverCertificateName string,
	iamTags []types.Tag,
) *iam.UploadServerCertificateInput {
	input := &iam.UploadServerCertificateInput{
		ServerCertificateName: aws.String(serverCertificateName),
	}

	if len(iamTags) > 0 {
		input.Tags = sortTagsByKey(iamTags)
	}

	valueSetters := []*pluginutils.ValueSetter[*iam.UploadServerCertificateInput]{
		pluginutils.NewValueSetter(
			"$.certificateBody",
			func(value *core.MappingNode, input *iam.UploadServerCertificateInput) {
				input.CertificateBody = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.privateKey",
			func(value *core.MappingNode, input *iam.UploadServerCertificateInput) {
				input.PrivateKey = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.certificateChain",
			func(value *core.MappingNode, input *iam.UploadServerCertificateInput) {
				input.CertificateChain = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.path",
			func(value *core.MappingNode, input *iam.UploadServerCertificateInput) {
				input.Path = aws.String(core.StringValue(value))
			},
		),
	}

	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
	}

	return input
}

func (s *serverCertificateCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	output, err := iamService.UploadServerCertificate(ctx, s.input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create server certificate: %w", err)
	}

	newSaveOpCtx.Data["uploadServerCertificateOutput"] = output
	return newSaveOpCtx, nil
}

func newServerCertificateCreate(generator utils.UniqueNameGenerator) *serverCertificateCreate {
	return &serverCertificateCreate{
		uniqueServerCertificateNameGenerator: generator,
	}
}
