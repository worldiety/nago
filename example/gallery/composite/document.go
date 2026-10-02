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
	"go.wdy.de/nago/presentation/ui/document"
)

func init() {
	app.Register("document", func(wnd core.Window) core.View {
		editing := core.AutoState[bool](wnd)
		title := core.AutoState[string](wnd).Init(func() string { return "Service agreement" })
		when := time.Date(2026, 3, 2, 9, 30, 0, 0, time.UTC)

		return document.Page(
			document.Editable(
				func() core.View {
					return Text(title.Get()).Font(HeadlineMedium)
				},
				func() core.View {
					return TextField("Title", title.Get()).InputValue(title)
				},
			).InputValue(editing).FullWidth(),
			Text("This agreement describes the services which the provider renders for the customer."),
		).Alignment(TopLeading).
			Size(document.Size{Width: L560, Height: L320}).
			Comment(
				document.LogEntry("Created the document", "Anna", when),
				document.LogEntry("Changed the title", "Ben", when.Add(2*time.Hour)),
			)
	})
}
