package generator

import (
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

// Languages is a map of supported language to a slice of the generator options
// it supports.
var Languages = map[string][]string{
	"go":   []string{"thrift_import", "frugal_import", "package_prefix"},
	"java": nil,
	"dart": nil,
}

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
	GenerateWrapper(f *os.File, p *parser.Frugal) error
	SetFrugal(*parser.Frugal)
}

// MultipleFileGenerator generates source code in a specified language in a
// multiple source files.
type MultipleFileGenerator interface {
	// Generic methods
	GenerateDependencies(f *parser.Frugal, dir string) error
	GenerateFile(name, outputDir string, fileType FileType) (*os.File, error)
	GenerateDocStringComment(*os.File) error
	GenerateConstants(f *os.File, name string) error
	GenerateNewline(*os.File, int) error
	GetOutputDir(dir string, f *parser.Frugal) string
	DefaultOutputDir() string
	SetFrugal(*parser.Frugal)

	// Service-specific methods
	GenerateThrift() bool // TODO: remove this once all languages impl rpc
	GenerateServicePackage(f *os.File, p *parser.Frugal, s *parser.Service) error
	GenerateServiceImports(*os.File, *parser.Service) error
	GenerateService(*os.File, *parser.Frugal, *parser.Service) error

	// Scope-specific methods
	GenerateScopePackage(f *os.File, p *parser.Frugal, s *parser.Scope) error
	GenerateScopeImports(*os.File, *parser.Frugal, *parser.Scope) error
	GeneratePublisher(*os.File, *parser.Scope) error
	GenerateSubscriber(*os.File, *parser.Scope) error
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
	o.SetFrugal(frugal)
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

	return o.GenerateSubscribers(file, frugal.Scopes)
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

// Generate the Frugal in the given directory.
func (o *MultipleFileProgramGenerator) Generate(frugal *parser.Frugal, outputDir string) error {
	o.SetFrugal(frugal)
	if err := o.GenerateDependencies(frugal, outputDir); err != nil {
		return err
	}
	if o.GenerateThrift() {
		for _, service := range frugal.Thrift.Services {
			if err := o.generateServiceFile(frugal, service, outputDir); err != nil {
				return err
			}
		}
	}
	for _, scope := range frugal.Scopes {
		if o.SplitPublisherSubscriber {
			if err := o.generateScopeFile(frugal, scope, outputDir, PublishFile); err != nil {
				return err
			}
			if err := o.generateScopeFile(frugal, scope, outputDir, SubscribeFile); err != nil {
				return err
			}
		} else {
			if err := o.generateScopeFile(frugal, scope, outputDir, CombinedFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o MultipleFileProgramGenerator) generateServiceFile(frugal *parser.Frugal, service *parser.Service,
	outputDir string) error {
	file, err := o.GenerateFile(service.Name, outputDir, CombinedFile)
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

	if err := o.GenerateServicePackage(file, frugal, service); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateServiceImports(file, service); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateService(file, frugal, service); err != nil {
		return err
	}

	return nil
}

func (o MultipleFileProgramGenerator) generateScopeFile(frugal *parser.Frugal, scope *parser.Scope,
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

	if err := o.GenerateScopePackage(file, frugal, scope); err != nil {
		return err
	}

	if err := o.GenerateNewline(file, 2); err != nil {
		return err
	}

	if err := o.GenerateScopeImports(file, frugal, scope); err != nil {
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
