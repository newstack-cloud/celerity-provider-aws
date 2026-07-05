//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsprovider "github.com/newstack-cloud/bluelink-provider-aws/provider"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	kmsservice "github.com/newstack-cloud/bluelink-provider-aws/services/kms/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	s3service "github.com/newstack-cloud/bluelink-provider-aws/services/s3/service"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

// nameCounter guarantees uniqueness of generated names within a single test
// process even when two harnesses are created in the same nanosecond.
var nameCounter atomic.Uint64

// Harness is a minimal direct-call integration harness for the few data sources
// that cannot be exercised through the blueprint-driven e2e suite (e.g. the Lambda
// layer version data source, whose layer content cannot be created by a blueprint
// and whose filter uses an integer version number). It wires the AWS provider in
// process so a data source can be looked up and fetched directly. The blueprint
// e2e suite under tests/e2e is the primary integration mechanism.
type Harness struct {
	T           *testing.T
	Ctx         context.Context
	Provider    provider.Provider
	ProviderCtx provider.Context
	AWSConfig   aws.Config
	Account     *AWSAccountInfo
	Region      string
	NamePrefix  string
}

// Setup builds the provider and resolves AWS account info. It skips the test when
// AWS_REGION is not configured, so the integration build can be compiled and run
// selectively without credentials.
func Setup(t *testing.T) *Harness {
	t.Helper()

	region := os.Getenv("AWS_REGION")
	if region == "" {
		t.Skip("AWS_REGION not set; skipping integration test")
	}

	ctx := context.Background()
	loader := &utils.DefaultAWSConfigLoader{}

	account, err := GetAWSAccountInfo(ctx, loader)
	if err != nil {
		t.Fatalf("failed to resolve AWS account info: %v", err)
	}

	awsConfig, err := loader.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("failed to load AWS config: %v", err)
	}

	configStore := utils.NewAWSConfigStore(
		os.Environ(),
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	suffix := uniqueSuffix()
	prov := awsprovider.NewProvider(
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
		configStore,
	)

	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString(region),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("it-session-" + suffix),
		},
	)

	return &Harness{
		T:           t,
		Ctx:         ctx,
		Provider:    prov,
		ProviderCtx: providerCtx,
		AWSConfig:   awsConfig,
		Account:     account,
		Region:      region,
		NamePrefix:  "bluelink-it-" + suffix,
	}
}

// Name returns a run-unique resource name combining the harness prefix with a
// per-resource label, e.g. "bluelink-it-<suffix>-ds-layer".
func (h *Harness) Name(label string) string {
	return fmt.Sprintf("%s-%s", h.NamePrefix, label)
}

func uniqueSuffix() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), nameCounter.Add(1))
}
