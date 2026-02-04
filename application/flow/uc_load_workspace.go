// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"errors"
	"os"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

func NewLoadWorkspace(handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID]) LoadWorkspace {
	return func(subject auth.Subject, id WorkspaceID) (option.Opt[*Workspace], error) {
		ws, err := handler.Aggregate(subject.Context(), id)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return option.None[*Workspace](), nil
			}

			return option.None[*Workspace](), err
		}

		if !(subject.HasPermission(PermFindWorkspaces) || ws.IsOwner(subject.ID())) {
			return option.None[*Workspace](), nil
		}

		return option.Some(ws), nil
	}
}
