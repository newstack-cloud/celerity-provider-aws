//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
)

// AWSAccountInfo holds the identity of the AWS account the integration tests
// run against, resolved at setup time via STS.
type AWSAccountInfo struct {
	AccountID string
	Region    string
}

// GetAWSAccountInfo resolves the current AWS account ID and region from the
// credentials/config loaded by the given loader. Integration tests use it to
// build the ARNs/URLs they expect real resources to be created with.
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
