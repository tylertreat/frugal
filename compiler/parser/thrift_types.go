package parser

import "fmt"

type Include struct {
	Name  string
	Value string
}

type Namespace struct {
	Scope string
	Value string
}

type Type struct {
	Name      string
	KeyType   *Type // If map
	ValueType *Type // If map, list, or set
}

type TypeDef struct {
	Comment []string
	Name    string
	Type    *Type
}

type EnumValue struct {
	Comment []string
	Name    string
	Value   int
}

type Enum struct {
	Comment []string
	Name    string
	Values  map[string]*EnumValue
}

type Constant struct {
	Comment []string
	Name    string
	Type    *Type
	Value   interface{}
}

type Field struct {
	Comment  []string
	ID       int
	Name     string
	Optional bool
	Type     *Type
	Default  interface{}
}

type Struct struct {
	Comment []string
	Name    string
	Fields  []*Field
}

type Method struct {
	Comment    []string
	Name       string
	Oneway     bool
	ReturnType *Type
	Arguments  []*Field
	Exceptions []*Field
}

type Service struct {
	Comment []string
	Name    string
	Extends string
	Methods map[string]*Method
}

type Thrift struct {
	Includes   []*Include
	Typedefs   []*TypeDef
	Namespaces []*Namespace
	Constants  []*Constant
	Enums      []*Enum
	Structs    []*Struct
	Exceptions []*Struct
	Unions     []*Struct
	Services   []*Service

	typedefIndex   map[string]*TypeDef
	namespaceIndex map[string]*Namespace
}

func (t *Thrift) UnderlyingType(typeName string) string {
	if typedef, ok := t.typedefIndex[typeName]; ok {
		typeName = typedef.Type.Name
	}
	return typeName
}

func (t *Thrift) Namespace(scope string) (string, bool) {
	namespace, ok := t.namespaceIndex[scope]
	value := ""
	if ok {
		value = namespace.Value
	}
	return value, ok
}

type Identifier string

type KeyValue struct {
	Key, Value interface{}
}

func (t *Type) String() string {
	switch t.Name {
	case "map":
		return fmt.Sprintf("map<%s,%s>", t.KeyType.String(), t.ValueType.String())
	case "list":
		return fmt.Sprintf("list<%s>", t.ValueType.String())
	case "set":
		return fmt.Sprintf("set<%s>", t.ValueType.String())
	}
	return t.Name
}
