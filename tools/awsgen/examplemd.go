package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/lang"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"gopkg.in/yaml.v3"
)

// Produces a bundled example file: a one-line description
// followed by the same blueprint in blueprintlang, YAML and JSONC.
func renderExampleMarkdown(blueprint *schema.Blueprint, description string) (string, error) {
	bp, err := lang.Emit(blueprint)
	if err != nil {
		return "", fmt.Errorf("emitting blueprintlang: %w", err)
	}

	// Gate: the emitted blueprintlang must parse back into a blueprint, so generated
	// examples can never be syntactically invalid.
	if _, err := lang.ParseString(bp); err != nil {
		return "", fmt.Errorf("generated blueprintlang example does not parse: %w", err)
	}

	yamlBytes, err := yaml.Marshal(blueprint)
	if err != nil {
		return "", fmt.Errorf("marshalling YAML: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(blueprint, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshalling JSON: %w", err)
	}
	// The blueprint JSON marshaller escapes forward slashes; unescape them so type
	// strings like "aws/sqs/queue" read naturally in documentation.
	jsonText := strings.ReplaceAll(string(jsonBytes), `\/`, `/`)

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", description)
	writeFencedBlock(&b, "blueprintlang", strings.TrimRight(bp, "\n"))
	b.WriteString("\n")
	writeFencedBlock(&b, "yaml", strings.TrimRight(string(yamlBytes), "\n"))
	b.WriteString("\n")
	writeFencedBlock(&b, "javascript", jsonText)
	return b.String(), nil
}

func writeFencedBlock(b *strings.Builder, language, content string) {
	fmt.Fprintf(b, "```%s\n%s\n```\n", language, content)
}
