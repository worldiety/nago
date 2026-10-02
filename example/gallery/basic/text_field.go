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
	icons "go.wdy.de/nago/presentation/icons/hero/solid"
	. "go.wdy.de/nago/presentation/ui"
)

func init() {
	app.Register("text-field", func(wnd core.Window) core.View {
		name := core.AutoState[string](wnd)
		age := core.AutoState[int64](wnd).Init(func() int64 { return 42 })

		return VStack(
			TextField("Name", "").
				InputValue(name).
				SupportingText("Your full name").
				Frame(Frame{Width: L320}),
			TextField("Search", "").
				Leading(ImageIcon(icons.MagnifyingGlass)).
				Placeholder("Find anything").
				Frame(Frame{Width: L320}),
			IntField("Age", age.Get(), age).
				Frame(Frame{Width: L320}),
			TextField("Email", "nago@").
				ErrorText("This is not a valid email address.").
				Frame(Frame{Width: L320}),
		).Alignment(Leading).Gap(L16)
	})
}
