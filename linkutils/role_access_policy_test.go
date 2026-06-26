//go:build unit

package linkutils

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type RoleAccessPolicySuite struct {
	suite.Suite
}

const testRoleName = "process-orders-role"

func accessStatement(sid string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   []any{"dynamodb:GetItem"},
		"Resource": "arn:aws:dynamodb:us-west-2:123456789012:table/orders",
	}
}

// Asserts the PutRolePolicy input targets the inline slot
// and its document contains exactly the expected Sids.
func putRolePolicyDocMatcher(s *suite.Suite, wantSids ...string) func(any) bool {
	return func(arg any) bool {
		input, ok := arg.(*iam.PutRolePolicyInput)
		if !ok {
			return false
		}
		if aws.ToString(input.PolicyName) != inlineAccessPolicyName {
			return false
		}
		statements, err := parseStatements(aws.ToString(input.PolicyDocument))
		s.Require().NoError(err)
		if len(statements) != len(wantSids) {
			return false
		}
		for _, sid := range wantSids {
			if _, ok := statements[sid]; !ok {
				return false
			}
		}
		return true
	}
}

func (s *RoleAccessPolicySuite) Test_create_places_grant_in_new_inline_policy() {
	mock := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	result, err := ReconcileRoleAccessPolicy(context.Background(), mock, RoleAccessGrant{
		RoleName:  testRoleName,
		SID:       "DynamoDBAccessordersTable",
		Statement: accessStatement("DynamoDBAccessordersTable"),
	})
	s.Require().NoError(err)
	s.Equal(inlineAccessPolicyName, result.PlacedSlot)

	mock.AssertCalledWith(&s.Suite, "PutRolePolicy", 0,
		plugintestutils.Any,
		putRolePolicyDocMatcher(&s.Suite, "DynamoDBAccessordersTable"),
	)
	mock.AssertNotCalled(&s.Suite, "DeleteRolePolicy")
	mock.AssertNotCalled(&s.Suite, "CreatePolicy")
}

func (s *RoleAccessPolicySuite) Test_overflow_places_grant_in_new_managed_policy_and_returns_arn() {
	const managedARN = "arn:aws:iam::123456789012:policy/bluelink-link-access-1"
	mock := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithCreatePolicyOutput(&iam.CreatePolicyOutput{Policy: &iamtypes.Policy{Arn: aws.String(managedARN)}}),
		iammock.WithAttachRolePolicyOutput(&iam.AttachRolePolicyOutput{}),
	)

	// A tiny inline budget forces the grant to overflow into a new managed policy.
	result, err := ReconcileRoleAccessPolicy(context.Background(), mock, RoleAccessGrant{
		RoleName:  testRoleName,
		SID:       "DynamoDBAccessordersTable",
		Statement: accessStatement("DynamoDBAccessordersTable"),
		Limits:    AccessPolicyLimits{MaxInlineBytes: 10, MaxManagedBytes: 9000, MaxManagedSlots: 5},
	})
	s.Require().NoError(err)
	s.False(result.PlacedSlotInline)
	s.Equal(managedAccessPolicyPrefix+"1", result.PlacedSlot)
	s.Equal(managedARN, result.PlacedSlotARN)
	mock.AssertNotCalled(&s.Suite, "PutRolePolicy")
}

func (s *RoleAccessPolicySuite) Test_update_existing_grant_rewrites_inline_policy() {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"DynamoDBAccessordersTable","Effect":"Allow","Action":["dynamodb:GetItem"],"Resource":"arn:old"}]}`
	mock := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{inlineAccessPolicyName}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	result, err := ReconcileRoleAccessPolicy(context.Background(), mock, RoleAccessGrant{
		RoleName:  testRoleName,
		SID:       "DynamoDBAccessordersTable",
		Statement: accessStatement("DynamoDBAccessordersTable"),
	})
	s.Require().NoError(err)
	s.Equal(inlineAccessPolicyName, result.PlacedSlot)

	// Still a single statement after the in-place replace.
	mock.AssertCalledWith(&s.Suite, "PutRolePolicy", 0,
		plugintestutils.Any,
		putRolePolicyDocMatcher(&s.Suite, "DynamoDBAccessordersTable"),
	)
	mock.AssertNotCalled(&s.Suite, "DeleteRolePolicy")
}

func (s *RoleAccessPolicySuite) Test_remove_only_statement_deletes_inline_policy() {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"DynamoDBAccessordersTable","Effect":"Allow","Action":["dynamodb:GetItem"],"Resource":"arn:x"}]}`
	mock := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{inlineAccessPolicyName}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
	)

	result, err := ReconcileRoleAccessPolicy(context.Background(), mock, RoleAccessGrant{
		RoleName: testRoleName,
		SID:      "DynamoDBAccessordersTable",
		// nil statement -> removal
	})
	s.Require().NoError(err)
	s.Empty(result.PlacedSlot)

	mock.AssertCalled(&s.Suite, "DeleteRolePolicy")
	mock.AssertNotCalled(&s.Suite, "PutRolePolicy")
}

func (s *RoleAccessPolicySuite) Test_remove_one_of_many_rewrites_inline_policy() {
	existing := `{"Version":"2012-10-17","Statement":[` +
		`{"Sid":"DynamoDBAccessordersTable","Effect":"Allow","Action":["dynamodb:GetItem"],"Resource":"arn:x"},` +
		`{"Sid":"DynamoDBAccessotherTable","Effect":"Allow","Action":["dynamodb:GetItem"],"Resource":"arn:y"}]}`
	mock := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{inlineAccessPolicyName}}),
		iammock.WithGetRolePolicyOutput(&iam.GetRolePolicyOutput{PolicyDocument: aws.String(existing)}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	_, err := ReconcileRoleAccessPolicy(context.Background(), mock, RoleAccessGrant{
		RoleName: testRoleName,
		SID:      "DynamoDBAccessordersTable",
	})
	s.Require().NoError(err)

	// The sibling statement remains, so the policy is rewritten, not deleted.
	mock.AssertCalledWith(&s.Suite, "PutRolePolicy", 0,
		plugintestutils.Any,
		putRolePolicyDocMatcher(&s.Suite, "DynamoDBAccessotherTable"),
	)
	mock.AssertNotCalled(&s.Suite, "DeleteRolePolicy")
}

func (s *RoleAccessPolicySuite) Test_parse_statements_url_decodes() {
	// IAM returns documents URL-encoded; ensure discovery decodes before parsing.
	encoded := "%7B%22Version%22%3A%222012-10-17%22%2C%22Statement%22%3A%5B%7B%22Sid%22%3A%22S1%22%2C%22Effect%22%3A%22Allow%22%7D%5D%7D"
	statements, err := parseStatements(encoded)
	s.Require().NoError(err)
	s.Contains(statements, "S1")
}

func (s *RoleAccessPolicySuite) Test_budget_exhausted_error_names_role_and_is_detectable() {
	limits := AccessPolicyLimits{MaxInlineBytes: 120, MaxManagedBytes: 120, MaxManagedSlots: 0}
	mock := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	_, err := ReconcileRoleAccessPolicy(context.Background(), mock, RoleAccessGrant{
		RoleName:  testRoleName,
		SID:       "BigStatement",
		Statement: accessStatement("BigStatement"),
		Limits:    limits,
	})
	s.Require().Error(err)
	s.True(errors.Is(err, ErrAccessPolicyBudgetExhausted), "should be detectable via errors.Is")
	s.Contains(err.Error(), testRoleName, "error should name the role")
}

func TestRoleAccessPolicySuite(t *testing.T) {
	suite.Run(t, new(RoleAccessPolicySuite))
}
