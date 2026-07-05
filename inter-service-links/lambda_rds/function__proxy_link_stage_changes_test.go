//go:build unit

package lambdards

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FunctionProxyLinkStageChangesSuite struct {
	suite.Suite
}

const fpStageHostPath = "apiFunction.environmentVariables[\"" + fpPrefix + "_HOST\"]"

func functionProxyStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
) provider.Link {
	return functionProxyLinkFactory(iammock.CreateIamServiceMock(), noopEC2ServiceFactory())
}

func proxyCaller(authMode string) provider.ResourceInfo {
	fields := map[string]*core.MappingNode{
		"aws.lambda.rds.ordersProxy.envVarPrefix": core.MappingNodeFromString(fpPrefix),
	}
	if authMode != "" {
		fields["aws.lambda.rds.ordersProxy.authMode"] = core.MappingNodeFromString(authMode)
	}
	return provider.ResourceInfo{
		ResourceName: "apiFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: fields},
			},
		},
	}
}

func proxyCallerVPC(authMode string) provider.ResourceInfo {
	caller := proxyCaller(authMode)
	caller.ResourceWithResolvedSubs.Spec = core.MappingNodeFields(
		"vpcConfig", core.MappingNodeFields(
			"subnetIds", &core.MappingNode{Items: []*core.MappingNode{
				core.MappingNodeFromString("subnet-1"),
			}},
		),
	)
	return caller
}

func proxyResourceBChanges(withNewField bool) *provider.Changes {
	changes := &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: "ordersProxy",
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: core.MappingNodeFields("endpoint", core.MappingNodeFromString(testProxyEndpoint)),
			},
		},
	}
	if withNewField {
		changes.NewFields = []provider.FieldChange{
			{FieldPath: "spec.endpoint", NewValue: core.MappingNodeFromString(testProxyEndpoint)},
		}
	}
	return changes
}

func (s *FunctionProxyLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		{
			Name: "stages host env var change from the proxy endpoint",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: proxyCaller("password")},
				ResourceBChanges: proxyResourceBChanges(true),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: fpStageHostPath, NewValue: core.MappingNodeFromString(testProxyEndpoint)},
					},
				},
			},
		},
		{
			Name: "adds exec-role known-on-deploy for iam auth on a new proxy",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: proxyCaller("iam")},
				ResourceBChanges: proxyResourceBChanges(true),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: fpStageHostPath, NewValue: core.MappingNodeFromString(testProxyEndpoint)},
					},
					FieldChangesKnownOnDeploy: []string{"apiFunctionExecutionRole"},
				},
			},
		},
		{
			Name: "surfaces network access known-on-deploy for a VPC-attached function",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: proxyCallerVPC("password")},
				ResourceBChanges: proxyResourceBChanges(true),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: fpStageHostPath, NewValue: core.MappingNodeFromString(testProxyEndpoint)},
					},
					FieldChangesKnownOnDeploy: []string{"apiFunctionNetworkAccess"},
				},
			},
		},
		{
			Name: "stages no change when the host env var already matches",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: proxyCaller("password")},
				ResourceBChanges: proxyResourceBChanges(false),
				CurrentLinkState: &state.LinkState{
					LinkID: "test-link",
					Data: map[string]*core.MappingNode{
						"apiFunction": {
							Fields: map[string]*core.MappingNode{
								"environmentVariables": {
									Fields: map[string]*core.MappingNode{
										fpPrefix + "_HOST": core.MappingNodeFromString(testProxyEndpoint),
									},
								},
							},
						},
					},
				},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					UnchangedFields: []string{fpStageHostPath},
				},
			},
		},
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionProxyStageLinkFactory(),
		&s.Suite,
	)
}

func TestFunctionProxyLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionProxyLinkStageChangesSuite))
}
