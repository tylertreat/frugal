package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var prefixVariable = regexp.MustCompile("{\\w*}")

type Program struct {
	Name       string
	Path       string
	Includes   []string
	Namespaces []*Namespace
}

func (p *Program) validate() error {
	if !identifier.MatchString(p.Name) {
		return fmt.Errorf("Invalid program name %s", p.Name)
	}

	includesMap := make(map[string]bool, len(p.Includes))
	for i, include := range p.Includes {
		include = strings.TrimSpace(include)
		if _, ok := includesMap[include]; ok {
			return fmt.Errorf("Duplicate include %s", include)
		}

		// Check for existence.
		includePath := filepath.Join(p.Path, include)
		if _, err := os.Stat(includePath); err != nil {
			return fmt.Errorf("Include not found %s", includePath)
		}

		includesMap[include] = true
		p.Includes[i] = include
	}

	namespacesMap := make(map[string]bool, len(p.Namespaces))
	for _, namespace := range p.Namespaces {
		namespaceName := strings.Title(strings.TrimSpace(namespace.Name))
		if !identifier.MatchString(namespaceName) {
			return fmt.Errorf("Invalid namespace name %s", namespaceName)
		}
		if _, ok := namespacesMap[namespaceName]; ok {
			return fmt.Errorf("Duplicate namespace %s", namespaceName)
		}
		namespacesMap[namespaceName] = true
		namespace.Name = namespaceName

		operationsMap := make(map[string]bool, len(namespace.Operations))
		for _, op := range namespace.Operations {
			operationName := strings.Title(strings.TrimSpace(op.Name))
			if !identifier.MatchString(operationName) {
				return fmt.Errorf("Invalid operation name %s in namespace %s",
					operationName, namespaceName)
			}
			if _, ok := operationsMap[operationName]; ok {
				return fmt.Errorf("Duplicate operation %s in namespace %s",
					operationName, namespaceName)
			}
			operationsMap[operationName] = true
			op.Name = operationName
			op.Param = strings.TrimSpace(op.Param)
		}
	}

	return nil
}

type Operation struct {
	Name  string
	Param string
}

type NamespacePrefix struct {
	String    string
	Variables []string
}

func newNamespacePrefix(prefix string) (*NamespacePrefix, error) {
	variables := []string{}
	for _, variable := range prefixVariable.FindAllString(prefix, -1) {
		variable = variable[1 : len(variable)-1]
		if len(variable) == 0 || !identifier.MatchString(variable) {
			return nil, fmt.Errorf("Invalid prefix variable '%s'", variable)
		}
		variables = append(variables, variable)
	}
	return &NamespacePrefix{String: prefix, Variables: variables}, nil
}

type Namespace struct {
	Name       string
	Prefix     *NamespacePrefix
	Operations []*Operation
}

func (n *Namespace) addOperation(op *Operation) {
	n.Operations = append(n.Operations, op)
}

func (n *Namespace) containsOperation(name string) bool {
	for _, op := range n.Operations {
		if op.Name == strings.Title(name) {
			return true
		}
	}
	return false
}
