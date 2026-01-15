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

func createTree(ws *flow.Workspace) *treeview.Node[any] {
	return (&treeview.Node[any]{
		Children: []*treeview.Node[any]{
			{
				Label:      "Packages",
				Expandable: true,
				Children: slices.Collect(func(yield func(*treeview.Node[any]) bool) {
					for p := range ws.Packages() {
						yield(&treeview.Node[any]{
							Label: string(p.Name()),
							Data:  p,
						})
					}
				}),
			},
			{
				Label:      "Types",
				Expandable: true,
				Children: slices.Collect(func(yield func(*treeview.Node[any]) bool) {
					for p := range ws.Types() {
						yield(&treeview.Node[any]{
							Label: string(p.Name()),
							Data:  p,
						})
					}
				}),
			},
			{
				Label:      "Repositories",
				Expandable: true,
			},
			{
				Label:      "Forms",
				Expandable: true,
			},
			{
				Label:      "Validations",
				Expandable: true,
			},
			{
				Label:      "Flows",
				Expandable: true,
			},
		},
	}).Expand(true)
}
