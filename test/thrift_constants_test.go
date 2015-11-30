package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
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
}
