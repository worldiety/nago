package main

import (
	"reflect"
)

type Import struct {
	Type string
	From string
}

type Field struct {
	Comment string
	Name    string
	Type    string
}

type Interface struct {
	Comment string
	Imports []Import
	Name    string
	Fields  []Field
}

func NewInterface(t any) *Interface {
	res := &Interface{}
	rType := reflect.TypeOf(t)
	rootTypeDef := resolveType(rType)
	res.Name = rootTypeDef.Name
	for i := range rType.NumField() {
		field := rType.Field(i)
		if field.Name == "_" {
			// we cannot use Tag if we span multiple lines with a fat comment.
			// An AST would be better and less redundant.
			res.Comment = string(field.Tag) //.Get("description")
			continue
		}

		if !field.IsExported() {
			continue
		}

		name := field.Tag.Get("json")
		value := field.Tag.Get("value")
		desc := field.Tag.Get("description")
		typeDef := &TSTypeDef{}
		if value != "" {
			typeDef.StrConsts = append(typeDef.StrConsts, value)
		} else {
			typeDef = resolveType(field.Type)
		}

		res.Import(typeDef)

		res.Fields = append(res.Fields, Field{
			Comment: desc,
			Name:    name,
			Type:    typeDef.String(),
		})
	}

	return res
}

func (i *Interface) GetName() string {
	return i.Name
}

func (i *Interface) Import(t *TSTypeDef) {
	importInto(t, &i.Imports)
}
