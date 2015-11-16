package compiler

import (
	"fmt"
	"os"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/generator/dartlang"
	"github.com/Workiva/frugal/compiler/generator/golang"
	"github.com/Workiva/frugal/compiler/generator/java"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

// Compile parses the respective Frugal and Thrift and generates code for them,
// returning an error if something failed.
func Compile(file, gen, out, delimiter string) error {
	globals.TopicDelimiter = delimiter

	// Ensure Frugal file exists.
	if !exists(file) {
		return fmt.Errorf("Frugal file not found: %s\n", file)
	}

	// Resolve Frugal generator.
	var g generator.ProgramGenerator
	switch gen {
	case "dart":
		g = generator.NewMultipleFileProgramGenerator(dartlang.NewGenerator(), false)
	case "go":
		g = generator.NewSingleFileProgramGenerator(golang.NewGenerator())
	case "java":
		g = generator.NewMultipleFileProgramGenerator(java.NewGenerator(), true)
	default:
		return fmt.Errorf("Invalid gen value %s", gen)
	}

	// Parse the Frugal file.
	frugal, err := parser.ParseFrugal(file)
	if err != nil {
		return err
	}

	// Generate intermediate Thrift IDL.
	idlFile, err := generateThriftIDL(frugal)
	if err != nil {
		return err
	}

	if out == "" {
		out = g.DefaultOutputDir()
	}
	fullOut := g.GetOutputDir(out, frugal)
	if err := os.MkdirAll(out, 0777); err != nil {
		return err
	}

	// Generate Thrift code.
	if err := generateThrift(out, gen, idlFile); err != nil {
		return err
	}

	// Generate Frugal code.
	return g.Generate(frugal, fullOut)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
