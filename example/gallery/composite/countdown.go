// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"time"

	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("countdown", func(wnd core.Window) core.View {
		done := core.AutoState[bool](wnd)

		return VStack(
			CountDown(2*time.Hour+15*time.Minute).
				Style(CountDownStyleClock).
				Days(false),
			CountDown(10*time.Minute).
				Style(CountDownStyleProgress).
				Done(done.Get()).
				Action(func() { done.Set(true) }).
				Frame(Frame{Width: L480}),
		).Alignment(Leading).Gap(L32)
	})
}
