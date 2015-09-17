package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"
)

var (
	identifier    = regexp.MustCompile("^[A-Za-z]+[A-Za-z0-9]")
	defaultPrefix = &NamespacePrefix{String: "", Variables: make([]string, 0)}
)

const (
	namespace          = "namespace"
	prefix             = "prefix"
	openDefinition     = "{"
	closeDefinition    = "}"
	operationDelimiter = ":"
)

const (
	stateStartNamespace = iota
	stateNamespaceName
	stateOpenDefinition
	statePrefix
	stateOperationName
	stateOperationDelimiter
	stateOperationParam
)

func Parse(filePath string) (*Program, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	name, err := getName(file)
	if err != nil {
		return nil, err
	}

	var (
		program          = &Program{Name: name, Path: filepath.Dir(file.Name())}
		namespacesMap    = make(map[string]*Namespace)
		namespaces       = []*Namespace{}
		state            = stateStartNamespace
		currentNamespace *Namespace
		currentOperation *Operation
	)

	// TODO: Add support for comments.

	var (
		s     scanner.Scanner
		token rune
	)
	s.Init(file)

	for token != scanner.EOF {
		token = s.Scan()
		tokenText := s.TokenText()
		if tokenText == "" {
			continue
		}
		switch state {
		case stateStartNamespace:
			if tokenText != namespace {
				goto parseErr
			}
			currentNamespace = &Namespace{Operations: []*Operation{}}
			state = stateNamespaceName
			continue
		case stateNamespaceName:
			name := identifier.FindString(tokenText)
			if name == "" {
				goto parseErr
			}
			if _, ok := namespacesMap[name]; ok {
				return nil, fmt.Errorf("Duplicate definition for namespace '%s': %s:%s",
					name, filePath, s.Pos())
			}
			currentNamespace.Name = strings.Title(name)
			state = stateOpenDefinition
			continue
		case stateOpenDefinition:
			if tokenText != openDefinition {
				goto parseErr
			}
			state = stateOperationName
			continue
		case statePrefix:
			if !strings.HasPrefix(tokenText, `"`) && !strings.HasSuffix(tokenText, `"`) {
				goto parseErr
			}
			prefixStr := tokenText[1 : len(tokenText)-1]
			nsPrefix, err := newNamespacePrefix(prefixStr)
			if err != nil {
				return nil, err
			}
			currentNamespace.Prefix = nsPrefix
			state = stateOperationName
			continue
		case stateOperationName:
			if tokenText == prefix {
				if currentNamespace.Prefix != nil {
					return nil, fmt.Errorf("Duplicate prefix definition for namespace '%s': %s:%s",
						currentNamespace.Name, filePath, s.Pos())
				}
				state = statePrefix
				continue
			}
			if tokenText == closeDefinition {
				namespacesMap[currentNamespace.Name] = currentNamespace
				state = stateStartNamespace
				continue
			}
			name := identifier.FindString(tokenText)
			if name == "" {
				goto parseErr
			}
			name = strings.Title(name)
			if currentNamespace.containsOperation(name) {
				return nil, fmt.Errorf("Duplicate definition for operation '%s': %s:%s",
					name, filePath, s.Pos())
			}
			currentOperation = &Operation{Name: name}
			state = stateOperationDelimiter
			continue
		case stateOperationDelimiter:
			if tokenText != operationDelimiter {
				goto parseErr
			}
			state = stateOperationParam
			continue
		case stateOperationParam:
			param := identifier.FindString(tokenText)
			if param == "" {
				goto parseErr
			}
			currentOperation.Param = param
			currentNamespace.addOperation(currentOperation)
			state = stateOperationName
			continue
		default:
			goto parseErr
		}
	}

	if state != stateStartNamespace {
		goto parseErr
	}

	for _, namespace := range namespacesMap {
		if namespace.Prefix == nil {
			namespace.Prefix = defaultPrefix
		}
		namespaces = append(namespaces, namespace)
	}
	program.Namespaces = namespaces

	return program, program.validate()

parseErr:
	return nil, fmt.Errorf("Invalid syntax: %s:%s", filePath, s.Pos())
}

func getName(f *os.File) (string, error) {
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	parts := strings.Split(info.Name(), ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid file: %s", f.Name())
	}
	return parts[0], nil
}
