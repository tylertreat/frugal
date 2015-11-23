package parser

import (
	"fmt"
	"sort"
)

type Type struct {
	Name      string
	KeyType   *Type // If map
	ValueType *Type // If map, list, or set
}

type EnumValue struct {
	Name  string
	Value int
}

type Enum struct {
	Name   string
	Values map[string]*EnumValue
}

type Constant struct {
	Name  string
	Type  *Type
	Value interface{}
}

type Field struct {
	ID       int
	Name     string
	Optional bool
	Type     *Type
	Default  interface{}
}

type Struct struct {
	Name   string
	Fields []*Field
}

type Method struct {
	Comment    string
	Name       string
	Oneway     bool
	ReturnType *Type
	Arguments  []*Field
	Exceptions []*Field
}

type Service struct {
	Name          string
	Extends       string
	Methods       map[string]*Method
	SortedMethods []*Method
}

type Thrift struct {
	Includes   map[string]string // name -> unique identifier (absolute path generally)
	Typedefs   map[string]*Type
	Namespaces map[string]string
	Constants  map[string]*Constant
	Enums      map[string]*Enum
	Structs    map[string]*Struct
	Exceptions map[string]*Struct
	Unions     map[string]*Struct
	Services   map[string]*Service
}

func (t *Thrift) sort() {
	for _, service := range t.Services {
		if service.SortedMethods == nil {
			service.SortedMethods = make([]*Method, 0, len(service.Methods))
		}
		for _, method := range service.Methods {
			service.SortedMethods = append(service.SortedMethods, method)
		}
		sort.Sort(MethodsByName(service.SortedMethods))
	}
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

type MethodsByName []*Method

func (b MethodsByName) Len() int {
	return len(b)
}

func (b MethodsByName) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b MethodsByName) Less(i, j int) bool {
	return b[i].Name < b[j].Name
}
