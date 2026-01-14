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
	"go.wdy.de/nago/pkg/xslices"
)

func NewFindWorkspace(repoName string, wsIndex *evs.StoreIndex[WorkspaceID, WorkspaceEvent]) FindWorkspaces {
	return func(subject auth.Subject) iter.Seq2[WorkspaceID, error] {
		it, err := wsIndex.GroupPrimary(context.Background())
		if err != nil {
			return xslices.ValuesWithError([]WorkspaceID{}, err)
		}

		return func(yield func(WorkspaceID, error) bool) {
			for id := range it {
				if subject.HasResourcePermission(repoName, string(id), PermFindWorkspaces) {
					if !yield(id, nil) {
						return
					}
				}
			}
		}
	}
}
