package compiler

import (
	"fmt"
	"os"
	"strings"

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

	// Process options for specific generators
	lang, options := cleanGenParam(gen)

	// Resolve Frugal generator.
	var g generator.ProgramGenerator
	switch lang {
	case "dart":
		g = generator.NewMultipleFileProgramGenerator(dartlang.NewGenerator(options), false)
	case "go":
		g = generator.NewSingleFileProgramGenerator(golang.NewGenerator(options))
	case "java":
		g = generator.NewMultipleFileProgramGenerator(java.NewGenerator(options), true)
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

// cleanGenParam processes a string that includes an optional trailing
// options set.  Format: <language>:<name>=<value>,<name>=<value>,...
func cleanGenParam(gen string) (lang string, options map[string]string) {
	lang = gen
	options = make(map[string]string)
	if strings.Contains(gen, ":") {
		s := strings.Split(gen, ":")
		lang = s[0]
		dirty := s[1]
		var optionArray []string
		if strings.Contains(dirty, ",") {
			optionArray = strings.Split(dirty, ",")
		} else {
			optionArray = append(optionArray, dirty)
		}
		for _, option := range optionArray {
			s := strings.Split(option, "=")
			options[s[0]] = s[1]
		}
	}
	return
}
