package main

import (
	"reflect"
)

type Union struct {
	Comment string
	Imports []Import
	Name    string
	Types   []string
}

func NewUnion(name string, types []reflect.Type) *Union {
	res := &Union{}
	res.Name = name
	for _, typ := range types {
		typeDef := resolveType(typ)
		res.Import(typeDef)
		res.Types = append(res.Types, typeDef.String())
	}

	return res
}

func (i *Union) GetName() string {
	return i.Name
}

func (i *Union) Import(t *TSTypeDef) {
	importInto(t, &i.Imports)
}
