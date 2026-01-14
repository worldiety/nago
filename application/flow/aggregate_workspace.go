// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"

	"go.wdy.de/nago/application/evs"
)

// Workspace is the aggregate root for all types and packages within a workspace.
type Workspace struct {
	ID          WorkspaceID
	Types       map[TypeID]*Type
	Packages    map[PackageID]*Package
	Name        string
	Description string
}

func (ws *Workspace) Identity() WorkspaceID {
	return ws.ID
}

func (ws *Workspace) CreatePackage(cmd CreatePackageCmd) (PackageCreated, error) {
	var zero PackageCreated
	if cmd.Package == "" {
		return zero, fmt.Errorf("package id cannot be empty")
	}

	if _, ok := ws.Packages[cmd.Package]; ok {
		return zero, fmt.Errorf("package %s already exists", cmd.Package)
	}

	if err := cmd.Path.Validate(); err != nil {
		return zero, err
	}

	if err := cmd.Name.Validate(); err != nil {
		return zero, err
	}

	return PackageCreated{
		Workspace:   ws.ID,
		Package:     cmd.Package,
		Path:        cmd.Path,
		Name:        cmd.Name,
		Description: cmd.Description,
	}, nil
}

func (ws *Workspace) ApplyEnvelope(evt evs.Envelope[WorkspaceEvent]) error {
	return ws.Apply(evt.Data)
}

func (ws *Workspace) Apply(evt WorkspaceEvent) error {
	switch evt := evt.(type) {
	case WorkspaceCreated:
		ws.ID = evt.Workspace
		ws.Types = map[TypeID]*Type{}
		ws.Packages = map[PackageID]*Package{}
		ws.Name = evt.Name
		ws.Description = evt.Description
	case PackageCreated:
		ws.Packages[evt.Package] = &Package{
			Package:     evt.Package,
			Path:        evt.Path,
			Name:        evt.Name,
			Description: evt.Description,
		}

	case TypeCreated:
		ws.Types[evt.Type] = &Type{
			ID:      evt.Type,
			Name:    evt.Name,
			BuildIn: evt.BuildIn,
		}

	default:
		return fmt.Errorf("unknown event type: %T", evt)
	}

	return nil
}
