package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	// Ensure corresponding Thrift and Frugal files exist.
	name := getName(file)
	if !exists(name + ".thrift") {
		return fmt.Errorf("Thrift file not found: %s.thrift\n", name)
	}
	if !exists(name + ".frugal") {
		return fmt.Errorf("Frugal file not found: %s.frugal\n", name)
	}

	// Process options (TODO: abstract to function)
	lang := gen
	var options map[string]string
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
	frugal, err := parser.ParseFrugal(name + ".frugal")
	if err != nil {
		return err
	}

	// Ensure Thrift file and parsed Frugal are valid (namespaces match,
	// struct references defined, etc.).
	thriftFile := name + ".thrift"
	if err := validate(thriftFile, frugal); err != nil {
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
	if err := generateThrift(out, gen, thriftFile); err != nil {
		return err
	}

	// Generate Frugal code.
	return g.Generate(frugal, fullOut)
}

func generateThrift(out, gen, file string) error {
	args := []string{"-r"}
	if out != "" {
		args = append(args, "-out", out)
	}
	args = append(args, "-gen", gen)
	args = append(args, file)
	if out, err := exec.Command("thrift", args...).CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	}
	return nil
}

func getName(path string) string {
	name := path
	extension := filepath.Ext(name)
	if extension != "" {
		name = name[0:strings.LastIndex(name, extension)]
	}
	return name
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
