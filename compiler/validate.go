package compiler

import (
	"fmt"

	"github.com/samuel/go-thrift/parser"

	frugParser "github.com/Workiva/frugal/compiler/parser"
)

// validate ensures the Thrift file and parsed Frugal are valid, meaning
// namespaces match and struct references are defined.
func validate(thriftFile string, frugal *frugParser.Frugal) error {
	parsed, err := parser.ParseFile(thriftFile)
	if err != nil {
		return err
	}
	thrift := parsed.(*parser.Thrift)

	// Ensure namespaces match.
	if err := validateNamespaces(thrift, frugal, thriftFile); err != nil {
		return err
	}

	// Ensure struct references are defined.
	if err := validateReferences(thrift, frugal, thriftFile); err != nil {
		return err
	}

	return nil
}

func validateNamespaces(tree *parser.Thrift, frugal *frugParser.Frugal, file string) error {
	if len(frugal.Namespaces) != len(tree.Namespaces) {
		return fmt.Errorf("Namespaces defined in %s do not match those defined in %s", frugal.Path, file)
	}
	for language, namespace := range frugal.Namespaces {
		thriftNamespace, ok := tree.Namespaces[language]
		if !ok {
			return fmt.Errorf("%s specifies namespace '%s' for language %s, but %s does not specify namespace",
				frugal.Path, namespace, language, file)
		}
		if namespace != thriftNamespace {
			return fmt.Errorf("%s specifies namespace '%s' for language %s, but %s specifies %s",
				frugal.Path, namespace, language, file, thriftNamespace)
		}
	}
	return nil
}

func validateReferences(tree *parser.Thrift, frugal *frugParser.Frugal, file string) error {
	structCache := make(map[string]bool)
	for _, scope := range frugal.Scopes {
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
				op.Param, frugal.Path, file)
		}
	}
	return nil
}
