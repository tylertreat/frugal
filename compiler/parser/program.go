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
	Scopes     []*Scope
	Namespaces map[string]string
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

	scopesMap := make(map[string]bool, len(p.Scopes))
	for _, scope := range p.Scopes {
		scopeName := strings.Title(strings.TrimSpace(scope.Name))
		if !identifier.MatchString(scopeName) {
			return fmt.Errorf("Invalid scope name %s", scopeName)
		}
		if _, ok := scopesMap[scopeName]; ok {
			return fmt.Errorf("Duplicate scope %s", scopeName)
		}
		scopesMap[scopeName] = true
		scope.Name = scopeName

		operationsMap := make(map[string]bool, len(scope.Operations))
		for _, op := range scope.Operations {
			operationName := strings.Title(strings.TrimSpace(op.Name))
			if !identifier.MatchString(operationName) {
				return fmt.Errorf("Invalid operation name %s in scope %s",
					operationName, scopeName)
			}
			if _, ok := operationsMap[operationName]; ok {
				return fmt.Errorf("Duplicate operation %s in scope %s",
					operationName, scopeName)
			}
			operationsMap[operationName] = true
			op.Name = operationName
			op.Param = strings.TrimSpace(op.Param)
		}
	}

	for namespace, definition := range p.Namespaces {
		definition = strings.TrimSpace(definition)
		if definition == "" {
			return fmt.Errorf("Invalid namespace definition %s", namespace)
		}
	}

	return nil
}

type Operation struct {
	Name  string
	Param string
}

type ScopePrefix struct {
	String    string
	Variables []string
}

func newScopePrefix(prefix string) (*ScopePrefix, error) {
	variables := []string{}
	for _, variable := range prefixVariable.FindAllString(prefix, -1) {
		variable = variable[1 : len(variable)-1]
		if len(variable) == 0 || !identifier.MatchString(variable) {
			return nil, fmt.Errorf("Invalid prefix variable '%s'", variable)
		}
		variables = append(variables, variable)
	}
	return &ScopePrefix{String: prefix, Variables: variables}, nil
}

func (n *ScopePrefix) Template() string {
	return prefixVariable.ReplaceAllString(n.String, "%s")
}

type Scope struct {
	Name       string
	Prefix     *ScopePrefix
	Operations []*Operation
}

func (s *Scope) addOperation(op *Operation) {
	s.Operations = append(s.Operations, op)
}

func (s *Scope) containsOperation(name string) bool {
	for _, op := range s.Operations {
		if op.Name == strings.Title(name) {
			return true
		}
	}
	return false
}
