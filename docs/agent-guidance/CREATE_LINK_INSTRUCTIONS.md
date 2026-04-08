## Celerity AWS Link Implementation Guide

### Overview

You need to create a new `${resourceTypeA}::${resourceTypeB}` link for the `${service}` service, if `${resourceTypeA}` and `${resourceTypeB}` and in the same service, it should be created in the `services/${service}` package directory. If `${resourceTypeA}` and `${resourceTypeB}` are in different services, it should be created in the `inter-service-links/` package directory.

A resource enables users of Celerity to connect resources using link selectors and labels in blueprint files, links provide a powerful abstraction that allows users to define both simple and complex relationships between resources through declarative link selectors.

The links needs to be implemented following existing patterns and conventions in link implementations so far either in the `services/${service}/links` package or in another service implementation in the `inter-service-links` directory or under the `services` directory in the `links` subdirectory.

### Plan

1. Create the link annotations definition file, carefully following existing link annotation definitions such as `services/lambda/function__function_link_annotations.go`. Annotations will be defined in the definitions schema file as mentioned in [sources](#sources).
2. Create the link implementation files, carefully following existing link implementations such as `services/lambda/links/function__function_link.go`.
3. Create the functionality for the `UpdateResourceA` method for the link, carefully following existing links, studying and reusing existing patterns and utils. See `services/lambda/links/function__function_link_update.go` for a guide. You must use the `pluginutils` package helpers to extract values from the spec data.
4. Implement an extensive test suite for the `UpdateResourceA` method, carefully following existing test suites for the `UpdateResourceA` method such as `services/lambda/links/function__function_link_update_test.go`. Study and reuse test utils and patterns from existing tests.
5. Create the functionality for the `UpdateResourceB` method for the link, carefully following existing links, studying and reusing existing patterns and utils. See `services/lambda/links/function__function_link_update.go` for a guide. You must use the `pluginutils` package helpers to extract values from the spec data.
6. Implement an extensive test suite for the `UpdateResourceB` method, carefully following existing test suites for the `UpdateResourceB` method such as `services/lambda/links/function__function_link_update_test.go`. Study and reuse test utils and patterns from existing tests.
7. Create the functionality for the `UpdateIntermediaryResources` method for the link, carefully following existing links, studying and reusing existing patterns and utils. See `services/lambda/links/function__function_link_update.go` for a guide. You must use the `pluginutils` package helpers to extract values from the spec data.
8. Implement an extensive test suite for the `UpdateIntermediaryResources` method, carefully following existing test suites for the `UpdateIntermediaryResources` method such as `services/lambda/links/function__function_link_update_test.go`. Study and reuse test utils and patterns from existing tests.
9. Create the functionality for the `StageChanges` method for the resource, carefully following existing links, studying and reusing existing patterns and utils. See `services/lambda/links/function__code_signing_config_link_stage_changes.go` for a guide. You must use the `linkhelpers` and `pluginutils` package helpers to extract values from the spec data and collect derived changes for the link data projection as a link is a projection of a subset of the state in resource A and resource B along with the full state of any intermediary resources required for the link.
10. Implement an extensive test suite for the `StageChanges` method, carefully following existing test suites for the `StageChanges` method such as `services/lambda/links/function__code_signing_config_link_stage_changes_test.go`. Study and reuse test utils and patterns from existing tests.
11. Validate the implementations and tests are complete and correct. Ensure that the tests are covering failures and edge cases as well as basic and complex use cases.
12. Check for any deviation from the patterns used in existing link implementations, study multiple existing link implementations.
13. Apply any corrections as a result of the analysis from step 11 and 12.
14. Study existing rich descriptions in the `services/${service}/links/descriptions` directory or in other `services/*/links/descriptions` directories.
15. Add a rich description for the new link to the `services/${service}/links/descriptions` directory and integrate the rich description into the main link definition file.

### Sources

You should use the service definition schema as a source of truth to guide the implementation of the link.
If `${resourceTypeA}` and `${resourceTypeB}` are a part of the same service, the service definition schema is located in the `definitions/services/${service}.yml` file. 
If `${resourceTypeA}` and `${resourceTypeB}` are a part of different services, the service definition schema is located in the `definitions/inter-service/${serviceA}-${serviceB}.yaml` file.
The structure of the service definition schema is defined in the `definitions/schema.yml` file.

To determine the service call actions that are required in the link implementation, you should use the AWS API Reference docs for the `${service}` service. You can also use the AWS SDK v2 for Go docs to accurately determine the service call actions required.

You should thoroughly review the existing link implementations, taking note of patterns for value extraction, collecting changes for the link data projection and methods to access fields in `*core.MappingNode` objects or `map[string]*core.MappingNode` maps.
You should also thoroughly review the existing tests for link implementations to understand how to implement the tests for the new link, using the `plugintestutils` package helpers where possible.

### Setting up new service

If `${service}` does not exist in the `services` directory, you need to create a new directory for it in the `services` directory.
You should create a new service interface in the `services/${service}/service.go` file, using existing service interfaces as a guide.
You should prepare a mock for the service to be used in tests in the `internal/testutils/${service}_mock/${service}_service_mock.go` file, using existing service mock implementations as a guide.

### Link File Structure

The link implementation should be structured as follows:

- `*_link.go` - The main link implementation file.
- `*_link_update.go` - The update resource A, resource B and intermediary resources implementation.
- `*_link_update_test.go` - The update resource A, resource B and intermediary resources test implementation.
- `*_link_stage_changes.go` - The stage changes implementation.
- `*_link_stage_changes_test.go` - The stage changes test implementation.
- `*_link_annotations.go` - The link annotations definition file.
- `descriptions_embed.go` - Embed directive for loading description markdown files.

### Intra-Service vs Inter-Service Links

**Intra-service links** (resources in the same AWS service):
- Location: `services/${service}/links/`
- Schema: `definitions/services/${service}.yml` under `linkDefinitions`
- Example: `aws/sqs/queue::aws/sqs/queue` (dead-letter queue configuration)

**Inter-service links** (resources in different AWS services):
- Location: `inter-service-links/${serviceA}_${serviceB}/`
- Schema: `definitions/inter-service/${serviceA}-${serviceB}.yml`
- Example: `aws/lambda/function::aws/dynamodb/table` (Lambda triggered by DynamoDB Streams)

Inter-service links require service factories from multiple services. The function signature should include all required service dependencies:

```go
// Example: Lambda-DynamoDB Stream link requires Lambda service
func FunctionTableStreamLink(
    lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
    awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Link

// Example: Lambda-DynamoDB access link requires Lambda, DynamoDB, and IAM services
func FunctionTableAccessLink(
    lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
    iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
    awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Link
```

For links that update intermediary resources in a third service (e.g., IAM roles), include that service factory as well.

### Inter-Service Link Registration

Inter-service links must be registered in `provider/provider.go` under the `Links` map with their full link type identifier:

```go
Links: map[string]provider.Link{
    // Intra-service link
    "aws/sqs/queue::aws/sqs/queue": sqslinks.QueueQueueLink(...),

    // Inter-service links - use resource order to differentiate link types
    // Lambda accessing DynamoDB (Lambda links TO DynamoDB)
    "aws/lambda/function::aws/dynamodb/table": lambdadynamodb.FunctionTableLink(...),
    // DynamoDB triggering Lambda via streams (DynamoDB links TO Lambda)
    "aws/dynamodb/table::aws/lambda/function": lambdadynamodb.TableFunctionLink(...),
},
```

When the same two services can have multiple relationship types, use resource type order to differentiate:
- `aws/lambda/function::aws/dynamodb/table` = Lambda needs access to DynamoDB
- `aws/dynamodb/table::aws/lambda/function` = DynamoDB triggers Lambda via streams

This allows bidirectional relationships where both resources can link to each other for different purposes.

### Link Methods

Across the files mentioned in the previous section, you should implement the following methods:

- `UpdateResourceA` - The update resource A operation implementation.
- `UpdateResourceB` - The update resource B operation implementation.
- `UpdateIntermediaryResources` - The update intermediary resources operation implementation.
- `StageChanges` - The stage changes operation implementation.

### Tests

You should implement thorough tests that cover both basic and complex uses of the links along with error cases for missing IDs (or other required fields) and when the service method call returns an error.

You must provide tests for all the methods mentioned in the [Link Methods](#link-methods) section using the files defined in the [Link File Structure](#link-file-structure) section.

**Testing ResourceDataMappings:**

When testing link update methods, ensure your test cases verify that the `ResourceDataMappings` field is correctly populated when the link modifies resource fields. Test cases should include:
- Verifying the correct mapping format (`{resourceName}::{resourceFieldPath}` -> `{linkFieldPath}`)
- Ensuring mappings are only included for fields that are actually modified
- Confirming that methods that don't modify resource fields don't include `ResourceDataMappings`

See existing test implementations such as `services/lambda/links/function__code_signing_config_link_update_test.go` for examples of how to test the `ResourceDataMappings` field in your expected outputs.

### Rich Descriptions

You should include a rich description that includes examples in the main link definition file.
Examples are defined in the `services/${service}/links/descriptions` directory and should be markdown files. You can use existing link examples as a guide.

Be sure to use the "```javascript ... ```" code block syntax for JSONC examples.

There is no need to add an explanation section at the bottom of the examples, only a description above the example code block(s).

Make sure you always close any code blocks that open with "\`\`\`" (usually followed by a language identifier) with "\`\`\`" on a new line.

You should inspect existing rich descriptions closely in the `services/${service}/links/descriptions` directory and other `services/*/links/descriptions` directories to understand how to structure the rich descriptions.

### About Link Annotations

Dynamic link annotation keys are defined in the form `aws.lambda.function.<targetFunction>.populateEnvVars` where `<targetFunction>` is the placeholder for the target function name.
There must only be one `<..>` placeholder in an annotation key and it can be located anywhere in the key.
The contents of the `<..>` placeholder can be anything but it must always point to a unique resource defined in the same blueprint file that is linked to the resource where the annotation is defined. This is then used to target a specific resource in the link that the given annotation should be applied to.

This must be considered when creating link implementations and using the annotations in the change staging and update operations.

See `services/lambda/links/function__function_link_annotations.go` for an example of how to define link annotations with dynamic keys.

#### Annotation Placement (resourceA vs resourceB)

When designing link annotations, carefully consider which resource the annotations should be placed on in the schema definition:

**Access links** (e.g., Lambda→DynamoDB, Lambda→SQS for read/write access):
- Annotations go on **resourceA** (the one doing the accessing)
- These annotations configure how resourceA accesses resourceB
- Example: `aws.lambda.dynamodb.<targetTable>.accessLevel` on Lambda configures Lambda's access to a specific table

**Event-driven trigger links** (e.g., DynamoDB→Lambda, SQS→Lambda for event triggers):
- Annotations go on **resourceB** (the target being triggered)
- These annotations configure the event source mapping for this specific target
- A source can trigger multiple targets, each with different configurations
- Example: `aws.dynamodb.lambda.batchSize` on Lambda (resourceB) configures this function's batch processing

**Exception - Source resource configuration:**
- Some annotations genuinely configure the source resource itself (not the relationship)
- Example: `aws.dynamodb.stream.viewType` configures the DynamoDB table's stream format
- These stay on resourceA because they affect the table regardless of which functions consume it

**Rationale:**
- A source resource can link to multiple targets, each with different configurations
- Each link relationship has its own configuration (e.g., batch size, filters)
- Putting configuration on the wrong resource creates ambiguity when one-to-many relationships exist

#### Link Annotation Naming Convention

Annotation names must clearly distinguish between **resource configuration** and **relationship configuration**:

**Resource configuration** (single service namespace):
- Pattern: `aws.{service}.{feature}.{setting}`
- Example: `aws.dynamodb.stream.viewType` - configures DynamoDB table's stream format
- These affect the resource regardless of which other resources link to it

**Intra-service relationship configuration** (same service):
- Pattern: `aws.{service}.{feature}.{setting}` or `aws.{service}.{feature}.<target>.{setting}`
- Examples:
  - `aws.lambda.invoke.<targetFunction>.populateEnvVars` - configures Lambda→Lambda invocation
  - `aws.sqs.dlq.maxReceiveCount` - configures SQS queue→DLQ relationship
- For intra-service links, repeating the service name would be redundant

**Inter-service relationship configuration** (different services):
- Pattern: `aws.{serviceA}.{serviceB}.{feature}.{setting}` or `aws.{serviceA}.{serviceB}.<target>.{setting}`
- Examples:
  - `aws.dynamodb.lambda.stream.batchSize` - configures DynamoDB→Lambda stream event source mapping
  - `aws.lambda.dynamodb.<targetTable>.accessLevel` - configures Lambda's access to a specific table
- Including both service names and the feature clarifies what aspect of the relationship is being configured

### About Intermediary Resources

As per the schema in `definitions/schema.yml`, intermediary resources are either `managed` or `existing`.
Managed resources are created and deleted by the link, while existing resources are expected to be present in the blueprint and are only updated by the link.

When creating a link implementation, the implementation of `UpdateIntermediaryResources` must fetch intermediary resources of the `existing` type from the external system and update them with the necessary changes. However, if the resource is not in the same blueprint as the link, the implementation should return an error, clearly explaining that the resource is not in the same blueprint as the link and that links can only update intermediary resources that are in the same blueprint as the link.
The format of the error message should be:

```
"intermediary resource of type '${intermediaryResourceType}' is not present in the same blueprint as this link (${linkTypeIdentifier}). Links can only update intermediary resources that are defined in the same blueprint. Please define the resource in this blueprint or remove the link and manually configure the relationship."
```

To determine whether or not an existing intermediary resource is present in the same blueprint as the link, you should use the `ResourceLookupService` in the input struct to fetch the resource by blueprint instance ID, type and external ID (as defined in the schema). The external ID will be derived from the spec of one of the two resources in the link relationship (e.g. the ARN of the role for the lambda function link).

For `managed` resources, the `ResourceDeployService` in the input struct should be used to manage creation, updates and the deletion of the resource.

_The "input struct" refers to the second argument of the `UpdateIntermediaryResources` method of a link._

### Link to Resource Data Mappings

As links are projections, in practise, they act as effects that update the state of the resources that are linked together or other existing resources in the same blueprint that are treated as intermediary resources that are required to "activate" the link. For this reason, drift detection would always detect changes to the state by link implementations as drift.

Updates have been made to the plugin and blueprint frameworks to allow for overlaying link data onto the state of the resource before drift checks are performed. This is made possible by the `ResourceDataMappings` field that is returned in the output of the `UpdateResourceA`, `UpdateResourceB` and `UpdateIntermediaryResources` methods.

The `ResourceDataMappings` field is a mapping of the form `{resourceName}::{resourceFieldPath}` to `{linkFieldPath}` where `{resourceName}` is the name of the resource as defined in the source blueprint file and `{resourceFieldPath}` is the path to the field in the resource that is being updated by the link. For example, `saveOrderFunction::spec.environment.variables.TABLE_NAME` -> `saveOrderFunction.environmentVariables.TABLE_NAME` indicates that the field `saveOrderFunction.environmentVariables.TABLE_NAME` modifies the `environment.variables.TABLE_NAME` field in the resource spec. `{resourceFieldPath}` is relative to the resource and will always start with `spec` as the root object.

**When to include ResourceDataMappings:**

- Include `ResourceDataMappings` when your link method modifies fields in a resource's spec that would otherwise be detected as drift
- Only include mappings for fields that are actually being modified by the link
- If a method doesn't modify any resource fields, omit the `ResourceDataMappings` field entirely (don't return an empty map)

**Example patterns:**

```go
// When modifying a resource field
return &provider.LinkUpdateResourceOutput{
    LinkData: &core.MappingNode{
        Fields: map[string]*core.MappingNode{
            input.ResourceInfo.ResourceName: {
                Fields: map[string]*core.MappingNode{
                    "codeSigningConfigArn": core.MappingNodeFromString(arn),
                },
            },
        },
    },
    ResourceDataMappings: map[string]string{
        fmt.Sprintf("%s::spec.codeSigningConfigArn", input.ResourceInfo.ResourceName): 
        fmt.Sprintf("%s.codeSigningConfigArn", input.ResourceInfo.ResourceName),
    },
}, nil

// When not modifying any resource fields
return &provider.LinkUpdateResourceOutput{
    LinkData: &core.MappingNode{
        Fields: map[string]*core.MappingNode{},
    },
    // No ResourceDataMappings field needed
}, nil
```

You can find examples of how these resource data mappings should be defined in existing link implementations such as `services/lambda/links/function__code_signing_config_link_update.go`.

### Extra notes

- You should avoid large code blocks with more than half a dozen nil checks for fields, instead, use the `linkhelpers`  and `pluginutils` package helpers to break down value extraction, as this will make the code more readable and maintainable.
- You must run the tests to ensure they are all passing before considering the task as complete.
- You must run existing tests in this project to ensure that regressions have not been introduced.
