package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

type serviceEntry struct {
	// Name is the CloudFormation service segment, e.g. "SQS", "DynamoDB".
	Name string
	// Exclude lists fully-qualified CFN type names within the service to skip
	// during sync (e.g. types whose published schema is malformed upstream).
	Exclude []string
	// TypeOverrides maps a fully-qualified CFN type name to an explicit Bluelink
	// type, for the cases the derived type is wrong (e.g. awkward acronym casing) or
	// where an established type name must be preserved.
	TypeOverrides map[string]string
	// Include, when non-empty, restricts sync and generation to exactly these
	// fully-qualified CFN type names within the service (an allowlist) instead of the
	// whole service. Used for large services where only a handful of types are wanted.
	Include []string
	// DataSourceOnly, when true, makes the service contribute only the opted-in data
	// sources from dataSourceConfigs: no managed resources are emitted or registered.
	// Used where another mechanism owns the resources (e.g. flex/vpc owns EC2
	// networking) but practitioners still need to look up existing infrastructure.
	DataSourceOnly bool
}

// The set of AWS services the generator onboards. Syncing discovers
// every provisionable public CloudFormation resource type under each service and
// vendors its schema; generation then runs over whatever is vendored. Add a
// service here and run `awsgen -sync` (requires AWS credentials) to pull its full
// resource set.
var services = []serviceEntry{
	{Name: "SQS"},
	{Name: "DynamoDB"},
	{
		Name: "IAM",
		// Preserve readable casing for the provider acronym types (the derived names
		// would be "oIDCProvider"/"sAMLProvider").
		TypeOverrides: map[string]string{
			"AWS::IAM::OIDCProvider": "aws/iam/oidcProvider",
			"AWS::IAM::SAMLProvider": "aws/iam/samlProvider",
		},
	},
	{
		Name: "Lambda",
		// Preserve the established, clearer type names (the derived names would be
		// the bare "version"/"url").
		TypeOverrides: map[string]string{
			"AWS::Lambda::Version": "aws/lambda/functionVersion",
			"AWS::Lambda::Url":     "aws/lambda/functionUrl",
		},
	},
	{Name: "Events"},
	{
		Name: "RDS",
		Include: []string{
			"AWS::RDS::DBInstance",
		},
		TypeOverrides: map[string]string{
			"AWS::RDS::DBInstance": "aws/rds/dbInstance",
		},
	},
	{
		// CloudWatch Logs: handler log groups.
		Name: "Logs",
		Include: []string{
			"AWS::Logs::LogGroup",
		},
	},
	{
		// SNS: pub/sub topics (celerity/topic) and their subscriptions. TopicPolicy is
		// not Cloud Control–provisionable, so topic access policies are handled via the
		// topic's own Policy attribute / link intermediaries where needed.
		Name: "SNS",
		Include: []string{
			"AWS::SNS::Topic",
			"AWS::SNS::Subscription",
		},
	},
	{
		// Kinesis: data streams as an external/stream event source for consumers
		// (celerity/consumer). Only the Stream type is needed; consumer groups and
		// stream consumers (enhanced fan-out) will come later.
		Name: "Kinesis",
		Include: []string{
			"AWS::Kinesis::Stream",
		},
	},
	{
		// EC2 is onboarded for data-source lookups only, the flex/vpc abstraction owns
		// the networking fabric, so no managed EC2 resources are emitted. Only the
		// types needed for existing-infrastructure lookups are synced.
		Name:           "EC2",
		DataSourceOnly: true,
		Include: []string{
			"AWS::EC2::VPC",
			"AWS::EC2::Subnet",
			"AWS::EC2::SecurityGroup",
		},
		// Preserve readable casing for the acronym type (the derived name would be "vPC").
		TypeOverrides: map[string]string{
			"AWS::EC2::VPC": "aws/ec2/vpc",
		},
	},
}

type dataSourceConfig struct {
	// FilterFields are the practitioner-facing filterable fields, matching the
	// hand-written data sources they replace. "region" selects the AWS region.
	FilterFields []string
	// DeriveIdentifierFromARN enables the GetResource fast path where a single `arn`
	// equality filter resolves to the primary identifier by ARN suffix. Set for
	// name-style identifiers; false where the identifier is the ARN itself (events
	// rule) or a non-derivable form (sqs queue URL).
	DeriveIdentifierFromARN bool
}

var dataSourceConfigs = map[string]dataSourceConfig{
	"AWS::Events::EventBus": {
		FilterFields:            []string{"name", "arn", "region"},
		DeriveIdentifierFromARN: true,
	},
	"AWS::Events::Rule": {
		FilterFields:            []string{"name", "eventBusName", "arn", "region"},
		DeriveIdentifierFromARN: false,
	},
	"AWS::SQS::Queue": {
		FilterFields:            []string{"queueName", "queueUrl", "arn", "region"},
		DeriveIdentifierFromARN: false,
	},
	"AWS::DynamoDB::Table": {
		FilterFields:            []string{"tableName", "arn", "region"},
		DeriveIdentifierFromARN: true,
	},
	"AWS::DynamoDB::GlobalTable": {
		FilterFields:            []string{"tableName", "arn", "region"},
		DeriveIdentifierFromARN: true,
	},
	"AWS::Logs::LogGroup": {
		FilterFields:            []string{"logGroupName", "arn", "region"},
		DeriveIdentifierFromARN: false,
	},
	"AWS::SNS::Topic": {
		FilterFields:            []string{"topicArn", "topicName", "region"},
		DeriveIdentifierFromARN: false,
	},
	"AWS::Kinesis::Stream": {
		FilterFields:            []string{"name", "arn", "region"},
		DeriveIdentifierFromARN: true,
	},
	// EC2 lookups (data-source-only service) for referencing existing networking
	// infrastructure from blueprints. Primary identifiers (vpcId/subnetId/id) take the
	// GetResource fast path; the friendlier fields resolve via ListResources + filter.
	"AWS::EC2::VPC": {
		FilterFields:            []string{"vpcId", "cidrBlock", "region"},
		DeriveIdentifierFromARN: false,
	},
	"AWS::EC2::Subnet": {
		FilterFields:            []string{"subnetId", "vpcId", "availabilityZone", "region"},
		DeriveIdentifierFromARN: false,
	},
	"AWS::EC2::SecurityGroup": {
		FilterFields:            []string{"id", "groupId", "groupName", "vpcId", "region"},
		DeriveIdentifierFromARN: false,
	},
}

func dataSourceConfigFor(cfnType string) (dataSourceConfig, bool) {
	cfg, ok := dataSourceConfigs[cfnType]
	return cfg, ok
}

func (s serviceEntry) excludes(cfnType string) bool {
	return slices.Contains(s.Exclude, cfnType)
}

func (s serviceEntry) includes(cfnType string) bool {
	return len(s.Include) == 0 || slices.Contains(s.Include, cfnType)
}

func dataSourceOnlyType(cfnType string) bool {
	svc, ok := serviceFor(cfnType)
	return ok && svc.DataSourceOnly
}

// cfnPrefix returns the type-name prefix for a service, e.g. "AWS::SQS::".
func (s serviceEntry) cfnPrefix() string {
	return fmt.Sprintf("AWS::%s::", s.Name)
}

// serviceFor returns the service entry whose segment matches a CFN type, if any.
func serviceFor(cfnType string) (serviceEntry, bool) {
	parts := strings.Split(cfnType, "::")
	if len(parts) < 2 {
		return serviceEntry{}, false
	}
	for _, svc := range services {
		if strings.EqualFold(svc.Name, parts[1]) {
			return svc, true
		}
	}
	return serviceEntry{}, false
}

func deriveBlueprintType(cfnType string) string {
	parts := strings.Split(cfnType, "::")
	if len(parts) < 3 {
		return strings.ToLower(strings.ReplaceAll(cfnType, "::", "/"))
	}
	return fmt.Sprintf("aws/%s/%s", strings.ToLower(parts[1]), lowerFirst(parts[2]))
}

func blueprintTypeFor(cfnType string) string {
	if svc, ok := serviceFor(cfnType); ok {
		if override, ok := svc.TypeOverrides[cfnType]; ok {
			return override
		}
	}
	return deriveBlueprintType(cfnType)
}

func vendoredSchemaFile(cfnType string) string {
	return strings.ToLower(strings.ReplaceAll(cfnType, "::", "-")) + ".json"
}

func vendoredSchemaFiles(schemasDir string) ([]string, error) {
	entries, err := os.ReadDir(schemasDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)
	return files, nil
}

// Returns the per-type output file name, e.g.
// "AWS::DynamoDB::Table" -> "dynamodb_table.go".
func schemaFileName(cfnType string) string {
	return exampleStem(cfnType) + ".go"
}

func exampleStem(cfnType string) string {
	parts := strings.Split(cfnType, "::")
	if len(parts) < 3 {
		return strings.ToLower(strings.ReplaceAll(cfnType, "::", "_"))
	}
	return fmt.Sprintf("%s_%s", strings.ToLower(parts[1]), strings.ToLower(parts[2]))
}
