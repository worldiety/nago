// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	"strings"

	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("qr-code-reader", func(wnd core.Window) core.View {
		cameras := core.AutoState[[]core.MediaDevice](wnd)
		scanned := core.AutoState[[]string](wnd)

		core.OnAppear(wnd, "list-cameras", func(ctx context.Context) {
			wnd.MediaDevices().List(core.MediaDeviceListOptions{WithVideo: true}).Observe(func(devices []core.MediaDevice, err error) {
				if err == nil {
					cameras.Set(devices)
				}
			})
		})

		var camera core.MediaDevice
		if len(cameras.Get()) > 0 {
			camera = cameras.Get()[0]
		}

		return VStack(
			QrCodeReader(camera).
				InputValue(scanned).
				ShowTracker(true).
				NoMediaDeviceContent(Text("No camera available")).
				Frame(Frame{}.Size(L320, L320)),
			Text("Scanned: "+strings.Join(scanned.Get(), ", ")),
		).Gap(L16)
	})
}
