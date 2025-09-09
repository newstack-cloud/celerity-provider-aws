package integration

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
)

type AWSAccountInfo struct {
	AccountID string
	Region    string
}

// GetAWSAccountInfo gets the current AWS account information from the given context and loader.
func GetAWSAccountInfo(ctx context.Context, loader utils.AWSConfigLoader) (*AWSAccountInfo, error) {
	awsConfig, err := loader.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsService := sts.NewFromConfig(awsConfig)
	callerIdentityOutput, err := stsService.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}

	return &AWSAccountInfo{
		AccountID: aws.ToString(callerIdentityOutput.Account),
		Region:    awsConfig.Region,
	}, nil
}
