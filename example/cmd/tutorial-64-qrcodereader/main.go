// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xmediadevice"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/picker"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_64")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		option.MustZero(cfg.StandardSystems())
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			valueRead := core.AutoState[string](wnd)
			allAvailableMediaDevices := core.AutoState[[]xmediadevice.MediaDevice](wnd)
			selectedMediaDevice := core.AutoState[[]xmediadevice.MediaDevice](wnd)
			if len(allAvailableMediaDevices.Get()) > 0 {
				// initially set first media device as selected if available
				firstMediaDevice := []xmediadevice.MediaDevice{allAvailableMediaDevices.Get()[0]}
				selectedMediaDevice.Set(firstMediaDevice)
			}
			return ui.VStack(
				ui.Text("qr code reader demo"),
				ui.MediaDevices().InputValue(allAvailableMediaDevices).WithAudio(false),
				picker.Picker[xmediadevice.MediaDevice]("Dieses Gerät zum Scannen verwenden", allAvailableMediaDevices.Get(), selectedMediaDevice).
					ItemRenderer(func(item xmediadevice.MediaDevice) core.View {
						return ui.Text(item.Label)
					}).
					ItemPickedRenderer(func(items []xmediadevice.MediaDevice) core.View {
						if len(items) > 0 {
							return ui.Text(items[0].Label)
						}
						return ui.Text("")
					}).
					SupportingText("Wähle eine Kamera aus").
					Title("Alle Geräte").
					Frame(ui.Frame{Width: ui.L320}),
				ui.QrCodeReader(getCurrentSelectedMediaDevice(selectedMediaDevice.Get())).InputValue(valueRead).Frame(ui.Frame{}.Size(ui.L320, ui.L320)),
				ui.Text("Letztes Ergebnis: "+valueRead.Get()),
			).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

func getCurrentSelectedMediaDevice(mediaDevices []xmediadevice.MediaDevice) xmediadevice.MediaDevice {
	if len(mediaDevices) > 0 {
		return mediaDevices[0]
	}

	return xmediadevice.MediaDevice{}
}
