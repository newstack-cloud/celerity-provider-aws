## Celerity AWS Resource Implementation Guide

### Overview

You need to create a new `${resource}` resource for the `${service}` service in the `services/${service}` package directory.
A resource enables users of Celerity to define and manage resources in blueprint files.
The resource needs to be implemented following existing patterns and conventions in resource implementations so far either in the `services/${service}` package or in another service implementation under the `services` directory.

### Plan

1. Create the resource schema definition file, carefully following existing resource schemas such as `services/lambda/function_resource_schema.go`.
2. Create the resource implementation files, carefully following existing resource implementations such as `services/lambda/function_resource.go`.
3. Create the functionality for the `Create` method for the resource, carefully following existing resources, studying and reusing existing patterns and utils. See `services/lambda/function_resource_create.go` for a guide. You must use the save operations approach and `pluginutils` package helpers to extract values from the spec data.
4. Implement an extensive test suite for the `Create` method, carefully following existing test suites for the `Create` method such as `services/lambda/function_resource_create_test.go`. Study and reuse test utils and patterns from existing tests.
5. Create the functionality for the `Update` method for the resource, carefully following existing resources, studying and reusing existing patterns and utils. See `services/lambda/function_resource_update.go` for a guide. You must use the save operations approach and `pluginutils` package helpers to extract values from the spec data.
6. Implement an extensive test suite for the `Update` method, carefully following existing test suites for the `Update` method such as `services/lambda/function_resource_update_test.go`. Study and reuse test utils and patterns from existing tests.
7. Create the functionality for the `Destroy` method for the resource, carefully following existing resources, studying and reusing existing patterns and utils. See `services/lambda/function_resource_destroy.go` for a guide. You must use the save operations approach and `pluginutils` package helpers to extract values from the spec data.
8. Implement an extensive test suite for the `Destroy` method, carefully following existing test suites for the `Destroy` method such as `services/lambda/function_resource_destroy_test.go`. Study and reuse test utils and patterns from existing tests.
9. Create the functionality for the `GetExternalState` method for the resource, carefully following existing resources, studying and reusing existing patterns and utils. See `services/lambda/function_resource_get_external_state.go` for a guide. You must use the save operations approach and `pluginutils` package helpers to extract values from the spec data.
10. Implement an extensive test suite for the `GetExternalState` method, carefully following existing test suites for the `GetExternalState` method such as `services/lambda/function_resource_get_external_state_test.go`. Study and reuse test utils and patterns from existing tests.
11. Create the functionality for the `Stabilised` method for the resource, carefully following existing resources, studying and reusing existing patterns and utils. See `services/lambda/function_resource_stabilised.go` for a guide. You must use the save operations approach and `pluginutils` package helpers to extract values from the spec data.
12. Implement an extensive test suite for the `Stabilised` method, carefully following existing test suites for the `Stabilised` method such as `services/lambda/function_resource_stabilised_test.go`. Study and reuse test utils and patterns from existing tests.
13. Validate the implementations and tests are complete and correct. Ensure that the tests are covering failures and edge cases as well as basic and complex use cases.
14. Check for any deviation from the patterns used in existing resource implementations, study multiple existing resource implementations.
15. Apply any corrections as a result of the analysis from step 13 and 14.
16. Study existing examples in the `services/${service}/examples/resources` directory.
17. Add examples for the new resource to the `services/${service}/examples/resources` directory and integrate the examples into the main resource definition file.
18. Ensure that the resource is registered with the provider in the `provider/provider.go` file.

### Sources

You should use the service definition schema as a source of truth to guide the implementation of the resource.
The service definition schema is located in the `definitions/services/${service}.yml` file and the structure of the schema is defined in the `definitions/schema.yml` file.

To determine the service call actions that are required in the resource implementation, you should use the AWS API Reference docs for the `${service}` service. You can also use the AWS SDK v2 for Go docs to accurately determine the service call actions required.

You should thoroughly review the existing resource implementations, taking note of patterns for save operations, destroy operations, value extraction and the method to access fields in `*core.MappingNode` objects or `map[string]*core.MappingNode` maps.
You should also thoroughly review the existing tests for resource implementations to understand how to implement the tests for the new resource, using the `plugintestutils` package helpers where possible.

### Setting up new service

If `${service}` does not exist in the `services` directory, you need to create a new directory for it in the `services` directory.
You should create a new service interface in the `services/${service}/service.go` file, using existing service interfaces as a guide.
You should prepare a mock for the service to be used in tests in the `internal/testutils/${service}_mock/${service}_service_mock.go` file, using existing service mock implementations as a guide.

### Resource File Structure

The resource implementation should be structured as follows:

- `*_resource.go` - The main resource implementation file.
- `*_resource_create.go` - The create operation implementation.
- `*_resource_create_test.go` - The create operation test implementation.
- `*_resource_update.go` - The update operation implementation.
- `*_resource_update_test.go` - The update operation test implementation.
- `*_resource_destroy.go` - The destroy operation implementation.
- `*_resource_destroy_test.go` - The destroy operation test implementation.
- `*_resource_get_external_state.go` - The get external state operation implementation.
- `*_resource_get_external_state_test.go` - The get external state operation test implementation.
- `*_resource_stabilised.go` - The stabilisation check operation implementation.
- `*_resource_stabilised_test.go` - The stabilisation check operation test implementation.
- `*_resource_schema.go` - The resource schema definition.

You should follow the existing structure in which update or create operations are implemented in a `*_ops.go` file.

### Resource Methods

Across the files mentioned in the previous section, you should implement the following methods:

- `Create` - The create operation implementation.
- `Update` - The update operation implementation.
- `Destroy` - The destroy operation implementation.
- `GetExternalState` - The get external state operation implementation.
- `Stabilised` - The stabilisation check operation implementation.

### Tests

You should implement thorough tests that cover both basic and complex uses of the resource along with error cases for missing IDs (or other required fields) and when the service method call returns an error.

You must provide tests for all the methods mentioned in the [Resource Methods](#resource-methods) section using the files defined in the [Resource File Structure](#resource-file-structure) section.

### Examples

You should include examples in the resource schema definition file.
Examples are defined in the `services/${service}/examples/resources` directory and should be markdown files. You can use existing resource examples as a guide.

Each resource example file (basic/complete) presents the SAME blueprint in three formats, in this order:

1. **Blueprintlang (PRIMARY):** the "```blueprintlang ... ```" code block, listed first as the canonical/primary format.
2. **YAML:** the "```yaml ... ```" code block.
3. **JSONC:** the "```javascript ... ```" code block.

Bundle all three formats in the same example file; do not create a separate per-format file (e.g. no standalone `*_jsonc.md` for resources).

There is no need to add an explanation section at the bottom of the examples, only a description above the example code blocks.

Make sure you always close any code blocks that open with "\`\`\`" (usually followed by a language identifier) with "\`\`\`" on a new line.

You should inspect existing examples closely in the `services/${service}/examples/resources` directory to understand how to structure the examples.

### Extra notes

- You should avoid large code blocks with more than half a dozen nil checks for fields, instead, use the `pluginutils` package helpers to break down value extraction, as this will make the code more readable and maintainable.
- You must run the tests to ensure they are all passing before considering the task as complete.
- You must run existing tests in this project to ensure that regressions have not been introduced.
- Opt for using `AllowedValues` in resource schemas instead of `Pattern` for string fields that have a static set of allowed values.
- Do not follow the SaveOperation pattern for the `Destroy` method of plugins as the SDK doesn't have an equivalent helper for destroying resources and in most cases destroying resources is a lot simpler than creating or updating them. If the functionality is complex, it can be broken down into multiple methods instead of the declarative SDK pattern.
- When embedding examples, reuse a single examples variable per service, there is no need to instantiate a new embedded file system for each resource. A single, shared examples variable should be defined in an `examples_embed.go` file in the `services/${service}` package.
- Include tags in the `Create*` AWS SDK call when the API for the resource supports tags instead of making separate calls to add tags to the resource.

### CommonTerminal Field

The `CommonTerminal` field in the ResourceDefinition indicates whether the resource is expected to link OUT to other resources:

- `CommonTerminal: true` - The resource is NOT expected to link out to other resources (it's a terminal/leaf node in the resource graph). Examples: Lambda alias, function URL, layer version permission, event source mapping.
- `CommonTerminal: false` - The resource IS expected to link out to other resources. Examples: Lambda function, SQS queue, DynamoDB table, IAM role.

This field is used to hint to users if they have forgotten to link a resource that should typically link to others. Choose the appropriate value based on whether the resource typically acts as an endpoint/consumer (terminal) or typically connects to other resources (non-terminal).
