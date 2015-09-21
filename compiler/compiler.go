package compiler

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/generator/golang"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

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

	// Resolve Frugal generator.
	var g generator.ProgramGenerator
	switch gen {
	case "go", "":
		g = generator.NewSingleFileProgramGenerator(golang.NewGenerator())
	default:
		return fmt.Errorf("Invalid gen value %s", gen)
	}

	// Parse the Frugal file.
	program, err := parser.Parse(name + ".frugal")
	if err != nil {
		return err
	}

	if len(program.Scopes) == 0 {
		return errors.New("No scopes to generate")
	}

	out = g.GetOutputDir(out, program)
	if err := os.MkdirAll(out, 0777); err != nil {
		return err
	}

	// Generate Thrift code.
	if err := generateThrift(filepath.Dir(out), gen, name+".thrift"); err != nil {
		return err
	}

	// TODO: Validate Frugal file against Thrift file (ensure namespaces match,
	// structs are defined, etc.)

	// Generate Frugal code.
	return g.Generate(program, out)
}

func generateThrift(out, gen, file string) error {
	args := []string{}
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
