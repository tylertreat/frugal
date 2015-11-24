package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/generator/dartlang"
	"github.com/Workiva/frugal/compiler/generator/golang"
	"github.com/Workiva/frugal/compiler/generator/java"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

// Options contains compiler options for code generation.
type Options struct {
	File               string
	Gen                string
	Out                string
	Delim              string
	RetainIntermediate bool
}

// Compile parses the respective Frugal and Thrift and generates code for them,
// returning an error if something failed.
func Compile(options Options) error {
	globals.TopicDelimiter = options.Delim
	globals.Gen = options.Gen
	globals.Out = options.Out
	globals.FileDir = filepath.Dir(options.File)

	defer func() {
		if !options.RetainIntermediate {
			// Clean up intermediate IDL.
			for _, file := range globals.IntermediateIDL {
				// Only try to remove if file still exists.
				if _, err := os.Stat(file); err == nil {
					if err := os.Remove(file); err != nil {
						fmt.Printf("Failed to remove intermediate IDL %s\n", file)
					}
				}
			}
		}
	}()

	absFile, err := filepath.Abs(options.File)
	if err != nil {
		return err
	}

	_, err = compile(absFile)
	return err
}

func compile(file string) (*parser.Frugal, error) {
	var (
		gen = globals.Gen
		out = globals.Out
		dir = filepath.Dir(file)
	)

	// Ensure Frugal file exists.
	if !exists(file) {
		return nil, fmt.Errorf("Frugal file not found: %s\n", file)
	}

	// Process options for specific generators.
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
		return nil, fmt.Errorf("Invalid gen value %s", gen)
	}

	// Parse the Frugal file.
	frugal, err := parser.ParseFrugal(file)
	if err != nil {
		return nil, err
	}

	if out == "" {
		out = g.DefaultOutputDir()
	}
	fullOut := g.GetOutputDir(out, frugal)
	if err := os.MkdirAll(out, 0777); err != nil {
		return nil, err
	}

	// Generate Thrift code.
	if err := generateThrift(frugal, dir, out, gen); err != nil {
		return nil, err
	}

	// Generate Frugal code.
	if frugal.ContainsFrugalDefinitions() {
		return frugal, g.Generate(frugal, fullOut)
	}
	return frugal, nil
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
