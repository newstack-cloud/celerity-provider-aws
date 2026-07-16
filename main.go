package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/newstack-cloud/bluelink-provider-aws/provider"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	elasticacheservice "github.com/newstack-cloud/bluelink-provider-aws/services/elasticache/service"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	kmsservice "github.com/newstack-cloud/bluelink-provider-aws/services/kms/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	s3service "github.com/newstack-cloud/bluelink-provider-aws/services/s3/service"
	secretsmanagerservice "github.com/newstack-cloud/bluelink-provider-aws/services/secretsmanager/service"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
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
			dynamodbservice.NewService,
			eventsservice.NewService,
			cloudcontrolservice.NewService,
			s3service.NewService,
			ssmservice.NewService,
			kmsservice.NewService,
			elasticacheservice.NewService,
			secretsmanagerservice.NewService,
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
			PluginVersion:        version,
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
