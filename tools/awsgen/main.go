// Command awsgen generates Cloud Control–backed Bluelink resource definitions from
// vendored CloudFormation registry schemas.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	schemasDir := flag.String("schemas", "tools/awsgen/schemas", "directory of vendored CFN schema JSON files")
	outDir := flag.String("out", "services/cloudcontrol/gen", "output directory for generated Go files")
	curatedDir := flag.String("curated", "tools/awsgen/curated_examples", "directory of curated example overrides")
	sync := flag.Bool("sync", false, "download and vendor the allowlisted CFN schemas instead of generating")
	region := flag.String("region", "us-east-1", "AWS region for the CloudFormation schema registry (with -sync)")
	flag.Parse()

	if *sync {
		if err := syncSchemas(*schemasDir, *region); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := run(*schemasDir, *outDir, *curatedDir); err != nil {
		log.Fatal(err)
	}
}

func run(schemasDir, outDir, curatedDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	files, err := vendoredSchemaFiles(schemasDir)
	if err != nil {
		return fmt.Errorf("listing vendored schemas: %w", err)
	}

	var resources []*irResource
	for _, file := range files {
		resource, err := generateOne(schemasDir, outDir, curatedDir, file)
		if err != nil {
			return fmt.Errorf("generating from %s: %w", file, err)
		}
		for _, warning := range resource.Warnings {
			fmt.Printf("warning [%s]: %s\n", resource.CFNType, warning)
		}
		resources = append(resources, resource)
	}

	// Data-source-only services (e.g. EC2) contribute lookup data sources but no
	// managed resources, so they are kept out of the resource registry.
	registryResources := make([]*irResource, 0, len(resources))
	for _, res := range resources {
		if dataSourceOnlyType(res.CFNType) {
			continue
		}
		registryResources = append(registryResources, res)
	}

	registry, err := emitRegistryFile(registryResources)
	if err != nil {
		return fmt.Errorf("emitting registry: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "registry.go"), registry, 0o644); err != nil {
		return err
	}

	dataSources, err := generateDataSources(outDir, curatedDir, resources)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(outDir, "examples_embed.go"), []byte(examplesEmbedFile), 0o644); err != nil {
		return err
	}

	fmt.Printf("generated %d resource(s) and %d data source(s) into %s\n", len(registryResources), len(dataSources), outDir)
	return nil
}

func generateDataSources(outDir, curatedDir string, resources []*irResource) ([]*irDataSource, error) {
	var dataSources []*irDataSource
	for _, resource := range resources {
		cfg, ok := dataSourceConfigFor(resource.CFNType)
		if !ok {
			continue
		}
		ds := deriveDataSource(resource, cfg)
		for _, warning := range ds.Warnings {
			fmt.Printf("warning [%s data source]: %s\n", ds.CFNType, warning)
		}

		source, err := emitDataSourceFile(ds, resource)
		if err != nil {
			return nil, fmt.Errorf("emitting %s data source: %w", ds.CFNType, err)
		}
		fileName := exampleStem(ds.CFNType) + "_data_source.go"
		if err := os.WriteFile(filepath.Join(outDir, fileName), source, 0o644); err != nil {
			return nil, err
		}

		if err := generateDataSourceExample(outDir, curatedDir, ds); err != nil {
			return nil, err
		}

		dataSources = append(dataSources, ds)
	}

	if len(dataSources) == 0 {
		return dataSources, nil
	}

	registry, err := emitDataSourceRegistryFile(dataSources)
	if err != nil {
		return nil, fmt.Errorf("emitting data source registry: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "data_sources.go"), registry, 0o644); err != nil {
		return nil, err
	}
	return dataSources, nil
}

func generateExamples(outDir, curatedDir string, res *irResource) error {
	resourcesDir := filepath.Join(outDir, "examples", "resources")
	if err := os.MkdirAll(resourcesDir, 0o755); err != nil {
		return err
	}

	for _, variant := range []exampleVariant{variantBasic, variantComplete} {
		fileName := fmt.Sprintf("%s_%s.md", exampleStem(res.CFNType), variant.name())
		content, err := exampleContent(curatedDir, fileName, res, variant)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(resourcesDir, fileName), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func exampleContent(curatedDir, fileName string, res *irResource, variant exampleVariant) (string, error) {
	if curatedDir != "" {
		if data, err := os.ReadFile(filepath.Join(curatedDir, fileName)); err == nil {
			return string(data), nil
		}
	}
	blueprint := buildExampleBlueprint(res, variant)
	return renderExampleMarkdown(blueprint, variant.description(res.Label))
}

func generateOne(schemasDir, outDir, curatedDir, schemaFile string) (*irResource, error) {
	data, err := os.ReadFile(filepath.Join(schemasDir, schemaFile))
	if err != nil {
		return nil, err
	}

	schema, err := loadCFNSchema(data)
	if err != nil {
		return nil, err
	}

	resource, err := convert(schema, blueprintTypeFor(schema.TypeName))
	if err != nil {
		return nil, err
	}

	// Data-source-only services emit no resource file or resource examples; the IR is
	// still returned so a lookup data source can be derived from it.
	if dataSourceOnlyType(schema.TypeName) {
		return resource, nil
	}

	source, err := emitResourceFile(resource)
	if err != nil {
		return nil, fmt.Errorf("emitting %s: %w", schema.TypeName, err)
	}

	outPath := filepath.Join(outDir, schemaFileName(schema.TypeName))
	if err := os.WriteFile(outPath, source, 0o644); err != nil {
		return nil, err
	}

	if err := generateExamples(outDir, curatedDir, resource); err != nil {
		return nil, fmt.Errorf("generating examples for %s: %w", schema.TypeName, err)
	}
	return resource, nil
}

const examplesEmbedFile = `// Code generated by tools/awsgen. DO NOT EDIT.

package gen

import "embed"

//go:embed examples
var examplesFS embed.FS

// readExamples loads the basic and complete example markdown for a resource type.
func readExamples(stem string) []string {
	var examples []string
	for _, variant := range []string{"basic", "complete"} {
		data, err := examplesFS.ReadFile("examples/resources/" + stem + "_" + variant + ".md")
		if err == nil {
			examples = append(examples, string(data))
		}
	}
	return examples
}

// readDataSourceExamples loads the example markdown for a data source type.
func readDataSourceExamples(stem string) []string {
	data, err := examplesFS.ReadFile("examples/datasources/" + stem + ".md")
	if err != nil {
		return nil
	}
	return []string{string(data)}
}
`
