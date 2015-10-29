package compiler

import (
	"fmt"

	"github.com/samuel/go-thrift/parser"

	frugParser "github.com/Workiva/frugal/compiler/parser"
)

// validate ensures the Thrift file and Frugal Program are valid, meaning
// namespaces match and struct references are defined.
func validate(thriftFile string, program *frugParser.Program) error {
	parsed, err := parser.ParseFile(thriftFile)
	if err != nil {
		return err
	}
	thrift := parsed.(*parser.Thrift)

	// Ensure namespaces match.
	if err := validateNamespaces(thrift, program, thriftFile); err != nil {
		return err
	}

	// Ensure struct references are defined.
	if err := validateReferences(thrift, program, thriftFile); err != nil {
		return err
	}

	return nil
}

func validateNamespaces(tree *parser.Thrift, program *frugParser.Program, file string) error {
	if len(program.Namespaces) != len(tree.Namespaces) {
		return fmt.Errorf("Namespaces defined in %s do not match those defined in %s", program.Path, file)
	}
	for language, namespace := range program.Namespaces {
		thriftNamespace, ok := tree.Namespaces[language]
		if !ok {
			return fmt.Errorf("%s specifies namespace '%s' for language %s, but %s does not specify namespace",
				program.Path, namespace, language, file)
		}
		if namespace != thriftNamespace {
			return fmt.Errorf("%s specifies namespace '%s' for language %s, but %s specifies %s",
				program.Path, namespace, language, file, thriftNamespace)
		}
	}
	return nil
}

func validateReferences(tree *parser.Thrift, program *frugParser.Program, file string) error {
	structCache := make(map[string]bool)
	for _, scope := range program.Scopes {
	opLoop:
		for _, op := range scope.Operations {
			if _, ok := structCache[op.Param]; ok {
				continue
			}
			for name, _ := range tree.Structs {
				structCache[name] = true
				if op.Param == name {
					continue opLoop
				}
			}
			return fmt.Errorf("Reference '%s' in %s not defined in %s",
				op.Param, program.Path, file)
		}
	}
	return nil
}
