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

const cssHeight = "calc(100dvh - 20rem)"

func PageInspector(wnd core.Window, uc inspector.UseCases) core.View {
	stores, err := uc.FindAll(wnd.Subject())
	if err != nil {
		return alert.BannerError(err)
	}

	selectedEntity := core.AutoState[inspector.PageEntry](wnd)

	selectedStore := core.AutoState[inspector.Store](wnd)
	selectedStore.Observe(func(newValue inspector.Store) {
		selectedEntity.Set(inspector.PageEntry{})
	})

	return ui.VStack(
		ui.HStack(ui.H1("Inspector")).Alignment(ui.Leading),
		ui.HStack(
			ui.TertiaryButton(func() {
				text := getTextState(wnd, selectedEntity).Get()
				if err := blob.Put(selectedStore.Get().Store, selectedEntity.Get().Key, []byte(text)); err != nil {
					alert.ShowBannerError(wnd, err)
				}
			}).PreIcon(flowbiteOutline.FloppyDisk).
				AccessibilityLabel(fmt.Sprintf("%s.%s speichern", selectedStore.Get().Name, selectedEntity.Get().Key)).
				Visible(selectedEntity.Get().Key != ""),
		).Alignment(ui.Trailing),
		ui.HLine().Padding(ui.Padding{Bottom: ui.L4}),
		ui.HStack(
			storagePicker(wnd, uc, stores, selectedStore).
				Frame(ui.Frame{MinWidth: "23rem", Height: cssHeight}),
			ui.VLine().Padding(ui.Padding{Left: ui.L8, Right: ui.L8}).Frame(ui.Frame{}),
			viewKeys(wnd, uc, selectedStore, selectedEntity).
				Frame(ui.Frame{MinWidth: "23rem", Height: cssHeight}),
			ui.VLine().Padding(ui.Padding{Left: ui.L8, Right: ui.L8}).Frame(ui.Frame{}),
			viewKeyContent(wnd, uc, selectedStore, selectedEntity),
		).Alignment(ui.Stretch).Frame(ui.Frame{Width: ui.Full, Height: ""}),
	).Alignment(ui.Stretch).Frame(ui.Frame{Width: ui.Full, Height: ""})
}

func presetSelectedStore(store inspector.Store, selected *core.State[inspector.Store]) ui.StylePreset {
	if store.Name == selected.Get().Name {
		return ui.StyleButtonSecondary
	}

	return ui.StyleButtonTertiary
}

func presetSelectedEntry(entry inspector.PageEntry, selected *core.State[inspector.PageEntry]) ui.StylePreset {
	if entry.Key == selected.Get().Key {
		return ui.StyleButtonSecondary
	}

	return ui.StyleButtonTertiary
}
