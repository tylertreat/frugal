package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidDart(t *testing.T) {
	options := compiler.Options{
		File:  validFile,
		Gen:   "dart",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	fooPath := filepath.Join(outputDir, "valid", "lib", "src", "frug_foo.dart")
	compareFiles(t, "expected/dart/frug_foo.dart", fooPath)
	barPath := filepath.Join(outputDir, "valid", "lib", "src", "frug_bar.dart")
	compareFiles(t, "expected/dart/frug_bar.dart", barPath)
}
