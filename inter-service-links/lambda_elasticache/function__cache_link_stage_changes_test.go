//go:build unit

package lambdaelasticache

import (
	"maps"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionCacheLinkStageChangesSuite struct {
	suite.Suite
}

const lcStageHostPath = "apiFunction.environmentVariables[\"" + lcPrefix + "_HOST\"]"

func cacheCaller(extra map[string]*core.MappingNode) provider.ResourceInfo {
	fields := map[string]*core.MappingNode{
		"aws.lambda.elasticache.sessionCache.envVarPrefix": core.MappingNodeFromString(lcPrefix),
	}
	maps.Copy(fields, extra)
	return provider.ResourceInfo{
		ResourceName: "apiFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: fields},
			},
		},
	}
}

func cacheCallerVPC() provider.ResourceInfo {
	caller := cacheCaller(nil)
	caller.ResourceWithResolvedSubs.Spec = core.MappingNodeFields(
		"vpcConfig", core.MappingNodeFields(
			"subnetIds", &core.MappingNode{Items: []*core.MappingNode{
				core.MappingNodeFromString("subnet-1"),
			}},
		),
	)
	return caller
}

// Resolves a single-instance cache (primary endpoint only).
func primaryCacheBChanges() *provider.Changes {
	return &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: "sessionCache",
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: core.MappingNodeFields(
					"primaryEndPoint", core.MappingNodeFields(
						"address", core.MappingNodeFromString(testCacheEndpoint),
					),
				),
			},
		},
		NewFields: []provider.FieldChange{
			{FieldPath: "spec.primaryEndPoint.address", NewValue: core.MappingNodeFromString(testCacheEndpoint)},
		},
	}
}

// Resolves a cluster-mode cache (configuration endpoint present).
func clusterCacheBChanges() *provider.Changes {
	return &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: "sessionCache",
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: core.MappingNodeFields(
					"configurationEndPoint", core.MappingNodeFields(
						"address", core.MappingNodeFromString(testCacheConfigEndpoint),
					),
				),
			},
		},
		NewFields: []provider.FieldChange{
			{FieldPath: "spec.configurationEndPoint.address", NewValue: core.MappingNodeFromString(testCacheConfigEndpoint)},
		},
	}
}

func (s *FunctionCacheLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		{
			Name: "stages host env var change from the primary endpoint",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: cacheCaller(nil)},
				ResourceBChanges: primaryCacheBChanges(),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: lcStageHostPath, NewValue: core.MappingNodeFromString(testCacheEndpoint)},
					},
				},
			},
		},
		{
			Name: "stages host env var change from the configuration endpoint in cluster mode",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: cacheCaller(nil)},
				ResourceBChanges: clusterCacheBChanges(),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: lcStageHostPath, NewValue: core.MappingNodeFromString(testCacheConfigEndpoint)},
					},
				},
			},
		},
		{
			Name: "adds exec-role known-on-deploy for iam auth on a new cache",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: cacheCaller(map[string]*core.MappingNode{
					"aws.lambda.elasticache.sessionCache.authMode": core.MappingNodeFromString("iam"),
				})},
				ResourceBChanges: primaryCacheBChanges(),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: lcStageHostPath, NewValue: core.MappingNodeFromString(testCacheEndpoint)},
					},
					FieldChangesKnownOnDeploy: []string{"apiFunctionExecutionRole"},
				},
			},
		},
		{
			Name: "surfaces network access known-on-deploy for a VPC-attached function",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: cacheCallerVPC()},
				ResourceBChanges: primaryCacheBChanges(),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: lcStageHostPath, NewValue: core.MappingNodeFromString(testCacheEndpoint)},
					},
					FieldChangesKnownOnDeploy: []string{"apiFunctionNetworkAccess"},
				},
			},
		},
		{
			Name: "stages no change when the host env var already matches",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: cacheCaller(nil)},
				ResourceBChanges: primaryCacheBChanges(),
				CurrentLinkState: &state.LinkState{
					LinkID: "test-link",
					Data: map[string]*core.MappingNode{
						"apiFunction": {
							Fields: map[string]*core.MappingNode{
								"environmentVariables": {
									Fields: map[string]*core.MappingNode{
										lcPrefix + "_HOST": core.MappingNodeFromString(testCacheEndpoint),
									},
								},
							},
						},
					},
				},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					UnchangedFields: []string{lcStageHostPath},
				},
			},
		},
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionCacheLinkFactory(iammock.CreateIamServiceMock(), noopEC2ServiceFactory()),
		&s.Suite,
	)
}

func TestFunctionCacheLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionCacheLinkStageChangesSuite))
}
