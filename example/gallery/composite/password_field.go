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
	app.Register("password-field", func(wnd core.Window) core.View {
		password := core.AutoState[string](wnd).Init(func() string { return "correct horse" })
		repeated := core.AutoState[string](wnd).Init(func() string { return "correct house" })

		return VStack(
			PasswordField("Password", password.Get()).
				InputValue(password).
				SupportingText("At least 12 characters").
				FullWidth(),
			PasswordField("Repeat password", repeated.Get()).
				InputValue(repeated).
				ErrorText("The passwords do not match").
				FullWidth(),
		).Gap(L16).Frame(Frame{Width: L320})
	})
}
