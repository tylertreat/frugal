package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParseFrugal parses the given Frugal file into its semantic representation.
func ParseFrugal(filePath string) (*Frugal, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	name, err := getName(file)
	if err != nil {
		return nil, err
	}

	parsed, err := ParseReader(filePath, file)
	if err != nil {
		return nil, err
	}

	frugal := parsed.(*Frugal)
	if err := frugal.validate(); err != nil {
		return nil, err
	}
	frugal.Name = name
	frugal.File = filePath
	frugal.Dir = filepath.Dir(file.Name())
	frugal.Path = filePath
	frugal.sort() // For determinism in generated code
	frugal.assignFrugal()

	return frugal, nil
}

func getName(f *os.File) (string, error) {
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	parts := strings.Split(info.Name(), ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid file: %s", f.Name())
	}
	return parts[0], nil
}
