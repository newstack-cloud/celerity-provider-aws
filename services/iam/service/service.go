package iamservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS IAM service
// used by the IAM resource implementations.
type Service interface {
	// CreateRole creates a new role for your AWS account.
	CreateRole(
		ctx context.Context,
		params *iam.CreateRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateRoleOutput, error)

	// GetRole retrieves information about the specified role, including the role's
	// path, GUID, ARN, and the role's trust policy that grants permission to assume the role.
	GetRole(
		ctx context.Context,
		params *iam.GetRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.GetRoleOutput, error)

	// UpdateRole updates the description or maximum session duration setting of a role.
	UpdateRole(
		ctx context.Context,
		params *iam.UpdateRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateRoleOutput, error)

	// UpdateAssumeRolePolicy updates the policy that grants an IAM entity permission
	// to assume a role.
	UpdateAssumeRolePolicy(
		ctx context.Context,
		params *iam.UpdateAssumeRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateAssumeRolePolicyOutput, error)

	// DeleteRole deletes the specified role. Unlike the AWS Management Console, when you
	// delete a role programmatically, you must delete the items attached to the role manually,
	// or the deletion fails.
	DeleteRole(
		ctx context.Context,
		params *iam.DeleteRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteRoleOutput, error)

	// AttachRolePolicy attaches the specified managed policy to the specified IAM role.
	AttachRolePolicy(
		ctx context.Context,
		params *iam.AttachRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.AttachRolePolicyOutput, error)

	// DetachRolePolicy removes the specified managed policy from the specified role.
	DetachRolePolicy(
		ctx context.Context,
		params *iam.DetachRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DetachRolePolicyOutput, error)

	// ListAttachedRolePolicies lists all managed policies that are attached to the specified IAM role.
	ListAttachedRolePolicies(
		ctx context.Context,
		params *iam.ListAttachedRolePoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListAttachedRolePoliciesOutput, error)

	// PutRolePolicy adds or updates an inline policy document that is embedded in the specified IAM role.
	PutRolePolicy(
		ctx context.Context,
		params *iam.PutRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.PutRolePolicyOutput, error)

	// DeleteRolePolicy deletes the specified inline policy that is embedded in the specified IAM role.
	DeleteRolePolicy(
		ctx context.Context,
		params *iam.DeleteRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteRolePolicyOutput, error)

	// ListRolePolicies lists the names of the inline policies that are embedded in the specified IAM role.
	ListRolePolicies(
		ctx context.Context,
		params *iam.ListRolePoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListRolePoliciesOutput, error)

	// TagRole adds one or more tags to an IAM role.
	TagRole(
		ctx context.Context,
		params *iam.TagRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.TagRoleOutput, error)

	// UntagRole removes the specified tags from the role.
	UntagRole(
		ctx context.Context,
		params *iam.UntagRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.UntagRoleOutput, error)

	// ListRoleTags lists the tags that are attached to the specified role.
	ListRoleTags(
		ctx context.Context,
		params *iam.ListRoleTagsInput,
		optFns ...func(*iam.Options),
	) (*iam.ListRoleTagsOutput, error)

	// PutRolePermissionsBoundary adds a permissions boundary to the IAM role.
	PutRolePermissionsBoundary(
		ctx context.Context,
		params *iam.PutRolePermissionsBoundaryInput,
		optFns ...func(*iam.Options),
	) (*iam.PutRolePermissionsBoundaryOutput, error)

	// DeleteRolePermissionsBoundary deletes the permissions boundary for the specified IAM role.
	DeleteRolePermissionsBoundary(
		ctx context.Context,
		params *iam.DeleteRolePermissionsBoundaryInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteRolePermissionsBoundaryOutput, error)

	// GetRolePolicy retrieves an inline policy document that is embedded in the specified IAM role.
	GetRolePolicy(
		ctx context.Context,
		params *iam.GetRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetRolePolicyOutput, error)

	// CreateUser creates a new user for your AWS account.
	CreateUser(
		ctx context.Context,
		params *iam.CreateUserInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateUserOutput, error)

	// GetUser retrieves information about the specified user, including the user's
	// path, GUID, ARN, and the user's creation date.
	GetUser(
		ctx context.Context,
		params *iam.GetUserInput,
		optFns ...func(*iam.Options),
	) (*iam.GetUserOutput, error)

	// UpdateUser updates the name and/or the path of the specified IAM user.
	UpdateUser(
		ctx context.Context,
		params *iam.UpdateUserInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateUserOutput, error)

	// DeleteUser deletes the specified IAM user. Unlike the AWS Management Console, when you
	// delete a user programmatically, you must delete the items attached to the user manually,
	// or the deletion fails.
	DeleteUser(
		ctx context.Context,
		params *iam.DeleteUserInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteUserOutput, error)

	// AttachUserPolicy attaches the specified managed policy to the specified user.
	AttachUserPolicy(
		ctx context.Context,
		params *iam.AttachUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.AttachUserPolicyOutput, error)

	// DetachUserPolicy removes the specified managed policy from the specified user.
	DetachUserPolicy(
		ctx context.Context,
		params *iam.DetachUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DetachUserPolicyOutput, error)

	// ListAttachedUserPolicies lists all managed policies that are attached to the specified IAM user.
	ListAttachedUserPolicies(
		ctx context.Context,
		params *iam.ListAttachedUserPoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListAttachedUserPoliciesOutput, error)

	// PutUserPolicy adds or updates an inline policy document that is embedded in the specified IAM user.
	PutUserPolicy(
		ctx context.Context,
		params *iam.PutUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.PutUserPolicyOutput, error)

	// DeleteUserPolicy deletes the specified inline policy that is embedded in the specified IAM user.
	DeleteUserPolicy(
		ctx context.Context,
		params *iam.DeleteUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteUserPolicyOutput, error)

	// ListUserPolicies lists the names of the inline policies that are embedded in the specified IAM user.
	ListUserPolicies(
		ctx context.Context,
		params *iam.ListUserPoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListUserPoliciesOutput, error)

	// TagUser adds one or more tags to an IAM user.
	TagUser(
		ctx context.Context,
		params *iam.TagUserInput,
		optFns ...func(*iam.Options),
	) (*iam.TagUserOutput, error)

	// UntagUser removes the specified tags from the user.
	UntagUser(
		ctx context.Context,
		params *iam.UntagUserInput,
		optFns ...func(*iam.Options),
	) (*iam.UntagUserOutput, error)

	// ListUserTags lists the tags that are attached to the specified user.
	ListUserTags(
		ctx context.Context,
		params *iam.ListUserTagsInput,
		optFns ...func(*iam.Options),
	) (*iam.ListUserTagsOutput, error)

	// PutUserPermissionsBoundary adds a permissions boundary to the IAM user.
	PutUserPermissionsBoundary(
		ctx context.Context,
		params *iam.PutUserPermissionsBoundaryInput,
		optFns ...func(*iam.Options),
	) (*iam.PutUserPermissionsBoundaryOutput, error)

	// DeleteUserPermissionsBoundary deletes the permissions boundary for the specified IAM user.
	DeleteUserPermissionsBoundary(
		ctx context.Context,
		params *iam.DeleteUserPermissionsBoundaryInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteUserPermissionsBoundaryOutput, error)

	// GetUserPolicy retrieves an inline policy document that is embedded in the specified IAM user.
	GetUserPolicy(
		ctx context.Context,
		params *iam.GetUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetUserPolicyOutput, error)

	// AddUserToGroup adds the specified user to the specified group.
	AddUserToGroup(
		ctx context.Context,
		params *iam.AddUserToGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.AddUserToGroupOutput, error)

	// RemoveUserFromGroup removes the specified user from the specified group.
	RemoveUserFromGroup(
		ctx context.Context,
		params *iam.RemoveUserFromGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.RemoveUserFromGroupOutput, error)

	// ListGroupsForUser lists the IAM groups that the specified IAM user belongs to.
	ListGroupsForUser(
		ctx context.Context,
		params *iam.ListGroupsForUserInput,
		optFns ...func(*iam.Options),
	) (*iam.ListGroupsForUserOutput, error)

	// CreateLoginProfile creates a password for the specified IAM user.
	CreateLoginProfile(
		ctx context.Context,
		params *iam.CreateLoginProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateLoginProfileOutput, error)

	// GetLoginProfile retrieves the login profile for the specified IAM user.
	GetLoginProfile(
		ctx context.Context,
		params *iam.GetLoginProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.GetLoginProfileOutput, error)

	// UpdateLoginProfile changes the password for the specified IAM user.
	UpdateLoginProfile(
		ctx context.Context,
		params *iam.UpdateLoginProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateLoginProfileOutput, error)

	// DeleteLoginProfile deletes the password for the specified IAM user.
	DeleteLoginProfile(
		ctx context.Context,
		params *iam.DeleteLoginProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteLoginProfileOutput, error)

	// CreateGroup creates a new group for your AWS account.
	CreateGroup(
		ctx context.Context,
		params *iam.CreateGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateGroupOutput, error)

	// GetGroup retrieves information about the specified group, including the group's
	// path, GUID, ARN, and the group's creation date.
	GetGroup(
		ctx context.Context,
		params *iam.GetGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.GetGroupOutput, error)

	// UpdateGroup updates the name and/or the path of the specified IAM group.
	UpdateGroup(
		ctx context.Context,
		params *iam.UpdateGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateGroupOutput, error)

	// DeleteGroup deletes the specified IAM group. Unlike the AWS Management Console, when you
	// delete a group programmatically, you must delete the items attached to the group manually,
	// or the deletion fails.
	DeleteGroup(
		ctx context.Context,
		params *iam.DeleteGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteGroupOutput, error)

	// AttachGroupPolicy attaches the specified managed policy to the specified group.
	AttachGroupPolicy(
		ctx context.Context,
		params *iam.AttachGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.AttachGroupPolicyOutput, error)

	// DetachGroupPolicy removes the specified managed policy from the specified group.
	DetachGroupPolicy(
		ctx context.Context,
		params *iam.DetachGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DetachGroupPolicyOutput, error)

	// ListAttachedGroupPolicies lists all managed policies that are attached to the specified IAM group.
	ListAttachedGroupPolicies(
		ctx context.Context,
		params *iam.ListAttachedGroupPoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListAttachedGroupPoliciesOutput, error)

	// PutGroupPolicy adds or updates an inline policy document that is embedded in the specified IAM group.
	PutGroupPolicy(
		ctx context.Context,
		params *iam.PutGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.PutGroupPolicyOutput, error)

	// DeleteGroupPolicy deletes the specified inline policy that is embedded in the specified IAM group.
	DeleteGroupPolicy(
		ctx context.Context,
		params *iam.DeleteGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteGroupPolicyOutput, error)

	// ListGroupPolicies lists the names of the inline policies that are embedded in the specified IAM group.
	ListGroupPolicies(
		ctx context.Context,
		params *iam.ListGroupPoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListGroupPoliciesOutput, error)

	// GetGroupPolicy retrieves an inline policy document that is embedded in the specified IAM group.
	GetGroupPolicy(
		ctx context.Context,
		params *iam.GetGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetGroupPolicyOutput, error)

	// CreateAccessKey creates a new access key for the specified IAM user.
	CreateAccessKey(
		ctx context.Context,
		params *iam.CreateAccessKeyInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateAccessKeyOutput, error)

	// UpdateAccessKey changes the status of the specified access key from Active to Inactive, or vice versa.
	UpdateAccessKey(
		ctx context.Context,
		params *iam.UpdateAccessKeyInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateAccessKeyOutput, error)

	// DeleteAccessKey deletes the specified access key.
	DeleteAccessKey(
		ctx context.Context,
		params *iam.DeleteAccessKeyInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteAccessKeyOutput, error)

	// ListAccessKeys lists the access keys associated with the specified IAM user.
	ListAccessKeys(
		ctx context.Context,
		params *iam.ListAccessKeysInput,
		optFns ...func(*iam.Options),
	) (*iam.ListAccessKeysOutput, error)

	// CreateInstanceProfile creates a new instance profile for your AWS account.
	CreateInstanceProfile(
		ctx context.Context,
		params *iam.CreateInstanceProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateInstanceProfileOutput, error)

	// GetInstanceProfile retrieves information about the specified instance profile, including the instance profile's
	// path, GUID, ARN, and the role associated with the instance profile.
	GetInstanceProfile(
		ctx context.Context,
		params *iam.GetInstanceProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.GetInstanceProfileOutput, error)

	// DeleteInstanceProfile deletes the specified instance profile. The instance profile must not have an associated role.
	DeleteInstanceProfile(
		ctx context.Context,
		params *iam.DeleteInstanceProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteInstanceProfileOutput, error)

	// AddRoleToInstanceProfile adds the specified IAM role to the specified instance profile.
	AddRoleToInstanceProfile(
		ctx context.Context,
		params *iam.AddRoleToInstanceProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.AddRoleToInstanceProfileOutput, error)

	// RemoveRoleFromInstanceProfile removes the specified IAM role from the specified instance profile.
	RemoveRoleFromInstanceProfile(
		ctx context.Context,
		params *iam.RemoveRoleFromInstanceProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.RemoveRoleFromInstanceProfileOutput, error)

	// CreatePolicy creates a new managed policy for your AWS account.
	CreatePolicy(
		ctx context.Context,
		params *iam.CreatePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.CreatePolicyOutput, error)

	// GetPolicy retrieves information about the specified managed policy, including the policy's
	// default version and the total number of IAM users, groups, and roles that the policy is attached to.
	GetPolicy(
		ctx context.Context,
		params *iam.GetPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetPolicyOutput, error)

	// DeletePolicy deletes the specified managed policy.
	DeletePolicy(
		ctx context.Context,
		params *iam.DeletePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DeletePolicyOutput, error)

	// CreatePolicyVersion creates a new version of the specified managed policy.
	CreatePolicyVersion(
		ctx context.Context,
		params *iam.CreatePolicyVersionInput,
		optFns ...func(*iam.Options),
	) (*iam.CreatePolicyVersionOutput, error)

	// DeletePolicyVersion deletes the specified version from the specified managed policy.
	DeletePolicyVersion(
		ctx context.Context,
		params *iam.DeletePolicyVersionInput,
		optFns ...func(*iam.Options),
	) (*iam.DeletePolicyVersionOutput, error)

	// ListPolicyVersions lists the versions of the specified managed policy.
	ListPolicyVersions(
		ctx context.Context,
		params *iam.ListPolicyVersionsInput,
		optFns ...func(*iam.Options),
	) (*iam.ListPolicyVersionsOutput, error)

	// GetPolicyVersion retrieves information about the specified version of the
	// specified managed policy, including the policy document.
	GetPolicyVersion(
		ctx context.Context,
		params *iam.GetPolicyVersionInput,
		optFns ...func(*iam.Options),
	) (*iam.GetPolicyVersionOutput, error)

	// TagPolicy adds one or more tags to an IAM managed policy.
	TagPolicy(
		ctx context.Context,
		params *iam.TagPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.TagPolicyOutput, error)

	// UntagPolicy removes the specified tags from the managed policy.
	UntagPolicy(
		ctx context.Context,
		params *iam.UntagPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.UntagPolicyOutput, error)

	// ListPolicyTags lists the tags that are attached to the specified managed policy.
	ListPolicyTags(
		ctx context.Context,
		params *iam.ListPolicyTagsInput,
		optFns ...func(*iam.Options),
	) (*iam.ListPolicyTagsOutput, error)

	// CreateOpenIDConnectProvider creates an IAM entity to describe an identity provider (IdP) that supports OpenID Connect (OIDC).
	CreateOpenIDConnectProvider(
		ctx context.Context,
		params *iam.CreateOpenIDConnectProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateOpenIDConnectProviderOutput, error)

	// GetOpenIDConnectProvider returns information about the specified OpenID Connect (OIDC) provider resource object in IAM.
	GetOpenIDConnectProvider(
		ctx context.Context,
		params *iam.GetOpenIDConnectProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.GetOpenIDConnectProviderOutput, error)

	// AddClientIDToOpenIDConnectProvider adds a new client ID (also known as audience) to the list of client IDs already registered for the specified IAM OpenID Connect (OIDC) provider resource.
	AddClientIDToOpenIDConnectProvider(
		ctx context.Context,
		params *iam.AddClientIDToOpenIDConnectProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.AddClientIDToOpenIDConnectProviderOutput, error)

	// RemoveClientIDFromOpenIDConnectProvider removes the specified client ID (also known as audience) from the list of client IDs registered for the specified IAM OpenID Connect (OIDC) provider resource object.
	RemoveClientIDFromOpenIDConnectProvider(
		ctx context.Context,
		params *iam.RemoveClientIDFromOpenIDConnectProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.RemoveClientIDFromOpenIDConnectProviderOutput, error)

	// UpdateOpenIDConnectProviderThumbprint replaces the existing list of server certificate thumbprints associated with an OpenID Connect (OIDC) provider resource object with a new list of thumbprints.
	UpdateOpenIDConnectProviderThumbprint(
		ctx context.Context,
		params *iam.UpdateOpenIDConnectProviderThumbprintInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error)

	// DeleteOpenIDConnectProvider deletes an OpenID Connect identity provider (IdP) resource object in IAM.
	DeleteOpenIDConnectProvider(
		ctx context.Context,
		params *iam.DeleteOpenIDConnectProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteOpenIDConnectProviderOutput, error)

	// TagOpenIDConnectProvider adds one or more tags to an OpenID Connect (OIDC)-compatible identity provider.
	TagOpenIDConnectProvider(
		ctx context.Context,
		params *iam.TagOpenIDConnectProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.TagOpenIDConnectProviderOutput, error)

	// UntagOpenIDConnectProvider removes the specified tags from the specified OpenID Connect (OIDC)-compatible identity provider.
	UntagOpenIDConnectProvider(
		ctx context.Context,
		params *iam.UntagOpenIDConnectProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.UntagOpenIDConnectProviderOutput, error)

	// ListOpenIDConnectProviderTags lists the tags that are attached to the specified OpenID Connect (OIDC)-compatible identity provider.
	ListOpenIDConnectProviderTags(
		ctx context.Context,
		params *iam.ListOpenIDConnectProviderTagsInput,
		optFns ...func(*iam.Options),
	) (*iam.ListOpenIDConnectProviderTagsOutput, error)

	// CreateSAMLProvider creates an IAM resource that describes an identity provider (IdP) that supports SAML 2.0.
	CreateSAMLProvider(
		ctx context.Context,
		params *iam.CreateSAMLProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateSAMLProviderOutput, error)

	// GetSAMLProvider returns the SAML provider metadocument that was uploaded when the IAM SAML provider resource object was created or updated.
	GetSAMLProvider(
		ctx context.Context,
		params *iam.GetSAMLProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.GetSAMLProviderOutput, error)

	// UpdateSAMLProvider updates the metadata document for an existing SAML provider resource object.
	UpdateSAMLProvider(
		ctx context.Context,
		params *iam.UpdateSAMLProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateSAMLProviderOutput, error)

	// DeleteSAMLProvider deletes a SAML provider resource in IAM.
	DeleteSAMLProvider(
		ctx context.Context,
		params *iam.DeleteSAMLProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteSAMLProviderOutput, error)

	// TagSAMLProvider adds one or more tags to a Security Assertion Markup Language (SAML) identity provider.
	TagSAMLProvider(
		ctx context.Context,
		params *iam.TagSAMLProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.TagSAMLProviderOutput, error)

	// UntagSAMLProvider removes the specified tags from the specified Security Assertion Markup Language (SAML) identity provider.
	UntagSAMLProvider(
		ctx context.Context,
		params *iam.UntagSAMLProviderInput,
		optFns ...func(*iam.Options),
	) (*iam.UntagSAMLProviderOutput, error)

	// ListSAMLProviderTags lists the tags that are attached to the specified Security Assertion Markup Language (SAML) identity provider.
	ListSAMLProviderTags(
		ctx context.Context,
		params *iam.ListSAMLProviderTagsInput,
		optFns ...func(*iam.Options),
	) (*iam.ListSAMLProviderTagsOutput, error)

	// Retrieves information about the specified server certificate stored in IAM.
	//
	// For more information about working with server certificates, see [Working with server certificates] in the IAM
	// User Guide. This topic includes a list of Amazon Web Services services that can
	// use the server certificates that you manage with IAM.
	//
	// [Working with server certificates]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_server-certs.html
	GetServerCertificate(
		ctx context.Context,
		params *iam.GetServerCertificateInput,
		optFns ...func(*iam.Options),
	) (*iam.GetServerCertificateOutput, error)

	// Lists the tags that are attached to the specified IAM server certificate. The
	// returned list of tags is sorted by tag key. For more information about tagging,
	// see [Tagging IAM resources]in the IAM User Guide.
	//
	// For certificates in a Region supported by Certificate Manager (ACM), we
	// recommend that you don't use IAM server certificates. Instead, use ACM to
	// provision, manage, and deploy your server certificates. For more information
	// about IAM server certificates, [Working with server certificates]in the IAM User Guide.
	//
	// [Working with server certificates]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_server-certs.html
	// [Tagging IAM resources]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html
	ListServerCertificateTags(
		ctx context.Context,
		params *iam.ListServerCertificateTagsInput,
		optFns ...func(*iam.Options),
	) (*iam.ListServerCertificateTagsOutput, error)

	// Deletes the specified server certificate.
	//
	// For more information about working with server certificates, see [Working with server certificates] in the IAM
	// User Guide. This topic also includes a list of Amazon Web Services services that
	// can use the server certificates that you manage with IAM.
	//
	// If you are using a server certificate with Elastic Load Balancing, deleting the
	// certificate could have implications for your application. If Elastic Load
	// Balancing doesn't detect the deletion of bound certificates, it may continue to
	// use the certificates. This could cause Elastic Load Balancing to stop accepting
	// traffic. We recommend that you remove the reference to the certificate from
	// Elastic Load Balancing before using this command to delete the certificate. For
	// more information, see [DeleteLoadBalancerListeners]in the Elastic Load Balancing API Reference.
	//
	// [Working with server certificates]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_server-certs.html
	// [DeleteLoadBalancerListeners]: https://docs.aws.amazon.com/ElasticLoadBalancing/latest/APIReference/API_DeleteLoadBalancerListeners.html
	DeleteServerCertificate(
		ctx context.Context,
		params *iam.DeleteServerCertificateInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteServerCertificateOutput, error)

	// Updates the name and/or the path of the specified server certificate stored in
	// IAM.
	//
	// For more information about working with server certificates, see [Working with server certificates] in the IAM
	// User Guide. This topic also includes a list of Amazon Web Services services that
	// can use the server certificates that you manage with IAM.
	//
	// You should understand the implications of changing a server certificate's path
	// or name. For more information, see [Renaming a server certificate]in the IAM User Guide.
	//
	// The person making the request (the principal), must have permission to change
	// the server certificate with the old name and the new name. For example, to
	// change the certificate named ProductionCert to ProdCert , the principal must
	// have a policy that allows them to update both certificates. If the principal has
	// permission to update the ProductionCert group, but not the ProdCert
	// certificate, then the update fails. For more information about permissions, see [Access management]
	// in the IAM User Guide.
	//
	// [Renaming a server certificate]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_server-certs_manage.html#RenamingServerCerts
	// [Access management]: https://docs.aws.amazon.com/IAM/latest/UserGuide/access.html
	// [Working with server certificates]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_server-certs.html
	UpdateServerCertificate(
		ctx context.Context,
		params *iam.UpdateServerCertificateInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateServerCertificateOutput, error)

	// TagServerCertificate adds one or more tags to an IAM server certificate.
	TagServerCertificate(
		ctx context.Context,
		params *iam.TagServerCertificateInput,
		optFns ...func(*iam.Options),
	) (*iam.TagServerCertificateOutput, error)

	// UntagServerCertificate removes the specified tags from the IAM server certificate.
	UntagServerCertificate(
		ctx context.Context,
		params *iam.UntagServerCertificateInput,
		optFns ...func(*iam.Options),
	) (*iam.UntagServerCertificateOutput, error)

	// Uploads a server certificate entity for the Amazon Web Services account. The
	// server certificate entity includes a public key certificate, a private key, and
	// an optional certificate chain, which should all be PEM-encoded.
	//
	// We recommend that you use [Certificate Manager] to provision, manage, and deploy your server
	// certificates. With ACM you can request a certificate, deploy it to Amazon Web
	// Services resources, and let ACM handle certificate renewals for you.
	// Certificates provided by ACM are free. For more information about using ACM, see
	// the [Certificate Manager User Guide].
	//
	// For more information about working with server certificates, see [Working with server certificates] in the IAM
	// User Guide. This topic includes a list of Amazon Web Services services that can
	// use the server certificates that you manage with IAM.
	//
	// For information about the number of server certificates you can upload, see [IAM and STS quotas] in
	// the IAM User Guide.
	//
	// Because the body of the public key certificate, private key, and the
	// certificate chain can be large, you should use POST rather than GET when calling
	// UploadServerCertificate . For information about setting up signatures and
	// authorization through the API, see [Signing Amazon Web Services API requests]in the Amazon Web Services General
	// Reference. For general information about using the Query API with IAM, see [Calling the API by making HTTP query requests]in
	// the IAM User Guide.
	//
	// [Certificate Manager]: https://docs.aws.amazon.com/acm/
	// [Certificate Manager User Guide]: https://docs.aws.amazon.com/acm/latest/userguide/
	// [IAM and STS quotas]: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-quotas.html
	// [Working with server certificates]: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_server-certs.html
	// [Signing Amazon Web Services API requests]: https://docs.aws.amazon.com/general/latest/gr/signing_aws_api_requests.html
	// [Calling the API by making HTTP query requests]: https://docs.aws.amazon.com/IAM/latest/UserGuide/programming.html
	UploadServerCertificate(
		ctx context.Context,
		params *iam.UploadServerCertificateInput,
		optFns ...func(*iam.Options),
	) (*iam.UploadServerCertificateOutput, error)
}

// NewService creates a new instance of the AWS IAM service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return iam.NewFromConfig(
		*awsConfig,
		iam.WithEndpointResolverV2(
			&iamEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type iamEndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *iamEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params iam.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	iamAliases := utils.Services["iam"]
	iamEndpoint, hasIamEndpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"iam",
		iamAliases,
	)
	if hasIamEndpoint && !core.IsScalarNil(iamEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(iamEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return iam.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
