// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"slices"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/ui/treeview"
)

func createTree(ws *flow.Workspace) *treeview.Node[any, string] {
	return &treeview.Node[any, string]{
		Children: []*treeview.Node[any, string]{
			{
				Label:      "Packages",
				Expandable: true,
				ID:         "packages",
				Children: slices.Collect(func(yield func(*treeview.Node[any, string]) bool) {
					for p := range ws.Packages.All() {
						yield(&treeview.Node[any, string]{
							Label: string(p.Name),
							ID:    string(p.Identity()),
							Data:  p,
							Children: slices.Collect(func(yield func(*treeview.Node[any, string]) bool) {
								for p := range p.Types.All() {
									yield(&treeview.Node[any, string]{
										Label: string(p.Name()),
										ID:    string(p.Identity()),
										Data:  p,
									})
								}
							}),
						})
					}
				}),
			},
			{
				Label:      "Types",
				Expandable: true,
				ID:         "types",
				Children: slices.Collect(func(yield func(*treeview.Node[any, string]) bool) {
					for p := range ws.Packages.Types() {
						yield(&treeview.Node[any, string]{
							Label: string(p.Name()),
							ID:    string(p.Identity()),
							Data:  p,
						})
					}
				}),
			},
			{
				Label:      "Repositories",
				ID:         "repositories",
				Expandable: true,
				Children: slices.Collect(func(yield func(*treeview.Node[any, string]) bool) {
					for p := range ws.Repositories.All() {
						yield(&treeview.Node[any, string]{
							ID:    string(p.Identity()),
							Label: string(p.Identity()),
							Data:  p,
						})
					}
				}),
			},
			{
				Label:      "Forms",
				Expandable: true,
				ID:         "forms",
				Children: slices.Collect(func(yield func(*treeview.Node[any, string]) bool) {
					for p := range ws.Forms.All() {
						yield(&treeview.Node[any, string]{
							ID:    string(p.Identity()),
							Label: string(p.Name()),
							Data:  p,
						})
					}
				}),
			},
			{
				Label:      "Validations",
				ID:         "validations",
				Expandable: true,
			},
			{
				Label:      "Flows",
				ID:         "flows",
				Expandable: true,
			},
		},
	}
}
