// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"fmt"
	"os"

	"go.wdy.de/nago/application/flow"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/treeview"
)

func PageEditor(wnd core.Window, uc flow.UseCases) core.View {
	optWs, err := uc.LoadWorkspace(wnd.Subject(), flow.WorkspaceID(wnd.Values()["workspace"]))
	if err != nil {
		return alert.BannerError(err)
	}

	if optWs.IsNone() {
		return alert.BannerError(fmt.Errorf("workspace not found: %s: %w", wnd.Values()["workspace"], os.ErrNotExist))
	}

	ws := optWs.Unwrap()

	tree := core.AutoState[*treeview.Node[string]](wnd).Init(func() *treeview.Node[string] {
		return &treeview.Node[string]{
			Children: []*treeview.Node[string]{
				{
					Label: "hallo",
					Icon:  icons.File,
				},
				{
					Label: "hallo2",
					Children: []*treeview.Node[string]{
						{Label: "hallo3"},
						{Label: "hallo4"},
					},
				},
			},
		}
	})

	return ui.VStack(
		ui.H1(ws.Name),
		treeview.TreeView(tree.Get()).Action(func(n *treeview.Node[string]) {
			fmt.Println(n)
			n.Selected = !n.Selected
			if n.Expandable {
				n.Expanded = !n.Expanded
			}
			tree.Invalidate()
		}),
	).FullWidth().Alignment(ui.Leading)
}
