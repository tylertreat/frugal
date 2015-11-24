package test

import (
	"path/filepath"
	"testing"

	"github.com/Workiva/frugal/compiler"
)

func TestValidGo(t *testing.T) {
	options := compiler.Options{
		File:  validFile,
		Gen:   "go",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	fooPath := filepath.Join(outputDir, "valid", "frug_foo_scope.go")
	compareFiles(t, "expected/go/frug_foo_scope.txt", fooPath)
	blahServPath := filepath.Join(outputDir, "valid", "frug_blah_service.go")
	compareFiles(t, "expected/go/frug_blah_service.txt", blahServPath)
	blahScopePath := filepath.Join(outputDir, "valid", "frug_blah_scope.go")
	compareFiles(t, "expected/go/frug_blah_scope.txt", blahScopePath)
}
