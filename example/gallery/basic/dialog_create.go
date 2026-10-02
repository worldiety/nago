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

// Contact is the model of the dialog_create and dialog_edit demos.
type Contact struct {
	Name  string `label:"Name"`
	Email string `label:"Email"`
	Notes string `label:"Notes" lines:"3"`
}

func init() {
	app.Register("dialog-create", func(wnd core.Window) core.View {
		presented := core.AutoState[bool](wnd)

		return VStack(
			form.DialogCreate(wnd, "New contact", presented, func(subject auth.Subject, c Contact) error {
				// store the contact
				return nil
			}),
			PrimaryButton(func() { presented.Set(true) }).Title("New contact"),
		)
	})
}
