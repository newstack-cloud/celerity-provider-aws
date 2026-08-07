//go:build unit

package ec2util

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// The real poll interval and timeout are sized for AWS; these tests only care about the
// decisions taken between polls.
func withFastGroupDeletePolling(t *testing.T) {
	originalInterval := SecurityGroupDeletePollInterval
	originalTimeout := SecurityGroupDeleteTimeout
	SecurityGroupDeletePollInterval = time.Millisecond
	SecurityGroupDeleteTimeout = 50 * time.Millisecond
	t.Cleanup(func() {
		SecurityGroupDeletePollInterval = originalInterval
		SecurityGroupDeleteTimeout = originalTimeout
	})
}

func dependencyViolationErr() error {
	return &smithy.GenericAPIError{
		Code:    "DependencyViolation",
		Message: "resource sg-workload has a dependent object",
	}
}

// The mock's call assertions take a testify suite; these tests are plain functions.
func testSuite(t *testing.T) *suite.Suite {
	s := new(suite.Suite)
	s.SetT(t)
	return s
}

// Deleting a security group races AWS releasing whatever still references it, so a
// dependency violation is retried rather than treated as failure.
func TestGroupDeleteRetriesWhileTheGroupIsStillReferenced(t *testing.T) {
	withFastGroupDeletePolling(t)

	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDeleteSecurityGroupErrorsThenSuccess(
			[]error{dependencyViolationErr(), dependencyViolationErr()},
		),
	)

	err := DeleteSecurityGroupWhenUnused(context.Background(), service, "sg-workload")
	require.NoError(t, err)
	require.Equal(
		t,
		3,
		service.DeleteSecurityGroupCallCount(),
		"expected two retries before the group was free",
	)
}

// A group already gone is success: teardown may be retried after a partial failure.
func TestGroupDeleteTreatsAMissingGroupAsDeleted(t *testing.T) {
	withFastGroupDeletePolling(t)

	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDeleteSecurityGroupErrorsThenSuccess(
			[]error{&smithy.GenericAPIError{Code: "InvalidGroup.NotFound"}},
		),
	)

	err := DeleteSecurityGroupWhenUnused(context.Background(), service, "sg-workload")
	require.NoError(t, err)
	require.Equal(t, 1, service.DeleteSecurityGroupCallCount())
}

// Anything other than a dependency AWS is still releasing is returned immediately.
func TestGroupDeleteDoesNotRetryOtherErrors(t *testing.T) {
	withFastGroupDeletePolling(t)

	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDeleteSecurityGroupErrorsThenSuccess(
			[]error{&smithy.GenericAPIError{Code: "UnauthorizedOperation"}},
		),
	)

	err := DeleteSecurityGroupWhenUnused(context.Background(), service, "sg-workload")
	require.Error(t, err)
	require.Equal(t, 1, service.DeleteSecurityGroupCallCount())
}

// A group still referenced when the short window closes is handed back to the
// deployment engine rather than blocking a worker or failing the destroy.
//
// The same operation was measured releasing its interfaces in 18 seconds on one run and
// 977 seconds on the next, so no fixed timeout is defensible: long enough for the slow
// case wastes a worker in the common one. Retrying later is the only shape that fits,
// and the destroy is idempotent so re-running is free.
func TestGroupDeleteYieldsWhenTheDependencyPersists(t *testing.T) {
	withFastGroupDeletePolling(t)

	service := ec2mock.CreateEc2ServiceMock(
		// Never released, which is the case the yield exists for.
		ec2mock.WithDeleteSecurityGroupError(dependencyViolationErr()),
	)

	err := DeleteSecurityGroupWhenUnused(context.Background(), service, "sg-workload")
	require.Error(t, err)

	var retryable *provider.RetryableError
	require.True(
		t,
		provider.AsRetryableError(err, &retryable),
		"a group still held must come back as retryable, not as a failed destroy",
	)
	require.Contains(t, retryable.ChildError.Error(), "sg-workload")
}

// A detached Lambda interface is deleted rather than waited on.
//
// AWS reclaims these on its own schedule, well after the function is gone, and until
// they are gone the security group cannot be deleted, and until that is gone neither
// can the VPC. Measured against a real account, leaving reclamation to AWS adds many
// minutes to a teardown that is already waiting on the detach.
func TestReleasedLambdaInterfacesAreReaped(t *testing.T) {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeNetworkInterfacesOutputs([]*ec2.DescribeNetworkInterfacesOutput{
			{
				NetworkInterfaces: []types.NetworkInterface{
					{
						NetworkInterfaceId: aws.String("eni-lambda"),
						Description:        aws.String("AWS Lambda VPC ENI-orders-fn"),
					},
				},
			},
		}),
	)

	require.NoError(t, ReapReleasedLambdaENIs(context.Background(), service, "sg-workload"))
	require.Equal(t, []string{"eni-lambda"}, service.DeletedNetworkInterfaceIDs())
}

// Interfaces belonging to other services are left alone. A VPC endpoint's interfaces go
// when the endpoint does; deleting them here would break something still live.
func TestOtherServicesInterfacesAreLeftAlone(t *testing.T) {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeNetworkInterfacesOutputs([]*ec2.DescribeNetworkInterfacesOutput{
			{
				NetworkInterfaces: []types.NetworkInterface{
					{
						NetworkInterfaceId: aws.String("eni-endpoint"),
						Description:        aws.String("VPC Endpoint Interface vpce-123"),
					},
				},
			},
		}),
	)

	require.NoError(t, ReapReleasedLambdaENIs(context.Background(), service, "sg-workload"))
	require.Empty(t, service.DeletedNetworkInterfaceIDs())
}

// The describe is filtered to detached interfaces: one still in use belongs to a running
// workload and must not be removed.
func TestOnlyDetachedInterfacesAreConsidered(t *testing.T) {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeNetworkInterfacesOutputs(
			[]*ec2.DescribeNetworkInterfacesOutput{{}},
		),
	)

	require.NoError(t, ReapReleasedLambdaENIs(context.Background(), service, "sg-workload"))

	service.AssertCalledWith(
		testSuite(t),
		"DescribeNetworkInterfaces",
		0,
		plugintestutils.Any,
		func(arg any) bool {
			in, ok := arg.(*ec2.DescribeNetworkInterfacesInput)
			if !ok {
				return false
			}
			byName := map[string][]string{}
			for _, filter := range in.Filters {
				byName[aws.ToString(filter.Name)] = filter.Values
			}
			return len(byName["group-id"]) == 1 &&
				byName["group-id"][0] == "sg-workload" &&
				len(byName["status"]) == 1 &&
				byName["status"][0] == "available"
		},
	)
}
