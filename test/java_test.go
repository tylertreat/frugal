package test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/globals"
)

func TestValidJava(t *testing.T) {
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 23, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:  validFile,
		Gen:   "java",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	pubPath := filepath.Join(outputDir, "foo", "FooPublisher.java")
	compareFiles(t, "expected/java/FooPublisher.java", pubPath)
	subPath := filepath.Join(outputDir, "foo", "FooSubscriber.java")
	compareFiles(t, "expected/java/FooSubscriber.java", subPath)
	pubPath = filepath.Join(outputDir, "foo", "BarPublisher.java")
	compareFiles(t, "expected/java/BarPublisher.java", pubPath)
	subPath = filepath.Join(outputDir, "foo", "BarSubscriber.java")
	compareFiles(t, "expected/java/BarSubscriber.java", subPath)
}
