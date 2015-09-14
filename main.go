package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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

	args := []string{}
	if *out != "" {
		args = append(args, "-out", *out)
	}
	args = append(args, "-gen", *gen)
	args = append(args, *file+".thrift")
	if out, err := exec.Command("thrift", args...).CombinedOutput(); err != nil {
		fmt.Println(string(out))
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

	namespaces, err := parser.Parse(*file + ".frugal")
	if err != nil {
		panic(err)
	}

	name, err := getName(*file + ".frugal")
	if err != nil {
		panic(err)
	}

	if err := g.Generate(name, *out, namespaces); err != nil {
		panic(err)
	}
}

func getName(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	parts := strings.Split(info.Name(), ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid file: %s", path)
	}
	return parts[0], nil
}
