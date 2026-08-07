//go:build integration

package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/newstack-cloud/bluelink-provider-aws/flex"
	"github.com/stretchr/testify/require"
)

// The shared inline policy the IAM access allocator
// writes link-granted statements into (mirrors linkutils.InlineAccessPolicyName).
const inlineAccessPolicyName = "bluelink-link-access"

// QueueURL builds the deterministic SQS queue URL for a queue name in the run's account/region.
func (h *Harness) QueueURL(queueName string) string {
	return fmt.Sprintf("https://sqs.%s.amazonaws.com/%s/%s", h.Region, h.Account.AccountID, queueName)
}

// QueueAttributes returns all attributes of a queue via the raw SQS API.
func (h *Harness) QueueAttributes(t *testing.T, queueName string) map[string]string {
	t.Helper()
	out, err := sqs.NewFromConfig(h.AWSConfig).GetQueueAttributes(h.Ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(h.QueueURL(queueName)),
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameAll},
	})
	require.NoError(t, err)
	return out.Attributes
}

// RoleInlinePolicy returns the decoded shared inline access policy of a role and
// whether it exists.
func (h *Harness) RoleInlinePolicy(t *testing.T, roleName string) (string, bool) {
	t.Helper()
	out, err := iam.NewFromConfig(h.AWSConfig).GetRolePolicy(h.Ctx, &iam.GetRolePolicyInput{
		RoleName:   aws.String(roleName),
		PolicyName: aws.String(inlineAccessPolicyName),
	})
	if err != nil {
		if _, ok := errors.AsType[*iamtypes.NoSuchEntityException](err); ok {
			return "", false
		}
		require.NoError(t, err)
	}
	decoded, err := url.QueryUnescape(aws.ToString(out.PolicyDocument))
	require.NoError(t, err)
	return decoded, true
}

// FunctionResourcePolicy returns the decoded resource-based policy of a Lambda
// function and whether one exists.
func (h *Harness) FunctionResourcePolicy(t *testing.T, functionName string) (string, bool) {
	t.Helper()
	out, err := lambda.NewFromConfig(h.AWSConfig).GetPolicy(h.Ctx, &lambda.GetPolicyInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		if _, ok := errors.AsType[*lambdatypes.ResourceNotFoundException](err); ok {
			return "", false
		}
		require.NoError(t, err)
	}
	return aws.ToString(out.Policy), true
}

// FunctionEnvValues returns the environment variable values of a Lambda function.
func (h *Harness) FunctionEnvValues(t *testing.T, functionName string) []string {
	t.Helper()
	out, err := lambda.NewFromConfig(h.AWSConfig).GetFunctionConfiguration(h.Ctx, &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
	})
	require.NoError(t, err)
	values := []string{}
	if out.Environment != nil {
		for _, v := range out.Environment.Variables {
			values = append(values, v)
		}
	}
	return values
}

// FunctionCodeSigningConfigArn returns the code signing config ARN attached to a
// Lambda function, or "" if none is configured.
func (h *Harness) FunctionCodeSigningConfigArn(t *testing.T, functionName string) string {
	t.Helper()
	out, err := lambda.NewFromConfig(h.AWSConfig).GetFunctionCodeSigningConfig(h.Ctx, &lambda.GetFunctionCodeSigningConfigInput{
		FunctionName: aws.String(functionName),
	})
	require.NoError(t, err)
	return aws.ToString(out.CodeSigningConfigArn)
}

// FunctionEventSourceMappings returns the event source mappings attached to a
// Lambda function via the raw Lambda API.
func (h *Harness) FunctionEventSourceMappings(t *testing.T, functionName string) []lambdatypes.EventSourceMappingConfiguration {
	t.Helper()
	out, err := lambda.NewFromConfig(h.AWSConfig).ListEventSourceMappings(h.Ctx, &lambda.ListEventSourceMappingsInput{
		FunctionName: aws.String(functionName),
	})
	require.NoError(t, err)
	return out.EventSourceMappings
}

// CompactJSON re-marshals a JSON document compactly so substring assertions are
// stable regardless of whitespace.
func CompactJSON(t *testing.T, raw string) string {
	t.Helper()
	var doc any
	require.NoError(t, json.Unmarshal([]byte(raw), &doc))
	compact, err := json.Marshal(doc)
	require.NoError(t, err)
	return string(compact)
}

// SecurityGroupRules returns the individual rules of a security group. Unlike the
// permissions on DescribeSecurityGroups these carry rule IDs and tags, which is
// what distinguishes a rule the flex VPC declared from one a link added.
func (h *Harness) SecurityGroupRules(
	t *testing.T,
	groupID string,
) []ec2types.SecurityGroupRule {
	t.Helper()
	out, err := ec2.NewFromConfig(h.AWSConfig).DescribeSecurityGroupRules(
		h.Ctx,
		&ec2.DescribeSecurityGroupRulesInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("group-id"), Values: []string{groupID}},
			},
		},
	)
	require.NoError(t, err)

	return out.SecurityGroupRules
}

// TagValue returns the value of a tag on an EC2 resource and whether it is set.
func TagValue(tags []ec2types.Tag, key string) (string, bool) {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value), true
		}
	}

	return "", false
}

// WorkloadSecurityGroups returns the per-workload security groups a placement link
// created in a flex VPC, keyed by the workload they belong to.
func (h *Harness) WorkloadSecurityGroups(
	t *testing.T,
	vpcName string,
) map[string]ec2types.SecurityGroup {
	t.Helper()
	out, err := ec2.NewFromConfig(h.AWSConfig).DescribeSecurityGroups(
		h.Ctx,
		&ec2.DescribeSecurityGroupsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("tag:" + flex.TagFlexVPCName),
					Values: []string{vpcName},
				},
				{
					Name:   aws.String("tag-key"),
					Values: []string{flex.TagFlexVPCWorkloadSecurityGroup},
				},
			},
		},
	)
	require.NoError(t, err)

	groups := map[string]ec2types.SecurityGroup{}
	for _, group := range out.SecurityGroups {
		if workload, ok := TagValue(group.Tags, flex.TagFlexVPCWorkloadSecurityGroup); ok {
			groups[workload] = group
		}
	}

	return groups
}

// NamedSecurityGroups returns the empty groups the flex VPC minted from its
// securityGroups list, keyed by the name they were declared under.
func (h *Harness) NamedSecurityGroups(
	t *testing.T,
	vpcName string,
) map[string]ec2types.SecurityGroup {
	t.Helper()
	out, err := ec2.NewFromConfig(h.AWSConfig).DescribeSecurityGroups(
		h.Ctx,
		&ec2.DescribeSecurityGroupsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("tag:" + flex.TagFlexVPCName),
					Values: []string{vpcName},
				},
				{
					Name:   aws.String("tag-key"),
					Values: []string{flex.TagFlexVPCSecurityGroupName},
				},
			},
		},
	)
	require.NoError(t, err)

	groups := map[string]ec2types.SecurityGroup{}
	for _, group := range out.SecurityGroups {
		if name, ok := TagValue(group.Tags, flex.TagFlexVPCSecurityGroupName); ok {
			groups[name] = group
		}
	}

	return groups
}

// VPCEndpointsForVPC returns every VPC endpoint in the given VPC.
func (h *Harness) VPCEndpointsForVPC(t *testing.T, vpcID string) []ec2types.VpcEndpoint {
	t.Helper()
	out, err := ec2.NewFromConfig(h.AWSConfig).DescribeVpcEndpoints(
		h.Ctx,
		&ec2.DescribeVpcEndpointsInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{vpcID}},
			},
		},
	)
	require.NoError(t, err)

	return out.VpcEndpoints
}

// SecurityGroupsInVPC returns every security group in the given VPC, including the
// default one AWS creates. Used after teardown to show what was left behind.
func (h *Harness) SecurityGroupsInVPC(t *testing.T, vpcID string) []ec2types.SecurityGroup {
	t.Helper()
	out, err := ec2.NewFromConfig(h.AWSConfig).DescribeSecurityGroups(
		h.Ctx,
		&ec2.DescribeSecurityGroupsInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{vpcID}},
			},
		},
	)
	require.NoError(t, err)

	return out.SecurityGroups
}

// VPCExists reports whether a VPC is still present, so a test can tell a clean
// teardown from one that stalled on a dependency it could not remove.
func (h *Harness) VPCExists(t *testing.T, vpcID string) bool {
	t.Helper()
	out, err := ec2.NewFromConfig(h.AWSConfig).DescribeVpcs(
		h.Ctx,
		&ec2.DescribeVpcsInput{
			Filters: []ec2types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{vpcID}},
			},
		},
	)
	if err != nil {
		return false
	}

	return len(out.Vpcs) > 0
}
