// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/web/vuejs"
	"strings"
	"time"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			firstname := core.AutoState[string](wnd)
			secret := core.AutoState[string](wnd)
			showAlert := core.AutoState[bool](wnd)
			myIntState := core.AutoState[int64](wnd)
			myFloatState := core.AutoState[float64](wnd)

			// add artificial latency
			time.Sleep(time.Millisecond * 500)

			return VStack(
				VStack(
					alert.Dialog("Achtung", Text(fmt.Sprintf("Deine Eingabe: %v\nsecret: %v\n int-field: %v\n float-field: %v", firstname, secret, myIntState, myFloatState)), showAlert, alert.Ok()),
					TextField("hello world", firstname.Get()).
						InputValue(firstname).
						Leading(ImageIcon(heroOutline.UserCircle)).
						Trailing(ImageIcon(heroOutline.UserCircle)).
						FullWidth(),
					// you can re-use the state, but be careful of the effects
					TextField("just numbers", numsOf(firstname.Get())).
						InputValue(firstname).
						KeyboardType(KeyboardInteger).
						Style(TextFieldReduced).
						FullWidth(),

					// learn task: take your time to understand what
					// the difference between value and input value is
					IntField("int-field", 42, myIntState).FullWidth(),
					FloatField("float-field", 42.5, myFloatState).FullWidth(),

					TextField("text area", "hello\nworld").Lines(3).FullWidth(),
					PrimaryButton(func() {
						showAlert.Set(true)
					}).Title("Check"),

					PasswordField("your secret", secret.Get()).InputValue(secret).FullWidth(),
				).
					Gap(L16).
					Frame(Frame{MaxWidth: L880}),
			).
				Frame(Frame{}.MatchScreen())
		})
	}).Run()
}

func numsOf(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		}
	}

	return sb.String()
}
