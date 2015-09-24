package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

const FilePrefix = "frug_"

// ProgramGenerator generates source code in a specified language for a Program
// produced by the parser.
type ProgramGenerator interface {
	// Generate the Program in the given directory.
	Generate(program *parser.Program, outputDir string) error

	// GetOutputDir returns the full output directory for generated code.
	GetOutputDir(dir string, p *parser.Program) string

	// DefaultOutputDir returns the default directory for generated code.
	DefaultOutputDir() string
}

// SingleFileGenerator generates source code in a specified language in a
// single source file.
type SingleFileGenerator interface {
	GenerateFile(name, outputDir string) (*os.File, error)
	GenerateDocStringComment(*os.File) error
	GeneratePackage(f *os.File, p *parser.Program) error
	GenerateImports(f *os.File, p *parser.Program) error
	GenerateConstants(f *os.File, name string) error
	GeneratePublishers(*os.File, []*parser.Scope) error
	GenerateSubscribers(*os.File, []*parser.Scope) error
	GenerateNewline(*os.File, int) error
	GetOutputDir(dir string, p *parser.Program) string
	DefaultOutputDir() string
	CheckCompile(path string) error
}

// MultipleFileGenerator generates source code in a specified language in a
// multiple source files.
type MultipleFileGenerator interface {
	GenerateDependencies(p *parser.Program, dir string) error
	GenerateFile(name, outputDir string) (*os.File, error)
	GenerateDocStringComment(*os.File) error
	GeneratePackage(f *os.File, p *parser.Program, s *parser.Scope) error
	GenerateImports(*os.File, *parser.Scope) error
	GenerateConstants(f *os.File, name string) error
	GeneratePublisher(*os.File, *parser.Scope) error
	GenerateSubscriber(*os.File, *parser.Scope) error
	GenerateNewline(*os.File, int) error
	GetOutputDir(dir string, p *parser.Program) string
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

	if err := o.GeneratePackage(file, program); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateImports(file, program); err != nil {
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
	return o.CheckCompile(fmt.Sprintf(".%s%s", string(os.PathSeparator), outputDir))
}

// GetOutputDir returns the full output directory for generated code.
func (o *SingleFileProgramGenerator) GetOutputDir(dir string, p *parser.Program) string {
	return o.SingleFileGenerator.GetOutputDir(dir, p)
}

// DefaultOutputDir returns the default directory for generated code.
func (o *SingleFileProgramGenerator) DefaultOutputDir() string {
	return o.SingleFileGenerator.DefaultOutputDir()
}

// MultipleFileProgramGenerator is an implementation of the ProgramGenerator
// interface which generates source code in one file.
type MultipleFileProgramGenerator struct {
	MultipleFileGenerator
}

func NewMultipleFileProgramGenerator(generator MultipleFileGenerator) ProgramGenerator {
	return &MultipleFileProgramGenerator{generator}
}

// Generate the Program in the given directory.
func (o *MultipleFileProgramGenerator) Generate(program *parser.Program, outputDir string) error {
	if err := o.GenerateDependencies(program, outputDir); err != nil {
		return err
	}
	for _, scope := range program.Scopes {
		if err := o.generateFile(program, scope, outputDir); err != nil {
			return err
		}
	}
	// Ensure code compiles. If it doesn't, it's likely because they didn't
	// generate the Thrift structs referenced in their Frugal file.
	return o.CheckCompile(fmt.Sprintf(".%s%s", string(os.PathSeparator), outputDir))
}

func (o MultipleFileProgramGenerator) generateFile(program *parser.Program, scope *parser.Scope,
	outputDir string) error {
	file, err := o.GenerateFile(strings.ToLower(scope.Name), outputDir)
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

	if err := o.GeneratePackage(file, program, scope); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateImports(file, scope); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateConstants(file, scope.Name); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GeneratePublisher(file, scope); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	return o.GenerateSubscriber(file, scope)
}

// GetOutputDir returns the full output directory for generated code.
func (o *MultipleFileProgramGenerator) GetOutputDir(dir string, p *parser.Program) string {
	return o.MultipleFileGenerator.GetOutputDir(dir, p)
}

// DefaultOutputDir returns the default directory for generated code.
func (o *MultipleFileProgramGenerator) DefaultOutputDir() string {
	return o.MultipleFileGenerator.DefaultOutputDir()
}
