package main

import (
	"fmt"

	"github.com/Workiva/frugal/example/go/gen/linking"
)

type LinkingPubSubHandler struct {
}

func NewLinkingPubSubHandler() *LinkingPubSubHandler {
	return &LinkingPubSubHandler{}
}

func (l *LinkingPubSubHandler) UpdateAtoms(a *linking.AtomUpdateRequest) error {
	fmt.Println("received", a)
	return nil
}

func (l *LinkingPubSubHandler) GetCurrentAtoms(a *linking.GetCurrentAtomsRequest) error {
	fmt.Println("received", a)
	return nil
}
