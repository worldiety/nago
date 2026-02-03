// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"iter"
	"maps"
	"slices"
	"strings"

	"go.wdy.de/nago/pkg/xmaps"
)

type Packages struct {
	packages map[PackageID]*Package
}

func NewPackages() *Packages {
	return &Packages{packages: make(map[PackageID]*Package)}
}

func (p *Packages) Clone() *Packages {
	return &Packages{packages: xmaps.Clone(p.packages)}
}

func (p *Packages) All() iter.Seq[*Package] {
	if len(p.packages) == 0 {
		return func(yield func(*Package) bool) {}
	}

	if len(p.packages) == 1 {
		return func(yield func(*Package) bool) {
			for _, p2 := range p.packages {
				yield(p2)
				return
			}
		}
	}

	return slices.Values(slices.SortedFunc(maps.Values(p.packages), func(e *Package, e2 *Package) int {
		return strings.Compare(string(e.ImportPath), string(e2.ImportPath))
	}))
}

func (p *Packages) Len() int {
	return len(p.packages)
}

func (p *Packages) First() (*Package, bool) {
	for p := range p.All() {
		return p, true
	}

	return nil, false
}

func (p *Packages) AddPackage(pkg *Package) {
	p.packages[pkg.ID] = pkg
}

func (p *Packages) ByID(id PackageID) (*Package, bool) {
	v, ok := p.packages[id]
	return v, ok
}

func (p *Packages) ByImportPath(path ImportPath) (*Package, bool) {
	for _, pkg := range p.packages {
		if pkg.ImportPath == path {
			return pkg, true
		}
	}

	return nil, false
}

func (p *Packages) StructTypeByID(id TypeID) (*StructType, bool) {
	for _, p := range p.packages {
		if s, ok := p.Types.ByID(id); ok {
			if s, ok := s.(*StructType); ok {
				return s, true
			}
		}
	}

	return nil, false
}

func (p *Packages) StructTypes() iter.Seq[*StructType] {
	return func(yield func(*StructType) bool) {
		for t := range p.Types() {
			if s, ok := t.(*StructType); ok {
				if !yield(s) {
					return
				}
			}
		}
	}
}

func (p *Packages) StringTypeByID(id TypeID) (*StringType, bool) {
	for _, p := range p.packages {
		if s, ok := p.Types.ByID(id); ok {
			if s, ok := s.(*StringType); ok {
				return s, true
			}
		}
	}

	return nil, false
}

func (p *Packages) TypeByID(id TypeID) (Type, bool) {
	for _, p := range p.packages {
		if t, ok := p.Types.ByID(id); ok {
			return t, true
		}
	}

	return nil, false
}

// Types returns all types in the workspace in a stable order.
func (p *Packages) Types() iter.Seq[Type] {
	return func(yield func(Type) bool) {
		for p := range p.All() {
			for t := range p.Types.All() {
				if !yield(t) {
					return
				}
			}
		}
	}
}
