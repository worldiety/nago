// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/dropdown"
)

func init() {
	app.Register("select", func(wnd core.Window) core.View {
		country := core.AutoState[string](wnd).Init(func() string { return "de" })

		return dropdown.Dropdown("Country", []dropdown.Option[string]{
			{Value: "de", Label: "Germany"},
			{Value: "fr", Label: "France"},
			{Value: "it", Label: "Italy"},
		}, country.Get()).
			InputValue(country).
			SupportingText("Where do you live?").
			Frame(Frame{Width: L320})
	})
}
