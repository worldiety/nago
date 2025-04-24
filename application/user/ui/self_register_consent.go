// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/markdown"
	"slices"
)

func consents(wnd core.Window, usrSettings user.Settings, consentStates map[consent.ID]*core.State[bool]) core.View {

	return ui.VStack(
		slices.Collect(func(yield func(view core.View) bool) {

			yield(ui.Space(ui.L48))
			yield(ui.Space(ui.L8)) // -8 due to gap

			for _, consentOpt := range usrSettings.Consents {
				state := consentStates[consentOpt.ID]
				hasErr := consentStates[consentOpt.ID+"err"]

				supportingText := consentOpt.Register.SupportingText
				if supportingText == "" {
					supportingText = "Ein Widerspruch ist jederzeit in den Einstellungen Ihres Benutzerkontos möglich."
				}

				label := consentOpt.Register.Label
				if !consentOpt.Required {
					label += " (optional)"
				}

				yield(
					ui.HStack(
						ui.Checkbox(state.Get()).InputChecked(state),
						ui.VStack(
							markdown.Render(markdown.Options{Window: wnd}, []byte(label)),
							ui.Text(consentOpt.Register.SupportingText).Font(ui.Small),
							ui.Text("Die Zustimmung ist erforderlich.").Color(ui.ColorError).Visible(hasErr.Get()),
						).Alignment(ui.TopLeading),
					).Alignment(ui.TopLeading),
				)

			}
		})...,
	).Alignment(ui.TopLeading).FullWidth().Gap(ui.L8)
}

func validateConsents(usrSettings user.Settings, consents map[consent.ID]*core.State[bool]) bool {
	validates := true
	for _, option := range usrSettings.Consents {
		consents[option.ID+"err"].Set(false)
		if option.Required && consents[option.ID].Get() == false {
			consents[option.ID+"err"].Set(true)
			validates = false
		}
	}

	return validates
}
