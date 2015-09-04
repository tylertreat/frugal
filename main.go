package main

import (
	"fmt"
	"os"

	"github.com/Workiva/frugal/generator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s file\n", os.Args[0])
		os.Exit(1)
	}

	namespaces, err := generator.Parse(os.Args[1])
	if err != nil {
		panic(err)
	}
	for _, ns := range namespaces {
		fmt.Println(ns.Name)
		for _, op := range ns.Operations {
			fmt.Printf("    %s: %s\n", op.Name, op.Param)
		}
	}
}
