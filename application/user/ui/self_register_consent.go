// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"slices"

	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/accessibility"
	"go.wdy.de/nago/presentation/ui/markdown"
)

func consents(wnd core.Window, usrSettings user.Settings, consentStates map[consent.ID]*core.State[bool]) core.View {

	return ui.VStack(
		slices.Collect(func(yield func(view core.View) bool) {
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

				// The texts are labels of the checkbox, so that a click on them toggles it, like in a CheckboxField.
				// The consent text names the checkbox and stays readable for assistive technology, because it may
				// contain links. The supporting and the error text are announced as description instead.
				id := "consent-" + string(consentOpt.ID)
				errText := ""
				if hasErr.Get() {
					errText = "Die Zustimmung ist erforderlich."
				}

				description := errText
				if description == "" {
					description = consentOpt.Register.SupportingText
				}

				yield(
					ui.HStack(
						accessibility.Accessible(
							ui.Checkbox(state.Get()).InputValue(state).ID(id),
						).
							Description(description).
							Invalid(errText != ""),
						ui.VStack(
							markdown.Render(markdown.Options{Window: wnd, LabelFor: id}, []byte(label)),
							accessibility.Accessible(
								ui.Text(consentOpt.Register.SupportingText).Font(ui.Small).LabelFor(id),
							).Hidden(true),
							accessibility.Accessible(
								ui.Text(errText).Color(ui.ColorError).LabelFor(id).Visible(errText != ""),
							).Hidden(true),
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
