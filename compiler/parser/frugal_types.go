package parser

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

//go:generate pigeon -o grammar.peg.go ./grammar.peg
//go:generate goimports -w ./grammar.peg.go

// LowercaseFirstLetter of the string.
func LowercaseFirstLetter(s string) string {
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Operation is a pub/sub scope operation. Corresponding publish and
// subscribe methods are generated from this for publishers and subscribers,
// respectively.
type Operation struct {
	Comment     []string
	Name        string
	Type        *Type
	Annotations Annotations
	Scope       *Scope // Pointer back to containing Scope
}

// ScopePrefix is the string prefix prepended to a pub/sub topic. The string
// can contain variables of the form {foo}, e.g. "foo.{bar}.baz" where "bar"
// is supplied at publish/subscribe time.
type ScopePrefix struct {
	String    string
	Variables []string
}

// Template returns the prefix where variables are replaced with the given
// string.
func (n *ScopePrefix) Template(s string) string {
	return prefixVariable.ReplaceAllString(n.String, s)
}

// Scope is a pub/sub namespace.
type Scope struct {
	Comment     []string
	Name        string
	Prefix      *ScopePrefix
	Operations  []*Operation
	Annotations Annotations
	Frugal      *Frugal // Pointer back to containing Frugal
}

// ReferencedIncludes returns a slice containing the referenced includes which
// will need to be imported in generated code for this Scope.
func (s *Scope) ReferencedIncludes() ([]*Include, error) {
	var err error
	includes := []*Include{}
	includesSet := make(map[string]*Include)
	for _, op := range s.Operations {
		includesSet, includes, err = addInclude(includesSet, includes, op.Type, s.Frugal.Thrift)
		if err != nil {
			return nil, err
		}
	}
	return includes, nil
}

func (s *Scope) assignScope() {
	for _, op := range s.Operations {
		op.Scope = s
	}
}

// IdentifierType indicates if a Identifier is a local/include constant/enum
type IdentifierType uint8

const (
	NonIdentifier IdentifierType = iota
	LocalConstant
	LocalEnum
	IncludeConstant
	IncludeEnum
)

// IdentifierContext gives information about the identifier for a given
// contextual reference
type IdentifierContext struct {
	Type      IdentifierType
	Constant  *Constant
	Enum      *Enum
	EnumValue *EnumValue
	Include   *Frugal
}

// Frugal contains the complete IDL parse tree.
type Frugal struct {
	Name           string
	File           string
	Dir            string
	Path           string
	Scopes         []*Scope
	Thrift         *Thrift
	ParsedIncludes map[string]*Frugal
}

// NamespaceForInclude returns the Namespace for the given inclue name and
// language.
func (f *Frugal) NamespaceForInclude(include, lang string) *Namespace {
	parsed, ok := f.ParsedIncludes[include]
	if !ok {
		return nil
	}
	return parsed.Thrift.Namespace(lang)
}

// ContainsFrugalDefinitions indicates if the parse tree contains any
// scope or service definitions.
func (f *Frugal) ContainsFrugalDefinitions() bool {
	return len(f.Scopes)+len(f.Thrift.Services) > 0
}

// ReferencedScopeIncludes returns a slice containing the referenced includes
// which will need to be imported in generated code for scopes.
func (f *Frugal) ReferencedScopeIncludes() ([]*Include, error) {
	includeNames := []string{}
	includesSet := make(map[string]*Include)
	for _, scope := range f.Scopes {
		scopeIncludes, err := scope.ReferencedIncludes()
		if err != nil {
			return nil, err
		}
		for _, include := range scopeIncludes {
			if _, ok := includesSet[include.Name]; !ok {
				includesSet[include.Name] = include
				includeNames = append(includeNames, include.Name)
			}
		}
	}
	sort.Strings(includeNames)
	includes := make([]*Include, len(includeNames))
	for i, include := range includeNames {
		includes[i] = includesSet[include]
	}
	return includes, nil
}

// ReferencedServiceIncludes returns a slice containing the referenced includes
// which will need to be imported in generated code for services.
func (f *Frugal) ReferencedServiceIncludes() ([]*Include, error) {
	includeNames := []string{}
	includesSet := make(map[string]*Include)
	for _, service := range f.Thrift.Services {
		servIncludes, err := service.ReferencedIncludes()
		if err != nil {
			return nil, err
		}
		for _, include := range servIncludes {
			if _, ok := includesSet[include.Name]; !ok {
				includesSet[include.Name] = include
				includeNames = append(includeNames, include.Name)
			}
		}
	}
	sort.Strings(includeNames)
	includes := make([]*Include, len(includeNames))
	for i, include := range includeNames {
		includes[i] = includesSet[include]
	}
	return includes, nil
}

// UnderlyingType follows any typedefs to get the base IDL type.
func (f *Frugal) UnderlyingType(t *Type) *Type {
	if t == nil {
		panic("Attempted to get underlying type of nil type")
	}
	typedefIndex := f.Thrift.typedefIndex
	include := t.IncludeName()
	if include != "" {
		parsed, ok := f.ParsedIncludes[include]
		if !ok {
			return t
		}
		typedefIndex = parsed.Thrift.typedefIndex
	}
	if typedef, ok := typedefIndex[t.ParamName()]; ok {
		// Recursively call underlying type to handle typedef nesting.
		return f.UnderlyingType(typedef.Type)
	}
	return t
}

// ConstantFromField returns a new Constant from the given Field and value.
func (f *Frugal) ConstantFromField(field *Field, value interface{}) *Constant {
	return &Constant{
		Name:  field.Name,
		Type:  field.Type,
		Value: value,
	}
}

// ContextFromIdentifier returns a IdentifierContext for the given Identifier.
func (f *Frugal) ContextFromIdentifier(identifier Identifier) *IdentifierContext {
	name := string(identifier)

	// Split based on '.', if present, it should be from an include.
	pieces := strings.Split(name, ".")
	if len(pieces) == 1 {
		// From this file.
		for _, constant := range f.Thrift.Constants {
			if name == constant.Name {
				return &IdentifierContext{
					Type:     LocalConstant,
					Constant: constant,
				}
			}
		}
	} else if len(pieces) == 2 {
		// Either from an include or part of an enum.
		for _, enum := range f.Thrift.Enums {
			if pieces[0] == enum.Name {
				for _, value := range enum.Values {
					if pieces[1] == value.Name {
						return &IdentifierContext{
							Type:      LocalEnum,
							Enum:      enum,
							EnumValue: value,
						}
					}
				}
			}
		}

		// If not part of an enum, it's from an include.
		include, ok := f.ParsedIncludes[pieces[0]]
		if !ok {
			panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
		}
		for _, constant := range include.Thrift.Constants {
			if pieces[1] == constant.Name {
				return &IdentifierContext{
					Type:     IncludeConstant,
					Constant: constant,
					Include:  include,
				}
			}
		}
	} else if len(pieces) == 3 {
		// Enum from an include.
		include, ok := f.ParsedIncludes[pieces[0]]
		if !ok {
			panic(fmt.Sprintf("referenced include '%s' in constant '%s' not present", pieces[0], name))
		}
		for _, enum := range include.Thrift.Enums {
			if pieces[1] == enum.Name {
				for _, value := range enum.Values {
					if pieces[2] == value.Name {
						return &IdentifierContext{
							Type:      IncludeEnum,
							Enum:      enum,
							EnumValue: value,
							Include:   include,
						}
					}
				}
				panic(fmt.Sprintf("referenced value '%s' of enum '%s' doesn't exist", pieces[1], pieces[0]))
			}
		}
	}

	panic("referenced constant doesn't exist: " + name)
}

// IsStruct indicates if the underlying Type is a struct.
func (f *Frugal) IsStruct(t *Type) bool {
	t = f.UnderlyingType(t)
	if _, ok := thriftBaseTypes[t.Name]; ok {
		return false
	}
	return t.KeyType == nil && t.ValueType == nil && !f.IsEnum(t)
}

// IsUnion indicates if the underlying types is a union.
func (f *Frugal) IsUnion(t *Type) bool {
	t = f.UnderlyingType(t)

	frugal := f
	if t.IncludeName() != "" {
		// The type is from an include
		frugal = f.ParsedIncludes[t.IncludeName()]
	}

	for _, union := range frugal.Thrift.Unions {
		if t.ParamName() == union.Name {
			return true
		}
	}
	return false
}

// IsEnum indicates if the underlying Type is an enum.
func (f *Frugal) IsEnum(t *Type) bool {
	include := t.IncludeName()
	containingFrugal := f
	if include != "" {
		if containing, ok := f.ParsedIncludes[include]; ok {
			containingFrugal = containing
		}
	}
	for _, enum := range containingFrugal.Thrift.Enums {
		if enum.Name == t.ParamName() {
			return true
		}
	}
	return false
}

func (f *Frugal) assignFrugal() {
	for _, scope := range f.Scopes {
		scope.assignScope()
	}
}

// validate parsed Frugal IDL by ensuring there are no duplicate service/scope
// names and the Thrift is valid.
func (f *Frugal) validate() error {
	// Ensure there are no duplicate names between services and scopes.
	names := make(map[string]bool)
	for _, service := range f.Thrift.Services {
		// Since not every language supports (exported) upper/lowercase
		// class first letters, index by lowercasing the first letter.
		serviceLower := LowercaseFirstLetter(service.Name)
		if _, ok := names[serviceLower]; ok {
			return fmt.Errorf("Duplicate service name %s", service.Name)
		}
		names[serviceLower] = true

		methodNames := make(map[string]bool)
		for _, method := range service.Methods {
			// Since not every language supports (exported) upper/lowercase
			// method first letters, index by lowercasing the first letter.
			methodLower := LowercaseFirstLetter(method.Name)
			if _, ok := methodNames[methodLower]; ok {
				return fmt.Errorf("Duplicate method name %s", method.Name)
			}
			methodNames[methodLower] = true
		}
	}

	names = make(map[string]bool)
	for _, scope := range f.Scopes {
		// Since not every language supports (exported) upper/lowercase
		// class first letters, index by lowercasing the first letter.
		scopeLower := LowercaseFirstLetter(scope.Name)
		if _, ok := names[scopeLower]; ok {
			return fmt.Errorf("Duplicate scope name %s", scope.Name)
		}
		names[scopeLower] = true

		opNames := make(map[string]bool)
		for _, op := range scope.Operations {
			// Since not every language supports (exported) upper/lowercase
			// method first letters, index by lowercasing the first letter.
			opLower := LowercaseFirstLetter(op.Name)
			if _, ok := opNames[opLower]; ok {
				return fmt.Errorf("Duplicate operation name %s", op.Name)
			}
			opNames[opLower] = true
		}
	}

	return f.Thrift.validate(f.ParsedIncludes)
}

func (f *Frugal) sort() {
	sort.Sort(scopesByName(f.Scopes))
}

type scopesByName []*Scope

func (b scopesByName) Len() int {
	return len(b)
}

func (b scopesByName) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b scopesByName) Less(i, j int) bool {
	return b[i].Name < b[j].Name
}
