// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"go/token"
	"iter"
	"maps"
	"slices"
	"strings"
	"sync"
	"unicode"
)

type PackageID string

type ImportPath string

func (p ImportPath) Validate() error {
	s := string(p)
	if s == "" {
		return fmt.Errorf("package path cannot be empty")
	}

	segments := strings.Split(s, "/")
	for i, segment := range segments {
		if segment == "" {
			return fmt.Errorf("empty segment at position %d", i)
		}
		if err := validateIdentifier(segment); err != nil {
			return fmt.Errorf("invalid segment %q at position %d: %w", segment, i, err)
		}
	}

	return nil
}

func validateIdentifier(s string) error {
	if s == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return fmt.Errorf("must start with letter or underscore")
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return fmt.Errorf("invalid character %q", r)
			}
		}
	}

	if token.Lookup(s).IsKeyword() {
		return fmt.Errorf("%q is a reserved keyword", s)
	}

	return nil
}

type PackageCreated struct {
	Workspace   WorkspaceID `json:"workspace,omitempty"`
	Package     PackageID   `json:"package,omitempty"`
	Path        ImportPath  `json:"path,omitempty"`
	Name        Ident       `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
}

func (e PackageCreated) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e PackageCreated) event() {}

type CreatePackageCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Path        ImportPath
	Name        Ident
	Description string `lines:"3"`
}

func (c CreatePackageCmd) WorkspaceID() WorkspaceID {
	return c.Workspace
}

func (c CreatePackageCmd) WithWorkspaceID(id WorkspaceID) CreatePackageCmd {
	c.Workspace = id
	return c
}

type Package struct {
	types       map[TypeID]Type
	pckage      PackageID
	path        ImportPath
	name        Ident
	description string
	mutex       *sync.Mutex
}

func (p *Package) String() string {
	return string(p.name) + " [" + string(p.path) + "]"
}

func (p *Package) Identity() PackageID {
	return p.pckage
}

func (p *Package) Path() ImportPath {
	return p.path
}

func (p *Package) Name() Ident {
	return p.name
}

func (p *Package) Description() string {
	return p.description
}

func (p *Package) TypeByName(name Ident) (Type, bool) {
	for _, t := range p.types {
		if t.Name() == name {
			return t, true
		}
	}

	return nil, false
}

func (p *Package) Types() iter.Seq[Type] {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return slices.Values(slices.SortedFunc(maps.Values(p.types), func(t Type, t2 Type) int {
		return strings.Compare(string(t.Name()), string(t2.Name()))
	}))
}
