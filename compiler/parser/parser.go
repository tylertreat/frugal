package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/scanner"
)

var (
	identifier    = regexp.MustCompile("^[A-Za-z]+[A-Za-z0-9]")
	defaultPrefix = &ScopePrefix{String: "", Variables: make([]string, 0)}
)

const (
	namespace          = "namespace"
	scope              = "scope"
	prefix             = "prefix"
	openDefinition     = "{"
	closeDefinition    = "}"
	operationDelimiter = ":"
)

const (
	stateStartNamespace = iota
	stateNamespaceLang
	stateNamespaceDefinition
	stateStartScope
	stateScopeName
	stateOpenDefinition
	statePrefix
	stateOperationName
	stateOperationDelimiter
	stateOperationParam
)

// Parse the Frugal file at the given path and produce a Program.
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
		program = &Program{
			Name:       name,
			Dir:        filepath.Dir(file.Name()),
			Path:       filePath,
			Namespaces: make(map[string]string),
		}
		scopesMap                  = make(map[string]*Scope)
		scopes                     = []*Scope{}
		state                      = stateStartScope
		currentNamespace           string
		currentNamespaceDefinition string
		currentScope               *Scope
		currentOperation           *Operation
	)

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
		case stateNamespaceLang:
			if _, ok := program.Namespaces[tokenText]; ok {
				return nil, fmt.Errorf("Duplicate namespace definition '%s': %s:%s",
					tokenText, filePath, s.Pos())
			}
			currentNamespace = tokenText
			state = stateNamespaceDefinition
			continue
		case stateNamespaceDefinition:
			if tokenText == scope {
				if currentNamespaceDefinition == "" {
					return nil, fmt.Errorf("Invalid namespace definition '%s': %s:%s",
						currentNamespace, filePath, s.Pos())
				}
				program.Namespaces[currentNamespace] = currentNamespaceDefinition
				currentNamespaceDefinition = ""
				currentScope = &Scope{Operations: []*Operation{}}
				state = stateScopeName
				continue
			}
			if tokenText == namespace {
				if currentNamespaceDefinition == "" {
					return nil, fmt.Errorf("Invalid namespace definition '%s': %s:%s",
						currentNamespace, filePath, s.Pos())
				}
				program.Namespaces[currentNamespace] = currentNamespaceDefinition
				currentNamespaceDefinition = ""
				state = stateNamespaceLang
				continue
			}
			currentNamespaceDefinition += tokenText
			continue
		case stateStartScope:
			if tokenText == namespace {
				state = stateNamespaceLang
				continue
			}
			if tokenText != scope {
				goto parseErr
			}
			currentScope = &Scope{Operations: []*Operation{}}
			state = stateScopeName
			continue
		case stateScopeName:
			name := identifier.FindString(tokenText)
			if name == "" {
				goto parseErr
			}
			if _, ok := scopesMap[name]; ok {
				return nil, fmt.Errorf("Duplicate definition for scope '%s': %s:%s",
					name, filePath, s.Pos())
			}
			currentScope.Name = strings.Title(name)
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
			nsPrefix, err := newScopePrefix(prefixStr)
			if err != nil {
				return nil, err
			}
			currentScope.Prefix = nsPrefix
			state = stateOperationName
			continue
		case stateOperationName:
			if tokenText == prefix {
				if currentScope.Prefix != nil {
					return nil, fmt.Errorf("Duplicate prefix definition for scope '%s': %s:%s",
						currentScope.Name, filePath, s.Pos())
				}
				state = statePrefix
				continue
			}
			if tokenText == closeDefinition {
				scopesMap[currentScope.Name] = currentScope
				state = stateStartScope
				continue
			}
			name := identifier.FindString(tokenText)
			if name == "" {
				goto parseErr
			}
			name = strings.Title(name)
			if currentScope.containsOperation(name) {
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
			currentScope.addOperation(currentOperation)
			state = stateOperationName
			continue
		default:
			goto parseErr
		}
	}

	if state == stateNamespaceDefinition {
		program.Namespaces[currentNamespace] = currentNamespaceDefinition
		state = stateStartScope
	}

	if state != stateStartScope {
		goto parseErr
	}

	for _, scope := range scopesMap {
		if scope.Prefix == nil {
			scope.Prefix = defaultPrefix
		}
		scopes = append(scopes, scope)
	}
	sort.Sort(ByName(scopes)) // For ordering determinism.
	program.Scopes = scopes

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

type ByName []*Scope

func (b ByName) Len() int {
	return len(b)
}

func (b ByName) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByName) Less(i, j int) bool {
	return strings.Compare(b[i].Name, b[j].Name) == -1
}
