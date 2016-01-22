package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/globals"
)

const (
	frugalFile   = "idl/constants.frugal"
	expectedFile = "expected/thrift/constants.thrift"
)

func TestThriftConstants(t *testing.T) {
	options := compiler.Options{
		File:               frugalFile,
		Gen:                "go",
		Out:                "out",
		Delim:              ".",
		DryRun:             true,
		RetainIntermediate: true,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}
	outputPath := filepath.Join("idl", "constants.thrift")
	compareFiles(t, expectedFile, outputPath)

	// Clean up intermediate IDL.
	for _, file := range globals.IntermediateIDL {
		// Only try to remove if file still exists.
		if _, err := os.Stat(file); err == nil {
			if err := os.Remove(file); err != nil {
				t.Fatal("Failed to remove intermediate IDL %s\n", file)
			}
		}
	}
}
