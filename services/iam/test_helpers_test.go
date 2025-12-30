//go:build unit

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	resgrouptagtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// mockResourceGroupTaggingService is a mock implementation of the Resource Groups Tagging API
// service for testing purposes. It returns empty results by default.
type mockResourceGroupTaggingService struct{}

func (m *mockResourceGroupTaggingService) GetResources(
	ctx context.Context,
	input *resourcegroupstaggingapi.GetResourcesInput,
	optFns ...func(*resourcegroupstaggingapi.Options),
) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	return &resourcegroupstaggingapi.GetResourcesOutput{
		ResourceTagMappingList: []resgrouptagtypes.ResourceTagMapping{},
	}, nil
}

// mockResourceGroupTaggingServiceFactory creates a mock Resource Groups Tagging API
// service factory for testing.
func mockResourceGroupTaggingServiceFactory(
	config *aws.Config,
	ctx provider.Context,
) resgrouptagservice.Service {
	return &mockResourceGroupTaggingService{}
}
