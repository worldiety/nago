// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package application_test

import (
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// An import id may contain any character, e.g. the label of a button, and the upload must still find it.
func TestUploadWithANonASCIIImportID(t *testing.T) {
	const id = "Datei wählen"
	base := nagotest.Serve(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.uploadumlauttest")
		cfg.RootView(".", func(wnd core.Window) core.View {
			name := core.AutoState[string](wnd)
			return ui.VStack(
				ui.PrimaryButton(func() {
					wnd.ImportFiles(core.ImportFilesOptions{
						ID: id,
						OnCompletion: func(files []core.File) {
							name.Set("received " + files[0].Name())
						},
					})
				}).Title("pick"),
				ui.Text(name.Get()),
			)
		})
	})

	w := nagotest.Dial(t, base, ".")
	w.Click(w.Find(nagotest.Text("pick")))
	w.Upload(id, nagotest.File("a.txt", []byte("hello")))
	w.WaitFor(nagotest.Text("received a.txt"), nagotest.RenderTimeout)
}
