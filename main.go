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
		file          = flag.String("file", "", "Thrift/Frugal file to generate")
		gen           = flag.String("gen", "", "Generate code for this language")
		out           = flag.String("out", "", "Output directory for generated code")
		thriftImport  = flag.String("thrift_import", "", "prefix for thrift import in generated code")
		frugalImport  = flag.String("frugal_import", "", "prefix for frugal import in generated code")
		packagePrefix = flag.String("package_prefix", "", "prefix for thrift generated package imports")
		delim         = flag.String("delim", defaultTopicDelim, "Topic token delimiter")
	)
	flag.Parse()

	if *file == "" || *gen == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := compiler.Compile(*file, *gen, *out, *delim, *thriftImport, *frugalImport, *packagePrefix); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
