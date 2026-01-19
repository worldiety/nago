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
	"sync/atomic"

	"go.wdy.de/nago/application/role"
	"go.wdy.de/nago/pkg/xslices"
)

type Kind int

const (
	String Kind = iota + 1
	Float
	Int
	Bool
	Struct
	Slice
	ForeignKey // TODO not required because model by relation?
)

/*type Type struct {
	ID            TypeID         `json:"id"`
	Package       Package        `json:"package"`
	Name          Typename       `json:"name"`
	Kind          Kind           `json:"kind"`
	Fields        []Field        `json:"fields,omitempty"`     // optional, only valid for Kind Struct
	Repository    *Repository    `json:"repository,omitempty"` // optional, only valid for Kind Struct
	BuildIn       bool           `json:"buildIn"`              // if true, the type is available universally
	RenderOptions *RenderOptions `json:"renderOptions,omitempty"`
}*/

type Type interface {
	Package() *Package
	Name() Ident
	Identity() TypeID
	Description() string
}

type StringEnumCase struct {
	name        Ident
	value       atomic.Pointer[string]
	description atomic.Pointer[string]
}

func (e *StringEnumCase) Name() Ident {
	return e.name
}

func (e *StringEnumCase) Value() string {
	return *e.value.Load()
}

func (e *StringEnumCase) Description() string {
	return *e.description.Load()
}

type StringType struct {
	parent      *Package
	name        Ident
	id          TypeID
	description string
	values      atomic.Pointer[xslices.Slice[*StringEnumCase]]
}

func (t *StringType) Values() iter.Seq[*StringEnumCase] {
	return t.values.Load().All()
}

func (t *StringType) Description() string {
	return t.description
}

func (t *StringType) Package() *Package {
	return t.parent
}

func (t *StringType) Name() Ident {
	return t.name
}

func (t *StringType) Identity() TypeID {
	return t.id
}

func (t *StringType) String() string {
	return fmt.Sprintf("%s.%s", t.Package().Name(), t.Name())
}

type StructType struct {
	parent      *Package
	name        Ident
	id          TypeID
	description string
	fields      []Field
}

func (t *StructType) Description() string {
	return t.description
}

func (t *StructType) Package() *Package {
	return t.parent
}

func (t *StructType) Name() Ident {
	return t.name
}

func (t *StructType) Identity() TypeID {
	return t.id
}

func (t *StructType) String() string {
	return fmt.Sprintf("%s.%s", t.Package().Name(), t.Name())
}

func (t *StructType) PrimaryKeyFields() iter.Seq[Field] {
	t.parent.mutex.Lock()
	defer t.parent.mutex.Unlock()

	return t.primaryKeyFields()
}

func (t *StructType) DocumentStoreReady() bool {
	t.parent.mutex.Lock()
	defer t.parent.mutex.Unlock()

	count := 0
	for _, f := range t.fields {
		if f.IsPrimaryKey() {
			count++
		}
	}

	return count == 1
}

func (t *StructType) primaryKeyFields() iter.Seq[Field] {
	var tmp []Field
	for _, field := range t.fields {
		if field.IsPrimaryKey() {
			tmp = append(tmp, field)
		}
	}

	slices.SortFunc(tmp, func(a, b Field) int {
		return strings.Compare(string(a.Name()), string(b.Name()))
	})

	return slices.Values(tmp)
}

func (t *StructType) Fields() iter.Seq[Field] {
	t.parent.mutex.Lock()
	defer t.parent.mutex.Unlock()

	return t.sortedFields()
}

func (t *StructType) sortedFields() iter.Seq[Field] {
	var tmp []Field
	for _, field := range t.fields {
		tmp = append(tmp, field)
	}

	slices.SortFunc(tmp, func(a, b Field) int {
		return strings.Compare(string(a.Name()), string(b.Name()))
	})

	return slices.Values(tmp)
}

func (t *StructType) NonPrimaryFields() iter.Seq[Field] {
	t.parent.mutex.Lock()
	defer t.parent.mutex.Unlock()

	return t.nonPrimaryFields()
}

func (t *StructType) nonPrimaryFields() iter.Seq[Field] {
	var tmp []Field
	for _, field := range t.fields {
		if !field.IsPrimaryKey() {
			tmp = append(tmp, field)
		}
	}

	slices.SortFunc(tmp, func(a, b Field) int {
		return strings.Compare(string(a.Name()), string(b.Name()))
	})

	return slices.Values(tmp)

}

type RenderOptions struct {
	Hidden         bool   `json:"hidden,omitempty"`
	Label          string `json:"label"`
	SupportingText string `json:"supportingText"`

	Card      *RenderOptionsCard      `json:"card,omitempty"`
	HStack    *RenderOptionsHStack    `json:"hstack,omitempty"`
	TextField *RenderOptionsTextField `json:"textField,omitempty"`
}

type RenderOptionsTextField struct {
	Lines int `json:"lines"`
}

type RenderOptionsCard struct {
	Label string `json:"label"`
}

type RenderOptionsHStack struct {
	Label   string    `json:"label"`
	Weights []float64 `json:"weights"`
}

type Driver string

const (
	NagoDriver Driver = "nagodb"
)

type Visibility struct {
	// self is always the case
	OneOf  []role.ID // TODO why a different bool expression for slice?
	Public bool      // ? extra or just use the nago anon-role feature?
}

type Approval struct {
	AllOf []role.ID // e.g. a C*O and a project manager
	// TODO what about tags which may be assigned by a (AI?) process without a role?
	// Specification for allowed transitions by role e.g.
	// first to legal person, then to ux, then to dev, then to pm and finally CEO.
	Transitions map[role.ID]role.ID

	// Which roles are allowed to see it?
	Read []role.ID

	// Which roles are allowed to edit it?
	Write []role.ID

	// TODO What about groups, do we need to duplicate this?
}
