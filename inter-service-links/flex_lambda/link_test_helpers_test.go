//go:build unit

package flexlambda

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func testLinkContext() provider.LinkContext {
	return plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {
				"region": core.ScalarFromString("us-west-2"),
			},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func testConfigStore(loader *testutils.MockAWSConfigLoader) pluginutils.ServiceConfigStore[*aws.Config] {
	return utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
}
