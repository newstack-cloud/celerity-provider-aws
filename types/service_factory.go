package types

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/two-hundred/celerity/libs/blueprint/provider"
)

// ServiceFactory is a function type that creates service instances
// to allow the provider to create clients on the fly
// based on provider configuration in the request to
// interact with AWS services.
type ServiceFactory[Service any] func(
	awsConfig *aws.Config,
	providerContext provider.Context,
) Service
