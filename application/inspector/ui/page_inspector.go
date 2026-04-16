// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiinspector

import (
	"net/url"

	"go.wdy.de/nago/application/inspector"
	"go.wdy.de/nago/application/inspector/rest"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/treeview"
)

func PageInspector(wnd core.Window, uc inspector.UseCases) core.View {
	stores, err := uc.FindAll(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}

	tree := createTree(stores)
	treeState := core.AutoState[treeview.TreeStateModel[string]](wnd).Init(func() treeview.TreeStateModel[string] {
		m := treeview.TreeStateModel[string]{
			Expanded: make(map[string]bool),
			Selected: make(map[string]bool),
		}
		for _, child := range tree.Children {
			m.Expanded[child.ID] = true
		}
		return m
	})

	selectedStore := core.AutoState[inspector.Store](wnd)

	q := url.Values{}
	var downloadAllPath string
	if selectedStore.Get().Store != nil {
		q.Set("store", selectedStore.Get().Name)
		store := selectedStore.Get()
		q.Set("store", store.Name)
		q.Set("id", rest.EncodeQuery([]string{"*"}))
		downloadAllPath = rest.PathDownloadAsZip + "?" + q.Encode()
	}

	return ui.VStack(
		ui.HStack(
			ui.IfElse(selectedStore.Get().Name != "",
				ui.Text(selectedStore.Get().Name),
				ui.Text("Inspector"),
			),

			ui.Spacer(),
			ui.TertiaryButton(nil).PreIcon(icons.Download).HRef(core.URI(downloadAllPath)).AccessibilityLabel(StrDownloadAllEntries.Get(wnd)).Enabled(selectedStore.Get().Store != nil),
		).FullWidth(),
		ui.HStack(
			ui.ScrollView(
				treeview.TreeView(tree, treeState).Action(func(n *treeview.Node[*inspector.Store, string]) {
					var store inspector.Store
					if n.Data != nil {
						store = *n.Data
					}

					selectedStore.Set(store)
				}),
			).Axis(ui.ScrollViewAxisHorizontal).Frame(ui.Frame{Width: ui.L320, MaxWidth: ui.L320}),
			ui.VLine().Frame(ui.Frame{}),
			ui.VStack(
				viewKeys(wnd, selectedStore.Get()),
			).Alignment(ui.Top).
				BackgroundColor(ui.ColorBackground).FullWidth(),
		).FullWidth().Alignment(ui.Stretch),
	).FullWidth().Alignment(ui.Leading)
}
