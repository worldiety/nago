package main

import (
	"strings"
)

// TSTypeDef the minimal part of our Typescript AST.
type TSTypeDef struct {
	Package    string
	Name       string
	StrConsts  []string
	TypeParams []*TSTypeDef
}

// String returns the typescript type definition (or declaration or whatever that is in their spec)
func (t *TSTypeDef) String() string {
	var sb strings.Builder

	if len(t.StrConsts) > 0 {
		// special case for string constant enums
		sb.WriteString("'")
		sb.WriteString(strings.Join(t.StrConsts, "' | '"))
		sb.WriteString("'")
		return sb.String()
	}

	if t.Name == "[]" && len(t.TypeParams) == 1 {
		// special case for arrays
		sb.WriteString(t.TypeParams[0].String())
		sb.WriteString("[]")
		return sb.String()
	}

	sb.WriteString(t.Name)

	if len(t.TypeParams) > 0 {
		// special case for generic declarations
		sb.WriteString("<")
		for i, p := range t.TypeParams {

			sb.WriteString(p.String()) // apply recursion for arbitrary nested generics
			if i < len(t.TypeParams)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteString(">")
	}

	return sb.String()
}

type NamedType interface {
	GetName() string
}
