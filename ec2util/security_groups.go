package ec2util

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

var (
	// Deliberately short. AWS releases the network interfaces holding a group anywhere
	// between seconds and a quarter of an hour for the very same operation, measured
	// against a real account, so there is no timeout worth picking: long enough for the
	// slow case wastes a worker in the common one, and short enough for the common case
	// fails constantly. This window covers only the fast case; anything longer is handed
	// back to the deployment engine.
	SecurityGroupDeleteTimeout      = 30 * time.Second
	SecurityGroupDeletePollInterval = 5 * time.Second
)

// DeleteSecurityGroupWhenUnused deletes a security group, giving whatever still holds
// it a brief chance to be released and handing the work back if it is not.
//
// A group cannot be deleted while a network interface references it, and interfaces
// outlive the thing that created them by an unpredictable margin: Lambda's outlive the
// function, and a VPC endpoint's outlive the endpoint. Rather than block a deployment
// worker for that margin, this tries for a short while and then returns a retryable
// error so the engine re-runs the destroy later. Destroys are idempotent, so re-running
// costs nothing.
func DeleteSecurityGroupWhenUnused(
	ctx context.Context,
	service ec2service.Service,
	groupID string,
) error {
	deadline := time.Now().Add(SecurityGroupDeleteTimeout)
	for {
		// AWS reclaims a detached Lambda interface on its own schedule, long after the
		// function is gone, so those are deleted here rather than waited on.
		if err := ReapReleasedLambdaENIs(ctx, service, groupID); err != nil {
			return err
		}

		_, err := service.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(groupID),
		})
		if err == nil || IsGroupNotFoundError(err) {
			return nil
		}
		if !IsDependencyViolationError(err) {
			return err
		}

		if time.Now().After(deadline) {
			// Retryable rather than fatal: the group is still referenced by something
			// AWS has not finished releasing, which resolves on its own given time.
			return &provider.RetryableError{
				ChildError: fmt.Errorf(
					"security group %s is still referenced, most likely by network interfaces "+
						"AWS has not released yet: %w",
					groupID,
					err,
				),
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(SecurityGroupDeletePollInterval):
		}
	}
}

// ReapReleasedLambdaENIs deletes the network interfaces Lambda has finished with on a
// security group.
//
// Only interfaces Lambda manages, and only once they are detached. An interface still
// in use belongs to a running workload, and one belonging to another service (a VPC
// endpoint, say) is that service's to remove; deleting either would break something
// that is still live.
//
// The detach itself cannot be hurried from here; this removes the separate delay
// between AWS detaching an interface and AWS deleting it.
func ReapReleasedLambdaENIs(
	ctx context.Context,
	service ec2service.Service,
	groupID string,
) error {
	output, err := service.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("group-id"),
				Values: []string{groupID},
			},
			{
				Name: aws.String("status"),
				Values: []string{
					string(types.NetworkInterfaceStatusAvailable),
				},
			},
		},
	})
	if err != nil {
		return err
	}
	if output == nil {
		return nil
	}

	for _, eni := range output.NetworkInterfaces {
		if !isLambdaManagedENI(eni) {
			continue
		}

		_, err := service.DeleteNetworkInterface(ctx, &ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: eni.NetworkInterfaceId,
		})
		// Racing AWS reclaiming the same interface is the expected outcome, not a failure.
		if err != nil && !isNetworkInterfaceGoneOrInUseError(err) {
			return err
		}
	}

	return nil
}

// Lambda names the interfaces it creates for a VPC-attached function, which is how they
// are told apart from interfaces belonging to VPC endpoints or to a customer workload.
func isLambdaManagedENI(eni types.NetworkInterface) bool {
	return strings.HasPrefix(aws.ToString(eni.Description), "AWS Lambda VPC ENI")
}

// IsDependencyViolationError reports whether an EC2 call failed because the resource is
// still referenced by something else.
func IsDependencyViolationError(err error) bool {
	return hasErrorCode(err, "DependencyViolation")
}

// IsGroupNotFoundError reports whether a security group is already gone, which is the
// expected outcome of a retried teardown.
func IsGroupNotFoundError(err error) bool {
	return hasErrorCode(err, "InvalidGroup.NotFound")
}

func isNetworkInterfaceGoneOrInUseError(err error) bool {
	return hasErrorCode(err, "InvalidNetworkInterfaceID.NotFound") ||
		hasErrorCode(err, "InvalidParameterValue")
}

func hasErrorCode(err error, code string) bool {
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		return apiErr.ErrorCode() == code
	}
	return false
}
