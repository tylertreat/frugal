package generator

import (
	"fmt"
	"os"
)

type BaseGenerator struct {
	Options map[string]string
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
