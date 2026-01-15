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
	"sync"

	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/pkg/xmaps"
)

// Workspace is the aggregate root for all types and packages within a workspace.
type Workspace struct {
	id          WorkspaceID
	packages    map[PackageID]*Package
	name        string
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
		ws.name = evt.Name
		ws.description = evt.Description
	case PackageCreated:
		ws.packages[evt.Package] = &Package{
			pckage:      evt.Package,
			types:       map[TypeID]Type{},
			path:        evt.Path,
			name:        evt.Name,
			description: evt.Description,
		}

	case StringTypeCreated:
		pkg := ws.packages[evt.Package]
		if pkg == nil {
			return fmt.Errorf("package %s not found", evt.Package)
		}

		pkg.types[evt.ID] = &StringType{
			name:        evt.Name,
			id:          evt.ID,
			description: evt.Description,
		}
	default:
		return fmt.Errorf("unknown event type: %T", evt)
	}

	return nil
}

func (ws *Workspace) Packages() iter.Seq[*Package] {
	return func(yield func(*Package) bool) {
		ws.mutex.Lock()
		defer ws.mutex.Unlock()

		for _, id := range xmaps.SortedKeys(ws.packages) {
			if !yield(ws.packages[id]) {
				return
			}
		}
	}
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

func (ws *Workspace) Name() string {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	return ws.name
}

func (ws *Workspace) Description() string {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	return ws.description
}
