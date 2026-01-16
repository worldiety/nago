// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"sync/atomic"
)

type FieldID string

type Field interface {
	Identity() FieldID
	Name() Ident
	Description() string
	IsPrimaryKey() bool
	SetPrimaryKey(bool)
	field()
	Typename() string
}

var (
	_ Field = (*StringField)(nil)
	_ Field = (*TypeField)(nil)
)

type StringField struct {
	name        Ident
	description string
	id          FieldID
	primaryKey  atomic.Bool
}

func (f *StringField) SetPrimaryKey(b bool) {
	f.primaryKey.Store(b)
}

func (f *StringField) Identity() FieldID {
	return f.id
}

func (f *StringField) Name() Ident {
	return f.name
}

func (f *StringField) Description() string {
	return f.description
}

func (f *StringField) field() {}

func (f *StringField) IsPrimaryKey() bool {
	return f.primaryKey.Load()
}

func (f *StringField) Typename() string {
	return "string"
}

type TypeField struct {
	id          FieldID
	name        Ident
	description string
	fieldType   Type
	primaryKey  atomic.Bool
}

func (f *TypeField) SetPrimaryKey(b bool) {
	f.primaryKey.Store(b)
}

func (f *TypeField) IsPrimaryKey() bool {
	return f.primaryKey.Load()
}

func (f *TypeField) Typename() string {
	return fmt.Sprintf("%s.%s", f.fieldType.Package().Name(), f.fieldType.Name())
}

func (f *TypeField) Identity() FieldID {
	return f.id
}

func (f *TypeField) Name() Ident {
	return f.name
}

func (f *TypeField) Description() string {
	return f.description
}

func (f *TypeField) field() {}

func (f *TypeField) Type() Type {
	return f.fieldType
}

type BoolField struct {
	name        Ident
	description string
	id          FieldID
}

func (f *BoolField) SetPrimaryKey(b bool) {
	panic("bool fields cannot be primary keys")
}

func (f *BoolField) Identity() FieldID {
	return f.id
}

func (f *BoolField) Name() Ident {
	return f.name
}

func (f *BoolField) Description() string {
	return f.description
}

func (f *BoolField) field() {}

func (f *BoolField) IsPrimaryKey() bool {
	return false // can never be a primary key
}

func (f *BoolField) Typename() string {
	return "bool"
}
