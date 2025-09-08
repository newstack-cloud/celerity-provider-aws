package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/newstack-cloud/bluelink-provider-aws/provider"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/plugin"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/pluginservicev1"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

//go:embed provider_description.md
var embedded embed.FS

func main() {
	serviceClient, closeService, err := pluginservicev1.NewEnvServiceClient()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer closeService()

	hostInfoContainer := pluginutils.NewHostInfoContainer()
	providerServer := providerv1.NewProviderPlugin(
		provider.NewProvider(
			iamservice.NewService,
			lambdaservice.NewService,
			ec2service.NewService,
			resgrouptagservice.NewService,
			sqsservice.NewService,
			utils.NewAWSConfigStore(
				os.Environ(),
				utils.AWSConfigFromProviderContext,
				&utils.DefaultAWSConfigLoader{},
				utils.AWSConfigCacheKey,
			),
		),
		hostInfoContainer,
		serviceClient,
	)

	providerDescription, _ := embedded.ReadFile("provider_description.md")
	config := plugin.ServePluginConfiguration{
		ID: "newstack-cloud/aws",
		PluginMetadata: &pluginservicev1.PluginMetadata{
			PluginVersion:        "0.1.0",
			DisplayName:          "AWS",
			FormattedDescription: string(providerDescription),
			RepositoryUrl:        "https://github.com/newstack-cloud/bluelink-provider-aws",
			Author:               "NewStack Cloud Limited",
		},
		ProtocolVersion: "1.0",
	}

	fmt.Println("Starting Bluelink AWS Provider Plugin Server...")
	close, err := plugin.ServeProviderV1(
		context.Background(),
		providerServer,
		serviceClient,
		hostInfoContainer,
		config,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	pluginutils.WaitForShutdown(close)
}
