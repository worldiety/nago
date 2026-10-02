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
)

func init() {
	app.Register("checkbox-field", func(wnd core.Window) core.View {
		accepted := core.AutoState[bool](wnd)
		newsletter := core.AutoState[bool](wnd)

		return VStack(
			CheckboxField("I accept the terms of use", accepted.Get()).
				InputValue(accepted).
				SupportingText("You can revoke your consent at any time."),
			CheckboxField("Subscribe to the newsletter", newsletter.Get()).
				InputValue(newsletter).
				ErrorText("Please confirm your email address first."),
		).Alignment(Leading).Gap(L16)
	})
}
