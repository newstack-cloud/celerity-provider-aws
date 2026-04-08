# Bluelink Provider for AWS

[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=newstack-cloud_bluelink-provider-aws&metric=coverage)](https://sonarcloud.io/summary/new_code?id=newstack-cloud_bluelink-provider-aws)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=newstack-cloud_bluelink-provider-aws&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=newstack-cloud_bluelink-provider-aws)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=newstack-cloud_bluelink-provider-aws&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=newstack-cloud_bluelink-provider-aws)

The Bluelink Provider plugin for AWS.
This plugin provides a collection of resources, data sources, custom variables, functions and links to interact with AWS services.

For the full list of supported resources, data sources, links and more, see the [Bluelink Registry](https://registry.bluelink.dev).

## Overview

The initial focus of the provider (powering Celerity v0) is to support FaaS-based (serverless functions) deployments on AWS. This includes core services like Lambda, IAM, SNS, SQS, DynamoDB, RDS and ElastiCache along with the links between them.

The provider also includes **Flex** resources, which are higher-level abstractions with presets that simplify what would otherwise be complex configuration. For example, `aws/flex/vpc` provides a streamlined way to set up networking without manually wiring individual VPC components.

## Provider Configuration

The provider accepts configuration for AWS authentication and connectivity including:

- **Authentication** - Access key ID, secret access key, session token, and AWS profile
- **Region** - AWS region for API operations
- **Assume Role** - Full support for `assumeRole` and `assumeRoleWithWebIdentity` with policy restrictions, session tags, and configurable duration
- **Endpoints** - Custom service endpoints, HTTP/HTTPS proxy settings
- **TLS** - Custom CA bundles, option to disable TLS verification
- **Retry** - Configurable max retries and retry mode (standard/adaptive)
- **S3** - Path-style addressing option
- **EC2 Metadata** - Custom metadata service endpoint and protocol mode

## Project Structure

```
definitions/
  services/          # YAML service definition schemas
  inter-service/     # Cross-service link definitions
  schema.yml         # JSON Schema for service definitions
services/            # AWS service implementations.
flex/                # Flex resources (higher-level abstractions with presets)
inter-service-links/ # Cross-service link implementations
provider/            # Provider registration and configuration
utils/               # Shared utilities and helpers
internal/testutils/  # Test mocks and integration test helpers
```

## Additional Documentation

- [Contributing](./docs/CONTRIBUTING.md)
- [Commit Guidelines](./docs/COMMIT_GUIDELINES.md)
