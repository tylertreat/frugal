package parser

import (
	"sort"
	"strings"
)

//go:generate pigeon -o grammar.peg.go ./grammar.peg
//go:generate goimports -w ./grammar.peg.go

type Operation struct {
	Comment []string
	Name    string
	Param   string
}

type ScopePrefix struct {
	String    string
	Variables []string
}

func (n *ScopePrefix) Template() string {
	return prefixVariable.ReplaceAllString(n.String, "%s")
}

type Scope struct {
	Comment    []string
	Name       string
	Prefix     *ScopePrefix
	Operations []*Operation
}

type Frugal struct {
	Name   string
	Dir    string
	Path   string
	Scopes []*Scope
	Thrift *Thrift
}

func (f *Frugal) ContainsFrugalDefinitions() bool {
	return len(f.Scopes) > 0
}

// Imports returns a slice containing the import packages required for the
// Frugal definitions.
func (f *Frugal) Imports() []string {
	importsSet := make(map[string]bool)
	for _, scope := range f.Scopes {
		for _, op := range scope.Operations {
			if strings.Contains(op.Param, ".") {
				importsSet[op.Param[0:strings.Index(op.Param, ".")]] = true
			}
		}
	}
	imports := make([]string, 0, len(importsSet))
	for imp, _ := range importsSet {
		imports = append(imports, imp)
	}
	return imports
}

func (f *Frugal) sort() {
	sort.Sort(ScopesByName(f.Scopes))
	for _, scope := range f.Scopes {
		sort.Sort(OperationsByName(scope.Operations))
	}
}

type ScopesByName []*Scope

func (b ScopesByName) Len() int {
	return len(b)
}

func (b ScopesByName) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ScopesByName) Less(i, j int) bool {
	return b[i].Name < b[j].Name
}

type OperationsByName []*Operation

func (b OperationsByName) Len() int {
	return len(b)
}

func (b OperationsByName) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b OperationsByName) Less(i, j int) bool {
	return b[i].Name < b[j].Name
}
