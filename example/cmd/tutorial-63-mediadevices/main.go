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
			mediaDevices := core.AutoState[[]xmediadevice.MediaDevice](wnd)
			hasGrantedPermissions := core.AutoState[bool](wnd)
			return ui.VStack(
				ui.Text("media devices demo"),
				ui.MediaDevices().InputValue(mediaDevices).HasGrantedPermissions(hasGrantedPermissions).WithAudio(false),
				ui.IfElse(!hasGrantedPermissions.Get(), ui.Text("Die benötigten Berechtigungen wurden nicht erteilt"), list.List(ui.ForEach(mediaDevices.Get(), func(mediaDevice xmediadevice.MediaDevice) core.View {
					return list.Entry().
						Headline(mediaDevice.Label).
						SupportingText("ID: " + mediaDevice.DeviceID + ", Gruppen-ID: " + mediaDevice.GroupID + ", Typ: " + mediaDevice.Kind.String())
				})...).Frame(ui.Frame{}.FullWidth()),
				)).Alignment(ui.Center).FullWidth()
		})
	}).Run()
}
