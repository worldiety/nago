// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiinspector

import (
	"fmt"
	"go.wdy.de/nago/application/inspector"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func viewKeys(wnd core.Window, uc inspector.UseCases, store *core.State[inspector.Store], entry *core.State[inspector.PageEntry]) ui.DecoredView {
	if store.Get().Name == "" {
		return ui.VStack(ui.Text("kein Store gewählt")).FullWidth()
	}

	activePage := core.AutoState[int](wnd)
	selectedKeys := core.AutoState[map[string]struct{}](wnd).Init(func() map[string]struct{} {
		return map[string]struct{}{}
	})

	store.Observe(func(newValue inspector.Store) {
		activePage.Set(0)
		selectedKeys.Set(map[string]struct{}{})
	})

	page, err := uc.Filter(wnd.Subject(), store.Get().Store, inspector.FilterOptions{
		DetectMimeTypes: true,
		OnlyKeys:        true,
		PageNo:          activePage.Get(),
		PageSize:        20,
	})

	if err != nil {
		return ui.VStack(alert.BannerError(err))
	}

	newEntryPresented := core.AutoState[bool](wnd)
	newEntryString := core.AutoState[string](wnd)
	deleteEntryPresented := core.AutoState[bool](wnd)

	var menuItems []ui.TMenuItem

	menuItems = append(menuItems, ui.MenuItem(func() {
		newEntryPresented.Set(true)
	}, ui.Text("neu")))

	if len(selectedKeys.Get()) > 0 {
		menuItems = append(menuItems, ui.MenuItem(func() {
			deleteEntryPresented.Set(true)
		}, ui.Text(fmt.Sprintf("%d Einträge löschen", len(selectedKeys.Get())))))
	}

	return ui.VStack(
		ui.IfFunc(newEntryPresented.Get(), func() core.View {
			return alert.Dialog(
				"Neuen Eintrag erstellen",
				ui.TextField("Key", newEntryString.Get()).InputValue(newEntryString).SupportingText(fmt.Sprintf("Der Key wird im Store %s erstellt. Er darf nicht leer sein.", store.Get().Name)),
				newEntryPresented,
				alert.Cancel(nil),
				alert.Save(func() (close bool) {
					if err := blob.Put(store.Get().Store, newEntryString.Get(), nil); err != nil {
						alert.ShowBannerError(wnd, err)
						return false
					}

					return true
				}),
			)
		}),

		ui.IfFunc(deleteEntryPresented.Get(), func() core.View {
			return alert.Dialog(
				"Einträge löschen",
				ui.Text(fmt.Sprintf("Sollen die %d Einträge aus %s gelöscht werden?", len(selectedKeys.Get()), store.Get().Name)),
				deleteEntryPresented,
				alert.Cancel(nil),
				alert.Delete(func() {
					for key := range selectedKeys.Get() {
						if err := blob.Delete(store.Get().Store, key); err != nil {
							alert.ShowBannerError(wnd, err)
						}
					}

				}),
			)
		}),

		ui.HStack(
			ui.Text(fmt.Sprintf("%d Einträge", page.Count)),
			ui.Menu(
				ui.TertiaryButton(nil).PreIcon(flowbiteOutline.DotsVertical),
				ui.MenuGroup(menuItems...),
			),
		).FullWidth().Alignment(ui.Trailing),
		ui.HLine().Padding(ui.Padding{Bottom: ui.L4}),
		ui.ScrollView(
			ui.VStack(
				ui.ForEach(page.Entries, func(t inspector.PageEntry) core.View {
					checked := core.StateOf[bool](wnd, "checked-"+t.Key).Init(func() bool {
						_, ok := selectedKeys.Get()[t.Key]
						return ok
					}).Observe(func(newValue bool) {
						if newValue {
							selectedKeys.Get()[t.Key] = struct{}{}
						} else {
							delete(selectedKeys.Get(), t.Key)
						}

					})

					return ui.HStack(
						ui.Checkbox(checked.Get()).InputChecked(checked),
						ui.PrimaryButton(func() {
							entry.Set(t)
							entry.Notify()
						}).Title(t.Key).Preset(presetSelectedEntry(t, entry)),
					)
				})...,
			).FullWidth().Alignment(ui.TopLeading).Font(ui.Monospace),
		).Axis(ui.ScrollViewAxisVertical).Frame(ui.Frame{Width: ui.Full}),
		ui.Spacer(),
		pager(wnd, activePage, page),
	).Alignment(ui.TopLeading)
}

func pager(wnd core.Window, activePage *core.State[int], page inspector.PageResult) core.View {
	if page.Pages <= 1 {
		return ui.HStack(
			ui.TertiaryButton(nil).PreIcon(flowbiteOutline.ChevronLeft).Enabled(false),
			ui.Text("1 von 1"),
			ui.TertiaryButton(nil).PreIcon(flowbiteOutline.ChevronRight).Enabled(false),
		).FullWidth().Gap(ui.L16)
	}

	return ui.HStack(
		ui.TertiaryButton(func() {
			activePage.Set(activePage.Get() - 1)
		}).PreIcon(flowbiteOutline.ChevronLeft).Enabled(activePage.Get() > 0),
		ui.Text(fmt.Sprintf("%d von %d", page.PageNo+1, page.Pages)),
		ui.TertiaryButton(func() {
			activePage.Set(activePage.Get() + 1)
		}).PreIcon(flowbiteOutline.ChevronRight).Enabled(activePage.Get() < page.Pages-1),
	).Gap(ui.L8).FullWidth()
}
