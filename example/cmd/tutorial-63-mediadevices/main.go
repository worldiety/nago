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
	"go.wdy.de/nago/web/vuejs"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_63")

		cfg.Serve(vuejs.Dist())
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		option.MustZero(cfg.StandardSystems())
		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "%6UbRsCuM8N$auy"))

		cfg.RootViewWithDecoration(".", func(wnd core.Window) core.View {
			mediaDevices := core.AutoState[[]core.MediaDevice](wnd)

			core.OnAppear(wnd, "list-devices", func(ctx context.Context) {
				wnd.MediaDevices().List(core.MediaDeviceListOptions{WithVideo: true}).Observe(func(t []core.MediaDevice, err error) {
					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					mediaDevices.Update(t)
					fmt.Println("got media devices", t)
				})
			})

			return list.List(
				ui.ForEach(mediaDevices.Get(), func(dev core.MediaDevice) core.View {
					return list.Entry().Headline(dev.Label()).SupportingText(string(dev.ID()))
				})...,
			).Caption(ui.Text("media devices demo"))

		})
	}).Run()
}
