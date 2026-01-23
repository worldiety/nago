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

type Types struct {
	types map[TypeID]Type
}

func NewTypes() *Types {
	return &Types{types: make(map[TypeID]Type)}
}

func (t *Types) Clone() *Types {
	return &Types{types: xmaps.Clone(t.types)}
}

func (t *Types) All() iter.Seq[Type] {
	return slices.Values(slices.SortedFunc(maps.Values(t.types), func(t Type, t2 Type) int {
		return strings.Compare(string(t.Name()), string(t2.Name()))
	}))
}

func (t *Types) ByName(name Ident) (Type, bool) {
	for _, t := range t.types {
		if t.Name() == name {
			return t, true
		}
	}

	return nil, false
}

func (t *Types) ByID(id TypeID) (Type, bool) {
	v, ok := t.types[id]
	return v, ok
}

func (t *Types) AddType(typ Type) bool {
	if _, ok := t.types[typ.Identity()]; ok {
		return false
	}

	t.types[typ.Identity()] = typ
	return true
}
