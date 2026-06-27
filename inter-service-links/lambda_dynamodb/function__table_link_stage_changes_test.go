//go:build unit

package lambdadynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FunctionTableLinkStageChangesSuite struct {
	suite.Suite
}

const ftEnvVarName = "ORDERS_TABLE_NAME"

func functionTableStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service],
) provider.Link {
	build := FunctionDynamoDBTableLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service {
			return iammock.CreateIamServiceMock()
		},
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, dynamodbservice.Service],
	) provider.Link {
		return build(FunctionToDynamoDBTableLinkDeps(deps))
	}
}

func callerFunctionWithEnvVarAnnotation() provider.ResourceInfo {
	return provider.ResourceInfo{
		ResourceName: "processOrdersFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.dynamodb.ordersTable.envVarName": core.MappingNodeFromString(
							ftEnvVarName,
						),
					},
				},
			},
		},
	}
}

func (s *FunctionTableLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		stageFunctionTableEnvVarChangeTestCase(),
		stageFunctionTableNoChangesTestCase(),
		stageFunctionTableDisableEnvVarsTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionTableStageLinkFactory(),
		&s.Suite,
	)
}

func stageFunctionTableEnvVarChangeTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name: "stages env var change for the linked table",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithEnvVarAnnotation(),
			},
			// ResourceBChanges represents the DynamoDB table from which the
			// table name is sourced into the function's environment variables.
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "ordersTable",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"tableName": core.MappingNodeFromString("orders-table"),
							},
						},
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.tableName",
						NewValue:  core.MappingNodeFromString("orders-table"),
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "processOrdersFunction.environmentVariables[\"" + ftEnvVarName + "\"]",
						NewValue:  core.MappingNodeFromString("orders-table"),
					},
				},
				// The table is new (it has NewFields), so the execution-role
				// permission data for this link is only known at deploy time.
				FieldChangesKnownOnDeploy: []string{"processOrdersFunctionExecutionRole"},
			},
		},
	}
}

func stageFunctionTableNoChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name: "stages no changes when the env var already matches",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithEnvVarAnnotation(),
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "ordersTable",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"tableName": core.MappingNodeFromString("orders-table"),
							},
						},
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"processOrdersFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									ftEnvVarName: core.MappingNodeFromString("orders-table"),
								},
							},
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				UnchangedFields: []string{
					"processOrdersFunction.environmentVariables[\"" + ftEnvVarName + "\"]",
				},
			},
		},
	}
}

func stageFunctionTableDisableEnvVarsTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	callerInfo := provider.ResourceInfo{
		ResourceName: "processOrdersFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.dynamodb.ordersTable.envVarName": core.MappingNodeFromString(
							ftEnvVarName,
						),
						"aws.lambda.dynamodb.ordersTable.populateEnvVars": core.MappingNodeFromBool(false),
					},
				},
			},
		},
	}

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name: "removes the env var when env var population is disabled",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerInfo,
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "ordersTable",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"tableName": core.MappingNodeFromString("orders-table"),
							},
						},
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"processOrdersFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									ftEnvVarName: core.MappingNodeFromString("orders-table"),
								},
							},
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				RemovedFields: []string{
					"processOrdersFunction.environmentVariables[\"" + ftEnvVarName + "\"]",
				},
			},
		},
	}
}

func TestFunctionTableLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionTableLinkStageChangesSuite))
}
