package parser

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
	Operations map[string]*Operation
}

type Frugal struct {
	Name       string
	Dir        string
	Path       string
	Namespaces map[string]string
	Scopes     map[string]*Scope
}

type Identifier string
