package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Workiva/frugal/compiler"
)

const defaultTopicDelim = "."

func main() {
	var (
		file  = flag.String("file", "", "Thrift/Frugal file to generate")
		gen   = flag.String("gen", "", "Generate code for this language")
		out   = flag.String("out", "", "Output directory for generated code")
		delim = flag.String("delim", defaultTopicDelim, "Topic token delimiter")
	)
	flag.Parse()

	if *file == "" || *gen == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := compiler.Compile(*file, *gen, *out, *delim); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
