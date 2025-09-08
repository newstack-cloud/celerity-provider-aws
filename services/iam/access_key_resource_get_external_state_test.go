//go:build unit

package iam

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMAccessKeyResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *IAMAccessKeyResourceGetExternalStateSuite) Test_get_external_state_iam_access_key() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		getExternalStateAccessKeyTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		AccessKeyResource,
		&s.Suite,
	)
}

func getExternalStateAccessKeyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service] {

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, iamservice.Service]{
		Name: "Get external state for IAM access key",
		ServiceFactory: iammock.CreateIamServiceMockFactory(
			iammock.WithListAccessKeysOutput(&iam.ListAccessKeysOutput{
				AccessKeyMetadata: []types.AccessKeyMetadata{
					{
						AccessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE"),
						Status:      types.StatusTypeActive,
						UserName:    aws.String("john.doe"),
					},
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-access-key-id",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id":       core.MappingNodeFromString("AKIAIOSFODNN7EXAMPLE"),
					"userName": core.MappingNodeFromString("john.doe"),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id":       core.MappingNodeFromString("AKIAIOSFODNN7EXAMPLE"),
					"userName": core.MappingNodeFromString("john.doe"),
					"status":   core.MappingNodeFromString("Active"),
				},
			},
		},
	}
}

func TestIAMAccessKeyResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(IAMAccessKeyResourceGetExternalStateSuite))
}
