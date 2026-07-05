//go:build unit

package lambdakms

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	kmsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/kms_mock"
	kmsservice "github.com/newstack-cloud/bluelink-provider-aws/services/kms/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionKeyLinkGrantsSuite struct {
	suite.Suite
}

const (
	fkGrantName = "bluelink-encryptFunction-dataKey"
	fkGrantID   = "grant-abc123"
)

func (s *FunctionKeyLinkGrantsSuite) actionsWithKMS(kmsSvc kmsservice.Service) (*functionKeyLinkActions, provider.Context) {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)
	actions := &functionKeyLinkActions{
		kmsServiceFactory: func(*aws.Config, provider.Context) kmsservice.Service { return kmsSvc },
		awsConfigStore:    testConfigStore(loader),
	}
	return actions, providerCtx
}

func decryptReconcile(manageGrant, isDestroy bool) keyGrantReconcile {
	return keyGrantReconcile{
		granteeRoleARN: fkRoleARN,
		keyID:          testKeyARN,
		grantName:      fkGrantName,
		accessLevel:    "decrypt",
		manageGrant:    manageGrant,
		isDestroy:      isDestroy,
	}
}

func (s *FunctionKeyLinkGrantsSuite) Test_creates_grant_when_managed_and_absent() {
	kmsSvc := kmsmock.CreateKMSServiceMock(
		kmsmock.WithListGrantsOutput(&kms.ListGrantsOutput{}),
		kmsmock.WithCreateGrantOutput(&kms.CreateGrantOutput{GrantId: aws.String(fkGrantID)}),
	)
	actions, providerCtx := s.actionsWithKMS(kmsSvc)

	err := actions.reconcileKeyGrant(context.Background(), providerCtx, decryptReconcile(true, false))
	s.Require().NoError(err)

	kmsSvc.AssertCalledWith(&s.Suite, "CreateGrant", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*kms.CreateGrantInput)
		if !ok {
			return false
		}
		return aws.ToString(in.KeyId) == testKeyARN &&
			aws.ToString(in.GranteePrincipal) == fkRoleARN &&
			aws.ToString(in.Name) == fkGrantName &&
			grantOperationsEqual(in.Operations, []kmstypes.GrantOperation{
				kmstypes.GrantOperationDecrypt, kmstypes.GrantOperationDescribeKey,
			})
	})
	kmsSvc.AssertNotCalled(&s.Suite, "RevokeGrant")
}

func (s *FunctionKeyLinkGrantsSuite) Test_no_op_when_managed_and_matching() {
	kmsSvc := kmsmock.CreateKMSServiceMock(
		kmsmock.WithListGrantsOutput(&kms.ListGrantsOutput{
			Grants: []kmstypes.GrantListEntry{
				{
					GrantId: aws.String(fkGrantID),
					Name:    aws.String(fkGrantName),
					Operations: []kmstypes.GrantOperation{
						kmstypes.GrantOperationDecrypt, kmstypes.GrantOperationDescribeKey,
					},
				},
			},
		}),
	)
	actions, providerCtx := s.actionsWithKMS(kmsSvc)

	err := actions.reconcileKeyGrant(context.Background(), providerCtx, decryptReconcile(true, false))
	s.Require().NoError(err)

	kmsSvc.AssertNotCalled(&s.Suite, "CreateGrant")
	kmsSvc.AssertNotCalled(&s.Suite, "RevokeGrant")
}

func (s *FunctionKeyLinkGrantsSuite) Test_revokes_and_recreates_when_operations_differ() {
	kmsSvc := kmsmock.CreateKMSServiceMock(
		kmsmock.WithListGrantsOutput(&kms.ListGrantsOutput{
			Grants: []kmstypes.GrantListEntry{
				{
					GrantId:    aws.String(fkGrantID),
					Name:       aws.String(fkGrantName),
					Operations: []kmstypes.GrantOperation{kmstypes.GrantOperationDecrypt},
				},
			},
		}),
		kmsmock.WithRevokeGrantOutput(&kms.RevokeGrantOutput{}),
		kmsmock.WithCreateGrantOutput(&kms.CreateGrantOutput{GrantId: aws.String(fkGrantID)}),
	)
	actions, providerCtx := s.actionsWithKMS(kmsSvc)

	err := actions.reconcileKeyGrant(context.Background(), providerCtx, decryptReconcile(true, false))
	s.Require().NoError(err)

	kmsSvc.AssertCalledWith(&s.Suite, "RevokeGrant", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*kms.RevokeGrantInput)
		return ok && aws.ToString(in.GrantId) == fkGrantID && aws.ToString(in.KeyId) == testKeyARN
	})
	kmsSvc.AssertCalledWith(&s.Suite, "CreateGrant", 0, plugintestutils.Any, func(arg any) bool {
		_, ok := arg.(*kms.CreateGrantInput)
		return ok
	})
}

func (s *FunctionKeyLinkGrantsSuite) Test_revokes_grant_on_destroy() {
	kmsSvc := kmsmock.CreateKMSServiceMock(
		kmsmock.WithListGrantsOutput(&kms.ListGrantsOutput{
			Grants: []kmstypes.GrantListEntry{
				{GrantId: aws.String(fkGrantID), Name: aws.String(fkGrantName)},
			},
		}),
		kmsmock.WithRevokeGrantOutput(&kms.RevokeGrantOutput{}),
	)
	actions, providerCtx := s.actionsWithKMS(kmsSvc)

	err := actions.reconcileKeyGrant(context.Background(), providerCtx, decryptReconcile(false, true))
	s.Require().NoError(err)

	kmsSvc.AssertCalledWith(&s.Suite, "RevokeGrant", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*kms.RevokeGrantInput)
		return ok && aws.ToString(in.GrantId) == fkGrantID
	})
	kmsSvc.AssertNotCalled(&s.Suite, "CreateGrant")
}

func (s *FunctionKeyLinkGrantsSuite) Test_revokes_grant_when_toggled_off() {
	kmsSvc := kmsmock.CreateKMSServiceMock(
		kmsmock.WithListGrantsOutput(&kms.ListGrantsOutput{
			Grants: []kmstypes.GrantListEntry{
				{GrantId: aws.String(fkGrantID), Name: aws.String(fkGrantName)},
			},
		}),
		kmsmock.WithRevokeGrantOutput(&kms.RevokeGrantOutput{}),
	)
	actions, providerCtx := s.actionsWithKMS(kmsSvc)

	// manageGrant false (toggled off), not a destroy: a previously-created grant is revoked.
	err := actions.reconcileKeyGrant(context.Background(), providerCtx, decryptReconcile(false, false))
	s.Require().NoError(err)

	kmsSvc.AssertCalledWith(&s.Suite, "RevokeGrant", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*kms.RevokeGrantInput)
		return ok && aws.ToString(in.GrantId) == fkGrantID
	})
	kmsSvc.AssertNotCalled(&s.Suite, "CreateGrant")
}

func (s *FunctionKeyLinkGrantsSuite) Test_no_action_when_unmanaged_and_absent() {
	kmsSvc := kmsmock.CreateKMSServiceMock(
		kmsmock.WithListGrantsOutput(&kms.ListGrantsOutput{}),
	)
	actions, providerCtx := s.actionsWithKMS(kmsSvc)

	err := actions.reconcileKeyGrant(context.Background(), providerCtx, decryptReconcile(false, false))
	s.Require().NoError(err)

	kmsSvc.AssertNotCalled(&s.Suite, "CreateGrant")
	kmsSvc.AssertNotCalled(&s.Suite, "RevokeGrant")
}

func (s *FunctionKeyLinkGrantsSuite) Test_grant_operations_for_access_level() {
	s.True(grantOperationsEqual(
		grantOperationsForAccessLevel("encryptDecrypt"),
		[]kmstypes.GrantOperation{
			kmstypes.GrantOperationDecrypt,
			kmstypes.GrantOperationDescribeKey,
			kmstypes.GrantOperationEncrypt,
			kmstypes.GrantOperationGenerateDataKey,
			kmstypes.GrantOperationGenerateDataKeyWithoutPlaintext,
		},
	))
	s.True(grantOperationsEqual(
		grantOperationsForAccessLevel("decrypt"),
		[]kmstypes.GrantOperation{kmstypes.GrantOperationDecrypt, kmstypes.GrantOperationDescribeKey},
	))
}

func (s *FunctionKeyLinkGrantsSuite) Test_grant_name_is_deterministic_and_scoped() {
	name := keyGrantName(
		&provider.ResourceInfo{ResourceName: "encryptFunction"},
		&provider.ResourceInfo{ResourceName: "dataKey"},
	)
	s.Equal(fkGrantName, name)
}

func TestFunctionKeyLinkGrantsSuite(t *testing.T) {
	suite.Run(t, new(FunctionKeyLinkGrantsSuite))
}
