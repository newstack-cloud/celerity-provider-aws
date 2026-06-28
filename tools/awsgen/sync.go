package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// The CloudFormation provisioning behaviours that Cloud
// Control can create — NON_PROVISIONABLE types are skipped.
var provisionableTypes = []cfntypes.ProvisioningType{
	cfntypes.ProvisioningTypeFullyMutable,
	cfntypes.ProvisioningTypeImmutable,
}

// The subset of the CloudFormation client the sync uses, so it can
// be faked in tests.
type cfnRegistry interface {
	ListTypes(context.Context, *cloudformation.ListTypesInput, ...func(*cloudformation.Options)) (*cloudformation.ListTypesOutput, error)
	DescribeType(context.Context, *cloudformation.DescribeTypeInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeTypeOutput, error)
}

// Discovers every provisionable public CloudFormation resource type
// for each configured service and vendors its schema into the schemas directory,
// validating each one before writing. Requires AWS credentials.
func syncSchemas(schemasDir, region, serviceFilter string) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("loading AWS config: %w", err)
	}
	return syncSchemasWith(ctx, cloudformation.NewFromConfig(cfg), schemasDir, region, serviceFilter)
}

// Vendors the onboarded services' schemas. When serviceFilter is non-empty,
// only the matching service is synced, so a single slice can vendor its new service without
// re-pulling (and potentially churning) the rest of the committed baseline.
func syncSchemasWith(ctx context.Context, client cfnRegistry, schemasDir, region, serviceFilter string) error {
	if err := os.MkdirAll(schemasDir, 0o755); err != nil {
		return err
	}

	total := 0
	for _, svc := range services {
		if serviceFilter != "" && !strings.EqualFold(svc.Name, serviceFilter) {
			continue
		}
		typeNames, err := discoverServiceTypes(ctx, client, svc)
		if err != nil {
			return fmt.Errorf("discovering %s types: %w", svc.Name, err)
		}
		for _, typeName := range typeNames {
			if err := syncType(ctx, client, schemasDir, typeName); err != nil {
				return fmt.Errorf("syncing %s: %w", typeName, err)
			}
			fmt.Printf("synced %s\n", typeName)
			total++
		}
		fmt.Printf("discovered %d provisionable type(s) for %s\n", len(typeNames), svc.Name)
	}

	fmt.Printf("synced %d schema(s) from the %s CloudFormation registry\n", total, region)
	return nil
}

// Lists the provisionable, non-deprecated public resource
// types under a service's "AWS::<Service>::" prefix, minus any excluded types.
func discoverServiceTypes(ctx context.Context, client cfnRegistry, svc serviceEntry) ([]string, error) {
	seen := map[string]bool{}
	for _, provisioning := range provisionableTypes {
		var nextToken *string
		for {
			out, err := client.ListTypes(ctx, &cloudformation.ListTypesInput{
				Type:             cfntypes.RegistryTypeResource,
				Visibility:       cfntypes.VisibilityPublic,
				ProvisioningType: provisioning,
				DeprecatedStatus: cfntypes.DeprecatedStatusLive,
				Filters: &cfntypes.TypeFilters{
					Category:       cfntypes.CategoryAwsTypes,
					TypeNamePrefix: aws.String(svc.cfnPrefix()),
				},
				MaxResults: aws.Int32(100),
				NextToken:  nextToken,
			})
			if err != nil {
				return nil, err
			}
			for _, summary := range out.TypeSummaries {
				name := aws.ToString(summary.TypeName)
				if name == "" || svc.excludes(name) || !svc.includes(name) {
					continue
				}
				seen[name] = true
			}
			if out.NextToken == nil {
				break
			}
			nextToken = out.NextToken
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func syncType(ctx context.Context, client cfnRegistry, schemasDir, typeName string) error {
	out, err := client.DescribeType(ctx, &cloudformation.DescribeTypeInput{
		Type:     cfntypes.RegistryTypeResource,
		TypeName: aws.String(typeName),
	})
	if err != nil {
		return err
	}
	schema := aws.ToString(out.Schema)
	if schema == "" {
		return fmt.Errorf("registry returned an empty schema")
	}

	// Validate before writing so a truncated or malformed schema never lands in
	// the vendored schemas.
	if _, err := loadCFNSchema([]byte(schema)); err != nil {
		return fmt.Errorf("schema is invalid: %w", err)
	}

	return os.WriteFile(filepath.Join(schemasDir, vendoredSchemaFile(typeName)), []byte(schema), 0o644)
}
