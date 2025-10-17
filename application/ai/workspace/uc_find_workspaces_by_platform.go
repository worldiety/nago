// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workspace

import (
	"iter"

	"go.wdy.de/nago/auth"
)

func NewFindWorkspacesByPlatform(repo Repository) FindWorkspacesByPlatform {
	return func(subject auth.Subject, platform Platform) iter.Seq2[Workspace, error] {
		return func(yield func(Workspace, error) bool) {
			for workspace, err := range repo.All() {
				if err != nil {
					if !yield(workspace, err) {
						return
					}

					continue
				}

				if workspace.Platform == platform {
					if !yield(workspace, nil) {
						return
					}
				}
			}
		}
	}
}
