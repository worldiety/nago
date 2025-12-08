// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"fmt"
	"strings"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	. "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			text := core.AutoState[string](wnd).Init(func() string {
				return "hello Thomas"
			})

			return VStack(
				VStack(
					TextField("hello world", text.Get()).
						InputValue(text).
						Lines(1).
						FullWidth().KeydownEnter(func() {
						fmt.Println("Hello world")
					}),
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
