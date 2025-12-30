package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *iamServerCertificateResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := a.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Try to get serverCertificateName from current spec first
	certName := ""
	serverCertName, hasServerCertName := pluginutils.GetValueByPath(
		"$.serverCertificateName",
		input.CurrentResourceSpec,
	)
	if hasServerCertName {
		certName = core.StringValue(serverCertName)
	}

	// If no name but we have ARN, extract name from ARN
	if certName == "" {
		arnValue, hasArn := pluginutils.GetValueByPath("$.arn", input.CurrentResourceSpec)
		if hasArn {
			arn := core.StringValue(arnValue)
			certName = extractCertificateNameFromARN(arn)
		}
	}

	// If still no name, attempt fallback lookup by Bluelink tags
	if certName == "" {
		fallbackName, err := a.lookupServerCertificateNameByTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup server certificate by tags: %w", err)
		}
		if fallbackName == "" {
			// Resource doesn't exist yet
			return &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				},
			}, nil
		}
		certName = fallbackName
	}

	serverCert, err := iamService.GetServerCertificate(ctx, &iam.GetServerCertificateInput{
		ServerCertificateName: aws.String(certName),
	})
	if err != nil {
		return nil, err
	}

	// The private key is not accessible via the AWS service call for security reasons,
	// once you upload the private key to AWS, it can no longer be retrieved via the API.
	// Private keys are stored in state but are marked as sensitive in the schema to prevent them
	// from being displayed in tool UIs and logs.
	privateKey, hasPrivateKey := pluginutils.GetValueByPath(
		"$.privateKey",
		input.CurrentResourceSpec,
	)
	if !hasPrivateKey {
		return nil, fmt.Errorf(
			"privateKey is expected to be in the current state of the server certificate resource",
		)
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"arn": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.ServerCertificateMetadata.Arn),
				),
				"certificateBody": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.CertificateBody),
				),
				"certificateChain": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.CertificateChain),
				),
				"path": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.ServerCertificateMetadata.Path),
				),
				"privateKey": privateKey,
				"serverCertificateName": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.ServerCertificateMetadata.ServerCertificateName),
				),
				"tags": extractUserIAMTags(
					serverCert.ServerCertificate.Tags,
					input.ProviderContext,
				),
			},
		},
	}, nil
}

// extractCertificateNameFromARN extracts the certificate name from an ARN.
// ARN format: arn:aws:iam::123456789012:server-certificate/certificate-name.
func extractCertificateNameFromARN(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// lookupServerCertificateNameByTags attempts to find an IAM server certificate by its Bluelink provenance tags.
// This is used as a fallback when the name is not available (e.g., interrupted resource creation).
// Returns empty string if no matching resource is found.
func (a *iamServerCertificateResourceActions) lookupServerCertificateNameByTags(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (string, error) {
	tagFilters := utils.BuildBluelinkTagFiltersForLookup(input)
	if tagFilters == nil {
		// Tagging is not enabled, cannot perform fallback lookup
		return "", nil
	}

	taggingService, err := a.getResourceGroupTaggingService(ctx, input.ProviderContext)
	if err != nil {
		return "", err
	}

	result, err := taggingService.GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters:          tagFilters,
		ResourceTypeFilters: []string{"iam:server-certificate"},
	})
	if err != nil {
		return "", err
	}

	if len(result.ResourceTagMappingList) == 0 {
		return "", nil
	}

	// Extract certificate name from ARN
	// ARN format: arn:aws:iam::123456789012:server-certificate/certificate-name
	arn := aws.ToString(result.ResourceTagMappingList[0].ResourceARN)
	parts := strings.Split(arn, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid server certificate ARN format: %s", arn)
	}
	return parts[len(parts)-1], nil
}
