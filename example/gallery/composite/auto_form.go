// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"go.wdy.de/nago/example/gallery/app"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/form"
)

type contact struct {
	Name       string     `label:"Name"`
	Email      string     `label:"E-mail" supportingText:"We never share your address."`
	Birthday   xtime.Date `label:"Birthday"`
	Notes      string     `label:"Notes" lines:"3"`
	Newsletter bool       `label:"Subscribe to the newsletter"`
}

func init() {
	app.Register("auto-form", func(wnd core.Window) core.View {
		state := core.AutoState[contact](wnd).Init(func() contact {
			return contact{Name: "Ada Lovelace", Email: "ada@example.com", Newsletter: true}
		})

		return form.Auto(form.AutoOptions{Window: wnd}, state).Frame(Frame{Width: L560})
	})
}
