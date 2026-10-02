// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/hero/outline"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/treeview"
)

func init() {
	app.Register("tree-view", func(wnd core.Window) core.View {
		type file = treeview.Node[string, string]

		root := &file{ID: "root", Label: "project", Icon: icons.Folder, Expandable: true, Children: []*file{
			{ID: "cmd", Label: "cmd", Icon: icons.Folder, Expandable: true, Children: []*file{
				{ID: "main.go", Label: "main.go", Icon: icons.DocumentText},
			}},
			{ID: "internal", Label: "internal", Icon: icons.Folder, Expandable: true, Children: []*file{
				{ID: "orders.go", Label: "orders.go", Icon: icons.DocumentText},
				{ID: "customers.go", Label: "customers.go", Icon: icons.DocumentText},
			}},
			{ID: "go.mod", Label: "go.mod", Icon: icons.DocumentText},
		}}

		state := core.AutoState[treeview.TreeStateModel[string]](wnd).Init(func() treeview.TreeStateModel[string] {
			return treeview.TreeStateModel[string]{
				Expanded: map[string]bool{"root": true, "internal": true},
				Selected: map[string]bool{"orders.go": true},
			}
		})

		return treeview.TreeView(root, state).
			Action(func(n *file) {
				// open the file n.ID
			}).
			Frame(Frame{Width: L320})
	})
}
