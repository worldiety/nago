// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/application/role"
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
	ID() TypeID
	Description() string
}

type StringType struct {
	parent      *Package
	name        Ident
	id          TypeID
	description string
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

func (t *StringType) ID() TypeID {
	return t.id
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

type Field struct {
	Name       Ident    `json:"name"`
	Type       Typename `json:"type"`
	Required   bool     `json:"required"`
	PrimaryKey bool     `json:"primaryKey"`
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

// TODO how to represent flows and roles and ownerships?
// TODO how to docusign?
type Repository struct {
	Driver Driver
	Source string // connection string, url or local nago database name

	// TODO
	Visibility *Visibility
	Approval   *Approval
}
