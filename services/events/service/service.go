package eventsservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the Amazon EventBridge
// service used by the EventBridge resource, data source and link implementations.
type Service interface {
	// CreateEventBus creates a new custom event bus within an account.
	CreateEventBus(
		ctx context.Context,
		params *eventbridge.CreateEventBusInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.CreateEventBusOutput, error)

	// UpdateEventBus updates the configuration of an existing custom event bus.
	UpdateEventBus(
		ctx context.Context,
		params *eventbridge.UpdateEventBusInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.UpdateEventBusOutput, error)

	// DeleteEventBus deletes the specified custom event bus.
	DeleteEventBus(
		ctx context.Context,
		params *eventbridge.DeleteEventBusInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DeleteEventBusOutput, error)

	// DescribeEventBus displays details about an event bus in an account.
	DescribeEventBus(
		ctx context.Context,
		params *eventbridge.DescribeEventBusInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DescribeEventBusOutput, error)

	// PutPermission revises the resource-based policy that grants another account
	// permission to put events to the specified event bus.
	PutPermission(
		ctx context.Context,
		params *eventbridge.PutPermissionInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.PutPermissionOutput, error)

	// RemovePermission revokes the permission of another account to put events to
	// the specified event bus.
	RemovePermission(
		ctx context.Context,
		params *eventbridge.RemovePermissionInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.RemovePermissionOutput, error)

	// PutRule creates or updates the specified rule.
	PutRule(
		ctx context.Context,
		params *eventbridge.PutRuleInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.PutRuleOutput, error)

	// DeleteRule deletes the specified rule.
	DeleteRule(
		ctx context.Context,
		params *eventbridge.DeleteRuleInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DeleteRuleOutput, error)

	// DescribeRule describes the specified rule.
	DescribeRule(
		ctx context.Context,
		params *eventbridge.DescribeRuleInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DescribeRuleOutput, error)

	// ListRules lists your rules, optionally filtered by event bus and name prefix.
	ListRules(
		ctx context.Context,
		params *eventbridge.ListRulesInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.ListRulesOutput, error)

	// PutTargets adds the specified targets to the specified rule, or updates the
	// targets if they are already associated with the rule.
	PutTargets(
		ctx context.Context,
		params *eventbridge.PutTargetsInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.PutTargetsOutput, error)

	// RemoveTargets removes the specified targets from the specified rule.
	RemoveTargets(
		ctx context.Context,
		params *eventbridge.RemoveTargetsInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.RemoveTargetsOutput, error)

	// ListTargetsByRule lists the targets assigned to the specified rule.
	ListTargetsByRule(
		ctx context.Context,
		params *eventbridge.ListTargetsByRuleInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.ListTargetsByRuleOutput, error)

	// TagResource adds the specified tags to the specified EventBridge resource.
	// Rules and event buses can be tagged.
	TagResource(
		ctx context.Context,
		params *eventbridge.TagResourceInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.TagResourceOutput, error)

	// UntagResource removes the specified tags from the specified EventBridge resource.
	UntagResource(
		ctx context.Context,
		params *eventbridge.UntagResourceInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.UntagResourceOutput, error)

	// ListTagsForResource displays the tags associated with an EventBridge resource.
	ListTagsForResource(
		ctx context.Context,
		params *eventbridge.ListTagsForResourceInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.ListTagsForResourceOutput, error)

	// CreateArchive creates an archive of events with the specified settings.
	CreateArchive(
		ctx context.Context,
		params *eventbridge.CreateArchiveInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.CreateArchiveOutput, error)

	// UpdateArchive updates the specified archive.
	UpdateArchive(
		ctx context.Context,
		params *eventbridge.UpdateArchiveInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.UpdateArchiveOutput, error)

	// DeleteArchive deletes the specified archive.
	DeleteArchive(
		ctx context.Context,
		params *eventbridge.DeleteArchiveInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DeleteArchiveOutput, error)

	// DescribeArchive retrieves details about an archive.
	DescribeArchive(
		ctx context.Context,
		params *eventbridge.DescribeArchiveInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DescribeArchiveOutput, error)

	// CreateConnection creates a connection that can be used with an API
	// destination to authorize against an HTTP endpoint.
	CreateConnection(
		ctx context.Context,
		params *eventbridge.CreateConnectionInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.CreateConnectionOutput, error)

	// UpdateConnection updates the settings of an existing connection.
	UpdateConnection(
		ctx context.Context,
		params *eventbridge.UpdateConnectionInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.UpdateConnectionOutput, error)

	// DeleteConnection deletes the specified connection.
	DeleteConnection(
		ctx context.Context,
		params *eventbridge.DeleteConnectionInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DeleteConnectionOutput, error)

	// DescribeConnection retrieves details about a connection.
	DescribeConnection(
		ctx context.Context,
		params *eventbridge.DescribeConnectionInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DescribeConnectionOutput, error)

	// CreateApiDestination creates an API destination, which is an HTTP
	// invocation endpoint configured as a target for events.
	CreateApiDestination(
		ctx context.Context,
		params *eventbridge.CreateApiDestinationInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.CreateApiDestinationOutput, error)

	// UpdateApiDestination updates an API destination.
	UpdateApiDestination(
		ctx context.Context,
		params *eventbridge.UpdateApiDestinationInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.UpdateApiDestinationOutput, error)

	// DeleteApiDestination deletes the specified API destination.
	DeleteApiDestination(
		ctx context.Context,
		params *eventbridge.DeleteApiDestinationInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DeleteApiDestinationOutput, error)

	// DescribeApiDestination retrieves details about an API destination.
	DescribeApiDestination(
		ctx context.Context,
		params *eventbridge.DescribeApiDestinationInput,
		optFns ...func(*eventbridge.Options),
	) (*eventbridge.DescribeApiDestinationOutput, error)
}

// NewService creates a new instance of the AWS EventBridge service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return eventbridge.NewFromConfig(
		*awsConfig,
		eventbridge.WithEndpointResolverV2(
			&eventsEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type eventsEndpointResolverV2 struct {
	providerContext provider.Context
}

func (e *eventsEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params eventbridge.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	eventsAliases := utils.Services["events"]
	eventsEndpoint, hasEventsEndpoint := utils.GetEndpointFromProviderConfig(
		e.providerContext,
		"events",
		eventsAliases,
	)
	if hasEventsEndpoint && !core.IsScalarNil(eventsEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(eventsEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return eventbridge.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
