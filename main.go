package main

import (
	"fmt"
	"os"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/generator"
	"github.com/Workiva/frugal/compiler/globals"
	"github.com/codegangsta/cli"
)

const defaultTopicDelim = "."

var (
	help               bool
	gen                string
	out                string
	delim              string
	retainIntermediate bool
	recurse            bool
)

func main() {
	app := cli.NewApp()
	app.Name = "frugal"
	app.Usage = "a tool for code generation"
	app.Version = globals.Version
	app.HideHelp = true

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "help, h",
			Usage:       "show help",
			Destination: &help,
		},
		cli.StringFlag{
			Name:        "gen",
			Usage:       genUsage(),
			Destination: &gen,
		},
		cli.StringFlag{
			Name:        "out",
			Usage:       "set the output location for generated files (no gen-* folder will be created)",
			Destination: &out,
		},
		cli.StringFlag{
			Name:        "delim",
			Value:       defaultTopicDelim,
			Usage:       "set the delimiter for pub/sub topic tokens",
			Destination: &delim,
		},
		cli.BoolFlag{
			Name:        "retain-intermediate",
			Usage:       "retain generated intermediate thrift files",
			Destination: &retainIntermediate,
		},
		cli.BoolFlag{
			Name:        "recurse, r",
			Usage:       "generate included files",
			Destination: &recurse,
		},
	}

	app.Action = func(c *cli.Context) {
		if help {
			cli.ShowAppHelp(c)
			os.Exit(0)
		}

		if len(c.Args()) == 0 {
			fmt.Printf("Usage: %s [options] file\n\n", app.Name)
			fmt.Printf("Use %s -help for a list of options\n", app.Name)
			os.Exit(1)
		}

		if gen == "" {
			fmt.Println("No output language specified")
			fmt.Printf("Usage: %s [options] file\n\n", app.Name)
			fmt.Printf("Use %s -help for a list of options\n", app.Name)
			os.Exit(1)
		}

		file := c.Args()[0]
		options := compiler.Options{
			File:               file,
			Gen:                gen,
			Out:                out,
			Delim:              delim,
			RetainIntermediate: retainIntermediate,
			Recurse:            recurse,
		}

		if err := compiler.Compile(options); err != nil {
			fmt.Printf("Failed to generate %s:\n\t%s\n", options.File, err.Error())
			os.Exit(1)
		}
	}

	app.Run(os.Args)
}

func genUsage() string {
	usage := "generate code with a registered generator and optional parameters " +
		"(lang[:key1=val1[,key2[,key3=val3]]])\n"
	prefix := ""
	for lang, options := range generator.Languages {
		optionsStr := ""
		optionsPrefix := ""
		for _, option := range options {
			optionsStr += optionsPrefix + option
			optionsPrefix = ", "
		}
		usage += fmt.Sprintf("%s\t    %s\t%s", prefix, lang, optionsStr)
		prefix = "\n"
	}
	return usage
}
