package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/Workiva/frugal/compiler/parser"
)

func generateThriftIDL(frugal *parser.Frugal) (string, error) {
	file := fmt.Sprintf("%s.thrift", frugal.Name)
	f, err := os.Create(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contents := ""
	thrift := frugal.Thrift

	// Generate namespaces.
	for lang, namespace := range thrift.Namespaces {
		contents += fmt.Sprintf("namespace %s %s\n", lang, namespace)
	}
	contents += "\n"

	// Generate includes.
	for _, include := range thrift.Includes {
		contents += fmt.Sprintf("include \"%s\"\n", include)
	}
	contents += "\n"

	// Generate constants.
	for _, constant := range thrift.Constants {
		value := constant.Value
		if constant.Type.Name == "string" {
			value = fmt.Sprintf(`"%s"`, value)
		}
		contents += fmt.Sprintf("const %s %s = %s\n", constant.Type, constant.Name, value)
	}
	contents += "\n"

	// Generate typedefs.
	for thriftType, typedef := range thrift.Typedefs {
		contents += fmt.Sprintf("typedef %s %s\n", thriftType, typedef)
	}
	contents += "\n"

	// Generate enums.
	for _, enum := range thrift.Enums {
		contents += fmt.Sprintf("enum %s {\n", enum.Name)
		values := make([]*parser.EnumValue, 0, len(enum.Values))
		for _, value := range enum.Values {
			values = append(values, value)
		}
		sort.Sort(enumValues(values))
		for _, value := range values {
			contents += fmt.Sprintf("\t%s,\n", value.Name)
		}
		contents += "}\n\n"
	}

	// Generate structs.
	for _, strct := range thrift.Structs {
		contents += fmt.Sprintf("struct %s {\n", strct.Name)
		for _, field := range strct.Fields {
			contents += fmt.Sprintf("\t%d: ", field.ID)
			if field.Optional {
				contents += "optional "
			} else {
				contents += "required "
			}
			contents += fmt.Sprintf("%s %s", field.Type.String(), field.Name)
			if field.Default != nil {
				def := field.Default
				defStr := ""
				switch d := def.(type) {
				case string:
					defStr = fmt.Sprintf(`"%s"`, d)
				default:
					defStr = fmt.Sprintf("%v", d)
				}
				contents += fmt.Sprintf(" = %s", defStr)
			}
			contents += ",\n"
		}
		contents += "}\n\n"
	}

	// Generate exceptions.
	for _, exception := range thrift.Exceptions {
		contents += fmt.Sprintf("exception %s {\n", exception.Name)
		for _, field := range exception.Fields {
			contents += fmt.Sprintf("\t%d: ", field.ID)
			if field.Optional {
				contents += "optional "
			} else {
				contents += "required "
			}
			contents += fmt.Sprintf("%s %s", field.Type.String(), field.Name)
			if field.Default != nil {
				def := field.Default
				defStr := ""
				switch d := def.(type) {
				case string:
					defStr = fmt.Sprintf(`"%s"`, d)
				default:
					defStr = fmt.Sprintf("%v", d)
				}
				contents += fmt.Sprintf(" = %s", defStr)
			}
			contents += ",\n"
		}
		contents += "}\n\n"
	}

	// Generate services.
	for _, service := range thrift.Services {
		contents += fmt.Sprintf("service %s ", service.Name)
		if service.Extends != "" {
			contents += fmt.Sprintf("%s ", service.Extends)
		}
		contents += "{\n"
		for _, method := range service.Methods {
			contents += "\t"
			if method.Oneway {
				contents += "oneway "
			}
			if method.ReturnType == nil {
				contents += "void "
			} else {
				contents += fmt.Sprintf("%s ", method.ReturnType.String())
			}
			contents += fmt.Sprintf("%s(", method.Name)
			prefix := ""
			for _, arg := range method.Arguments {
				modifier := "required"
				if arg.Optional {
					modifier = "optional"
				}
				contents += fmt.Sprintf("%s%d:%s %s %s", prefix, arg.ID,
					modifier, arg.Type.String(), arg.Name)
				if arg.Default != nil {
					def := arg.Default
					defStr := ""
					switch d := def.(type) {
					case string:
						defStr = fmt.Sprintf(`"%s"`, d)
					default:
						defStr = fmt.Sprintf("%v", d)
					}
					contents += fmt.Sprintf(" = %s", defStr)
				}
				prefix = ", "
			}
			contents += ")"
			if len(method.Exceptions) > 0 {
				contents += " throws ("
				prefix := ""
				for _, exception := range method.Exceptions {
					contents += fmt.Sprintf("%s%d:%s %s", prefix, exception.ID,
						exception.Type.String(), exception.Name)
					prefix = ", "
				}
				contents += ")"
			}
			contents += ",\n\n"
		}
		contents += "}\n\n"
	}

	// TODO: Generate unions.

	_, err = f.WriteString(contents)
	return file, err
}

func generateThrift(out, gen, file string) error {
	args := []string{"-r"}
	if out != "" {
		args = append(args, "-out", out)
	}
	args = append(args, "-gen", gen)
	args = append(args, file)
	if out, err := exec.Command("thrift", args...).CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	}

	// Remove the intermediate Thrift file.
	return os.Remove(file)
}

type enumValues []*parser.EnumValue

func (e enumValues) Len() int {
	return len(e)
}

func (e enumValues) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e enumValues) Less(i, j int) bool {
	return e[i].Value < e[j].Value
}
