package java

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/Workiva/frugal/compiler/parser"
	"github.com/Workiva/frugal/compiler/plugin"
)

const (
	writeEnumValue        = "WriteEnumValue"
	writeEnumFields       = "WriteEnumFields"
	writeEnumConstructors = "WriteEnumConstructors"
	writeEnumMethods      = "WriteEnumMethods"
)

type line struct {
	indent  uint
	content string
}

// Printer prints code to a buffer.
// TODO: this is probably generic enough for all languages. May want to move
// it.
type Printer struct {
	output []line
	indent uint
}

func (p *Printer) ScopeUp() *Printer {
	p.indent++
	return p
}

func (p *Printer) ScopeDown() *Printer {
	p.indent--
	return p
}

func (p *Printer) Println(ln string) *Printer {
	p.output = append(p.output, line{indent: p.indent, content: ln + "\n"})
	return p
}

func (p Printer) Output(indentCount uint, indent string) string {
	output := ""
	baseIndent := ""
	for i := uint(0); i < indentCount; i++ {
		baseIndent += indent
	}
	for _, line := range p.output {
		output += baseIndent
		for i := uint(0); i < line.indent+1; i++ {
			output += indent
		}
		output += line.content
	}
	return output
}

// Field is a Java class member variable.
type Field struct {
	Modifier     string
	Final        bool
	Type         string
	Name         string
	InitialValue string
}

func (f Field) format() string {
	modifier := ""
	if f.Modifier != "" {
		modifier = f.Modifier
	}
	if f.Final {
		if modifier != "" {
			modifier += " "
		}
		modifier += "final"
	}
	if modifier != "" {
		modifier += " "
	}
	formatted := fmt.Sprintf("%s%s %s", modifier, f.Type, f.Name)
	if f.InitialValue != "" {
		formatted += fmt.Sprintf(" = %s", f.InitialValue)
	}
	return formatted
}

// Argument is a Java method or constructor parameter.
type Argument struct {
	Final bool
	Type  string
	Name  string
}

func (a Argument) format() string {
	modifier := ""
	if a.Final {
		modifier = "final "
	}
	return fmt.Sprintf("%s%s %s", modifier, a.Type, a.Name)
}

// Constructor is a Java class constructor.
type Constructor struct {
	Modifier  string
	Arguments []Argument
	Body      *Printer
}

func (c Constructor) format(indentCount uint, indent, name string) string {
	modifier := ""
	if c.Modifier != "" {
		modifier = c.Modifier + " "
	}
	output := fmt.Sprintf("%s%s%s(", indent, modifier, name)
	prefix := ""
	for _, arg := range c.Arguments {
		output += fmt.Sprintf("%s%s", prefix, arg.format())
		prefix = ", "
	}
	output += ") {\n"
	output += c.Body.Output(indentCount, indent)
	output += indent + "}\n\n"
	return output
}

// EnumConstructor is a Java enum constructor (Modifier is ignored).
type EnumConstructor Constructor

func (e EnumConstructor) format(indentCount uint, indent, name string) string {
	output := fmt.Sprintf("%sprivate %s(", indent, name)
	prefix := ""
	for _, arg := range e.Arguments {
		output += fmt.Sprintf("%s%s", prefix, arg.format())
		prefix = ", "
	}
	output += ") {\n"
	output += e.Body.Output(indentCount, indent)
	output += indent + "}\n\n"
	return output
}

// Method is a Java class method.
type Method struct {
	Modifier   string
	Name       string
	ReturnType string
	Arguments  []Argument
	Body       *Printer
}

func (m Method) format(indentCount uint, indent string) string {
	modifier := ""
	if m.Modifier != "" {
		modifier = m.Modifier + " "
	}
	output := fmt.Sprintf("%s%s%s %s(", indent, modifier, m.ReturnType, m.Name)
	prefix := ""
	for _, arg := range m.Arguments {
		output += fmt.Sprintf("%s%s", prefix, arg.format())
		prefix = ", "
	}
	output += ") {\n"
	output += m.Body.Output(indentCount, indent)
	output += indent + "}\n\n"
	return output
}

// EnumValue is a Java enum value.
type EnumValue struct {
	Name      string
	Arguments []string
}

func (e EnumValue) format() string {
	formatted := e.Name
	if len(e.Arguments) > 0 {
		formatted += "("
		prefix := ""
		for _, arg := range e.Arguments {
			formatted += prefix + arg
			prefix = ", "
		}
		formatted += ")"
	}
	return formatted
}

// Writer exposes hooks for Java code generation.
// TODO: "Writer" isn't a great name for this since it doesn't actually write
// anything, it just modifies the generated output. Think of a better name...
type Writer interface {
	// WriteEnumValue is called when generating enum values, e.g.
	//
	// public enum MyEnum implements TEnum {
	//     FOO(1), // produced by a call to WriteEnumValue
	//     BAR(2),
	//     BAZ(3);
	//
	//     ...
	//
	// Note that Frugal requires the int value for wire-level compatibility.
	WriteEnumValue(enum *parser.Enum, value *parser.EnumValue) (EnumValue, error)

	// WriteEnumFields is called when generating enum fields, e.g.
	//
	// private final int value; // produced by a call to WriteEnumFields
	//
	// Note that Frugal requires the int value for wire-level compatibility.
	WriteEnumFields(enum *parser.Enum) ([]Field, error)

	// WriteEnumConstructors is called when generating enum constructors, e.g.
	//
	// MyEnum(int value) {      // produced by
	//     this.value = value;  // a call to
	// }                        // WriteEnumConstructors
	//
	// Note that Frugal requires the int value for wire-level compatibility.
	WriteEnumConstructors(enum *parser.Enum) ([]EnumConstructor, error)

	// WriteEnumMethods is called when generating enum methods, e.g.
	//
	// public int getValue() {
	//     return value;
	// }
	//
	// Note that Frugal requires the int value for wire-level compatibility.
	WriteEnumMethods(enum *parser.Enum) ([]Method, error)
}

// defaultWriter implements the Writer interface and provides default generate
// behavior for Java.
type defaultWriter struct{}

// WriteEnumValue is called when generating enum values, i.e.
//
// public enum MyEnum implements TEnum {
//     FOO(1), // produced by a call to WriteEnumValue
//     BAR(2),
//     BAZ(3);
//
//     ...
//
// Note that Frugal requires the int value for wire-level compatibility.
func (d *defaultWriter) WriteEnumValue(enum *parser.Enum, value *parser.EnumValue) (EnumValue, error) {
	// FOO(1), BAR(2), etc.
	return EnumValue{Name: value.Name, Arguments: []string{strconv.Itoa(value.Value)}}, nil
}

// WriteEnumFields is called when generating enum fields, e.g.
//
// private final int value; // produced by a call to WriteEnumFields
//
// Note that Frugal requires the int value for wire-level compatibility.
func (d *defaultWriter) WriteEnumFields(enum *parser.Enum) ([]Field, error) {
	// private final int value
	value := Field{
		Modifier: "private",
		Final:    true,
		Type:     "int",
		Name:     "value",
	}
	return []Field{value}, nil
}

// WriteEnumConstructors is called when generating enum constructors, e.g.
//
// MyEnum(int value) {      // produced by
//     this.value = value;  // a call to
// }                        // WriteEnumConstructors
//
// Note that Frugal requires the int value for wire-level compatibility.
func (d *defaultWriter) WriteEnumConstructors(enum *parser.Enum) ([]EnumConstructor, error) {
	// Default value constructor.
	printer := new(Printer)
	printer.Println("this.value = value;")
	constructor := EnumConstructor{
		Modifier:  "private",
		Arguments: []Argument{Argument{Type: "int", Name: "value"}},
		Body:      printer,
	}
	return []EnumConstructor{constructor}, nil
}

// WriteEnumMethods is called when generating enum methods, e.g.
//
// public int getValue() {
//     return value;
// }
//
// Note that Frugal requires the int value for wire-level compatibility.
func (d *defaultWriter) WriteEnumMethods(enum *parser.Enum) ([]Method, error) {
	// getValue() method.
	printer := new(Printer)
	printer.Println("return value;")
	method := Method{
		Modifier:   "public",
		Name:       "getValue",
		ReturnType: "int",
		Body:       printer,
	}
	return []Method{method}, nil
}

// pluginWriter implements the Writer interface and calls out to a
// user-provided plugin to provide generate behavior for Java.
type pluginWriter struct {
	plugin        *plugin.FrugalPlugin
	defaultWriter Writer
}

// WriteEnumValue is called when generating enum values, i.e.
//
// public enum MyEnum implements TEnum {
//     FOO(1), // produced by a call to WriteEnumValue
//     BAR(2),
//     BAZ(3);
//
//     ...
//
// Note that Frugal requires the int value for wire-level compatibility.
func (p *pluginWriter) WriteEnumValue(enum *parser.Enum, value *parser.EnumValue) (EnumValue, error) {
	f := p.plugin.Lookup(writeEnumValue)
	if f == nil {
		return p.defaultWriter.WriteEnumValue(enum, value)
	}
	writer, ok := f.(func(*parser.Enum, *parser.EnumValue) (EnumValue, error))
	if !ok {
		return EnumValue{}, fmt.Errorf("%s is %s, not func(*parser.Enum, *parser.Value) (string, error)",
			writeEnumValue, reflect.TypeOf(f))
	}
	return writer(enum, value)
}

// WriteEnumFields is called when generating enum fields, e.g.
//
// private final int value; // produced by a call to WriteEnumFields
//
// Note that Frugal requires the int value for wire-level compatibility.
func (p *pluginWriter) WriteEnumFields(enum *parser.Enum) ([]Field, error) {
	defaultFields, err := p.defaultWriter.WriteEnumFields(enum)
	if err != nil {
		return nil, err
	}
	f := p.plugin.Lookup(writeEnumFields)
	if f == nil {
		return defaultFields, nil
	}
	writer, ok := f.(func(*parser.Enum) ([]Field, error))
	if !ok {
		return nil, fmt.Errorf("%s is %s, not func(*parser.Enum) ([]java.Field, error)",
			writeEnumFields, reflect.TypeOf(f))
	}
	fields, err := writer(enum)
	if err != nil {
		return nil, err
	}
	// Include default fields (private final int value).
	return append(fields, defaultFields...), nil
}

// WriteEnumConstructors is called when generating enum constructors, e.g.
//
// MyEnum(int value) {      // produced by
//     this.value = value;  // a call to
// }                        // WriteEnumConstructors
//
// Note that Frugal requires the int value for wire-level compatibility.
func (p *pluginWriter) WriteEnumConstructors(enum *parser.Enum) ([]EnumConstructor, error) {
	defaultConstructors, err := p.defaultWriter.WriteEnumConstructors(enum)
	if err != nil {
		return nil, err
	}
	f := p.plugin.Lookup(writeEnumConstructors)
	if f == nil {
		return defaultConstructors, nil
	}
	writer, ok := f.(func(*parser.Enum) ([]EnumConstructor, error))
	if !ok {
		return nil, fmt.Errorf("%s is %s, not func(*parser.Enum) ([]java.EnumConstructor, error)",
			writeEnumConstructors, reflect.TypeOf(f))
	}
	constructors, err := writer(enum)
	if err != nil {
		return nil, err
	}
	// Include default constructors (private value constructor).
	return append(constructors, defaultConstructors...), nil
}

// WriteEnumMethods is called when generating enum methods, e.g.
//
// public int getValue() {
//     return value;
// }
//
// Note that Frugal requires the int value for wire-level compatibility.
func (p *pluginWriter) WriteEnumMethods(enum *parser.Enum) ([]Method, error) {
	defaultMethods, err := p.defaultWriter.WriteEnumMethods(enum)
	if err != nil {
		return nil, err
	}
	f := p.plugin.Lookup(writeEnumMethods)
	if f == nil {
		return defaultMethods, nil
	}
	writer, ok := f.(func(*parser.Enum) ([]Method, error))
	if !ok {
		return nil, fmt.Errorf("%s is %s, not func(*parser.Enum) ([]java.Method, error)",
			writeEnumMethods, reflect.TypeOf(f))
	}
	methods, err := writer(enum)
	if err != nil {
		return nil, err
	}
	// Include default methods (public int getValue()).
	return append(methods, defaultMethods...), nil
}
