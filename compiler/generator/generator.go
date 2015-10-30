package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/Workiva/frugal/compiler/parser"
)

const FilePrefix = "frug_"

type FileType string

const (
	CombinedFile  FileType = "combined"
	PublishFile   FileType = "publish"
	SubscribeFile FileType = "subscribe"
)

// ProgramGenerator generates source code in a specified language for a Frugal
// produced by the parser.
type ProgramGenerator interface {
	// Generate the Frugal in the given directory.
	Generate(frugal *parser.Frugal, outputDir string) error

	// GetOutputDir returns the full output directory for generated code.
	GetOutputDir(dir string, f *parser.Frugal) string

	// DefaultOutputDir returns the default directory for generated code.
	DefaultOutputDir() string
}

// SingleFileGenerator generates source code in a specified language in a
// single source file.
type SingleFileGenerator interface {
	GenerateFile(name, outputDir string) (*os.File, error)
	GenerateDocStringComment(*os.File) error
	GeneratePackage(f *os.File, p *parser.Frugal) error
	GenerateImports(f *os.File, p *parser.Frugal) error
	GenerateConstants(f *os.File, name string) error
	GeneratePublishers(*os.File, []*parser.Scope) error
	GenerateSubscribers(*os.File, []*parser.Scope) error
	GenerateNewline(*os.File, int) error
	GetOutputDir(dir string, f *parser.Frugal) string
	DefaultOutputDir() string
	CheckCompile(path string) error
}

// MultipleFileGenerator generates source code in a specified language in a
// multiple source files.
type MultipleFileGenerator interface {
	GenerateDependencies(f *parser.Frugal, dir string) error
	GenerateFile(name, outputDir string, fileType FileType) (*os.File, error)
	GenerateDocStringComment(*os.File) error
	GeneratePackage(f *os.File, p *parser.Frugal, s *parser.Scope) error
	GenerateImports(*os.File, *parser.Scope) error
	GenerateConstants(f *os.File, name string) error
	GeneratePublisher(*os.File, *parser.Scope) error
	GenerateSubscriber(*os.File, *parser.Scope) error
	GenerateNewline(*os.File, int) error
	GetOutputDir(dir string, f *parser.Frugal) string
	DefaultOutputDir() string
	CheckCompile(path string) error
}

func GetPackageComponents(pkg string) []string {
	return strings.Split(pkg, ".")
}

// SingleFileProgramGenerator is an implementation of the ProgramGenerator
// interface which generates source code in one file.
type SingleFileProgramGenerator struct {
	SingleFileGenerator
}

func NewSingleFileProgramGenerator(generator SingleFileGenerator) ProgramGenerator {
	return &SingleFileProgramGenerator{generator}
}

// Generate the Frugal in the given directory.
func (o *SingleFileProgramGenerator) Generate(frugal *parser.Frugal, outputDir string) error {
	file, err := o.GenerateFile(frugal.Name, outputDir)
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

	if err := o.GeneratePackage(file, frugal); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateImports(file, frugal); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateConstants(file, frugal.Name); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GeneratePublishers(file, frugal.Scopes); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateSubscribers(file, frugal.Scopes); err != nil {
		return err
	}

	// Ensure code compiles. If it doesn't, it's likely because they didn't
	// generate the Thrift structs referenced in their Frugal file.
	return o.CheckCompile(fmt.Sprintf(".%s%s", string(os.PathSeparator), outputDir))
}

// GetOutputDir returns the full output directory for generated code.
func (o *SingleFileProgramGenerator) GetOutputDir(dir string, f *parser.Frugal) string {
	return o.SingleFileGenerator.GetOutputDir(dir, f)
}

// DefaultOutputDir returns the default directory for generated code.
func (o *SingleFileProgramGenerator) DefaultOutputDir() string {
	return o.SingleFileGenerator.DefaultOutputDir()
}

// MultipleFileProgramGenerator is an implementation of the ProgramGenerator
// interface which generates source code in one file.
type MultipleFileProgramGenerator struct {
	MultipleFileGenerator
	SplitPublisherSubscriber bool
}

func NewMultipleFileProgramGenerator(generator MultipleFileGenerator,
	splitPublisherSubscriber bool) ProgramGenerator {
	return &MultipleFileProgramGenerator{generator, splitPublisherSubscriber}
}

// Generate the Program in the given directory.
func (o *MultipleFileProgramGenerator) Generate(frugal *parser.Frugal, outputDir string) error {
	if err := o.GenerateDependencies(frugal, outputDir); err != nil {
		return err
	}
	for _, scope := range frugal.Scopes {
		if o.SplitPublisherSubscriber {
			if err := o.generateFile(frugal, scope, outputDir, PublishFile); err != nil {
				return err
			}
			if err := o.generateFile(frugal, scope, outputDir, SubscribeFile); err != nil {
				return err
			}
		} else {
			if err := o.generateFile(frugal, scope, outputDir, CombinedFile); err != nil {
				return err
			}
		}
	}
	// Ensure code compiles. If it doesn't, it's likely because they didn't
	// generate the Thrift structs referenced in their Frugal file.
	return o.CheckCompile(fmt.Sprintf(".%s%s", string(os.PathSeparator), outputDir))
}

func (o MultipleFileProgramGenerator) generateFile(frugal *parser.Frugal, scope *parser.Scope,
	outputDir string, fileType FileType) error {
	file, err := o.GenerateFile(scope.Name, outputDir, fileType)
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

	if err := o.GeneratePackage(file, frugal, scope); err != nil {
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

	if fileType == CombinedFile || fileType == PublishFile {
		if err := o.GeneratePublisher(file, scope); err != nil {
			return err
		}
	}

	if fileType == CombinedFile {
		if err := o.GenerateNewline(file, 2); err != nil {
			return err
		}
	}

	if fileType == CombinedFile || fileType == SubscribeFile {
		if err := o.GenerateSubscriber(file, scope); err != nil {
			return err
		}
	}

	if err := o.GenerateNewline(file, 1); err != nil {
		return err
	}

	return nil
}

// GetOutputDir returns the full output directory for generated code.
func (o *MultipleFileProgramGenerator) GetOutputDir(dir string, f *parser.Frugal) string {
	return o.MultipleFileGenerator.GetOutputDir(dir, f)
}

// DefaultOutputDir returns the default directory for generated code.
func (o *MultipleFileProgramGenerator) DefaultOutputDir() string {
	return o.MultipleFileGenerator.DefaultOutputDir()
}
