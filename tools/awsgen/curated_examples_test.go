package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/lang"
)

// TestCuratedExamplesParse asserts every curated example override's blueprintlang
// block parses. Curated files are embedded verbatim (overlay Layer 2), bypassing the
// generator's round-trip gate, so this guards them against drift from the
// blueprintlang grammar.
func TestCuratedExamplesParse(t *testing.T) {
	dir := "curated_examples"
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("no curated examples directory")
		}
		t.Fatal(err)
	}

	found := false
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		found = true
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			block := extractFencedBlock(string(data), "blueprintlang")
			if block == "" {
				t.Fatalf("%s has no blueprintlang block", entry.Name())
			}
			if _, err := lang.ParseString(block); err != nil {
				t.Fatalf("curated blueprintlang does not parse: %v", err)
			}
		})
	}
	if !found {
		t.Skip("no curated example files")
	}
}

// Returns the contents of the first ```<lang> fenced code block,
// or an empty string when none is present.
func extractFencedBlock(markdown, language string) string {
	open := "```" + language + "\n"
	start := strings.Index(markdown, open)
	if start < 0 {
		return ""
	}
	start += len(open)
	end := strings.Index(markdown[start:], "\n```")
	if end < 0 {
		return ""
	}
	return markdown[start : start+end]
}
