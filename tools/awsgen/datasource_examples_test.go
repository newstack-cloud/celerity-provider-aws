package main

import (
	"strings"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/lang"
)

// TestDataSourceExampleGeneration renders the example for each generated data source
// and asserts it has all three format blocks and that the blueprintlang block parses
// (the round-trip gate that resource examples also pass through).
func TestDataSourceExampleGeneration(t *testing.T) {
	files, err := vendoredSchemaFiles("schemas")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		resource := loadResource(t, file, "")
		cfg, ok := dataSourceConfigFor(resource.CFNType)
		if !ok {
			continue
		}
		t.Run(resource.CFNType, func(t *testing.T) {
			ds := deriveDataSource(resource, cfg)
			content, err := renderDataSourceExample(ds)
			if err != nil {
				t.Fatalf("rendering data source example: %v", err)
			}
			for _, language := range []string{"```blueprintlang", "```yaml", "```javascript"} {
				if !strings.Contains(content, language) {
					t.Errorf("data source example missing %s block", language)
				}
			}
			block := extractFencedBlock(content, "blueprintlang")
			if block == "" {
				t.Fatal("data source example has no blueprintlang block")
			}
			if _, err := lang.ParseString(block); err != nil {
				t.Fatalf("data source blueprintlang does not parse: %v\n%s", err, block)
			}
		})
	}
}
