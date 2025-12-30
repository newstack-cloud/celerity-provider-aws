## Celerity AWS Resource Scaffolding Guide

### Overview

This guide describes how to scaffold a new `${resource}` resource for the `${service}` service in the `services/${service}` package directory. The goal is to generate only the resource definition, resource schema, file structure, and method/test stubs—leaving all operation logic to be implemented later.

### Plan

1. **Create the resource schema definition file**
   - Follow patterns from existing schemas (e.g., `services/lambda/function_resource_schema.go`).
   - Define the resource fields, types, and validation, referencing the service definition schema in `definitions/services/${service}.yml` and `definitions/schema.yml`.
   - Add a placeholder for examples.

2. **Create the resource implementation files as stubs**
   - Main resource file: `${resource}_resource.go` with a factory function that produces a resource using the plugin framework SDK ResourceDefinition helper type. This should contain an actions struct of the form `{resource}ResourceActions` that contains the methods for the resource actions and the methods for the actions should be defined in the individual action files. See existing resource implementation files for an example.
   There should be stubs for the properties that represent resources, `GetExternalStateFunc`, `CreateFunc`, `UpdateFunc`, `DestroyFunc`, `StabilisedFunc` that point to the action methods.
   - Action files: `${resource}_resource_create.go`, `${resource}_resource_update.go`, `${resource}_resource_destroy.go`, `${resource}_resource_get_external_state.go`, `${resource}_resource_stabilised.go`—each containing only the method signature on the `{resource}ResourceActions` struct and returning empty values that fulfil the return type of the action method.
   - If following the ops pattern, also create `${resource}_resource_create_ops.go` and `${resource}_resource_update_ops.go` as empty files or with TODOs.

3. **Create test scaffolding for each operation**
   - For each operation, create a corresponding test file (e.g., `${resource}_resource_create_test.go`).
   - Each test file should include a test suite struct and a suite registration function for the suite of tests for the action. The test suite struct should be named `{Resource}Resource{Action}Suite` (e.g. `ServerCertificateResourceCreateSuite`) and the suite registration function should be named `Test{Resource}Resource{Action}Suite` (e.g. `TestServerCertificateResourceCreateSuite`). Use the existing resource test files for reference.
   - Import the testing and testify packages and add a TODO comment for future test logic.

4. **Set up examples**
   - In `services/${service}/examples/resources/`, create markdown files for the resource examples (e.g., `${service}_${resource}_basic.md`) based on the schema that you have generated for the resource. Ensure these examples are embedded in the resource implementation file. Add basic, complex and JSONC examples. See examples for already implemented resources for reference and see existing resource implementation files for a guide on how to embed examples.
   - Be sure to use the "```javascript ... ```" code block syntax for JSONC examples.
   - There is no need to add an explanation section at the bottom of the examples, only a description above the example code block(s).
   - Make sure you always close any code blocks that open with "\`\`\`" (usually followed by a language identifier) with "\`\`\`" on a new line.
   - Do not forget to use "```yaml ... ```" code block syntax for YAML examples.
   - You should inspect existing examples closely in the `services/${service}/examples/resources` directory to understand how to structure the examples.


5. **Register the resource**
   - Register the resource definition using the factory function in the resource implementation file with the provider in `provider/provider.go`.

### File Structure

The resource scaffolding should include the following files:

- `${resource}_resource.go` - Main resource file
- `${resource}_resource_schema.go` - Resource schema definition
- `${resource}_resource_create.go` - Create action stub
- `${resource}_resource_create_test.go` - Create action test stub
- `${resource}_resource_update.go` - Update action stub
- `${resource}_resource_update_test.go` - Update action test stub
- `${resource}_resource_destroy.go` - Destroy operation stub
- `${resource}_resource_destroy_test.go` - Destroy action test stub
- `${resource}_resource_get_external_state.go` - Get external state action stub
- `${resource}_resource_get_external_state_test.go` - Get external state action test stub
- `${resource}_resource_stabilised.go` - Stabilised action stub
- `${resource}_resource_stabilised_test.go` - Stabilised action test stub
- `${resource}_resource_create_ops.go` - (if needed) Create ops stub
- `${resource}_resource_update_ops.go` - (if needed) Update ops stub
- `examples/resources/${service}_${resource}_*.md` - Example files

### Method and Test Stubs

- Each operation in the main resource file resource definition and operation files should have the correct signature and a `TODO` comment.
- Each test file should import `testing` and `testify` and have a test suite struct along with a test suite registration function.
- Each test file should be implemented with a skeleton based on the structure used for existing test files for the same resource operation and the high level definition of the service in the `definitions/services/${service}.yml` file. For example, the `services/iam/server_certificate_resource_create_test.go` file is a good reference for the structure of the test file for the `Create` operation.

### Examples

- Add example markdown files in the examples directory for the service based on the schema that you have generated for the resource.
- Ensure that the examples are embedded in the resource implementation file.
- There is a single embed.FS variable in the service in `examples_embed.go`, you should not embed each file individually in the resource implementation file. Instead, you should reference the embed.FS variable in the resource implementation file for the example files you have created like in the IAM user resource implementation file in `services/iam/user_resource.go`.
- Examples must include the full resources section of blueprint file, see existing examples for reference.

Example YAML format:

```yaml
resources:
   <resourceName>:
      type: <resourceType>
      metadata:
         displayName: <displayName>
         description: <description>
      spec:
         <resourceSpec>
```

Example JSONC format:

```javascript
{
   "resources": {
      "<resourceName>": {
         "type": "<resourceType>",
         "metadata": {
            "displayName": "<displayName>",
            "description": "<description>"
         },
         "spec": {
            // The resource spec ...
         }
      }
   }
}
```

### Notes

- Do not implement any operation logic at this stage—focus only on structure and signatures.
- Follow naming and file structure conventions from existing resources.
- Use TODO comments to indicate where logic will be added later.
- Ensure all files are created, even if empty or with only a stub.
- This approach enables iterative, logic-driven development after scaffolding is complete.
- Ensure that the correct methods, function and type placeholders are inserted into each file.
- Always return empty values that fulfil the return type of action methods that will be implemented later.
- Do not try to use convenience helpers for creating `core.MappingNode` array or map values. Stick to using the `core.MappingNode` type directly, populating the `Items` property for arrays and `Fields` property for maps.
- Always check the signatures of the equivalent action methods for existing resource implementations for the `{resource}ResourceActions` struct methods.
- In each resource implementation file, the "github.com/newstack-cloud/bluelink-provider-aws/services/{package}/service" package should be imported as `{package}service` (e.g. `iamservice`) and then referenced as the package containing the service for the generic types in the file.

### CommonTerminal Field

The `CommonTerminal` field in the ResourceDefinition indicates whether the resource is expected to link OUT to other resources:

- `CommonTerminal: true` - The resource is NOT expected to link out to other resources (it's a terminal/leaf node in the resource graph). Examples: Lambda alias, function URL, layer version permission, event source mapping.
- `CommonTerminal: false` - The resource IS expected to link out to other resources. Examples: Lambda function, SQS queue, DynamoDB table, IAM role.

This field is used to hint to users if they have forgotten to link a resource that should typically link to others. Choose the appropriate value based on whether the resource being scaffolded typically acts as an endpoint/consumer (terminal) or typically connects to other resources (non-terminal).
