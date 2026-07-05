package lambdakms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	kmsservice "github.com/newstack-cloud/bluelink-provider-aws/services/kms/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type keyGrantReconcile struct {
	granteeRoleARN string
	keyID          string
	grantName      string
	accessLevel    string
	// The desired end state: true means the grant should exist, false means it
	// should not (either never requested, or toggled off).
	manageGrant bool
	// isDestroy forces removal regardless of manageGrant.
	isDestroy bool
}

// Brings this link's KMS grant into the desired state. It is stateless: the
// existing grant is discovered by name (scoped to the grantee role), so create/update/revoke
// decisions are made against the live grant list rather than persisted link data.
func (l *functionKeyLinkActions) reconcileKeyGrant(
	ctx context.Context,
	providerCtx provider.Context,
	input keyGrantReconcile,
) error {
	kmsService, err := l.getKMSService(ctx, providerCtx)
	if err != nil {
		return err
	}

	existing, err := findGrantByName(ctx, kmsService, input.keyID, input.granteeRoleARN, input.grantName)
	if err != nil {
		return err
	}

	// The desired state is for there to be no grant (destroyed or unmanaged).
	// Revoke a previously-created one if present.
	if input.isDestroy || !input.manageGrant {
		if existing == nil {
			return nil
		}
		return revokeGrant(ctx, kmsService, input.keyID, aws.ToString(existing.GrantId))
	}

	// The desired state is for the grant to be present with the correct operations.
	desiredOperations := grantOperationsForAccessLevel(input.accessLevel)
	if existing == nil {
		return createGrant(ctx, kmsService, input.keyID, input.granteeRoleARN, input.grantName, desiredOperations)
	}
	if !grantOperationsEqual(existing.Operations, desiredOperations) {
		if err := revokeGrant(ctx, kmsService, input.keyID, aws.ToString(existing.GrantId)); err != nil {
			return err
		}
		return createGrant(ctx, kmsService, input.keyID, input.granteeRoleARN, input.grantName, desiredOperations)
	}
	return nil
}

func findGrantByName(
	ctx context.Context,
	service kmsservice.Service,
	keyID, granteeRoleARN, grantName string,
) (*kmstypes.GrantListEntry, error) {
	var marker *string
	for {
		output, err := service.ListGrants(ctx, &kms.ListGrantsInput{
			KeyId:            aws.String(keyID),
			GranteePrincipal: aws.String(granteeRoleARN),
			Marker:           marker,
		})
		if err != nil {
			return nil, err
		}
		if output == nil {
			return nil, nil
		}

		for i := range output.Grants {
			if aws.ToString(output.Grants[i].Name) == grantName {
				return &output.Grants[i], nil
			}
		}

		if !output.Truncated || output.NextMarker == nil {
			return nil, nil
		}
		marker = output.NextMarker
	}
}

func createGrant(
	ctx context.Context,
	service kmsservice.Service,
	keyID, granteeRoleARN, grantName string,
	operations []kmstypes.GrantOperation,
) error {
	_, err := service.CreateGrant(ctx, &kms.CreateGrantInput{
		KeyId:            aws.String(keyID),
		GranteePrincipal: aws.String(granteeRoleARN),
		Operations:       operations,
		Name:             aws.String(grantName),
	})
	return err
}

func revokeGrant(
	ctx context.Context,
	service kmsservice.Service,
	keyID, grantID string,
) error {
	_, err := service.RevokeGrant(ctx, &kms.RevokeGrantInput{
		KeyId:   aws.String(keyID),
		GrantId: aws.String(grantID),
	})
	return err
}

func keyGrantName(
	functionInfo *provider.ResourceInfo,
	keyInfo *provider.ResourceInfo,
) string {
	return fmt.Sprintf(
		"bluelink-%s-%s",
		pluginutils.StripNonAlphaNumericChars(functionInfo.ResourceName),
		pluginutils.StripNonAlphaNumericChars(keyInfo.ResourceName),
	)
}

// The link-data field under which this link records the KMS grant it
// manages, so grant creation/revocation is surfaced in staged changes and tracked in link
// state. The name and operations are known at plan time (derived from the resource names and
// accessLevel), independent of deployment.
const keyGrantLinkDataField = "keyGrant"

// keyGrantLinkDataNode is the link-data representation of the managed KMS grant.
func keyGrantLinkDataNode(
	functionInfo *provider.ResourceInfo,
	keyInfo *provider.ResourceInfo,
	accessLevel string,
) *core.MappingNode {
	operations := grantOperationsForAccessLevel(accessLevel)
	operationItems := make([]*core.MappingNode, len(operations))
	for i, operation := range operations {
		operationItems[i] = core.MappingNodeFromString(string(operation))
	}
	return core.MappingNodeFields(
		"name", core.MappingNodeFromString(keyGrantName(functionInfo, keyInfo)),
		"operations", &core.MappingNode{Items: operationItems},
	)
}

func grantOperationsForAccessLevel(accessLevel string) []kmstypes.GrantOperation {
	switch accessLevel {
	case "encryptDecrypt":
		return []kmstypes.GrantOperation{
			kmstypes.GrantOperationDecrypt,
			kmstypes.GrantOperationDescribeKey,
			kmstypes.GrantOperationEncrypt,
			kmstypes.GrantOperationGenerateDataKey,
			kmstypes.GrantOperationGenerateDataKeyWithoutPlaintext,
		}
	case "decrypt":
		fallthrough
	default:
		return []kmstypes.GrantOperation{
			kmstypes.GrantOperationDecrypt,
			kmstypes.GrantOperationDescribeKey,
		}
	}
}

func grantOperationsEqual(a, b []kmstypes.GrantOperation) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[kmstypes.GrantOperation]bool, len(a))
	for _, op := range a {
		set[op] = true
	}
	for _, op := range b {
		if !set[op] {
			return false
		}
	}
	return true
}
