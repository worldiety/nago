// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"strconv"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_28")
		cfg.Serve(vuejs.Dist())

		cfg.RootView(".", func(wnd core.Window) core.View {
			stateCount := core.AutoState[int](wnd).Init(func() int { return 6 })
			listEntries := make([]core.View, 0)
			lastEntryID := ""
			for i := range stateCount.Get() {
				entryID := "entry-" + strconv.Itoa(i)
				lastEntryID = entryID
				listEntries = append(listEntries, ui.VStack(ui.Text(strconv.Itoa(i+1)).Padding(ui.Padding{}.All(ui.L8))).ID(entryID))
			}

			return ui.VStack(
				ui.HStack(
					ui.SecondaryButton(func() { stateCount.Set(stateCount.Get() - 1) }).Title("Remove Entry").Disabled(stateCount.Get() <= 0),
					ui.PrimaryButton(func() { stateCount.Set(stateCount.Get() + 1) }).Title("Add Entry"),
				).Gap(ui.L8).FullWidth(),
				ui.VStack(
					ui.ScrollView(
						ui.VStack(
							ui.VStack(listEntries...),
						),
					).
						BackgroundColor(ui.M2).
						ScrollToView(lastEntryID, ui.ScrollAnimationSmooth).
						ScrollBehavior(ui.ScrollBehaviorAuto).
						ScrollButtonLabel("Neuer Eintrag").
						Frame(ui.Frame{Width: ui.L480, Height: ui.L320}),
				),
			).Gap(ui.L32).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}
