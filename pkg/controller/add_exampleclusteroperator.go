package controller

import (
	"github.com/djzager/example-cluster-operator/pkg/controller/exampleclusteroperator"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, exampleclusteroperator.Add)
}
