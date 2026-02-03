// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"go.wdy.de/nago/pkg/xslices"
)

var _ Type = (*StructType)(nil)

type StructType struct {
	Parent      PackageID
	ID          TypeID
	Fields      *Fields
	name        Ident
	description string
}

func NewStructType(parent PackageID, id TypeID, name Ident) *StructType {
	return &StructType{
		Parent: parent,
		name:   name,
		ID:     id,
		Fields: NewFields(),
	}
}

func (t *StructType) Name() Ident {
	return t.name
}

func (t *StructType) SetName(ident Ident) {
	t.name = ident
}

func (t *StructType) Description() string {
	return t.description
}

func (t *StructType) SetDescription(s string) {
	t.description = s
}

func (t *StructType) Clone() Type {
	return &StructType{
		Parent:      t.Parent,
		ID:          t.ID,
		Fields:      t.Fields.Clone(),
		name:        t.name,
		description: t.description,
	}
}

func (t *StructType) Identity() TypeID {
	return t.ID
}

func (t *StructType) String() string {
	return fmt.Sprintf("%s", t.name)
}

func (t *StructType) DocumentStoreReady() bool {
	count := 0
	for f := range t.Fields.All() {
		if f, ok := f.(PKField); ok {
			if f.PrimaryKey() {
				count++
			}
		}

	}

	return count == 1
}

type Fields struct {
	fields []Field
}

func NewFields() *Fields {
	return &Fields{}
}

func (f *Fields) Add(field Field) {
	f.fields = append(f.fields, field)
}

func (f *Fields) ByName(name Ident) (Field, bool) {
	for _, field := range f.fields {
		if field.Name() == name {
			return field, true
		}
	}

	return nil, false
}

func (f *Fields) ByID(id FieldID) (Field, bool) {
	for _, field := range f.fields {
		if field.Identity() == id {
			return field, true
		}
	}

	return nil, false
}

func (f *Fields) Len() int {
	return len(f.fields)
}

// All returns the fields in stable alphabetical order.
func (f *Fields) All() iter.Seq[Field] {
	var tmp []Field
	for _, field := range f.fields {
		tmp = append(tmp, field)
	}

	slices.SortFunc(tmp, func(a, b Field) int {
		return strings.Compare(string(a.Name()), string(b.Name()))
	})

	return slices.Values(tmp)
}

func (f *Fields) NonPrimaryFields() iter.Seq[Field] {
	return f.Filter(func(field Field) bool {
		if field, ok := field.(PKField); ok && field.PrimaryKey() {
			return false
		}

		return true
	})
}

// Filter returns the fields in stable alphabetical order after applying the predicate.
func (f *Fields) Filter(predicate func(Field) bool) iter.Seq[Field] {
	var tmp []Field
	for _, field := range f.fields {
		if predicate(field) {
			tmp = append(tmp, field)
		}
	}

	slices.SortFunc(tmp, func(a, b Field) int {
		return strings.Compare(string(a.Name()), string(b.Name()))
	})

	return slices.Values(tmp)

}

func (f *Fields) PrimaryKeys() iter.Seq[Field] {
	return f.Filter(func(field Field) bool {
		if field, ok := field.(PKField); ok && field.PrimaryKey() {
			return true
		}

		return false
	})
}

func (f *Fields) Clone() *Fields {
	return &Fields{fields: xslices.Clone(f.fields)}
}

type FieldID string

type Field interface {
	Identity() FieldID
	Name() Ident
	JSONName() string
	SetName(Ident)
	Description() string
	SetDescription(string)
	Typename() string
	field()
	Clone() Field
}

var (
	_ Field   = (*StringField)(nil)
	_ PKField = (*StringField)(nil)
	_ Field   = (*TypeField)(nil)
	_ PKField = (*TypeField)(nil)
	_ Field   = (*BoolField)(nil)
)

type StringField struct {
	Parent      TypeID
	ID          FieldID
	description string
	primaryKey  bool
	name        Ident
}

func NewStringField(parent TypeID, id FieldID, name Ident) *StringField {
	return &StringField{
		Parent: parent,
		ID:     id,
		name:   name,
	}
}

func (f *StringField) SuitableAsPrimaryKey(ws *Workspace) bool {
	return true
}

func (f *StringField) PrimaryKey() bool {
	return f.primaryKey
}

func (f *StringField) SetPrimaryKey(b bool) {
	f.primaryKey = b
}

func (f *StringField) JSONName() string {
	return string(f.name)
}

func (f *StringField) Clone() Field {
	c := *f
	return &c
}

func (f *StringField) Identity() FieldID {
	return f.ID
}

func (f *StringField) Name() Ident {
	return f.name
}

func (f *StringField) SetName(ident Ident) {
	f.name = ident
}

func (f *StringField) Description() string {
	return f.description
}

func (f *StringField) SetDescription(s string) {
	f.description = s
}

func (f *StringField) field() {}

func (f *StringField) Typename() string {
	return "string"
}

type PKField interface {
	Field
	SuitableAsPrimaryKey(ws *Workspace) bool
	PrimaryKey() bool
	SetPrimaryKey(b bool)
}
type TypeField struct {
	Parent      TypeID
	ID          FieldID
	name        Ident
	description string
	Type        TypeID
	primaryKey  bool
}

func NewTypeField(parent TypeID, id FieldID, name Ident, fieldType TypeID) *TypeField {
	return &TypeField{
		Parent: parent,
		ID:     id,
		name:   name,
		Type:   fieldType,
	}
}

func (f *TypeField) Name() Ident {
	return f.name
}

func (f *TypeField) SetName(ident Ident) {
	f.name = ident
}

func (f *TypeField) SetDescription(s string) {
	f.description = s
}

func (f *TypeField) Clone() Field {
	c := *f
	return &c
}

func (f *TypeField) JSONName() string {
	return string(f.name)
}

func (f *TypeField) Description() string {
	return f.description
}

func (f *TypeField) SuitableAsPrimaryKey(ws *Workspace) bool {
	ref, ok := ws.Packages.TypeByID(f.Type)
	if !ok {
		return false
	}

	_, ok = ref.(*StringType)
	return ok
}

func (f *TypeField) StringType(ws *Workspace) bool {
	ref, ok := ws.Packages.TypeByID(f.Type)
	if !ok {
		return false
	}

	_, ok = ref.(*StringType)
	return ok
}

func (f *TypeField) String() string {
	return string(f.name)
}

func (f *TypeField) PrimaryKey() bool {
	return f.primaryKey
}

func (f *TypeField) SetPrimaryKey(b bool) {
	f.primaryKey = b
}

func (f *TypeField) Typename() string {
	return string(f.name)
}

func (f *TypeField) Identity() FieldID {
	return f.ID
}

func (f *TypeField) field() {}

type BoolField struct {
	Parent      TypeID
	ID          FieldID
	name        Ident
	description string
}

func NewBoolField(parent TypeID, id FieldID, name Ident) *BoolField {
	return &BoolField{
		Parent: parent,
		ID:     id,
		name:   name,
	}
}

func (f *BoolField) Name() Ident {
	return f.name
}

func (f *BoolField) SetName(ident Ident) {
	f.name = ident
}

func (f *BoolField) Description() string {
	return f.description
}

func (f *BoolField) SetDescription(s string) {
	f.description = s
}

func (f *BoolField) Clone() Field {
	c := *f
	return &c
}

func (f *BoolField) Identity() FieldID {
	return f.ID
}

func (f *BoolField) field() {}

func (f *BoolField) Typename() string {
	return "bool"
}

func (f *BoolField) String() string {
	return string(f.name)
}

func (f *BoolField) JSONName() string {
	return string(f.name)
}
