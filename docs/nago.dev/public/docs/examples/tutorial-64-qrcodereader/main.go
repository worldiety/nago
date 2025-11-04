// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	"fmt"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
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
			valuesRead := core.AutoState[[]string](wnd)
			selectedMediaDevices := core.AutoState[[]core.MediaDevice](wnd)
			allAvailableMediaDevices := core.AutoState[[]core.MediaDevice](wnd)

			core.OnAppear(wnd, "list-devices", func(ctx context.Context) {
				wnd.MediaDevices().List(core.MediaDeviceListOptions{WithVideo: true}).Observe(func(t []core.MediaDevice, err error) {
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					allAvailableMediaDevices.Update(t)
					fmt.Println("got media devices", t)
				})
			})

			return ui.VStack(
				ui.Text("qr code reader demo"),
				picker.Picker[core.MediaDevice]("Dieses Gerät zum Scannen verwenden", allAvailableMediaDevices.Get(), selectedMediaDevices).
					ItemRenderer(func(item core.MediaDevice) core.View {
						return ui.Text(item.Label())
					}).
					ItemPickedRenderer(func(items []core.MediaDevice) core.View {
						if len(items) > 0 {
							return ui.Text(items[0].Label())
						}
						return ui.Text("")
					}).
					SupportingText("Wähle eine Kamera aus").
					Title("Alle Geräte").
					Frame(ui.Frame{Width: ui.L320}),
				ui.QrCodeReader(getCurrentSelectedMediaDevice(selectedMediaDevices.Get())).InputValue(valuesRead).NoMediaDeviceContent(
					ui.HStack(ui.Text("Kein Gerät ausgewählt")).FullWidth(),
				).Frame(ui.Frame{}.Size(ui.L320, ui.L320)),
				list.List(ui.ForEach(valuesRead.Get(), func(value string) core.View {
					return ui.Text(value)
				})...),
			).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

func getCurrentSelectedMediaDevice(mediaDevices []core.MediaDevice) core.MediaDevice {
	if len(mediaDevices) > 0 {
		return mediaDevices[0]
	}

	return core.MediaDevice{}
}
