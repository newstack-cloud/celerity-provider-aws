package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/two-hundred/celerity-provider-aws/provider"
	"github.com/two-hundred/celerity-provider-aws/services/lambda"
	"github.com/two-hundred/celerity-provider-aws/utils"
	"github.com/two-hundred/celerity/libs/plugin-framework/plugin"
	"github.com/two-hundred/celerity/libs/plugin-framework/pluginservicev1"
	"github.com/two-hundred/celerity/libs/plugin-framework/sdk/pluginutils"
	"github.com/two-hundred/celerity/libs/plugin-framework/sdk/providerv1"
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
			lambda.NewService,
			utils.NewAWSConfigStore(
				os.Environ(),
				utils.AWSConfigFromProviderContext,
				&utils.DefaultAWSConfigLoader{},
			),
		),
		hostInfoContainer,
		serviceClient,
	)

	providerDescription, _ := embedded.ReadFile("provider_description.md")
	config := plugin.ServePluginConfiguration{
		ID: "two-hundred/aws",
		PluginMetadata: &pluginservicev1.PluginMetadata{
			PluginVersion:        "1.0.0",
			DisplayName:          "AWS",
			FormattedDescription: string(providerDescription),
			RepositoryUrl:        "https://github.com/two-hundred/celerity-provider-aws",
			Author:               "Two Hundred",
		},
		ProtocolVersion: "1.0",
	}

	fmt.Println("Starting Celerity AWS Provider Plugin Server...")
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
