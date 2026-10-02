// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/form"
)

func init() {
	app.Register("dialog-edit", func(wnd core.Window) core.View {
		presented := core.AutoState[bool](wnd)
		contact := core.AutoState[Contact](wnd).Init(func() Contact {
			return Contact{Name: "Ada Lovelace", Email: "ada@example.com"}
		})

		return VStack(
			form.DialogEdit(wnd, "Edit contact", presented, contact, func(subject auth.Subject) error {
				// store contact.Get()
				return nil
			}),
			PrimaryButton(func() { presented.Set(true) }).Title("Edit contact"),
		)
	})
}
