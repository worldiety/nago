// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"context"
	"iter"

	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

func NewFindWorkspace(handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID]) FindWorkspaces {
	return func(subject auth.Subject) iter.Seq2[WorkspaceID, error] {

		return func(yield func(WorkspaceID, error) bool) {
			it, err := handler.All(subject.Context())
			if err != nil {
				yield("", err)
				return
			}

			for id := range it {
				ws, err := handler.Aggregate(context.Background(), id)
				if err != nil {
					if !yield("", err) {
						return
					}

					continue
				}

				if subject.HasPermission(PermFindWorkspaces) || ws.IsOwner(subject.ID()) {
					if !yield(id, nil) {
						return
					}
				}
			}
		}
	}
}
