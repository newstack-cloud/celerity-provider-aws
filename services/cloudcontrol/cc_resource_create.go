package cloudcontrol

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/overlays"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *ccResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	service, err := a.getCloudControlService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	desiredNode, err := a.buildDesiredCFNNode(input)
	if err != nil {
		return nil, err
	}
	desiredState, err := mappingNodeToCFNJSON(desiredNode)
	if err != nil {
		return nil, err
	}

	requestToken, identifier, err := a.submitCreate(ctx, service, input, desiredState)
	if err != nil {
		return nil, err
	}

	if overlays.IsStabiliseRequired(a.config.BlueprintType) {
		// Slow-to-stabilise resources (e.g. an RDS instance) can take minutes to
		// provision and their computed fields (endpoint, ARN) are not readable until the
		// operation succeeds. Rather than block Deploy, return as soon as the identifier
		// is known; the request token drives HasStabilised, which captures every computed
		// field once the operation succeeds. Consumers wait for stabilisation via the
		// stabilised-dependency declaration, so the absence of computed fields at
		// config-complete is safe.
		return a.deferredComputedFieldsOutput(ctx, service, requestToken, identifier)
	}

	return a.captureComputedFields(
		ctx,
		service,
		input,
		requestToken,
		identifier,
	)
}

func (a *ccResourceActions) submitCreate(
	ctx context.Context,
	service cloudcontrolservice.Service,
	input *provider.ResourceDeployInput,
	desiredState string,
) (requestToken string, identifier string, err error) {
	createResource := pluginutils.RetryableReturnValue(
		func(
			ctx context.Context,
			in *awscc.CreateResourceInput,
		) (*awscc.CreateResourceOutput, error) {
			return service.CreateResource(ctx, in)
		},
		isCCErrorRetryable,
	)

	output, err := createResource(ctx, &awscc.CreateResourceInput{
		TypeName:     aws.String(a.config.CFNType),
		DesiredState: aws.String(desiredState),
		ClientToken:  aws.String(clientToken("create", input.InstanceID, input.ResourceID)),
	})
	if err != nil {
		return "", "", err
	}

	if output.ProgressEvent == nil {
		return "", "", fmt.Errorf(
			"cloud control create for %s returned no progress event",
			a.config.CFNType,
		)
	}

	return aws.ToString(output.ProgressEvent.RequestToken),
		aws.ToString(output.ProgressEvent.Identifier),
		nil
}

func (a *ccResourceActions) deferredComputedFieldsOutput(
	ctx context.Context,
	service cloudcontrolservice.Service,
	requestToken string,
	identifier string,
) (*provider.ResourceDeployOutput, error) {
	resolvedID, err := a.waitForIdentifier(ctx, service, requestToken, identifier)
	if err != nil {
		return nil, err
	}
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: a.computedFieldValues(nil, resolvedID, requestToken, nil),
	}, nil
}
