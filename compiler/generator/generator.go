package generator

import (
	"errors"
	"fmt"
	"os"

	"github.com/Workiva/frugal/compiler/parser"
)

const filePrefix = "frug_"

type OOGenerator interface {
	GenerateFile(name, outputDir string, namespaces []*parser.Namespace) (*os.File, error)
	GenerateDocStringComment(*os.File) error
	GeneratePackage(f *os.File, name, outputDir string) error
	GenerateImports(*os.File) error
	GenerateConstants(f *os.File, name string) error
	GeneratePublishers(*os.File, []*parser.Namespace) error
	GenerateSubscribers(*os.File, []*parser.Namespace) error
	GenerateNewline(*os.File, int) error
	DefaultOutputDir() string
	CheckCompile(path string) error
}

type BaseGenerator struct{}

func (b *BaseGenerator) CreateFile(name, outputDir, suffix string, namespaces []*parser.Namespace) (*os.File, error) {
	if len(namespaces) == 0 {
		return nil, errors.New("No namespaces to generate")
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/%s", outputDir, name), 0777); err != nil {
		return nil, err
	}
	return os.Create(fmt.Sprintf("%s/%s/%s%s.%s", outputDir, name, filePrefix, name, suffix))
}

func (b *BaseGenerator) GenerateNewline(file *os.File, count int) error {
	str := ""
	for i := 0; i < count; i++ {
		str += "\n"
	}
	_, err := file.WriteString(str)
	return err
}

type ProgramGenerator interface {
	// Generate the Program in the given directory.
	Generate(program *parser.Program, outputDir string) error

	// DefaultOutputDir is the default directory to generate in.
	DefaultOutputDir() string
}

type OOProgramGenerator struct {
	OOGenerator
}

func NewOOProgramGenerator(generator OOGenerator) ProgramGenerator {
	return &OOProgramGenerator{generator}
}

// Generate the Program in the given directory.
func (o *OOProgramGenerator) Generate(program *parser.Program, outputDir string) error {
	if outputDir == "" {
		outputDir = o.DefaultOutputDir()
	}

	file, err := o.GenerateFile(program.Name, outputDir, program.Namespaces)
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

	if err := o.GeneratePublishers(file, program.Namespaces); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateSubscribers(file, program.Namespaces); err != nil {
		return err
	}

	// Ensure code compiles. If it doesn't, it's likely because they didn't
	// generate the Thrift structs referenced in their Frugal file.
	path := fmt.Sprintf(".%s%s%s%s", string(os.PathSeparator), outputDir, string(os.PathSeparator), program.Name)
	return o.CheckCompile(path)
}

// DefaultOutputDir is the default directory to generate in.
func (o *OOProgramGenerator) DefaultOutputDir() string {
	return o.OOGenerator.DefaultOutputDir()
}
