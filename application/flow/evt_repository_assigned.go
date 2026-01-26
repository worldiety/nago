// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"context"
	"fmt"
	"strings"

	"go.wdy.de/nago/application/evs"
)

type RepositoryID string

func (id RepositoryID) Validate() error {
	s := string(id)
	if s == "" {
		return fmt.Errorf("repository id cannot be empty")
	}

	if strings.HasPrefix(s, "nago.") {
		// security note: this is very important to keep: without this check, an attacker with only flow privileges
		// can escalate privileges by assigning repositories with names of the system (like user, roles, groups)
		// and mutate them in arbitrary ways.
		return fmt.Errorf("repository id cannot start with 'nago.'")
	}

	first := rune(s[0])
	if first >= '0' && first <= '9' {
		return fmt.Errorf("repository id cannot start with a digit")
	}

	for i, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '.' && r != '_' {
			return fmt.Errorf("repository id contains invalid character '%c' at position %d", r, i)
		}
	}

	return nil
}

type RepositoryAssigned struct {
	Workspace   WorkspaceID  `json:"workspace"`
	Struct      TypeID       `json:"struct"`
	Repository  RepositoryID `json:"repository"`
	Description string       `json:"description"`
}

func (evt RepositoryAssigned) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt RepositoryAssigned) Discriminator() evs.Discriminator {
	return "RepositoryAssigned"
}

func (evt RepositoryAssigned) Evolve(ctx context.Context, ws *Workspace) error {
	st, ok := ws.Packages.TypeByID(evt.Struct)
	if !ok {
		return fmt.Errorf("struct %s not found", evt.Struct)
	}

	if _, ok := ws.Repositories.ByID(evt.Repository); ok {
		return fmt.Errorf("repository %s already exists", evt.Repository)
	}

	r := NewRepository(evt.Repository, st.Identity())
	r.Description = evt.Description
	ws.Repositories.Add(r)

	return nil
}
