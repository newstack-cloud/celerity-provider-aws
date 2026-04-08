# Link Implementation Checklist

This checklist is for use by background agents (or human contributors) to track progress and validate the quality of a new link implementation. Mark each item as complete as you progress.

---

## 1. File and Structure Validation
- [ ] All required files are present:
  - `*_link.go`
  - `*_link_update.go`
  - `*_link_update_test.go`
  - `*_link_stage_changes.go`
  - `*_link_stage_changes_test.go`
  - `*_link_annotations.go`
  - `descriptions_embed.go`
- [ ] Files are placed in the correct directory:
  - Intra-service links: `services/${service}/links/`
  - Inter-service links: `inter-service-links/${serviceA}_${serviceB}/`
- [ ] For inter-service links: function signature includes all required service factories

## 2. Schema and Method Validation
- [ ] Link schema matches the service definition schema (`definitions/services/${service}.yml` or `definitions/inter-service/${serviceA}-${serviceB}.yaml`).
- [ ] All required methods are implemented:
  - `UpdateResourceA`, `UpdateResourceB`, `UpdateIntermediaryResources`, `StageChanges`
- [ ] Method signatures match existing patterns.
- [ ] Uses `linkhelpers` and `pluginutils` helpers for value extraction and nil checks.

## 3. ResourceDataMappings Implementation
- [ ] `ResourceDataMappings` field is included in `UpdateResourceA`, `UpdateResourceB`, and `UpdateIntermediaryResources` outputs when resource fields are modified.
- [ ] ResourceDataMappings format follows the pattern: `{resourceName}::{resourceFieldPath}` -> `{linkFieldPath}`
- [ ] Resource field paths start with `spec.` and are relative to the resource.
- [ ] ResourceDataMappings are only included for fields that are actually modified by the link.
- [ ] Methods that don't modify resource fields omit the `ResourceDataMappings` field entirely (not empty map).
- [ ] Tests verify correct ResourceDataMappings format and content.

## 4. Intermediary Resources Handling
- [ ] `UpdateIntermediaryResources` correctly handles both `managed` and `existing` intermediary resource types.
- [ ] For `existing` resources, implementation uses `ResourceLookupService` to fetch resources by blueprint instance ID, type, and external ID.
- [ ] For `managed` resources, implementation uses `ResourceDeployService` to manage creation, updates, and deletion.
- [ ] Proper error message format when existing intermediary resource is not in the same blueprint:
  ```
  "intermediary resource of type '${intermediaryResourceType}' is not present in the same blueprint as this link (${linkTypeIdentifier}). Links can only update intermediary resources that are defined in the same blueprint. Please define the resource in this blueprint or remove the link and manually configure the relationship."
  ```

## 5. Annotation Handling
- [ ] Link annotations are defined and handled as per the schema and existing patterns.
- [ ] Any dynamic annotation keys (e.g., with <placeholder>) are implemented as required.
- [ ] Dynamic annotation keys have only one <placeholder> and point to a unique resource in the same blueprint.
- [ ] Tests cover both static and dynamic annotation scenarios.

## 6. Test Coverage
- [ ] Test files exist for each method.
- [ ] Tests cover:
  - [ ] Basic use cases
  - [ ] Complex/edge cases
  - [ ] Error cases (e.g., missing required fields, service call failures)
  - [ ] ResourceDataMappings validation (correct format, content, and omission when appropriate)
  - [ ] Intermediary resource scenarios (both managed and existing types)
- [ ] Tests use mocks and test utils as per existing patterns.
- [ ] All tests pass locally.

## 7. Rich Descriptions and Examples
- [ ] At least one rich description is added to `services/${service}/links/descriptions/`.
- [ ] Rich description is referenced in the main link definition file.
- [ ] Example uses correct code block formatting (e.g., ```javascript for JSONC).
- [ ] No unclosed code blocks.
- [ ] Description is present above the example.

## 8. Consistency and Pattern Adherence
- [ ] Implementation follows patterns from existing links in the same or similar service.
- [ ] No large blocks of repetitive nil checks—uses helpers instead.
- [ ] No deviation from naming conventions or file structure.
- [ ] No deprecated or discouraged patterns present.
- [ ] Code is readable and maintainable with proper use of `linkhelpers` and `pluginutils` package helpers.

## 9. Source of Truth and Reference Validation
- [ ] API calls used match those specified in the service definition schema and doc links.
- [ ] All required fields and annotations are handled as per the schema.
- [ ] No missing or extra fields in the link spec.
- [ ] AWS API Reference docs are used as the authoritative source for service call actions.

## 10. Regression and Integration
- [ ] All existing tests in the project pass (no regressions introduced).
- [ ] Link integrates cleanly with the rest of the provider (no import or dependency issues).
- [ ] Link registered in `provider/provider.go` under the `Links` map.
- [ ] For inter-service links with bidirectional relationships: resource type order differentiates link purpose (e.g., `aws/lambda/function::aws/dynamodb/table` vs `aws/dynamodb/table::aws/lambda/function`).

## 11. Linting and Formatting
- [ ] Code passes linting (e.g., `gofmt`, `golint`, or project-specific linter).
- [ ] No TODOs, commented-out code, or debug prints left in output.

## 12. Notes and Special Cases
- [ ] If any `notes` are present in the service definition, they are addressed in the implementation or tests.
- [ ] Any known quirks or edge cases are documented in comments or test cases.
- [ ] Link data projection correctly represents a subset of state in resource A and resource B along with full state of intermediary resources.

---

**Instructions:**
- Use this checklist as a progress tracker during implementation.
- Mark each item as complete as you finish it.
- If any item cannot be completed, flag it for review or request clarification.
- Pay special attention to ResourceDataMappings implementation as this is critical for avoiding drift detection issues. 