## Celerity AWS Data Source Implementation Guide

### Overview

You need to create a new `${dataSource}` resource for the `${service}` service in the `services/${service}` package directory.
A data source enables users of Celerity to define data sources to pull data from external sources into a blueprint.
The data source needs to be implemented following existing patterns and conventions in data source implementations so far either in the `services/${service}` package or in another service implementation under the `services` directory.

### Sources

You should use the service definition schema as a source of truth to guide the implementation of the data source.
The service definition schema is located in the `definitions/services/${service}.yml` file and the structure of the schema is defined in the `definitions/schema.yml` file.
You should use the `docLinks` field in the service definition schema to use the Terraform data source equivalent docs to determine the output fields of the data source while cross-referencing with provided AWS API reference doc links as the authorititave source if truth about data available when fetching data from the service.

If you have trouble finding the data source type reference docs, please do not continue with the implementation, instead, refine your approach to find the correct data source type reference docs or suggest a change to the instructions.

You should thoroughly review the existing data source implementations, taking note of patterns for value extraction from the service API call results and the method to access fields in `*core.MappingNode` objects or `map[string]*core.MappingNode` maps.
You should also thoroughly review the existing tests for data source implementations to understand how to implement the tests for the new data source, using the `plugintestutils` package helpers where possible.

### Data Source File Structure

The data source implementation should be structured as follows:

- `*_data_source.go` - The main data source implementation file.
- `*_data_source_test.go` - The data source test implementation.
- `*_data_source_schema.go` - The data source schema definition.

### Data Source Methods

Across the files mentioned in the previous section, you should implement the following methods:

- `Fetch` - The fetch operation implementation.

### Tests

You should implement thorough tests that cover both basic and complex uses of the data source along with error cases for missing IDs (or other required filter fields) and when the service method call returns an error.

You must provide tests for all the methods mentioned in the [Data Source Methods](#data-source-methods) section using the files defined in the [Data Source File Structure](#data-source-file-structure) section.

### Examples

You should include examples in the data source schema definition file.
Examples are defined in the `services/${service}/examples/datasources` directory and should be markdown files. You can use existing data source examples as a guide.

Each data source example file presents the same blueprint in three formats, in this order, bundled in one file:

1. **Blueprintlang (PRIMARY):** the "```blueprintlang ... ```" code block, listed first. Data sources use the `data <name>: <type> { filter "<field>" == "<value>"  export <field>: <type> }` form (and top-level `export name: type { field = datasources.<name>.<field> }`).
2. **YAML:** the "```yaml ... ```" code block, using the `datasources:` section with `filter` (field/operator/search) and typed `exports`.
3. **JSONC:** the "```javascript ... ```" code block.

Use `services/events/examples/datasources/events_event_bus.md` as the canonical reference.

There is no need to add an explanation section at the bottom of the examples, only a description above the example code blocks.

Make sure you always close open code blocks in the example markdown files.

You should inspect existing examples closely in the `services/${service}/examples/datasources` directory to understand how to structure the examples.

### Extra notes

- You should avoid large code blocks with more than half a dozen nil checks for fields, instead, use the `pluginutils` package helpers to break down value extraction, as this will make the code more readable and maintainable.
- You must run the tests to ensure they are all passing before considering the task as complete.
- You must run existing tests in this project to ensure that regressions have not been introduced.
