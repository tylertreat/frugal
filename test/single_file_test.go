package test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

const (
	outputDir   = "out"
	delim       = "."
	validFile   = "valid.frugal"
	invalidFile = "invalid.frugal"
)

var (
	languages           = []string{"go"}
	expectedOutputPaths = map[string]string{
		"go": "expected/go/frug_valid.txt",
	}
)

func TestSingleFileValid(t *testing.T) {
	for _, language := range languages {
		testSingleFileLanguage(t, language)
	}
}

func TestInvalid(t *testing.T) {
	options := compiler.Options{
		File:  invalidFile,
		Gen:   languages[0],
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func testSingleFileLanguage(t *testing.T, language string) {
	options := compiler.Options{
		File:  validFile,
		Gen:   language,
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}
	outputPath := filepath.Join(outputDir, "valid", fmt.Sprintf("frug_%s.%s", "valid", language))
	expectedPath := expectedOutputPaths[language]
	compareFiles(t, expectedPath, outputPath)
}

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
