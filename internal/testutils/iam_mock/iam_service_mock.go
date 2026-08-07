package iammock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type iamServiceMock struct {
	plugintestutils.MockCalls

	// Role-related mock fields
	createRoleOutput *iam.CreateRoleOutput
	createRoleError  error
	getRoleOutput    *iam.GetRoleOutput
	getRoleError     error
	updateRoleOutput *iam.UpdateRoleOutput
	updateRoleError  error
	deleteRoleOutput *iam.DeleteRoleOutput
	deleteRoleError  error

	// Role policy document-related mock fields
	updateAssumeRolePolicyOutput *iam.UpdateAssumeRolePolicyOutput
	updateAssumeRolePolicyError  error

	// Role policy attachment-related mock fields
	attachRolePolicyOutput         *iam.AttachRolePolicyOutput
	attachRolePolicyError          error
	detachRolePolicyOutput         *iam.DetachRolePolicyOutput
	detachRolePolicyError          error
	listAttachedRolePoliciesOutput *iam.ListAttachedRolePoliciesOutput
	listAttachedRolePoliciesError  error

	// Inline policy-related mock fields
	putRolePolicyOutput    *iam.PutRolePolicyOutput
	putRolePolicyError     error
	deleteRolePolicyOutput *iam.DeleteRolePolicyOutput
	deleteRolePolicyError  error
	listRolePoliciesOutput *iam.ListRolePoliciesOutput
	listRolePoliciesError  error
	getRolePolicyOutput    *iam.GetRolePolicyOutput
	getRolePolicyError     error

	// Tag-related mock fields
	tagRoleOutput      *iam.TagRoleOutput
	tagRoleError       error
	untagRoleOutput    *iam.UntagRoleOutput
	untagRoleError     error
	listRoleTagsOutput *iam.ListRoleTagsOutput
	listRoleTagsError  error

	// Permissions boundary-related mock fields
	putRolePermissionsBoundaryOutput    *iam.PutRolePermissionsBoundaryOutput
	putRolePermissionsBoundaryError     error
	deleteRolePermissionsBoundaryOutput *iam.DeleteRolePermissionsBoundaryOutput
	deleteRolePermissionsBoundaryError  error

	// User-related mock fields
	createUserOutput *iam.CreateUserOutput
	createUserError  error
	getUserOutput    *iam.GetUserOutput
	getUserError     error
	updateUserOutput *iam.UpdateUserOutput
	updateUserError  error
	deleteUserOutput *iam.DeleteUserOutput
	deleteUserError  error

	// User policy attachment-related mock fields
	attachUserPolicyOutput         *iam.AttachUserPolicyOutput
	attachUserPolicyError          error
	detachUserPolicyOutput         *iam.DetachUserPolicyOutput
	detachUserPolicyError          error
	listAttachedUserPoliciesOutput *iam.ListAttachedUserPoliciesOutput
	listAttachedUserPoliciesError  error

	// User policy operations
	putUserPolicyOutput    *iam.PutUserPolicyOutput
	putUserPolicyError     error
	deleteUserPolicyOutput *iam.DeleteUserPolicyOutput
	deleteUserPolicyError  error
	listUserPoliciesOutput *iam.ListUserPoliciesOutput
	listUserPoliciesError  error

	// User tag operations
	tagUserOutput      *iam.TagUserOutput
	tagUserError       error
	untagUserOutput    *iam.UntagUserOutput
	untagUserError     error
	listUserTagsOutput *iam.ListUserTagsOutput
	listUserTagsError  error

	// User permissions boundary operations
	putUserPermissionsBoundaryOutput    *iam.PutUserPermissionsBoundaryOutput
	putUserPermissionsBoundaryError     error
	deleteUserPermissionsBoundaryOutput *iam.DeleteUserPermissionsBoundaryOutput
	deleteUserPermissionsBoundaryError  error

	// User policy operations
	getUserPolicyOutput *iam.GetUserPolicyOutput
	getUserPolicyError  error

	// Group operations
	addUserToGroupOutput      *iam.AddUserToGroupOutput
	addUserToGroupError       error
	removeUserFromGroupOutput *iam.RemoveUserFromGroupOutput
	removeUserFromGroupError  error
	listGroupsForUserOutput   *iam.ListGroupsForUserOutput
	listGroupsForUserError    error

	// Login profile operations
	createLoginProfileOutput *iam.CreateLoginProfileOutput
	createLoginProfileError  error
	getLoginProfileOutput    *iam.GetLoginProfileOutput
	getLoginProfileError     error
	updateLoginProfileOutput *iam.UpdateLoginProfileOutput
	updateLoginProfileError  error
	deleteLoginProfileOutput *iam.DeleteLoginProfileOutput
	deleteLoginProfileError  error

	// Group-related mock fields
	createGroupOutput *iam.CreateGroupOutput
	createGroupError  error
	getGroupOutput    *iam.GetGroupOutput
	getGroupError     error
	updateGroupOutput *iam.UpdateGroupOutput
	updateGroupError  error
	deleteGroupOutput *iam.DeleteGroupOutput
	deleteGroupError  error

	// Group policy attachment-related mock fields
	attachGroupPolicyOutput         *iam.AttachGroupPolicyOutput
	attachGroupPolicyError          error
	detachGroupPolicyOutput         *iam.DetachGroupPolicyOutput
	detachGroupPolicyError          error
	listAttachedGroupPoliciesOutput *iam.ListAttachedGroupPoliciesOutput
	listAttachedGroupPoliciesError  error

	// Group inline policy-related mock fields
	putGroupPolicyOutput    *iam.PutGroupPolicyOutput
	putGroupPolicyError     error
	deleteGroupPolicyOutput *iam.DeleteGroupPolicyOutput
	deleteGroupPolicyError  error
	listGroupPoliciesOutput *iam.ListGroupPoliciesOutput
	listGroupPoliciesError  error
	getGroupPolicyOutput    *iam.GetGroupPolicyOutput
	getGroupPolicyError     error

	// Access key-related mock fields
	createAccessKeyOutput *iam.CreateAccessKeyOutput
	createAccessKeyError  error
	updateAccessKeyOutput *iam.UpdateAccessKeyOutput
	updateAccessKeyError  error
	deleteAccessKeyOutput *iam.DeleteAccessKeyOutput
	deleteAccessKeyError  error
	listAccessKeysOutput  *iam.ListAccessKeysOutput
	listAccessKeysError   error

	// Instance profile-related mock fields
	createInstanceProfileOutput         *iam.CreateInstanceProfileOutput
	createInstanceProfileError          error
	getInstanceProfileOutput            *iam.GetInstanceProfileOutput
	getInstanceProfileError             error
	deleteInstanceProfileOutput         *iam.DeleteInstanceProfileOutput
	deleteInstanceProfileError          error
	addRoleToInstanceProfileOutput      *iam.AddRoleToInstanceProfileOutput
	addRoleToInstanceProfileError       error
	removeRoleFromInstanceProfileOutput *iam.RemoveRoleFromInstanceProfileOutput
	removeRoleFromInstanceProfileError  error

	// Managed policy-related mock fields
	createPolicyOutput *iam.CreatePolicyOutput
	createPolicyError  error
	getPolicyOutput    *iam.GetPolicyOutput
	getPolicyError     error
	deletePolicyOutput *iam.DeletePolicyOutput
	deletePolicyError  error

	// Policy version-related mock fields
	createPolicyVersionOutput *iam.CreatePolicyVersionOutput
	createPolicyVersionError  error
	deletePolicyVersionOutput *iam.DeletePolicyVersionOutput
	deletePolicyVersionError  error
	listPolicyVersionsOutput  *iam.ListPolicyVersionsOutput
	listPolicyVersionsError   error
	getPolicyVersionOutput    *iam.GetPolicyVersionOutput
	getPolicyVersionError     error

	// Policy tag-related mock fields
	tagPolicyOutput      *iam.TagPolicyOutput
	tagPolicyError       error
	untagPolicyOutput    *iam.UntagPolicyOutput
	untagPolicyError     error
	listPolicyTagsOutput *iam.ListPolicyTagsOutput
	listPolicyTagsError  error

	// OIDC provider-related mock fields
	createOpenIDConnectProviderOutput             *iam.CreateOpenIDConnectProviderOutput
	createOpenIDConnectProviderError              error
	getOpenIDConnectProviderOutput                *iam.GetOpenIDConnectProviderOutput
	getOpenIDConnectProviderError                 error
	addClientIDToOpenIDConnectProviderOutput      *iam.AddClientIDToOpenIDConnectProviderOutput
	addClientIDToOpenIDConnectProviderError       error
	removeClientIDFromOpenIDConnectProviderOutput *iam.RemoveClientIDFromOpenIDConnectProviderOutput
	removeClientIDFromOpenIDConnectProviderError  error
	updateOpenIDConnectProviderThumbprintOutput   *iam.UpdateOpenIDConnectProviderThumbprintOutput
	updateOpenIDConnectProviderThumbprintError    error
	deleteOpenIDConnectProviderOutput             *iam.DeleteOpenIDConnectProviderOutput
	deleteOpenIDConnectProviderError              error
	tagOpenIDConnectProviderOutput                *iam.TagOpenIDConnectProviderOutput
	tagOpenIDConnectProviderError                 error
	untagOpenIDConnectProviderOutput              *iam.UntagOpenIDConnectProviderOutput
	untagOpenIDConnectProviderError               error
	listOpenIDConnectProviderTagsOutput           *iam.ListOpenIDConnectProviderTagsOutput
	listOpenIDConnectProviderTagsError            error

	// SAML Provider fields
	createSAMLProviderOutput   *iam.CreateSAMLProviderOutput
	createSAMLProviderError    error
	getSAMLProviderOutput      *iam.GetSAMLProviderOutput
	getSAMLProviderError       error
	updateSAMLProviderOutput   *iam.UpdateSAMLProviderOutput
	updateSAMLProviderError    error
	deleteSAMLProviderOutput   *iam.DeleteSAMLProviderOutput
	deleteSAMLProviderError    error
	tagSAMLProviderOutput      *iam.TagSAMLProviderOutput
	tagSAMLProviderError       error
	untagSAMLProviderOutput    *iam.UntagSAMLProviderOutput
	untagSAMLProviderError     error
	listSAMLProviderTagsOutput *iam.ListSAMLProviderTagsOutput
	listSAMLProviderTagsError  error

	// Server certificate fields
	getServerCertificateOutput      *iam.GetServerCertificateOutput
	getServerCertificateError       error
	listServerCertificateTagsOutput *iam.ListServerCertificateTagsOutput
	listServerCertificateTagsError  error
	deleteServerCertificateOutput   *iam.DeleteServerCertificateOutput
	deleteServerCertificateError    error
	updateServerCertificateOutput   *iam.UpdateServerCertificateOutput
	updateServerCertificateError    error
	tagServerCertificateOutput      *iam.TagServerCertificateOutput
	tagServerCertificateError       error
	untagServerCertificateOutput    *iam.UntagServerCertificateOutput
	untagServerCertificateError     error
	uploadServerCertificateOutput   *iam.UploadServerCertificateOutput
	uploadServerCertificateError    error
}

type iamServiceMockOption func(*iamServiceMock)

// Mock is an exported alias for the mock type, so tests in other packages can hold a
// reference to a configured mock and assert on the calls it recorded.
type Mock = iamServiceMock

func CreateIamServiceMockFactory(
	opts ...iamServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
	mock := CreateIamServiceMock(opts...)
	return func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
		return mock
	}
}

func CreateIamServiceMock(
	opts ...iamServiceMockOption,
) *iamServiceMock {
	mock := &iamServiceMock{}

	for _, opt := range opts {
		opt(mock)
	}

	return mock
}

// NewMockService creates a new mock Service for testing.
func NewMockService(t any) *iamServiceMock {
	return CreateIamServiceMock()
}

// Mock configuration options for Role operations

func WithCreateRoleOutput(output *iam.CreateRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createRoleOutput = output
	}
}

func WithCreateRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createRoleError = err
	}
}

func WithGetRoleOutput(output *iam.GetRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getRoleOutput = output
	}
}

func WithGetRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getRoleError = err
	}
}

func WithUpdateRoleOutput(output *iam.UpdateRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateRoleOutput = output
	}
}

func WithUpdateRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateRoleError = err
	}
}

func WithDeleteRoleOutput(output *iam.DeleteRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRoleOutput = output
	}
}

func WithDeleteRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRoleError = err
	}
}

// Mock configuration options for Role Policy Document operations

func WithUpdateAssumeRolePolicyOutput(output *iam.UpdateAssumeRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateAssumeRolePolicyOutput = output
	}
}

func WithUpdateAssumeRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateAssumeRolePolicyError = err
	}
}

// Mock configuration options for Role Policy Attachment operations

func WithAttachRolePolicyOutput(output *iam.AttachRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachRolePolicyOutput = output
	}
}

func WithAttachRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachRolePolicyError = err
	}
}

func WithDetachRolePolicyOutput(output *iam.DetachRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachRolePolicyOutput = output
	}
}

func WithDetachRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachRolePolicyError = err
	}
}

func WithListAttachedRolePoliciesOutput(output *iam.ListAttachedRolePoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedRolePoliciesOutput = output
	}
}

func WithListAttachedRolePoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedRolePoliciesError = err
	}
}

// Mock configuration options for Inline Policy operations

func WithPutRolePolicyOutput(output *iam.PutRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putRolePolicyOutput = output
	}
}

func WithPutRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putRolePolicyError = err
	}
}

func WithDeleteRolePolicyOutput(output *iam.DeleteRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRolePolicyOutput = output
	}
}

func WithDeleteRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRolePolicyError = err
	}
}

func WithListRolePoliciesOutput(output *iam.ListRolePoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listRolePoliciesOutput = output
	}
}

func WithListRolePoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listRolePoliciesError = err
	}
}

func WithGetRolePolicyOutput(output *iam.GetRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getRolePolicyOutput = output
	}
}

func WithGetRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getRolePolicyError = err
	}
}

// Mock configuration options for Tag operations

func WithTagRoleOutput(output *iam.TagRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagRoleOutput = output
	}
}

func WithTagRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagRoleError = err
	}
}

func WithUntagRoleOutput(output *iam.UntagRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagRoleOutput = output
	}
}

func WithUntagRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagRoleError = err
	}
}

func WithListRoleTagsOutput(output *iam.ListRoleTagsOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listRoleTagsOutput = output
	}
}

func WithListRoleTagsError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listRoleTagsError = err
	}
}

// Mock configuration options for Permissions Boundary operations

func WithPutRolePermissionsBoundaryOutput(output *iam.PutRolePermissionsBoundaryOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putRolePermissionsBoundaryOutput = output
	}
}

func WithPutRolePermissionsBoundaryError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putRolePermissionsBoundaryError = err
	}
}

func WithDeleteRolePermissionsBoundaryOutput(output *iam.DeleteRolePermissionsBoundaryOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRolePermissionsBoundaryOutput = output
	}
}

func WithDeleteRolePermissionsBoundaryError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRolePermissionsBoundaryError = err
	}
}

// Mock configuration options for User operations

func WithCreateUserOutput(output *iam.CreateUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createUserOutput = output
	}
}

func WithCreateUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createUserError = err
	}
}

func WithGetUserOutput(output *iam.GetUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getUserOutput = output
	}
}

func WithGetUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getUserError = err
	}
}

func WithUpdateUserOutput(output *iam.UpdateUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateUserOutput = output
	}
}

func WithUpdateUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateUserError = err
	}
}

func WithDeleteUserOutput(output *iam.DeleteUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserOutput = output
	}
}

func WithDeleteUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserError = err
	}
}

// Mock configuration options for User Policy Attachment operations

func WithAttachUserPolicyOutput(output *iam.AttachUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachUserPolicyOutput = output
	}
}

func WithAttachUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachUserPolicyError = err
	}
}

func WithDetachUserPolicyOutput(output *iam.DetachUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachUserPolicyOutput = output
	}
}

func WithDetachUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachUserPolicyError = err
	}
}

func WithListAttachedUserPoliciesOutput(output *iam.ListAttachedUserPoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedUserPoliciesOutput = output
	}
}

func WithListAttachedUserPoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedUserPoliciesError = err
	}
}

// Mock configuration options for User Policy operations

func WithPutUserPolicyOutput(output *iam.PutUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putUserPolicyOutput = output
	}
}

func WithPutUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putUserPolicyError = err
	}
}

func WithDeleteUserPolicyOutput(output *iam.DeleteUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserPolicyOutput = output
	}
}

func WithDeleteUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserPolicyError = err
	}
}

func WithListUserPoliciesOutput(output *iam.ListUserPoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listUserPoliciesOutput = output
	}
}

func WithListUserPoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listUserPoliciesError = err
	}
}

// Mock configuration options for User Tag operations

func WithTagUserOutput(output *iam.TagUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagUserOutput = output
	}
}

func WithTagUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagUserError = err
	}
}

func WithUntagUserOutput(output *iam.UntagUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagUserOutput = output
	}
}

func WithUntagUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagUserError = err
	}
}

func WithListUserTagsOutput(output *iam.ListUserTagsOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listUserTagsOutput = output
	}
}

func WithListUserTagsError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listUserTagsError = err
	}
}

// Mock configuration options for User Permissions Boundary operations

func WithPutUserPermissionsBoundaryOutput(output *iam.PutUserPermissionsBoundaryOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putUserPermissionsBoundaryOutput = output
	}
}

func WithPutUserPermissionsBoundaryError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putUserPermissionsBoundaryError = err
	}
}

func WithDeleteUserPermissionsBoundaryOutput(output *iam.DeleteUserPermissionsBoundaryOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserPermissionsBoundaryOutput = output
	}
}

func WithDeleteUserPermissionsBoundaryError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserPermissionsBoundaryError = err
	}
}

// Mock configuration options for User Policy operations

func WithGetUserPolicyOutput(output *iam.GetUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getUserPolicyOutput = output
	}
}

func WithGetUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getUserPolicyError = err
	}
}

// Mock configuration options for Group operations

func WithAddUserToGroupOutput(output *iam.AddUserToGroupOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.addUserToGroupOutput = output
	}
}

func WithAddUserToGroupError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.addUserToGroupError = err
	}
}

func WithRemoveUserFromGroupOutput(output *iam.RemoveUserFromGroupOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.removeUserFromGroupOutput = output
	}
}

func WithRemoveUserFromGroupError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.removeUserFromGroupError = err
	}
}

func WithListGroupsForUserOutput(output *iam.ListGroupsForUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listGroupsForUserOutput = output
	}
}

func WithListGroupsForUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listGroupsForUserError = err
	}
}

// Mock configuration options for Login Profile operations

func WithCreateLoginProfileOutput(output *iam.CreateLoginProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createLoginProfileOutput = output
	}
}

func WithCreateLoginProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createLoginProfileError = err
	}
}

func WithGetLoginProfileOutput(output *iam.GetLoginProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getLoginProfileOutput = output
	}
}

func WithGetLoginProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getLoginProfileError = err
	}
}

func WithUpdateLoginProfileOutput(output *iam.UpdateLoginProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateLoginProfileOutput = output
	}
}

func WithUpdateLoginProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateLoginProfileError = err
	}
}

func WithDeleteLoginProfileOutput(output *iam.DeleteLoginProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteLoginProfileOutput = output
	}
}

func WithDeleteLoginProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteLoginProfileError = err
	}
}

// Mock configuration options for Group operations

func WithCreateGroupOutput(output *iam.CreateGroupOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createGroupOutput = output
	}
}

func WithCreateGroupError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createGroupError = err
	}
}

func WithGetGroupOutput(output *iam.GetGroupOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getGroupOutput = output
	}
}

func WithGetGroupError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getGroupError = err
	}
}

func WithUpdateGroupOutput(output *iam.UpdateGroupOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateGroupOutput = output
	}
}

func WithUpdateGroupError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateGroupError = err
	}
}

func WithDeleteGroupOutput(output *iam.DeleteGroupOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteGroupOutput = output
	}
}

func WithDeleteGroupError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteGroupError = err
	}
}

func WithAttachGroupPolicyOutput(output *iam.AttachGroupPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachGroupPolicyOutput = output
	}
}

func WithAttachGroupPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachGroupPolicyError = err
	}
}

func WithDetachGroupPolicyOutput(output *iam.DetachGroupPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachGroupPolicyOutput = output
	}
}

func WithDetachGroupPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachGroupPolicyError = err
	}
}

func WithListAttachedGroupPoliciesOutput(output *iam.ListAttachedGroupPoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedGroupPoliciesOutput = output
	}
}

func WithListAttachedGroupPoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedGroupPoliciesError = err
	}
}

func WithPutGroupPolicyOutput(output *iam.PutGroupPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putGroupPolicyOutput = output
	}
}

func WithPutGroupPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putGroupPolicyError = err
	}
}

func WithDeleteGroupPolicyOutput(output *iam.DeleteGroupPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteGroupPolicyOutput = output
	}
}

func WithDeleteGroupPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteGroupPolicyError = err
	}
}

func WithListGroupPoliciesOutput(output *iam.ListGroupPoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listGroupPoliciesOutput = output
	}
}

func WithListGroupPoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listGroupPoliciesError = err
	}
}

func WithGetGroupPolicyOutput(output *iam.GetGroupPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getGroupPolicyOutput = output
	}
}

func WithGetGroupPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getGroupPolicyError = err
	}
}

// Access key mock options.
func WithCreateAccessKeyOutput(output *iam.CreateAccessKeyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createAccessKeyOutput = output
	}
}

func WithCreateAccessKeyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createAccessKeyError = err
	}
}

func WithUpdateAccessKeyOutput(output *iam.UpdateAccessKeyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateAccessKeyOutput = output
	}
}

func WithUpdateAccessKeyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateAccessKeyError = err
	}
}

func WithDeleteAccessKeyOutput(output *iam.DeleteAccessKeyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteAccessKeyOutput = output
	}
}

func WithDeleteAccessKeyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteAccessKeyError = err
	}
}

func WithListAccessKeysOutput(output *iam.ListAccessKeysOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAccessKeysOutput = output
	}
}

func WithListAccessKeysError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAccessKeysError = err
	}
}

// Instance profile mock configuration options.
func WithCreateInstanceProfileOutput(output *iam.CreateInstanceProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createInstanceProfileOutput = output
	}
}

func WithCreateInstanceProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createInstanceProfileError = err
	}
}

func WithGetInstanceProfileOutput(output *iam.GetInstanceProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getInstanceProfileOutput = output
	}
}

func WithGetInstanceProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getInstanceProfileError = err
	}
}

func WithDeleteInstanceProfileOutput(output *iam.DeleteInstanceProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteInstanceProfileOutput = output
	}
}

func WithDeleteInstanceProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteInstanceProfileError = err
	}
}

func WithAddRoleToInstanceProfileOutput(output *iam.AddRoleToInstanceProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.addRoleToInstanceProfileOutput = output
	}
}

func WithAddRoleToInstanceProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.addRoleToInstanceProfileError = err
	}
}

func WithRemoveRoleFromInstanceProfileOutput(output *iam.RemoveRoleFromInstanceProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.removeRoleFromInstanceProfileOutput = output
	}
}

func WithRemoveRoleFromInstanceProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.removeRoleFromInstanceProfileError = err
	}
}

// Managed policy mock options.
func WithCreatePolicyOutput(output *iam.CreatePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createPolicyOutput = output
	}
}

func WithCreatePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createPolicyError = err
	}
}

func WithGetPolicyOutput(output *iam.GetPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getPolicyOutput = output
	}
}

func WithGetPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getPolicyError = err
	}
}

func WithDeletePolicyOutput(output *iam.DeletePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deletePolicyOutput = output
	}
}

func WithDeletePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deletePolicyError = err
	}
}

func WithCreatePolicyVersionOutput(output *iam.CreatePolicyVersionOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createPolicyVersionOutput = output
	}
}

func WithCreatePolicyVersionError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createPolicyVersionError = err
	}
}

func WithDeletePolicyVersionOutput(output *iam.DeletePolicyVersionOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deletePolicyVersionOutput = output
	}
}

func WithDeletePolicyVersionError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deletePolicyVersionError = err
	}
}

func WithListPolicyVersionsOutput(output *iam.ListPolicyVersionsOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listPolicyVersionsOutput = output
	}
}

func WithListPolicyVersionsError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listPolicyVersionsError = err
	}
}

func WithGetPolicyVersionOutput(output *iam.GetPolicyVersionOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getPolicyVersionOutput = output
	}
}

func WithGetPolicyVersionError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getPolicyVersionError = err
	}
}

func WithTagPolicyOutput(output *iam.TagPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagPolicyOutput = output
	}
}

func WithTagPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagPolicyError = err
	}
}

func WithUntagPolicyOutput(output *iam.UntagPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagPolicyOutput = output
	}
}

func WithUntagPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagPolicyError = err
	}
}

func WithListPolicyTagsOutput(output *iam.ListPolicyTagsOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listPolicyTagsOutput = output
	}
}

func WithListPolicyTagsError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listPolicyTagsError = err
	}
}

// OIDC provider mock options.
func WithCreateOpenIDConnectProviderOutput(output *iam.CreateOpenIDConnectProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createOpenIDConnectProviderOutput = output
	}
}

func WithCreateOpenIDConnectProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createOpenIDConnectProviderError = err
	}
}

func WithGetOpenIDConnectProviderOutput(output *iam.GetOpenIDConnectProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getOpenIDConnectProviderOutput = output
	}
}

func WithGetOpenIDConnectProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getOpenIDConnectProviderError = err
	}
}

func WithAddClientIDToOpenIDConnectProviderOutput(output *iam.AddClientIDToOpenIDConnectProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.addClientIDToOpenIDConnectProviderOutput = output
	}
}

func WithAddClientIDToOpenIDConnectProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.addClientIDToOpenIDConnectProviderError = err
	}
}

func WithRemoveClientIDFromOpenIDConnectProviderOutput(output *iam.RemoveClientIDFromOpenIDConnectProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.removeClientIDFromOpenIDConnectProviderOutput = output
	}
}

func WithRemoveClientIDFromOpenIDConnectProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.removeClientIDFromOpenIDConnectProviderError = err
	}
}

func WithUpdateOpenIDConnectProviderThumbprintOutput(output *iam.UpdateOpenIDConnectProviderThumbprintOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateOpenIDConnectProviderThumbprintOutput = output
	}
}

func WithUpdateOpenIDConnectProviderThumbprintError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateOpenIDConnectProviderThumbprintError = err
	}
}

func WithDeleteOpenIDConnectProviderOutput(output *iam.DeleteOpenIDConnectProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteOpenIDConnectProviderOutput = output
	}
}

func WithDeleteOpenIDConnectProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteOpenIDConnectProviderError = err
	}
}

func WithTagOpenIDConnectProviderOutput(output *iam.TagOpenIDConnectProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagOpenIDConnectProviderOutput = output
	}
}

func WithTagOpenIDConnectProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagOpenIDConnectProviderError = err
	}
}

func WithUntagOpenIDConnectProviderOutput(output *iam.UntagOpenIDConnectProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagOpenIDConnectProviderOutput = output
	}
}

func WithUntagOpenIDConnectProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagOpenIDConnectProviderError = err
	}
}

func WithListOpenIDConnectProviderTagsOutput(output *iam.ListOpenIDConnectProviderTagsOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listOpenIDConnectProviderTagsOutput = output
	}
}

func WithListOpenIDConnectProviderTagsError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listOpenIDConnectProviderTagsError = err
	}
}

// SAML Provider mock options.
func WithCreateSAMLProviderOutput(output *iam.CreateSAMLProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createSAMLProviderOutput = output
	}
}

func WithCreateSAMLProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createSAMLProviderError = err
	}
}

func WithGetSAMLProviderOutput(output *iam.GetSAMLProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getSAMLProviderOutput = output
	}
}

func WithGetSAMLProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getSAMLProviderError = err
	}
}

func WithUpdateSAMLProviderOutput(output *iam.UpdateSAMLProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateSAMLProviderOutput = output
	}
}

func WithUpdateSAMLProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateSAMLProviderError = err
	}
}

func WithDeleteSAMLProviderOutput(output *iam.DeleteSAMLProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteSAMLProviderOutput = output
	}
}

func WithDeleteSAMLProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteSAMLProviderError = err
	}
}

func WithTagSAMLProviderOutput(output *iam.TagSAMLProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagSAMLProviderOutput = output
	}
}

func WithTagSAMLProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagSAMLProviderError = err
	}
}

func WithUntagSAMLProviderOutput(output *iam.UntagSAMLProviderOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagSAMLProviderOutput = output
	}
}

func WithUntagSAMLProviderError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagSAMLProviderError = err
	}
}

func WithListSAMLProviderTagsOutput(output *iam.ListSAMLProviderTagsOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listSAMLProviderTagsOutput = output
	}
}

func WithListSAMLProviderTagsError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listSAMLProviderTagsError = err
	}
}

// Server certificate mock options.
func WithGetServerCertificateOutput(output *iam.GetServerCertificateOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getServerCertificateOutput = output
	}
}

func WithGetServerCertificateError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getServerCertificateError = err
	}
}

func WithListServerCertificateTagsOutput(output *iam.ListServerCertificateTagsOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listServerCertificateTagsOutput = output
	}
}

func WithListServerCertificateTagsError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listServerCertificateTagsError = err
	}
}

func WithDeleteServerCertificateOutput(output *iam.DeleteServerCertificateOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteServerCertificateOutput = output
	}
}

func WithDeleteServerCertificateError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteServerCertificateError = err
	}
}

func WithUpdateServerCertificateOutput(output *iam.UpdateServerCertificateOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateServerCertificateOutput = output
	}
}

func WithUpdateServerCertificateError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateServerCertificateError = err
	}
}

func WithTagServerCertificateOutput(output *iam.TagServerCertificateOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagServerCertificateOutput = output
	}
}

func WithTagServerCertificateError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagServerCertificateError = err
	}
}

func WithUntagServerCertificateOutput(output *iam.UntagServerCertificateOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagServerCertificateOutput = output
	}
}

func WithUntagServerCertificateError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagServerCertificateError = err
	}
}

func WithUploadServerCertificateOutput(output *iam.UploadServerCertificateOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.uploadServerCertificateOutput = output
	}
}

func WithUploadServerCertificateError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.uploadServerCertificateError = err
	}
}

// Group operation implementations.
func (m *iamServiceMock) CreateGroup(
	ctx context.Context,
	params *iam.CreateGroupInput,
	optFns ...func(*iam.Options),
) (*iam.CreateGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createGroupOutput, m.createGroupError
}

func (m *iamServiceMock) GetGroup(
	ctx context.Context,
	params *iam.GetGroupInput,
	optFns ...func(*iam.Options),
) (*iam.GetGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getGroupOutput, m.getGroupError
}

func (m *iamServiceMock) UpdateGroup(
	ctx context.Context,
	params *iam.UpdateGroupInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateGroupOutput, m.updateGroupError
}

func (m *iamServiceMock) DeleteGroup(
	ctx context.Context,
	params *iam.DeleteGroupInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteGroupOutput, m.deleteGroupError
}

func (m *iamServiceMock) AttachGroupPolicy(
	ctx context.Context,
	params *iam.AttachGroupPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.AttachGroupPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.attachGroupPolicyOutput, m.attachGroupPolicyError
}

func (m *iamServiceMock) DetachGroupPolicy(
	ctx context.Context,
	params *iam.DetachGroupPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DetachGroupPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.detachGroupPolicyOutput, m.detachGroupPolicyError
}

func (m *iamServiceMock) ListAttachedGroupPolicies(
	ctx context.Context,
	params *iam.ListAttachedGroupPoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListAttachedGroupPoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listAttachedGroupPoliciesOutput, m.listAttachedGroupPoliciesError
}

func (m *iamServiceMock) PutGroupPolicy(
	ctx context.Context,
	params *iam.PutGroupPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.PutGroupPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putGroupPolicyOutput, m.putGroupPolicyError
}

func (m *iamServiceMock) DeleteGroupPolicy(
	ctx context.Context,
	params *iam.DeleteGroupPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteGroupPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteGroupPolicyOutput, m.deleteGroupPolicyError
}

func (m *iamServiceMock) ListGroupPolicies(
	ctx context.Context,
	params *iam.ListGroupPoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListGroupPoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listGroupPoliciesOutput, m.listGroupPoliciesError
}

func (m *iamServiceMock) GetGroupPolicy(
	ctx context.Context,
	params *iam.GetGroupPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.GetGroupPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getGroupPolicyOutput, m.getGroupPolicyError
}

// Service interface implementation methods

func (m *iamServiceMock) CreateRole(
	ctx context.Context,
	params *iam.CreateRoleInput,
	optFns ...func(*iam.Options),
) (*iam.CreateRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createRoleOutput, m.createRoleError
}

func (m *iamServiceMock) GetRole(
	ctx context.Context,
	params *iam.GetRoleInput,
	optFns ...func(*iam.Options),
) (*iam.GetRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getRoleOutput, m.getRoleError
}

func (m *iamServiceMock) UpdateRole(
	ctx context.Context,
	params *iam.UpdateRoleInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateRoleOutput, m.updateRoleError
}

func (m *iamServiceMock) DeleteRole(
	ctx context.Context,
	params *iam.DeleteRoleInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteRoleOutput, m.deleteRoleError
}

func (m *iamServiceMock) UpdateAssumeRolePolicy(
	ctx context.Context,
	params *iam.UpdateAssumeRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateAssumeRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateAssumeRolePolicyOutput, m.updateAssumeRolePolicyError
}

func (m *iamServiceMock) AttachRolePolicy(
	ctx context.Context,
	params *iam.AttachRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.AttachRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.attachRolePolicyOutput, m.attachRolePolicyError
}

func (m *iamServiceMock) DetachRolePolicy(
	ctx context.Context,
	params *iam.DetachRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DetachRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.detachRolePolicyOutput, m.detachRolePolicyError
}

func (m *iamServiceMock) ListAttachedRolePolicies(
	ctx context.Context,
	params *iam.ListAttachedRolePoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListAttachedRolePoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listAttachedRolePoliciesOutput, m.listAttachedRolePoliciesError
}

func (m *iamServiceMock) PutRolePolicy(
	ctx context.Context,
	params *iam.PutRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.PutRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putRolePolicyOutput, m.putRolePolicyError
}

func (m *iamServiceMock) DeleteRolePolicy(
	ctx context.Context,
	params *iam.DeleteRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteRolePolicyOutput, m.deleteRolePolicyError
}

func (m *iamServiceMock) ListRolePolicies(
	ctx context.Context,
	params *iam.ListRolePoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListRolePoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listRolePoliciesOutput, m.listRolePoliciesError
}

func (m *iamServiceMock) GetRolePolicy(
	ctx context.Context,
	params *iam.GetRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.GetRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getRolePolicyOutput, m.getRolePolicyError
}

func (m *iamServiceMock) TagRole(
	ctx context.Context,
	params *iam.TagRoleInput,
	optFns ...func(*iam.Options),
) (*iam.TagRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.tagRoleOutput, m.tagRoleError
}

func (m *iamServiceMock) UntagRole(
	ctx context.Context,
	params *iam.UntagRoleInput,
	optFns ...func(*iam.Options),
) (*iam.UntagRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.untagRoleOutput, m.untagRoleError
}

func (m *iamServiceMock) ListRoleTags(
	ctx context.Context,
	params *iam.ListRoleTagsInput,
	optFns ...func(*iam.Options),
) (*iam.ListRoleTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listRoleTagsOutput, m.listRoleTagsError
}

func (m *iamServiceMock) PutRolePermissionsBoundary(
	ctx context.Context,
	params *iam.PutRolePermissionsBoundaryInput,
	optFns ...func(*iam.Options),
) (*iam.PutRolePermissionsBoundaryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putRolePermissionsBoundaryOutput, m.putRolePermissionsBoundaryError
}

func (m *iamServiceMock) DeleteRolePermissionsBoundary(
	ctx context.Context,
	params *iam.DeleteRolePermissionsBoundaryInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteRolePermissionsBoundaryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteRolePermissionsBoundaryOutput, m.deleteRolePermissionsBoundaryError
}

func (m *iamServiceMock) CreateUser(
	ctx context.Context,
	params *iam.CreateUserInput,
	optFns ...func(*iam.Options),
) (*iam.CreateUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createUserOutput, m.createUserError
}

func (m *iamServiceMock) GetUser(
	ctx context.Context,
	params *iam.GetUserInput,
	optFns ...func(*iam.Options),
) (*iam.GetUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getUserOutput, m.getUserError
}

func (m *iamServiceMock) UpdateUser(
	ctx context.Context,
	params *iam.UpdateUserInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateUserOutput, m.updateUserError
}

func (m *iamServiceMock) DeleteUser(
	ctx context.Context,
	params *iam.DeleteUserInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteUserOutput, m.deleteUserError
}

func (m *iamServiceMock) AttachUserPolicy(
	ctx context.Context,
	params *iam.AttachUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.AttachUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.attachUserPolicyOutput, m.attachUserPolicyError
}

func (m *iamServiceMock) DetachUserPolicy(
	ctx context.Context,
	params *iam.DetachUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DetachUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.detachUserPolicyOutput, m.detachUserPolicyError
}

func (m *iamServiceMock) ListAttachedUserPolicies(
	ctx context.Context,
	params *iam.ListAttachedUserPoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListAttachedUserPoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listAttachedUserPoliciesOutput, m.listAttachedUserPoliciesError
}

func (m *iamServiceMock) PutUserPolicy(
	ctx context.Context,
	params *iam.PutUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.PutUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putUserPolicyOutput, m.putUserPolicyError
}

func (m *iamServiceMock) DeleteUserPolicy(
	ctx context.Context,
	params *iam.DeleteUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteUserPolicyOutput, m.deleteUserPolicyError
}

func (m *iamServiceMock) ListUserPolicies(
	ctx context.Context,
	params *iam.ListUserPoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListUserPoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listUserPoliciesOutput, m.listUserPoliciesError
}

func (m *iamServiceMock) TagUser(
	ctx context.Context,
	params *iam.TagUserInput,
	optFns ...func(*iam.Options),
) (*iam.TagUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.tagUserOutput, m.tagUserError
}

func (m *iamServiceMock) UntagUser(
	ctx context.Context,
	params *iam.UntagUserInput,
	optFns ...func(*iam.Options),
) (*iam.UntagUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.untagUserOutput, m.untagUserError
}

func (m *iamServiceMock) ListUserTags(
	ctx context.Context,
	params *iam.ListUserTagsInput,
	optFns ...func(*iam.Options),
) (*iam.ListUserTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listUserTagsOutput, m.listUserTagsError
}

func (m *iamServiceMock) PutUserPermissionsBoundary(
	ctx context.Context,
	params *iam.PutUserPermissionsBoundaryInput,
	optFns ...func(*iam.Options),
) (*iam.PutUserPermissionsBoundaryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putUserPermissionsBoundaryOutput, m.putUserPermissionsBoundaryError
}

func (m *iamServiceMock) DeleteUserPermissionsBoundary(
	ctx context.Context,
	params *iam.DeleteUserPermissionsBoundaryInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteUserPermissionsBoundaryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteUserPermissionsBoundaryOutput, m.deleteUserPermissionsBoundaryError
}

func (m *iamServiceMock) GetUserPolicy(
	ctx context.Context,
	params *iam.GetUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.GetUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)

	// Handle different policy names
	if params.PolicyName != nil {
		switch aws.ToString(params.PolicyName) {
		case "S3Access":
			return &iam.GetUserPolicyOutput{
				UserName:       params.UserName,
				PolicyName:     aws.String("S3Access"),
				PolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject","s3:PutObject"],"Resource":["arn:aws:s3:::my-bucket/*"]}]}`),
			}, nil
		case "DynamoDBAccess":
			return &iam.GetUserPolicyOutput{
				UserName:       params.UserName,
				PolicyName:     aws.String("DynamoDBAccess"),
				PolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["dynamodb:PutItem","dynamodb:GetItem"],"Resource":["arn:aws:dynamodb:::table/my-table"]}]}`),
			}, nil
		}
	}

	// Fallback to the default mock response
	return m.getUserPolicyOutput, m.getUserPolicyError
}

func (m *iamServiceMock) AddUserToGroup(
	ctx context.Context,
	params *iam.AddUserToGroupInput,
	optFns ...func(*iam.Options),
) (*iam.AddUserToGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.addUserToGroupOutput, m.addUserToGroupError
}

func (m *iamServiceMock) RemoveUserFromGroup(
	ctx context.Context,
	params *iam.RemoveUserFromGroupInput,
	optFns ...func(*iam.Options),
) (*iam.RemoveUserFromGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.removeUserFromGroupOutput, m.removeUserFromGroupError
}

func (m *iamServiceMock) ListGroupsForUser(
	ctx context.Context,
	params *iam.ListGroupsForUserInput,
	optFns ...func(*iam.Options),
) (*iam.ListGroupsForUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listGroupsForUserOutput, m.listGroupsForUserError
}

func (m *iamServiceMock) CreateLoginProfile(
	ctx context.Context,
	params *iam.CreateLoginProfileInput,
	optFns ...func(*iam.Options),
) (*iam.CreateLoginProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createLoginProfileOutput, m.createLoginProfileError
}

func (m *iamServiceMock) GetLoginProfile(
	ctx context.Context,
	params *iam.GetLoginProfileInput,
	optFns ...func(*iam.Options),
) (*iam.GetLoginProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getLoginProfileOutput, m.getLoginProfileError
}

func (m *iamServiceMock) UpdateLoginProfile(
	ctx context.Context,
	params *iam.UpdateLoginProfileInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateLoginProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateLoginProfileOutput, m.updateLoginProfileError
}

func (m *iamServiceMock) DeleteLoginProfile(
	ctx context.Context,
	params *iam.DeleteLoginProfileInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteLoginProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteLoginProfileOutput, m.deleteLoginProfileError
}

// Access key methods.
func (m *iamServiceMock) CreateAccessKey(
	ctx context.Context,
	params *iam.CreateAccessKeyInput,
	optFns ...func(*iam.Options),
) (*iam.CreateAccessKeyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createAccessKeyOutput, m.createAccessKeyError
}

func (m *iamServiceMock) UpdateAccessKey(
	ctx context.Context,
	params *iam.UpdateAccessKeyInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateAccessKeyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateAccessKeyOutput, m.updateAccessKeyError
}

func (m *iamServiceMock) DeleteAccessKey(
	ctx context.Context,
	params *iam.DeleteAccessKeyInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteAccessKeyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteAccessKeyOutput, m.deleteAccessKeyError
}

func (m *iamServiceMock) ListAccessKeys(
	ctx context.Context,
	params *iam.ListAccessKeysInput,
	optFns ...func(*iam.Options),
) (*iam.ListAccessKeysOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listAccessKeysOutput, m.listAccessKeysError
}

// Instance profile methods.
func (m *iamServiceMock) CreateInstanceProfile(
	ctx context.Context,
	params *iam.CreateInstanceProfileInput,
	optFns ...func(*iam.Options),
) (*iam.CreateInstanceProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createInstanceProfileOutput, m.createInstanceProfileError
}

func (m *iamServiceMock) GetInstanceProfile(
	ctx context.Context,
	params *iam.GetInstanceProfileInput,
	optFns ...func(*iam.Options),
) (*iam.GetInstanceProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getInstanceProfileOutput, m.getInstanceProfileError
}

func (m *iamServiceMock) DeleteInstanceProfile(
	ctx context.Context,
	params *iam.DeleteInstanceProfileInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteInstanceProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteInstanceProfileOutput, m.deleteInstanceProfileError
}

func (m *iamServiceMock) AddRoleToInstanceProfile(
	ctx context.Context,
	params *iam.AddRoleToInstanceProfileInput,
	optFns ...func(*iam.Options),
) (*iam.AddRoleToInstanceProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.addRoleToInstanceProfileOutput, m.addRoleToInstanceProfileError
}

func (m *iamServiceMock) RemoveRoleFromInstanceProfile(
	ctx context.Context,
	params *iam.RemoveRoleFromInstanceProfileInput,
	optFns ...func(*iam.Options),
) (*iam.RemoveRoleFromInstanceProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.removeRoleFromInstanceProfileOutput, m.removeRoleFromInstanceProfileError
}

// Managed policy methods.
func (m *iamServiceMock) CreatePolicy(
	ctx context.Context,
	params *iam.CreatePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.CreatePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createPolicyOutput, m.createPolicyError
}

func (m *iamServiceMock) GetPolicy(
	ctx context.Context,
	params *iam.GetPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.GetPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getPolicyOutput, m.getPolicyError
}

func (m *iamServiceMock) DeletePolicy(
	ctx context.Context,
	params *iam.DeletePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DeletePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deletePolicyOutput, m.deletePolicyError
}

func (m *iamServiceMock) CreatePolicyVersion(
	ctx context.Context,
	params *iam.CreatePolicyVersionInput,
	optFns ...func(*iam.Options),
) (*iam.CreatePolicyVersionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createPolicyVersionOutput, m.createPolicyVersionError
}

func (m *iamServiceMock) DeletePolicyVersion(
	ctx context.Context,
	params *iam.DeletePolicyVersionInput,
	optFns ...func(*iam.Options),
) (*iam.DeletePolicyVersionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deletePolicyVersionOutput, m.deletePolicyVersionError
}

func (m *iamServiceMock) ListPolicyVersions(
	ctx context.Context,
	params *iam.ListPolicyVersionsInput,
	optFns ...func(*iam.Options),
) (*iam.ListPolicyVersionsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listPolicyVersionsOutput, m.listPolicyVersionsError
}

func (m *iamServiceMock) GetPolicyVersion(
	ctx context.Context,
	params *iam.GetPolicyVersionInput,
	optFns ...func(*iam.Options),
) (*iam.GetPolicyVersionOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getPolicyVersionOutput, m.getPolicyVersionError
}

func (m *iamServiceMock) TagPolicy(
	ctx context.Context,
	params *iam.TagPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.TagPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.tagPolicyOutput, m.tagPolicyError
}

// SAML Provider operation implementations.
func (m *iamServiceMock) CreateSAMLProvider(
	ctx context.Context,
	params *iam.CreateSAMLProviderInput,
	optFns ...func(*iam.Options),
) (*iam.CreateSAMLProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createSAMLProviderOutput, m.createSAMLProviderError
}

func (m *iamServiceMock) GetSAMLProvider(
	ctx context.Context,
	params *iam.GetSAMLProviderInput,
	optFns ...func(*iam.Options),
) (*iam.GetSAMLProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getSAMLProviderOutput, m.getSAMLProviderError
}

func (m *iamServiceMock) UpdateSAMLProvider(
	ctx context.Context,
	params *iam.UpdateSAMLProviderInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateSAMLProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateSAMLProviderOutput, m.updateSAMLProviderError
}

func (m *iamServiceMock) DeleteSAMLProvider(
	ctx context.Context,
	params *iam.DeleteSAMLProviderInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteSAMLProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteSAMLProviderOutput, m.deleteSAMLProviderError
}

func (m *iamServiceMock) TagSAMLProvider(
	ctx context.Context,
	params *iam.TagSAMLProviderInput,
	optFns ...func(*iam.Options),
) (*iam.TagSAMLProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.tagSAMLProviderOutput, m.tagSAMLProviderError
}

func (m *iamServiceMock) UntagSAMLProvider(
	ctx context.Context,
	params *iam.UntagSAMLProviderInput,
	optFns ...func(*iam.Options),
) (*iam.UntagSAMLProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.untagSAMLProviderOutput, m.untagSAMLProviderError
}

func (m *iamServiceMock) ListSAMLProviderTags(
	ctx context.Context,
	params *iam.ListSAMLProviderTagsInput,
	optFns ...func(*iam.Options),
) (*iam.ListSAMLProviderTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listSAMLProviderTagsOutput, m.listSAMLProviderTagsError
}

func (m *iamServiceMock) UntagPolicy(
	ctx context.Context,
	params *iam.UntagPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.UntagPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.untagPolicyOutput, m.untagPolicyError
}

func (m *iamServiceMock) ListPolicyTags(
	ctx context.Context,
	params *iam.ListPolicyTagsInput,
	optFns ...func(*iam.Options),
) (*iam.ListPolicyTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listPolicyTagsOutput, m.listPolicyTagsError
}

// OIDC provider methods.
func (m *iamServiceMock) CreateOpenIDConnectProvider(
	ctx context.Context,
	params *iam.CreateOpenIDConnectProviderInput,
	optFns ...func(*iam.Options),
) (*iam.CreateOpenIDConnectProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createOpenIDConnectProviderOutput, m.createOpenIDConnectProviderError
}

func (m *iamServiceMock) GetOpenIDConnectProvider(
	ctx context.Context,
	params *iam.GetOpenIDConnectProviderInput,
	optFns ...func(*iam.Options),
) (*iam.GetOpenIDConnectProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getOpenIDConnectProviderOutput, m.getOpenIDConnectProviderError
}

func (m *iamServiceMock) AddClientIDToOpenIDConnectProvider(
	ctx context.Context,
	params *iam.AddClientIDToOpenIDConnectProviderInput,
	optFns ...func(*iam.Options),
) (*iam.AddClientIDToOpenIDConnectProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.addClientIDToOpenIDConnectProviderOutput, m.addClientIDToOpenIDConnectProviderError
}

func (m *iamServiceMock) RemoveClientIDFromOpenIDConnectProvider(
	ctx context.Context,
	params *iam.RemoveClientIDFromOpenIDConnectProviderInput,
	optFns ...func(*iam.Options),
) (*iam.RemoveClientIDFromOpenIDConnectProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.removeClientIDFromOpenIDConnectProviderOutput, m.removeClientIDFromOpenIDConnectProviderError
}

func (m *iamServiceMock) UpdateOpenIDConnectProviderThumbprint(
	ctx context.Context,
	params *iam.UpdateOpenIDConnectProviderThumbprintInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateOpenIDConnectProviderThumbprintOutput, m.updateOpenIDConnectProviderThumbprintError
}

func (m *iamServiceMock) DeleteOpenIDConnectProvider(
	ctx context.Context,
	params *iam.DeleteOpenIDConnectProviderInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteOpenIDConnectProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteOpenIDConnectProviderOutput, m.deleteOpenIDConnectProviderError
}

func (m *iamServiceMock) TagOpenIDConnectProvider(
	ctx context.Context,
	params *iam.TagOpenIDConnectProviderInput,
	optFns ...func(*iam.Options),
) (*iam.TagOpenIDConnectProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.tagOpenIDConnectProviderOutput, m.tagOpenIDConnectProviderError
}

func (m *iamServiceMock) UntagOpenIDConnectProvider(
	ctx context.Context,
	params *iam.UntagOpenIDConnectProviderInput,
	optFns ...func(*iam.Options),
) (*iam.UntagOpenIDConnectProviderOutput, error) {
	m.RegisterCall(ctx, params)
	return m.untagOpenIDConnectProviderOutput, m.untagOpenIDConnectProviderError
}

func (m *iamServiceMock) ListOpenIDConnectProviderTags(
	ctx context.Context,
	params *iam.ListOpenIDConnectProviderTagsInput,
	optFns ...func(*iam.Options),
) (*iam.ListOpenIDConnectProviderTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listOpenIDConnectProviderTagsOutput, m.listOpenIDConnectProviderTagsError
}

// Server certificate methods.
func (m *iamServiceMock) GetServerCertificate(
	ctx context.Context,
	params *iam.GetServerCertificateInput,
	optFns ...func(*iam.Options),
) (*iam.GetServerCertificateOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getServerCertificateOutput, m.getServerCertificateError
}

func (m *iamServiceMock) ListServerCertificateTags(
	ctx context.Context,
	params *iam.ListServerCertificateTagsInput,
	optFns ...func(*iam.Options),
) (*iam.ListServerCertificateTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listServerCertificateTagsOutput, m.listServerCertificateTagsError
}

func (m *iamServiceMock) DeleteServerCertificate(
	ctx context.Context,
	params *iam.DeleteServerCertificateInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteServerCertificateOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteServerCertificateOutput, m.deleteServerCertificateError
}

func (m *iamServiceMock) UpdateServerCertificate(
	ctx context.Context,
	params *iam.UpdateServerCertificateInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateServerCertificateOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateServerCertificateOutput, m.updateServerCertificateError
}

func (m *iamServiceMock) TagServerCertificate(
	ctx context.Context,
	params *iam.TagServerCertificateInput,
	optFns ...func(*iam.Options),
) (*iam.TagServerCertificateOutput, error) {
	m.RegisterCall(ctx, params)
	return m.tagServerCertificateOutput, m.tagServerCertificateError
}

func (m *iamServiceMock) UntagServerCertificate(
	ctx context.Context,
	params *iam.UntagServerCertificateInput,
	optFns ...func(*iam.Options),
) (*iam.UntagServerCertificateOutput, error) {
	m.RegisterCall(ctx, params)
	return m.untagServerCertificateOutput, m.untagServerCertificateError
}

func (m *iamServiceMock) UploadServerCertificate(
	ctx context.Context,
	params *iam.UploadServerCertificateInput,
	optFns ...func(*iam.Options),
) (*iam.UploadServerCertificateOutput, error) {
	m.RegisterCall(ctx, params)
	return m.uploadServerCertificateOutput, m.uploadServerCertificateError
}
