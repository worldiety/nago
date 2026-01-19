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
	"maps"
	"slices"
	"strings"
	"sync"

	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/pkg/xslices"
)

// Workspace is the aggregate root for all types and packages within a workspace.
type Workspace struct {
	id          WorkspaceID
	packages    map[PackageID]*Package
	repos       map[RepositoryID]*Repository
	forms       map[FormID]*Form
	name        Ident
	description string
	mutex       sync.Mutex
	valid       bool
}

func (ws *Workspace) Identity() WorkspaceID {
	return ws.id
}

func (ws *Workspace) applyEnvelope(evt evs.Envelope[WorkspaceEvent]) error {
	return ws.apply(evt.Data)
}

func (ws *Workspace) apply(evt WorkspaceEvent) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if !ws.valid {
		return fmt.Errorf("workspace is not valid: %s", ws.id)
	}

	switch evt := evt.(type) {
	case WorkspaceCreated:
		ws.id = evt.Workspace
		ws.packages = map[PackageID]*Package{}
		ws.repos = map[RepositoryID]*Repository{}
		ws.forms = map[FormID]*Form{}
		ws.name = evt.Name
		ws.description = evt.Description
	case PackageCreated:
		ws.packages[evt.Package] = &Package{
			pckage:      evt.Package,
			types:       map[TypeID]Type{},
			path:        evt.Path,
			name:        evt.Name,
			description: evt.Description,
			mutex:       &ws.mutex,
		}

	case StringTypeCreated:
		pkg := ws.packages[evt.Package]
		if pkg == nil {
			return fmt.Errorf("package %s not found", evt.Package)
		}

		t := &StringType{
			parent:      pkg,
			name:        evt.Name,
			id:          evt.ID,
			description: evt.Description,
		}
		t.values.Store(&xslices.Slice[*StringEnumCase]{})

		pkg.types[evt.ID] = t
	case StructTypeCreated:
		pkg := ws.packages[evt.Package]
		if pkg == nil {
			return fmt.Errorf("package %s not found", evt.Package)
		}

		pkg.types[evt.ID] = &StructType{
			name:        evt.Name,
			id:          evt.ID,
			description: evt.Description,
			parent:      pkg,
		}
	case StringFieldAppended:
		st, ok := ws.structTypeByID(evt.Struct)
		if !ok {
			return fmt.Errorf("struct %s not found", evt.Struct)
		}

		st.fields = append(st.fields, &StringField{
			name:        evt.Name,
			description: evt.Description,
			id:          evt.ID,
		})
	case BoolFieldAppended:
		st, ok := ws.structTypeByID(evt.Struct)
		if !ok {
			return fmt.Errorf("struct %s not found", evt.Struct)
		}

		st.fields = append(st.fields, &BoolField{
			name:        evt.Name,
			description: evt.Description,
			id:          evt.ID,
		})

	case TypeFieldAppended:
		st, ok := ws.structTypeByID(evt.Struct)
		if !ok {
			return fmt.Errorf("struct %s not found", evt.Struct)
		}

		t, ok := ws.typeByID(evt.Type)
		if !ok {
			return fmt.Errorf("type %s not found", evt.Type)
		}

		st.fields = append(st.fields, &TypeField{
			name:        evt.Name,
			description: evt.Description,
			id:          evt.ID,
			fieldType:   t,
		})
	case RepositoryAssigned:
		st, ok := ws.structTypeByID(evt.Struct)
		if !ok {
			return fmt.Errorf("struct %s not found", evt.Struct)
		}

		repo := &Repository{
			parent:     ws,
			id:         evt.Repository,
			structType: st,
		}

		ws.repos[repo.id] = repo

	case PrimaryKeySelected:
		st, ok := ws.structTypeByID(evt.Struct)
		if !ok {
			return fmt.Errorf("struct %s not found", evt.Struct)
		}

		for _, f := range st.fields {
			if f.Identity() == evt.Field {
				f.SetPrimaryKey(true)
			} else {
				// keep it, it is defined like this
				f.SetPrimaryKey(false)
			}
		}

	case FormCreated:
		form := &Form{
			parent: ws,
			id:     evt.ID,
		}

		form.name.Store(&evt.Name)
		form.description.Store(&evt.Description)
		ws.forms[evt.ID] = form

	case StringEnumCaseAdded:
		st, ok := ws.stringTypeByID(evt.String)
		if !ok {
			return fmt.Errorf("string %s not found", evt.String)
		}

		eCase := &StringEnumCase{
			name: evt.Name,
		}
		eCase.value.Store(&evt.Value)
		eCase.description.Store(&evt.Description)

		slice := st.values.Load().Append(eCase)
		st.values.Store(&slice)

	default:
		return fmt.Errorf("unknown event type: %T", evt)
	}

	return nil
}

func (ws *Workspace) Forms() iter.Seq[*Form] {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	tmp := slices.SortedFunc(maps.Values(ws.forms), func(a, b *Form) int {
		return strings.Compare(string(a.id), string(b.id))
	})

	return slices.Values(tmp)
}

func (ws *Workspace) Repositories() iter.Seq[*Repository] {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	tmp := slices.SortedFunc(maps.Values(ws.repos), func(a, b *Repository) int {
		return strings.Compare(string(a.id), string(b.id))
	})

	return slices.Values(tmp)
}

func (ws *Workspace) RepositoryByID(id RepositoryID) (*Repository, bool) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	return ws.repositoryByID(id)
}

func (ws *Workspace) repositoryByID(id RepositoryID) (*Repository, bool) {
	v, ok := ws.repos[id]
	return v, ok
}

func (ws *Workspace) StructTypes() iter.Seq[*StructType] {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	var tmp []*StructType
	for _, p := range ws.packages {
		for _, t := range p.types {
			if st, ok := t.(*StructType); ok {
				tmp = append(tmp, st)
			}
		}
	}

	slices.SortFunc(tmp, func(a, b *StructType) int {
		return strings.Compare(string(a.name), string(b.Name()))
	})

	return slices.Values(tmp)
}

func (ws *Workspace) StructTypeByID(id TypeID) (*StructType, bool) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	return ws.structTypeByID(id)
}

func (ws *Workspace) structTypeByID(id TypeID) (*StructType, bool) {

	for _, p := range ws.packages {
		if s, ok := p.types[id]; ok {
			if s, ok := s.(*StructType); ok {
				return s, true
			}
		}
	}

	return nil, false
}

func (ws *Workspace) stringTypeByID(id TypeID) (*StringType, bool) {

	for _, p := range ws.packages {
		if s, ok := p.types[id]; ok {
			if s, ok := s.(*StringType); ok {
				return s, true
			}
		}
	}

	return nil, false
}

func (ws *Workspace) Packages() iter.Seq[*Package] {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	var tmp []*Package
	for _, p := range ws.packages {
		tmp = append(tmp, p)
	}

	slices.SortFunc(tmp, func(a, b *Package) int {
		return strings.Compare(string(a.Path()), string(b.path))
	})

	return slices.Values(tmp)
}

func (ws *Workspace) typeByID(id TypeID) (Type, bool) {
	for _, p := range ws.packages {
		for _, t := range p.types {
			if t.Identity() == id {
				return t, true
			}
		}
	}

	return nil, false
}

func (ws *Workspace) Types() iter.Seq[Type] {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	var tmp []Type
	for _, p := range ws.packages {
		for _, t := range p.types {
			tmp = append(tmp, t)
		}
	}

	slices.SortFunc(tmp, func(a, b Type) int {
		return strings.Compare(string(a.Name()), string(b.Name()))
	})

	return slices.Values(tmp)
}

func (ws *Workspace) PackageByPath(path ImportPath) (*Package, bool) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	return ws.packageByPath(path)
}

func (ws *Workspace) packageByPath(path ImportPath) (*Package, bool) {
	for _, p := range ws.packages {
		if p.path == path {
			return p, true
		}
	}

	return nil, false
}

func (ws *Workspace) Name() Ident {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	return ws.name
}

func (ws *Workspace) Description() string {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	return ws.description
}
