package generator

import (
	"fmt"
	"os"

	"github.com/Workiva/frugal/compiler/parser"
)

type BaseGenerator struct {
	Options map[string]string
	Frugal  *parser.Frugal
}

func (b *BaseGenerator) CreateFile(name, outputDir, suffix string, usePrefix bool) (*os.File, error) {
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return nil, err
	}
	prefix := ""
	if usePrefix {
		prefix = FilePrefix
	}
	return os.Create(fmt.Sprintf("%s/%s%s.%s", outputDir, prefix, name, suffix))
}

func (b *BaseGenerator) GenerateNewline(file *os.File, count int) error {
	str := ""
	for i := 0; i < count; i++ {
		str += "\n"
	}
	_, err := file.WriteString(str)
	return err
}

func (b *BaseGenerator) GenerateInlineComment(comment []string, indent string) string {
	inline := ""
	for _, line := range comment {
		inline += indent + "// " + line + "\n"
	}
	return inline
}

func (b *BaseGenerator) GenerateBlockComment(comment []string, indent string) string {
	block := indent + "/**\n"
	for _, line := range comment {
		block += indent + " * " + line + "\n"
	}
	block += indent + " */\n"
	return block
}

func (b *BaseGenerator) SetFrugal(f *parser.Frugal) {
	b.Frugal = f
}
