package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Workiva/frugal/generator"
	"github.com/Workiva/frugal/generator/golang"
	"github.com/Workiva/frugal/parser"
)

func main() {
	var (
		file = flag.String("file", "", "Thrift/Frugal file to generate")
		gen  = flag.String("gen", "", "Generate code for this language")
		out  = flag.String("out", "", "Output directory for generated code")
	)
	flag.Parse()

	if *file == "" || *gen == "" {
		flag.Usage()
		os.Exit(1)
	}

	name := getName(*file)
	if !exists(name + ".thrift") {
		fmt.Printf("Thrift file not found: %s.thrift\n", name)
		os.Exit(1)
	}
	if !exists(name + ".frugal") {
		fmt.Printf("Frugal file not found: %s.frugal\n", name)
		os.Exit(1)
	}

	program, err := parser.Parse(name + ".frugal")
	if err != nil {
		panic(err)
	}

	if len(program.Namespaces) == 0 {
		fmt.Println("No namespaces to generate")
		os.Exit(1)
	}

	if err := generateThrift(*out, *gen, name+".thrift"); err != nil {
		os.Exit(1)
	}

	var g generator.ProgramGenerator
	switch *gen {
	case "go":
		g = generator.NewOOProgramGenerator(golang.NewGenerator())
	default:
		flag.Usage()
		os.Exit(1)
	}

	if err := g.Generate(program, *out); err != nil {
		panic(err)
	}
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
