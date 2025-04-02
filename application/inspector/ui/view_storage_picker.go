// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiinspector

import (
	"go.wdy.de/nago/application/backup"
	"go.wdy.de/nago/application/inspector"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func storagePicker(wnd core.Window, uc inspector.UseCases, stores []inspector.Store, selected *core.State[inspector.Store]) ui.DecoredView {
	var tmp []core.View
	tmp = append(tmp, ui.Text("Entities").Font(ui.SubTitle))
	for _, store := range stores {
		if store.Stereotype == backup.StereotypeDocument {
			tmp = append(tmp, ui.TertiaryButton(func() {
				selected.Set(store)
				selected.Notify()
			}).Title(store.Name).Preset(presetSelectedStore(store, selected)))
		}
	}

	tmp = append(tmp, ui.Text("Blobs").Font(ui.SubTitle))
	for _, store := range stores {
		if store.Stereotype == backup.StereotypeBlob {
			tmp = append(tmp, ui.TertiaryButton(func() {
				selected.Set(store)
				selected.Notify()
			}).Title(store.Name).Preset(presetSelectedStore(store, selected)))
		}
	}

	return ui.VStack(
		ui.ScrollView(
			ui.VStack(
				tmp...,
			).Alignment(ui.TopLeading).Font(ui.Monospace).Frame(ui.Frame{Width: ui.Full}),
		).Axis(ui.ScrollViewAxisVertical).Frame(ui.Frame{}.FullWidth()),
	).Alignment(ui.TopLeading).Frame(ui.Frame{Width: ui.Full})
}
