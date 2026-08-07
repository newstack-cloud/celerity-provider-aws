//go:build unit

package linkutils

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MockCalls asserts through a testify suite; these tests are plain functions, so this
// adapts one to the running test.
func suiteFor(t *testing.T) *suite.Suite {
	t.Helper()

	testSuite := &suite.Suite{}
	testSuite.SetT(t)

	return testSuite
}

func existingRulesOutput(
	count int,
	isEgress bool,
	port int32,
) *ec2.DescribeSecurityGroupRulesOutput {
	rules := make([]ec2types.SecurityGroupRule, count)
	for i := range rules {
		rules[i] = ec2types.SecurityGroupRule{
			SecurityGroupRuleId: aws.String(fmt.Sprintf("sgr-%d", i)),
			IsEgress:            aws.Bool(isEgress),
			ToPort:              aws.Int32(port),
			ReferencedGroupInfo: &ec2types.ReferencedSecurityGroup{
				GroupId: aws.String(fmt.Sprintf("sg-%d", i)),
			},
		}
	}

	return &ec2.DescribeSecurityGroupRulesOutput{SecurityGroupRules: rules}
}

func TestAuthorizeRuleWithinBudgetAddsAnAbsentRule(t *testing.T) {
	ec2Service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupRulesOutput(existingRulesOutput(2, true, 443)),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)

	err := AuthorizeRuleWithinBudget(
		context.Background(),
		ec2Service,
		"sg-caller",
		RuleEgress,
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-endpoint", Port: 443},
		"link-1",
	)

	require.NoError(t, err)
	ec2Service.AssertCalled(suiteFor(t), "AuthorizeSecurityGroupEgress")
}

// The group has to be read to count its rules, and that same read settles whether the
// rule is already there. Authorising it again would rely on the duplicate error, which
// tells the allocator nothing about how full the group is.
func TestAuthorizeRuleWithinBudgetSkipsARuleThatIsAlreadyPresent(t *testing.T) {
	existing := existingRulesOutput(2, true, 443)
	existing.SecurityGroupRules = append(
		existing.SecurityGroupRules,
		ec2types.SecurityGroupRule{
			SecurityGroupRuleId: aws.String("sgr-existing"),
			IsEgress:            aws.Bool(true),
			ToPort:              aws.Int32(443),
			ReferencedGroupInfo: &ec2types.ReferencedSecurityGroup{
				GroupId: aws.String("sg-endpoint"),
			},
		},
	)

	ec2Service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupRulesOutput(existing),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)

	err := AuthorizeRuleWithinBudget(
		context.Background(),
		ec2Service,
		"sg-caller",
		RuleEgress,
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-endpoint", Port: 443},
		"link-1",
	)

	require.NoError(t, err)
	ec2Service.AssertNotCalled(suiteFor(t), "AuthorizeSecurityGroupEgress")
}

func TestAuthorizeRuleWithinBudgetRefusesToOverfillTheGroup(t *testing.T) {
	limits := DefaultSecurityGroupRuleLimits()
	ec2Service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupRulesOutput(
			existingRulesOutput(limits.MaxRulesPerDirection, true, 443),
		),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)

	err := AuthorizeRuleWithinBudget(
		context.Background(),
		ec2Service,
		"sg-caller",
		RuleEgress,
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-endpoint", Port: 443},
		"link-1",
	)

	require.ErrorIs(t, err, ErrSecurityGroupRuleBudgetExhausted)
	ec2Service.AssertNotCalled(suiteFor(t), "AuthorizeSecurityGroupEgress")
}

// Egress rules must not be counted against the ingress budget or the other way round.
// AWS budgets the two directions separately, so folding them together would report a
// group as full at half its actual capacity.
func TestAuthorizeRuleWithinBudgetCountsEachDirectionSeparately(t *testing.T) {
	limits := DefaultSecurityGroupRuleLimits()
	// A group full of egress rules has an entirely empty ingress side.
	ec2Service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupRulesOutput(
			existingRulesOutput(limits.MaxRulesPerDirection, true, 443),
		),
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
	)

	err := AuthorizeRuleWithinBudget(
		context.Background(),
		ec2Service,
		"sg-endpoint",
		RuleIngress,
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-caller", Port: 443},
		"link-1",
	)

	require.NoError(t, err)
	ec2Service.AssertCalled(suiteFor(t), "AuthorizeSecurityGroupIngress")
}

// A rule permitting a CIDR carries no referenced group. It still occupies the group's
// budget, and reading it must not panic on the absent group reference.
func TestAuthorizeRuleWithinBudgetCountsCidrRulesWithoutAGroupReference(t *testing.T) {
	ec2Service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupRulesOutput(&ec2.DescribeSecurityGroupRulesOutput{
			SecurityGroupRules: []ec2types.SecurityGroupRule{
				{
					SecurityGroupRuleId: aws.String("sgr-cidr"),
					IsEgress:            aws.Bool(true),
					ToPort:              aws.Int32(443),
					CidrIpv4:            aws.String("0.0.0.0/0"),
				},
			},
		}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)

	err := AuthorizeRuleWithinBudget(
		context.Background(),
		ec2Service,
		"sg-caller",
		RuleEgress,
		SecurityGroupRuleRef{PairedSecurityGroupID: "sg-endpoint", Port: 443},
		"link-1",
	)

	require.NoError(t, err)
}
