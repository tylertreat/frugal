package generator

import (
	"fmt"
	"os"

	"github.com/Workiva/frugal/compiler/parser"
)

const filePrefix = "frug_"

// ProgramGenerator generates source code in a specified language for a Program
// produced by the parser.
type ProgramGenerator interface {
	// Generate the Program in the given directory.
	Generate(program *parser.Program, outputDir string) error

	// DefaultOutputDir is the default directory to generate in.
	DefaultOutputDir() string
}

// SingleFileGenerator generates source code in a specified language in a
// single source file.
type SingleFileGenerator interface {
	GenerateFile(name, outputDir string) (*os.File, error)
	GenerateDocStringComment(*os.File) error
	GeneratePackage(f *os.File, name, outputDir string) error
	GenerateImports(*os.File) error
	GenerateConstants(f *os.File, name string) error
	GeneratePublishers(*os.File, []*parser.Scope) error
	GenerateSubscribers(*os.File, []*parser.Scope) error
	GenerateNewline(*os.File, int) error
	DefaultOutputDir() string
	CheckCompile(path string) error
}

// SingleFileProgramGenerator is an implementation of the ProgramGenerator
// interface which generates source code in one file.
type SingleFileProgramGenerator struct {
	SingleFileGenerator
}

func NewSingleFileProgramGenerator(generator SingleFileGenerator) ProgramGenerator {
	return &SingleFileProgramGenerator{generator}
}

// Generate the Program in the given directory.
func (o *SingleFileProgramGenerator) Generate(program *parser.Program, outputDir string) error {
	if outputDir == "" {
		outputDir = o.DefaultOutputDir()
	}

	file, err := o.GenerateFile(program.Name, outputDir)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := o.GenerateDocStringComment(file); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GeneratePackage(file, program.Name, outputDir); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateImports(file); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateConstants(file, program.Name); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GeneratePublishers(file, program.Scopes); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateSubscribers(file, program.Scopes); err != nil {
		return err
	}

	// Ensure code compiles. If it doesn't, it's likely because they didn't
	// generate the Thrift structs referenced in their Frugal file.
	path := fmt.Sprintf(".%s%s%s%s", string(os.PathSeparator), outputDir, string(os.PathSeparator), program.Name)
	return o.CheckCompile(path)
}

// DefaultOutputDir is the default directory to generate in.
func (o *SingleFileProgramGenerator) DefaultOutputDir() string {
	return o.SingleFileGenerator.DefaultOutputDir()
}
