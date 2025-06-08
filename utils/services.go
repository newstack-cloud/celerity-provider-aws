package utils

import (
	"fmt"
	"strings"

	"github.com/two-hundred/celerity/libs/blueprint/core"
	"github.com/two-hundred/celerity/libs/blueprint/provider"
)

// AWSServiceList returns a string of all the AWS services and their aliases.
func AWSServiceList() string {
	servicesSB := strings.Builder{}

	for service, aliases := range Services {
		servicesSB.WriteString(fmt.Sprintf("- %s", service))
		if len(aliases) > 0 {
			servicesSB.WriteString(
				fmt.Sprintf(" (%s)\n", strings.Join(aliases, ", ")),
			)
		}
	}

	return servicesSB.String()
}

// Services is a map of AWS services and their aliases.
var Services = map[string][]string{
	"account":  {},
	"lambda":   {},
	"dynamodb": {},
	"sqs":      {},
}

// GetEndpointFromProviderConfig returns the endpoint for a given service or one of its aliases.
func GetEndpointFromProviderConfig(
	providerContext provider.Context,
	service string,
	aliases []string,
) (*core.ScalarValue, bool) {
	endpoint, hasEndpoint := providerContext.ProviderConfigVariable(
		fmt.Sprintf("endpoint.%s", service),
	)
	if hasEndpoint && !core.IsScalarNil(endpoint) {
		return endpoint, true
	}

	for _, alias := range aliases {
		endpoint, hasEndpoint = providerContext.ProviderConfigVariable(
			fmt.Sprintf("endpoint.%s", alias),
		)
		if hasEndpoint && !core.IsScalarNil(endpoint) {
			return endpoint, true
		}
	}

	return nil, false
}
