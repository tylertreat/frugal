package parser

import "sort"

//go:generate pigeon -o grammar.peg.go ./grammar.peg
//go:generate goimports -w ./grammar.peg.go

type Operation struct {
	Comment string
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
	Name       string
	Prefix     *ScopePrefix
	Operations []*Operation
}

type Frugal struct {
	Name       string
	Dir        string
	Path       string
	Namespaces map[string]string
	Scopes     []*Scope
}

func (f *Frugal) sort() {
	sort.Sort(ScopesByName(f.Scopes))
	for _, scope := range f.Scopes {
		sort.Sort(OperationsByName(scope.Operations))
	}
}

type Identifier string

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
