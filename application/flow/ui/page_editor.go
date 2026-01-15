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

	tree := core.AutoState[*treeview.Node[any]](wnd).Init(func() *treeview.Node[any] {
		return createTree(ws)
	})

	presentCreatePackage := core.AutoState[bool](wnd).Observe(func(newValue bool) {
		tree.Reset()
	})

	presentCreateStringType := core.AutoState[bool](wnd).Observe(func(newValue bool) {
		tree.Reset()
	})

	presentCreateStructType := core.AutoState[bool](wnd).Observe(func(newValue bool) {
		tree.Reset()
	})

	selected := core.AutoState[any](wnd).Observe(func(newValue any) {

	})

	return ui.VStack(

		dialogCmd(wnd, ws, "Create new package", presentCreatePackage, uc.CreatePackage),
		dialogCmd(wnd, ws, "Create new string type", presentCreateStringType, uc.CreateStringType),
		dialogCmd(wnd, ws, "Create new struct type", presentCreateStructType, uc.CreateStructType),

		ui.HStack(
			ui.Text(ws.Name()),
			ui.Spacer(),

			ui.TertiaryButton(func() {
				presentCreatePackage.Set(true)
			}).PreIcon(icons.Folder).AccessibilityLabel(StrActionCreatePackage.Get(wnd)),

			ui.Menu(
				ui.TertiaryButton(nil).PreIcon(icons.TableRow).AccessibilityLabel(StrActionCreateType.Get(wnd)),
				ui.MenuGroup(
					ui.MenuItem(func() {
						presentCreateStringType.Set(true)
					}, ui.Text("Create String Type")),

					ui.MenuItem(func() {
						presentCreateStructType.Set(true)
					}, ui.Text("Create Struct Type")),
				),
			),

			ui.TertiaryButton(func() {

			}).PreIcon(icons.Database).AccessibilityLabel(StrActionCreateRepository.Get(wnd)),
		).FullWidth().Border(ui.Border{BottomColor: ui.ColorIconsMuted, BottomWidth: ui.L1}).Padding(ui.Padding{}.Vertical(ui.L2)),

		ui.HStack(
			ui.ScrollView(
				treeview.TreeView(tree.Get()).Action(func(n *treeview.Node[any]) {
					if n.Expandable {
						n.Expanded = !n.Expanded
					} else {
						tree.Get().Select(false)
						n.Selected = true
					}

					selected.Set(n.Data)
					selected.Invalidate()

					tree.Invalidate()
				})).Axis(ui.ScrollViewAxisBoth).Frame(ui.Frame{Width: ui.L200, MaxWidth: ui.L200}),
			ui.VLine().Frame(ui.Frame{}),
			ui.VStack(
				renderSelected(wnd, uc, ws, selected.Get()),
			).Alignment(ui.Top).
				BackgroundColor(ui.ColorBackground).FullWidth(),
		).FullWidth().Alignment(ui.Stretch),
	).FullWidth().Alignment(ui.Leading)
}

func renderSelected(wnd core.Window, uc flow.UseCases, ws *flow.Workspace, selected any) core.View {
	if selected == nil {
		return ui.Text("Nothing selected")
	}

	switch t := selected.(type) {
	case *flow.StringType:
		return viewTypeString(wnd, uc, ws, t)
	case *flow.StructType:
		return viewTypeStruct(wnd, uc, ws, t)
	default:
		return ui.Text(fmt.Sprintf("%T", selected))
	}

}
