// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"slices"
	"strings"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xerrors"
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

func (e RepositoryAssigned) WorkspaceID() WorkspaceID {
	return e.Workspace
}

func (e RepositoryAssigned) event() {}

type AssignRepositoryCmd struct {
	Workspace   WorkspaceID `visible:"false"`
	Struct      TypeID      `source:"nago.flow.pkstructs"`
	Repository  RepositoryID
	Description string `lines:"3"`
}

func (c AssignRepositoryCmd) cmd() {}

type AssignRepository func(subject auth.Subject, cmd AssignRepositoryCmd) (RepositoryAssigned, error)

func NewAssignRepository(hnd handleCmd[RepositoryAssigned]) AssignRepository {
	return func(subject auth.Subject, cmd AssignRepositoryCmd) (RepositoryAssigned, error) {
		return hnd(subject, cmd.Workspace, func(ws *Workspace) (RepositoryAssigned, error) {
			var zero RepositoryAssigned
			if err := cmd.Repository.Validate(); err != nil {
				return zero, xerrors.WithFields("Validation", "Repository", err.Error())
			}

			s, ok := ws.structTypeByID(cmd.Struct)
			if !ok {
				return zero, xerrors.WithFields("Validation", "Struct", "Struct not found")
			}

			if v := slices.Collect(s.PrimaryKeyFields()); len(v) != 1 {
				return zero, xerrors.WithFields("Validation", "Struct", "Struct must have exactly one primary key field")
			}

			return RepositoryAssigned{
				Workspace:   cmd.Workspace,
				Struct:      cmd.Struct,
				Repository:  cmd.Repository,
				Description: cmd.Description,
			}, nil
		})
	}
}
