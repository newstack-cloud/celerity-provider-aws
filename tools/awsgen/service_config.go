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
		// RDS relational databases. DBInstance is the standalone primary (and read replicas /
		// Aurora member instances); DBCluster is the Aurora (Serverless v2) cluster; DBProxy +
		// DBProxyTargetGroup provide Lambda connection pooling for the standalone-instance path;
		// DBSubnetGroup groups the private subnets a database is placed in.
		// TypeOverrides preserve readable casing (the derived names would be
		// "dBInstance"/"dBCluster"/"dBProxy"/... from the DB acronym).
		Name: "RDS",
		Include: []string{
			"AWS::RDS::DBInstance",
			"AWS::RDS::DBCluster",
			"AWS::RDS::DBProxy",
			"AWS::RDS::DBProxyTargetGroup",
			"AWS::RDS::DBSubnetGroup",
		},
		TypeOverrides: map[string]string{
			"AWS::RDS::DBInstance":         "aws/rds/dbInstance",
			"AWS::RDS::DBCluster":          "aws/rds/dbCluster",
			"AWS::RDS::DBProxy":            "aws/rds/dbProxy",
			"AWS::RDS::DBProxyTargetGroup": "aws/rds/dbProxyTargetGroup",
			"AWS::RDS::DBSubnetGroup":      "aws/rds/dbSubnetGroup",
		},
	},
	{
		// CloudWatch Logs: log groups.
		Name: "Logs",
		Include: []string{
			"AWS::Logs::LogGroup",
		},
	},
	{
		// SNS: pub/sub topics and their subscriptions. The legacy
		// AWS::SNS::TopicPolicy is not Cloud Control–provisionable; TopicInlinePolicy is
		// the Cloud Control–friendly variant (the SNS analogue of QueueInlinePolicy) used
		// as a managed link intermediary for resource-based topic policies (e.g. s3→sns).
		Name: "SNS",
		Include: []string{
			"AWS::SNS::Topic",
			"AWS::SNS::Subscription",
			"AWS::SNS::TopicInlinePolicy",
		},
	},
	{
		// S3: object storage buckets. Bucket carries inline
		// notification, versioning, encryption, lifecycle and CORS config; BucketPolicy
		// is the resource-based access policy. Bucket notifications to Lambda/SQS/SNS are
		// modelled inline on the bucket's NotificationConfiguration, with links handling
		// the target permission side-effects.
		Name: "S3",
		Include: []string{
			"AWS::S3::Bucket",
			"AWS::S3::BucketPolicy",
		},
	},
	{
		// Kinesis: data streams as a stream event source. Only the Stream type is needed;
		// consumer groups and stream consumers (enhanced fan-out) will come later.
		Name: "Kinesis",
		Include: []string{
			"AWS::Kinesis::Stream",
		},
	},
	{
		// SecretsManager: encrypted secrets. Secret holds
		// the value; ResourcePolicy is the resource-based access policy (the SecretsManager
		// analogue of BucketPolicy/TopicInlinePolicy), available as a managed intermediary or
		// for manual cross-account grants. Identity-based grants (lambda->secret) go through
		// the function's execution role, not the resource policy.
		Name: "SecretsManager",
		Include: []string{
			"AWS::SecretsManager::Secret",
			"AWS::SecretsManager::ResourcePolicy",
		},
	},
	// SSM parameters are intentionally NOT onboarded via Cloud Control: CloudFormation (and
	// therefore Cloud Control) cannot create SecureString parameters (AWS::SSM::Parameter Type
	// allows only String | StringList), and SecureString is a low-cost alternative to Secrets
	// Manager that practitioners will want. aws/ssm/parameter is a custom implementation against
	// the SSM SDK in services/ssm/, supporting String | StringList | SecureString with a
	// `secureValue` field marked Sensitive.
	{
		// KMS: customer managed keys for encrypting secrets, parameters, queues and buckets,
		// plus human-friendly Aliases used to look up keys by a stable name.
		Name: "KMS",
		Include: []string{
			"AWS::KMS::Key",
			"AWS::KMS::Alias",
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
	"AWS::S3::Bucket": {
		FilterFields:            []string{"bucketName", "arn", "region"},
		DeriveIdentifierFromARN: false,
	},
	// RDS instance identifier is the DBInstanceIdentifier; the ARN suffix (db:<id>) derives it.
	"AWS::RDS::DBInstance": {
		FilterFields:            []string{"dbInstanceIdentifier", "arn", "region"},
		DeriveIdentifierFromARN: true,
	},
	// RDS cluster identifier is the DBClusterIdentifier; the ARN suffix (cluster:<id>) derives it.
	"AWS::RDS::DBCluster": {
		FilterFields:            []string{"dbClusterIdentifier", "arn", "region"},
		DeriveIdentifierFromARN: true,
	},
	// RDS proxy identifier is the DBProxyName; the ARN is a separate computed field.
	"AWS::RDS::DBProxy": {
		FilterFields:            []string{"dbProxyName", "arn", "region"},
		DeriveIdentifierFromARN: false,
	},
	// SecretsManager secret's primary identifier IS its ARN, so an `arn` filter takes the
	// GetResource fast path directly; `name` resolves via ListResources + filter.
	"AWS::SecretsManager::Secret": {
		FilterFields:            []string{"name", "arn", "region"},
		DeriveIdentifierFromARN: false,
	},
	// KMS key's identifier is the KeyId; the ARN suffix (key/<keyId>) derives it, so a single
	// `arn` filter resolves via GetResource.
	"AWS::KMS::Key": {
		FilterFields:            []string{"keyId", "arn", "region"},
		DeriveIdentifierFromARN: true,
	},
	// KMS alias lookup: resolve a key by its stable human-friendly alias name. The identifier
	// is the AliasName (e.g. "alias/my-key").
	"AWS::KMS::Alias": {
		FilterFields:            []string{"aliasName", "region"},
		DeriveIdentifierFromARN: false,
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
