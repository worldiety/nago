// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"fmt"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

func adoption(wnd core.Window, usrSettings user.Settings, showError, adoptGDPR, adoptGTC, adoptNewsletter, adoptMinAge, adoptSendSMS *core.State[bool]) core.View {

	return ui.VStack(
		ui.Space(ui.L48),
		ui.Space(ui.L8), // -8 due to gap
		ui.IfFunc(usrSettings.RequireDataProtectionConditions, func() core.View {
			return ui.HStack(
				ui.Checkbox(adoptGDPR.Get()).InputChecked(adoptGDPR),
				ui.TextLayout(
					ui.Text("Ja, ich habe die "),
					ui.Link(wnd, "Datenschutzbestimmungen", "#", "_blank"),
					ui.Text(" gelesen und akzeptiert"),
				),
			).Alignment(ui.TopLeading)
		}),

		ui.IfFunc(usrSettings.RequireTermsAndConditions, func() core.View {
			return ui.HStack(
				ui.Checkbox(adoptGTC.Get()).InputChecked(adoptGTC),
				ui.TextLayout(
					ui.Text("Ja, ich habe die "),
					ui.Link(wnd, "Geschäftsbedingungen", "#", "_blank"),
					ui.Text(" gelesen und akzeptiert"),
				),
			).Alignment(ui.TopLeading)
		}),

		ui.IfFunc(usrSettings.RequireMinAge > 0, func() core.View {
			return ui.HStack(
				ui.Checkbox(adoptMinAge.Get()).InputChecked(adoptMinAge),
				ui.Text(fmt.Sprintf("Ja, ich bestätige, dass ich mindestens %d Jahre alt bin.", usrSettings.RequireMinAge)),
			).Alignment(ui.TopLeading)
		}),

		ui.IfFunc(usrSettings.CanAcceptNewsletter, func() core.View {
			return ui.HStack(
				ui.Checkbox(adoptNewsletter.Get()).InputChecked(adoptNewsletter),
				ui.VStack(
					ui.Text("Ja, ich melde mich zum Newsletter an. Eine Abbestellung ist jederzeit möglich. (optional)"),
					ui.Text("Ein Widerspruch ist jederzeit in den Einstellungen Ihres Benutzerkontos/über Abmeldelink in den E-Mails möglich, ohne dass weitere (Übermittlungs-)Kosten als die nach den Basistarifen entstehen.").Font(ui.Small),
				),
			).Alignment(ui.TopLeading)
		}),

		ui.IfFunc(usrSettings.CanReceiveSMS, func() core.View {
			return ui.HStack(
				ui.Checkbox(adoptSendSMS.Get()).InputChecked(adoptSendSMS),
				ui.VStack(
					ui.Text("Ja, ich melde mich zum SMS Versand an. Eine Abbestellung ist jederzeit möglich. (optional)"),
					ui.Text("Ein Widerspruch ist jederzeit in den Einstellungen Ihres Benutzerkontos möglich.").Font(ui.Small),
				),
			).Alignment(ui.TopLeading)
		}),

		ui.IfFunc(showError.Get() && !validateAdoption(usrSettings, adoptGDPR, adoptGTC, adoptMinAge), func() core.View {
			var msg string
			if usrSettings.RequireDataProtectionConditions && !adoptGDPR.Get() {
				msg += "Bitte akzeptieren Sie unsere Datenschutzbestimmungen. "
			}

			if usrSettings.RequireTermsAndConditions && !adoptGTC.Get() {
				msg += "Bitte akzeptieren Sie unsere Geschäftsbedingungen. "
			}

			if usrSettings.RequireMinAge > 0 && !adoptMinAge.Get() {
				msg += "Bitte bestätigen Sie Ihr Alter. "
			}

			return alert.Banner("Zustimmung erforderlich", msg).Frame(ui.Frame{}.FullWidth())
		}),
	).Alignment(ui.TopLeading).FullWidth().Gap(ui.L8)
}

func validateAdoption(usrSettings user.Settings, adoptGDPR, adoptGTC, adoptMinAge *core.State[bool]) bool {
	if usrSettings.RequireTermsAndConditions && !adoptGTC.Get() {
		return false
	}

	if usrSettings.RequireMinAge > 0 && !adoptMinAge.Get() {
		return false
	}

	if usrSettings.RequireDataProtectionConditions && !adoptGDPR.Get() {
		return false
	}

	return true
}
