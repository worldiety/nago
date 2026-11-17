// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/worldiety/i18n/date"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// requestRegistration is the first step of the self registration: the visitor enters the mail address and gets a
// link, which opens the rest of the registration.
type requestRegistration struct {
	wnd      core.Window
	email    *core.State[string]
	errEmail *core.State[string]
	sentTo   *core.State[string]
	notice   *core.State[string]
}

func newRequestRegistration(wnd core.Window) requestRegistration {
	return requestRegistration{
		wnd:      wnd,
		email:    core.AutoState[string](wnd),
		errEmail: core.AutoState[string](wnd),
		sentTo:   core.AutoState[string](wnd),
		notice:   core.AutoState[string](wnd),
	}
}

// send requests a link for the address and tells, if it has been sent. A failure is shown at the address field, or
// as notice, if a link has been sent before, because that view has no field.
func (r requestRegistration) send(request user.RequestRegistration, email string) bool {
	r.errEmail.Set("")
	r.notice.Set("")

	// the session limits the links a single visitor can request
	err := request(string(r.wnd.Session().ID()), user.Email(email))
	if err == nil {
		r.sentTo.Set(email)
		return true
	}

	var msg string
	var cooldown user.RegistrationCooldownError
	if errors.As(err, &cooldown) {
		retryAt := date.Format(r.wnd.Locale(), date.TimeMinute, nextFullMinute(cooldown.RetryAt).In(r.wnd.Location()))
		msg = fmt.Sprintf("Bitte warten Sie bis %s Uhr, bevor Sie einen neuen Link anfordern.", retryAt)
	} else if localized, ok := localizedMessage(err); ok {
		msg = localized
	} else {
		slog.Error("cannot request registration", "err", err)
		msg = "Der Link konnte nicht gesendet werden. Bitte versuchen Sie es später erneut."
	}

	if r.sentTo.Get() != "" {
		r.notice.Set(msg)
	} else {
		r.errEmail.Set(msg)
	}

	return false
}

// nextFullMinute rounds the time up to a full minute. A waiting time is shown by the minute, and at the shown
// minute the waiting must be over.
func nextFullMinute(t time.Time) time.Time {
	if full := t.Truncate(time.Minute); full.Before(t) {
		return full.Add(time.Minute)
	}

	return t
}

// view returns the subcaption, the body and the actions of the card.
func (r requestRegistration) view(request user.RequestRegistration) (string, core.View, []core.View) {
	if r.sentTo.Get() != "" {
		body := ui.VStack(
			ui.TextLayout(
				ui.Text("Wir haben eine E-Mail an "),
				inlineCode(r.sentTo.Get()),
				ui.Text(" gesendet. Öffnen Sie den Link darin, um Ihre Angaben zu ergänzen und das Konto zu erstellen."),
			),
			ui.Text(fmt.Sprintf("Der Link ist %d Stunden gültig. Keine E-Mail erhalten? Prüfen Sie auch den Spam-Ordner.", int(user.DefaultRegistrationLifeTime.Hours()))).Font(ui.BodySmall),
			ui.If(r.notice.Get() != "", ui.Text(r.notice.Get()).Font(ui.BodySmall)),
		).Alignment(ui.Leading).FullWidth().Gap(ui.L16)

		actions := []core.View{
			ui.TertiaryButton(func() {
				r.sentTo.Set("")
				r.notice.Set("")
			}).Title("Andere Adresse"),
			ui.PrimaryButton(func() {
				if r.send(request, r.sentTo.Get()) {
					r.notice.Set("Ein neuer Link wurde gesendet.")
				}
			}).Title("Erneut senden"),
		}

		return "Postfach prüfen", body, actions
	}

	body := ui.VStack(
		ui.Text("Wir senden Ihnen einen Link an diese Adresse. Über den Link ergänzen Sie Ihre Angaben und erstellen das Konto.").Padding(ui.Padding{Bottom: ui.L16}),
		ui.TextField("E-Mail Adresse", r.email.Get()).
			ErrorText(r.errEmail.Get()).
			InputValue(r.email).
			KeyboardType(ui.KeyboardEMail).
			FullWidth(),
	).FullWidth().Gap(ui.L8)

	actions := []core.View{
		ui.PrimaryButton(func() {
			r.send(request, r.email.Get())
		}).Title("Link senden"),
	}

	return "Bitte die E-Mail eingeben", body, actions
}

// inlineCode highlights a value within a running text, the way inline code is shown, so that e.g. a mail address
// stands out and can be checked for typos.
func inlineCode(value string) core.View {
	return ui.Text(value).
		WordBreak(ui.WordBreakBreakWord).
		Font(ui.MonoMedium).
		BackgroundColor(ui.ColorBackground).
		Border(ui.Border{}.Radius(ui.L4)).
		Padding(ui.Padding{}.Horizontal(ui.L4))
}
