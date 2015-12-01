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

	fooPath := filepath.Join(outputDir, "valid", "lib", "src", "f_foo.dart")
	compareFiles(t, "expected/dart/f_foo.dart", fooPath)
	blahPath := filepath.Join(outputDir, "valid", "lib", "src", "f_blah.dart")
	compareFiles(t, "expected/dart/f_blah.dart", blahPath)
}
