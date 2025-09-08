//go:build unit

package iam

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type IAMManagedPolicyResourceUpdateSuite struct {
	suite.Suite
}

func (s *IAMManagedPolicyResourceUpdateSuite) Test_update_iam_managed_policy() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		recreateManagedPolicyOnNameOrPathChangeTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		func(serviceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service], configStore pluginutils.ServiceConfigStore[*aws.Config]) provider.Resource {
			return ManagedPolicyResource(serviceFactory, configStore)
		},
		&s.Suite,
	)
}

func recreateManagedPolicyOnNameOrPathChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	resourceARN := "arn:aws:iam::123456789012:policy/OldPolicy"
	policyId := "ANPA1234567890123456"
	staticTime := time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC)
	timestamp := staticTime.Format("2006-01-02T15:04:05Z")

	service := iammock.CreateIamServiceMock(
		iammock.WithDeletePolicyOutput(&iam.DeletePolicyOutput{}),
		iammock.WithCreatePolicyOutput(&iam.CreatePolicyOutput{
			Policy: &types.Policy{
				Arn:                           aws.String(resourceARN),
				PolicyId:                      aws.String(policyId),
				PolicyName:                    aws.String("NewPolicy"),
				Path:                          aws.String("/newpath/"),
				CreateDate:                    aws.Time(staticTime),
				UpdateDate:                    aws.Time(staticTime),
				AttachmentCount:               aws.Int32(0),
				DefaultVersionId:              aws.String("v1"),
				IsAttachable:                  true,
				PermissionsBoundaryUsageCount: aws.Int32(0),
			},
		}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"policyName": core.MappingNodeFromString("OldPolicy"),
			"path":       core.MappingNodeFromString("/oldpath/"),
			"policyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Action": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("s3:GetObject"),
										},
									},
									"Resource": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("*"),
										},
									},
								},
							},
						},
					},
				},
			},
			"arn": core.MappingNodeFromString(resourceARN),
			"id":  core.MappingNodeFromString(policyId),
		},
	}
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"policyName": core.MappingNodeFromString("NewPolicy"),
			"path":       core.MappingNodeFromString("/newpath/"),
			"policyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Action": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("s3:GetObject"),
										},
									},
									"Resource": {
										Items: []*core.MappingNode{
											core.MappingNodeFromString("*"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "recreate managed policy on policyName or path change",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-policy-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-policy-id",
					ResourceName: "TestPolicy",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-policy-id",
						Name:       "TestPolicy",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/managedPolicy",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.policyName",
						PrevValue: core.MappingNodeFromString("OldPolicy"),
						NewValue:  core.MappingNodeFromString("NewPolicy"),
					},
					{
						FieldPath: "spec.path",
						PrevValue: core.MappingNodeFromString("/oldpath/"),
						NewValue:  core.MappingNodeFromString("/newpath/"),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":                           core.MappingNodeFromString(resourceARN),
				"spec.id":                            core.MappingNodeFromString(policyId),
				"spec.attachmentCount":               core.MappingNodeFromInt(0),
				"spec.createDate":                    core.MappingNodeFromString(timestamp),
				"spec.defaultVersionId":              core.MappingNodeFromString("v1"),
				"spec.isAttachable":                  core.MappingNodeFromBool(true),
				"spec.permissionsBoundaryUsageCount": core.MappingNodeFromInt(0),
				"spec.updateDate":                    core.MappingNodeFromString(timestamp),
			},
		},
		SaveActionsCalled: map[string]any{
			"DeletePolicy": &iam.DeletePolicyInput{
				PolicyArn: aws.String(resourceARN),
			},
			"CreatePolicy": func(actual any) (plugintestutils.EqualityCheckValues, error) {
				input, ok := actual.(*iam.CreatePolicyInput)
				if !ok {
					return plugintestutils.EqualityCheckValues{}, fmt.Errorf("input is not an *iam.CreatePolicyInput")
				}
				// Unmarshal the policy document for comparison
				var actualDoc map[string]any
				if err := json.Unmarshal([]byte(*input.PolicyDocument), &actualDoc); err != nil {
					return plugintestutils.EqualityCheckValues{}, err
				}
				expectedDoc := map[string]any{
					"Version": "2012-10-17",
					"Statement": []any{
						map[string]any{
							"Effect":   "Allow",
							"Action":   []any{"s3:GetObject"},
							"Resource": []any{"*"},
						},
					},
				}
				expectedInput := &iam.CreatePolicyInput{
					PolicyName: aws.String("NewPolicy"),
					Path:       aws.String("/newpath/"),
				}
				actualInput := &iam.CreatePolicyInput{
					PolicyName: input.PolicyName,
					Path:       input.Path,
				}
				expectedMap := map[string]any{
					"PolicyName":     *expectedInput.PolicyName,
					"Path":           *expectedInput.Path,
					"PolicyDocument": expectedDoc,
				}
				actualMap := map[string]any{
					"PolicyName":     *actualInput.PolicyName,
					"Path":           *actualInput.Path,
					"PolicyDocument": actualDoc,
				}
				return plugintestutils.EqualityCheckValues{
					Expected: expectedMap,
					Actual:   actualMap,
				}, nil
			},
		},
	}
}

func TestIAMManagedPolicyResourceUpdate(t *testing.T) {
	suite.Run(t, new(IAMManagedPolicyResourceUpdateSuite))
}
