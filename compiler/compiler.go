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
		g = generator.NewOOProgramGenerator(golang.NewGenerator())
	default:
		return fmt.Errorf("Invalid gen value %s", gen)
	}

	// Parse the Frugal file.
	program, err := parser.Parse(name + ".frugal")
	if err != nil {
		return err
	}

	if len(program.Namespaces) == 0 {
		return errors.New("No namespaces to generate")
	}

	if out == "" {
		out = g.DefaultOutputDir()
	}
	if err := os.MkdirAll(out, 0777); err != nil {
		return err
	}

	// Generate Thrift code.
	if err := generateThrift(out, gen, name+".thrift"); err != nil {
		return err
	}

	// Generate Frugal code.
	if err := g.Generate(program, out); err != nil {
		return err
	}

	// Ensure code compiles. If it doesn't, it's likely because they didn't
	// generate the Thrift structs referenced in their Frugal file.
	path := fmt.Sprintf(".%s%s%s%s", string(os.PathSeparator), out, string(os.PathSeparator), program.Name)
	return checkCompile(path)
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

func checkCompile(path string) error {
	if out, err := exec.Command("go", "build", path).CombinedOutput(); err != nil {
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
