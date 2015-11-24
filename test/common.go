package test

import (
	"bufio"
	"os"
	"testing"
)

const (
	outputDir   = "out"
	delim       = "."
	validFile   = "valid.frugal"
	invalidFile = "invalid.frugal"
)

func compareFiles(t *testing.T, expectedPath, generatedPath string) {
	expected, err := os.Open(expectedPath)
	if err != nil {
		t.Fatalf("Failed to open file %s", expectedPath)
	}
	defer expected.Close()
	generated, err := os.Open(generatedPath)
	if err != nil {
		t.Fatalf("Failed to open file %s", generatedPath)
	}
	defer generated.Close()

	expectedScanner := bufio.NewScanner(expected)
	generatedScanner := bufio.NewScanner(generated)
	line := 1
	for expectedScanner.Scan() {
		generatedScanner.Scan()
		expectedLine := expectedScanner.Text()
		generatedLine := generatedScanner.Text()
		if expectedLine != generatedLine {
			t.Fatalf("Expected line <%s>, generated line <%s> at line %d", expectedLine, generatedLine, line)
		}
		line++
	}

	if generatedScanner.Scan() {
		t.Fatal("Generated has more lines than expected")
	}
}
