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
	globals.Now = time.Date(2015, 10, 5, 0, 0, 0, 0, time.UTC)

	if err := compiler.Compile(validFile, "java", outputDir, delim); err != nil {
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
