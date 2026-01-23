// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"iter"
	"slices"

	"go.wdy.de/nago/pkg/xslices"
)

var _ Type = (*StringType)(nil)

type StringType struct {
	Parent      PackageID
	ID          TypeID
	Enumeration *Literals
	name        Ident
	description string
}

func NewStringType(parent PackageID, id TypeID, name Ident) *StringType {
	return &StringType{
		Parent:      parent,
		ID:          id,
		name:        name,
		Enumeration: NewLiterals(),
	}
}

func (t *StringType) SetName(ident Ident) {
	t.name = ident
}

func (t *StringType) SetDescription(s string) {
	t.description = s
}

func (t *StringType) Name() Ident {
	return t.name
}

func (t *StringType) Description() string {
	return t.description
}

func (t *StringType) Clone() Type {
	return &StringType{
		Parent:      t.Parent,
		ID:          t.ID,
		Enumeration: t.Enumeration.Clone(),
		name:        t.name,
		description: t.description,
	}
}

func (t *StringType) IsEnum() bool {
	return t.Enumeration.Len() != 0
}

func (t *StringType) Identity() TypeID {
	return t.ID
}

func (t *StringType) String() string {
	return string(t.name)
}

type Literals struct {
	values []*Literal
}

func NewLiterals() *Literals {
	return &Literals{}
}

func (l *Literals) Add(literal *Literal) {
	l.values = append(l.values, literal)
}

func (l *Literals) All() iter.Seq[*Literal] {
	return slices.Values(l.values)
}

func (l *Literals) ByName(name Ident) (*Literal, bool) {
	for _, literal := range l.values {
		if literal.Name == name {
			return literal, true
		}
	}

	return nil, false
}

func (l *Literals) ByValue(value string) (*Literal, bool) {
	for _, literal := range l.values {
		if literal.Value == value {
			return literal, true
		}
	}

	return nil, false
}

func (l *Literals) Len() int {
	return len(l.values)
}

func (l *Literals) Clone() *Literals {
	return &Literals{
		values: xslices.Clone(l.values),
	}
}

type Literal struct {
	Name        Ident
	Value       string
	Description string
}

func NewLiteral(name Ident, value string) *Literal {
	return &Literal{
		Name:  name,
		Value: value,
	}
}

func (l *Literal) Clone() *Literal {
	tmp := *l
	return &tmp
}
