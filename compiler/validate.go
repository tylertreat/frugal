package compiler

import (
	"fmt"

	"github.com/edmodo/frugal/parser"

	frugParser "github.com/Workiva/frugal/compiler/parser"
)

type errParse struct {
	errors []*parser.CompileError
}

func (e *errParse) Error() string {
	str := ""
	for _, err := range e.errors {
		str += fmt.Sprintf("%s (line %d, col %d): %s\n", err.File, err.Pos.Line, err.Pos.Col, err.Message)
	}
	return str
}

// validate ensures the Thrift file and Frugal Program are valid, meaning
// namespaces match and struct references are defined.
func validate(thriftFile string, program *frugParser.Program) error {
	context := parser.NewCompileContext()
	tree := context.ParseRecursive(thriftFile)
	if tree == nil {
		return &errParse{context.Errors}
	}

	// Ensure namespaces match.
	if err := validateNamespaces(tree, program); err != nil {
		return err
	}

	// Ensure struct references are defined.
	if err := validateReferences(tree, program); err != nil {
		return err
	}

	return nil
}

func validateNamespaces(tree *parser.ParseTree, program *frugParser.Program) error {
	if len(program.Namespaces) != len(tree.Namespaces) {
		return fmt.Errorf("Namespaces defined in %s do not match those defined in %s", program.Path, tree.Path)
	}
	for language, namespace := range program.Namespaces {
		thriftNamespace, ok := tree.Namespaces[language]
		if !ok {
			return fmt.Errorf("%s specifies namespace '%s' for language %s, but %s does not specify namespace",
				program.Path, namespace, language, tree.Path)
		}
		if namespace != thriftNamespace {
			return fmt.Errorf("%s specifies namespace '%s' for language %s, but %s specifies %s",
				program.Path, namespace, language, tree.Path, thriftNamespace)
		}
	}
	return nil
}

func validateReferences(tree *parser.ParseTree, program *frugParser.Program) error {
	structCache := make(map[string]bool)
	for _, scope := range program.Scopes {
	opLoop:
		for _, op := range scope.Operations {
			if _, ok := structCache[op.Param]; ok {
				continue
			}
			for _, node := range tree.Nodes {
				switch n := node.(type) {
				case *parser.StructNode:
					structCache[n.Name.Data.(string)] = true
					if op.Param == n.Name.Data.(string) {
						continue opLoop
					}
				default:
					continue
				}
			}
			return fmt.Errorf("Reference '%s' in %s not defined in %s",
				op.Param, program.Path, tree.Path)
		}
	}
	return nil
}
