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

	// Ensure corresponding Thrift and Frugal files exist.
	name := getName(*file)
	if !exists(name + ".thrift") {
		fmt.Printf("Thrift file not found: %s.thrift\n", name)
		os.Exit(1)
	}
	if !exists(name + ".frugal") {
		fmt.Printf("Frugal file not found: %s.frugal\n", name)
		os.Exit(1)
	}

	// Resolve Frugal generator.
	var g generator.ProgramGenerator
	switch *gen {
	case "go":
		g = generator.NewOOProgramGenerator(golang.NewGenerator())
	default:
		flag.Usage()
		os.Exit(1)
	}

	// Parse the Frugal file.
	program, err := parser.Parse(name + ".frugal")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(program.Namespaces) == 0 {
		fmt.Println("No namespaces to generate")
		os.Exit(1)
	}

	output := *out
	if output == "" {
		output = g.DefaultOutputDir()
	}
	if err := os.MkdirAll(output, 0777); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Generate Thrift code.
	if err := generateThrift(output, *gen, name+".thrift"); err != nil {
		os.Exit(1)
	}

	// Generate Frugal code.
	if err := g.Generate(program, output); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Ensure code compiles. If it doesn't, it's likely because they didn't
	// generate the Thrift structs referenced in their Frugal file.
	path := fmt.Sprintf(".%s%s%s%s", string(os.PathSeparator), output, string(os.PathSeparator), program.Name)
	if err := checkCompile(path); err != nil {
		os.Exit(1)
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
