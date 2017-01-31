package test

import (
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestInvalid(t *testing.T) {
	options := compiler.Options{
		File:  invalidFile,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateServices(t *testing.T) {
	options := compiler.Options{
		File:  duplicateServices,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateScopes(t *testing.T) {
	options := compiler.Options{
		File:  duplicateScopes,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateMethods(t *testing.T) {
	options := compiler.Options{
		File:  duplicateMethods,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateOperations(t *testing.T) {
	options := compiler.Options{
		File:  duplicateOperations,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateMethodArgIds(t *testing.T) {
	options := compiler.Options{
		File:  duplicateMethodArgIds,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

func TestDuplicateStructFieldIds(t *testing.T) {
	options := compiler.Options{
		File:  duplicateStructFieldIds,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if compiler.Compile(options) == nil {
		t.Fatal("Expected error")
	}
}

// Ensures an error is returned when a "*" namespace has a vendor annotation.
func TestWildcardNamespaceWithVendorAnnotation(t *testing.T) {
	options := compiler.Options{
		File:  badNamespace,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err == nil {
		t.Fatal("Expected error")
	}
}
