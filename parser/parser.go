package parser

import (
	"fmt"
	"os"
	"regexp"
	"text/scanner"
)

var identifier = regexp.MustCompile("^[A-Za-z]+[A-Za-z0-9]")

const (
	namespace          = "namespace"
	openDefinition     = "{"
	closeDefinition    = "}"
	operationDelimiter = ":"
)

const (
	stateStartNamespace = iota
	stateNamespaceName
	stateOpenDefinition
	stateOperationName
	stateOperationDelimiter
	stateOperationParam
)

type Operation struct {
	Name  string
	Param string
}

type Namespace struct {
	Name       string
	Operations []*Operation
}

func (n *Namespace) addOperation(op *Operation) {
	n.Operations = append(n.Operations, op)
}

func (n *Namespace) containsOperation(name string) bool {
	for _, op := range n.Operations {
		if op.Name == name {
			return true
		}
	}
	return false
}

func Parse(filePath string) ([]*Namespace, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var s scanner.Scanner
	s.Init(file)

	var (
		namespacesMap    = make(map[string]*Namespace)
		namespaces       = []*Namespace{}
		state            = stateStartNamespace
		currentNamespace *Namespace
		currentOperation *Operation
	)

	// TODO: Add support for comments.

	var token rune
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
			currentNamespace.Name = name
			state = stateOpenDefinition
			continue
		case stateOpenDefinition:
			if tokenText != openDefinition {
				goto parseErr
			}
			state = stateOperationName
			continue
		case stateOperationName:
			tokenText := tokenText
			if tokenText == closeDefinition {
				namespacesMap[currentNamespace.Name] = currentNamespace
				state = stateStartNamespace
				continue
			}
			name := identifier.FindString(tokenText)
			if name == "" {
				goto parseErr
			}
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
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil

parseErr:
	return nil, fmt.Errorf("Invalid syntax: %s:%s", filePath, s.Pos())
}
