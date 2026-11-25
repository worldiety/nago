// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uipush

import (
	"context"
	"errors"
	"fmt"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/push"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

var (
	StrTitle           = i18n.MustString("nago.push.title", i18n.Values{language.German: "Benachrichtigungen", language.English: "Notifications"})
	StrToggle          = i18n.MustString("nago.push.toggle", i18n.Values{language.German: "Benachrichtigungen auf diesem Gerät", language.English: "Notifications on this device"})
	StrLoading         = i18n.MustString("nago.push.loading", i18n.Values{language.German: "Der Browser wird geprüft …", language.English: "Checking the browser …"})
	StrUnsupported     = i18n.MustString("nago.push.unsupported", i18n.Values{language.German: "Dieser Browser kann keine Benachrichtigungen empfangen.", language.English: "This browser cannot receive notifications."})
	StrRequiresInstall = i18n.MustString("nago.push.requires_install", i18n.Values{language.German: "Auf iPhone und iPad gibt es Benachrichtigungen nur für eine installierte App: Tippe auf „Teilen“ und „Zum Home-Bildschirm“ und öffne die App dann vom Home-Bildschirm.", language.English: "On iPhone and iPad, only an installed app gets notifications: tap “Share” and “Add to Home Screen” and open the app from the home screen."})
	StrDenied          = i18n.MustString("nago.push.denied", i18n.Values{language.German: "Du hast Benachrichtigungen für diese Seite im Browser blockiert. Du kannst sie in den Website-Einstellungen des Browsers wieder erlauben.", language.English: "You have blocked notifications for this site in the browser. You can allow them again in the site settings of the browser."})
	StrInactive        = i18n.MustString("nago.push.inactive", i18n.Values{language.German: "Lass dich auf diesem Gerät benachrichtigen, auch wenn die App geschlossen ist. Der Browser fragt dich dann nach der Erlaubnis.", language.English: "Get notified on this device, even if the app is closed. The browser then asks you for the permission."})
	StrSignIn          = i18n.MustString("nago.push.sign_in", i18n.Values{language.German: "Melde dich an, um auf diesem Gerät Benachrichtigungen zu erhalten.", language.English: "Sign in to get notifications on this device."})
	StrActive          = i18n.MustString("nago.push.active", i18n.Values{language.German: "Dieses Gerät erhält deine Benachrichtigungen.", language.English: "This device gets your notifications."})
	StrNotAllowed      = i18n.MustString("nago.push.not_allowed", i18n.Values{language.German: "Benachrichtigungen wurden nicht erlaubt.", language.English: "Notifications have not been allowed."})
	StrNoAnswer        = i18n.MustString("nago.push.no_answer", i18n.Values{language.German: "Der Push-Dienst dieses Geräts antwortet nicht. Versuche es später noch einmal.", language.English: "The push service of this device does not answer. Try again later."})
	StrEnable          = i18n.MustString("nago.push.enable", i18n.Values{language.German: "Benachrichtigungen aktivieren", language.English: "Enable notifications"})
	StrDisable         = i18n.MustString("nago.push.disable", i18n.Values{language.German: "Auf diesem Gerät deaktivieren", language.English: "Disable on this device"})
	StrSendTest        = i18n.MustString("nago.push.send_test", i18n.Values{language.German: "Test senden", language.English: "Send a test"})
	StrTestTitle       = i18n.MustString("nago.push.test_title", i18n.Values{language.German: "Benachrichtigungen sind aktiv", language.English: "Notifications are active"})
	StrTestBody        = i18n.MustString("nago.push.test_body", i18n.Values{language.German: "So sehen Benachrichtigungen dieser App aus.", language.English: "This is how notifications of this app look."})
	StrTestSentX       = i18n.MustVarString("nago.push.test_sent_x", i18n.Values{language.German: "Die Benachrichtigung ging an {count} Gerät(e).", language.English: "The notification went to {count} device(s)."})
)

// model is the state of a view, which shows the push notifications.
type model struct {
	Loaded bool
	State  State
}

// load reads the state of the browser, as soon as the view appears.
func load(wnd core.Window, id string) *core.State[model] {
	m := core.StateOf[model](wnd, id)
	core.OnAppear(wnd, id+".appear", func(ctx context.Context) {
		ReadState(wnd, func(s State, err error) {
			if err != nil {
				alert.ShowBannerError(wnd, err)
			}

			m.Update(model{Loaded: true, State: s})
		})
	})

	return m
}

// hint explains the state, which keeps the user from subscribing.
func hint(wnd core.Window, m model) string {
	switch {
	case !m.Loaded:
		return StrLoading.Get(wnd)
	case m.State.RequiresInstall:
		return StrRequiresInstall.Get(wnd)
	case !m.State.Supported:
		return StrUnsupported.Get(wnd)
	case m.State.Permission == PermissionDenied:
		return StrDenied.Get(wnd)
	case !wnd.Subject().Valid():
		return StrSignIn.Get(wnd)
	case m.State.Registered:
		return StrActive.Get(wnd)
	default:
		return StrInactive.Get(wnd)
	}
}

// subscribe subscribes the browser and shows, why it did not work.
func subscribe(wnd core.Window, m *core.State[model], after func(State)) {
	Subscribe(wnd, func(s State, err error) {
		// an error tells nothing about the browser, so the last known state stays
		if err != nil {
			if errors.Is(err, ErrNoAnswer) {
				alert.ShowBannerMessage(wnd, alert.Message{Title: StrTitle.Get(wnd), Message: StrNoAnswer.Get(wnd), Intent: alert.IntentWarning})
			} else {
				alert.ShowBannerError(wnd, err)
			}

			after(m.Get().State)
			return
		}

		if !s.Registered && s.CanSubscribe() {
			alert.ShowBannerMessage(wnd, alert.Message{Title: StrTitle.Get(wnd), Message: StrNotAllowed.Get(wnd), Intent: alert.IntentWarning})
		}

		m.Update(model{Loaded: true, State: s})
		after(s)
	})
}

func unsubscribe(wnd core.Window, m *core.State[model], after func(State)) {
	Unsubscribe(wnd, func(s State, err error) {
		if err != nil {
			alert.ShowBannerError(wnd, err)
			after(m.Get().State)
			return
		}

		m.Update(model{Loaded: true, State: s})
		after(s)
	})
}

func nothing(State) {}

// SendTest sends a test notification to all devices of the user of the window.
func SendTest(wnd core.Window) {
	uc, ok := useCases(wnd)
	if !ok {
		alert.ShowBannerError(wnd, ErrNotEnabled)
		return
	}

	report, err := uc.SendToMe(wnd.Subject(), push.Message{
		Title: StrTestTitle.Get(wnd),
		Body:  StrTestBody.Get(wnd),
		Tag:   "nago.push.test",
		Route: wnd.Path(),
		TTL:   -1,
	})
	if err != nil {
		alert.ShowBannerError(wnd, err)
		return
	}

	alert.ShowBannerMessage(wnd, alert.Message{
		Title:   StrTitle.Get(wnd),
		Message: StrTestSentX.Get(wnd, i18n.String("count", fmt.Sprint(report.Sent))),
		Intent:  alert.IntentOk,
	})
}

// Card shows, whether the browser gets the notifications of the user, and offers to enable, disable and try them.
// Only its button to enable the notifications lets the browser ask for the permission. It shows nothing, if the
// push module is not enabled.
func Card(wnd core.Window) core.View {
	if !Enabled(wnd) {
		return nil
	}

	m := load(wnd, "nago.push.card")
	cur := m.Get()

	var actions []core.View
	switch {
	case !cur.Loaded || !cur.State.CanSubscribe() || !wnd.Subject().Valid():
	case cur.State.Registered:
		actions = append(actions,
			ui.SecondaryButton(func() { unsubscribe(wnd, m, nothing) }).Title(StrDisable.Get(wnd)),
			ui.PrimaryButton(func() { SendTest(wnd) }).Title(StrSendTest.Get(wnd)),
		)
	default:
		actions = append(actions,
			ui.PrimaryButton(func() { subscribe(wnd, m, nothing) }).PreIcon(flowbiteOutline.Bell).Title(StrEnable.Get(wnd)),
		)
	}

	return ui.VStack(
		ui.HStack(
			ui.ImageIcon(flowbiteOutline.Bell),
			ui.Text(StrTitle.Get(wnd)).Font(ui.TitleMedium),
		).Gap(ui.L8),
		ui.Text(hint(wnd, cur)),
		ui.If(len(actions) > 0, ui.HStack(actions...).Gap(ui.L8).Alignment(ui.Trailing).FullWidth()),
	).Gap(ui.L16).
		Alignment(ui.Leading).
		BackgroundColor(ui.ColorCardBody).
		Padding(ui.Padding{}.All(ui.L16)).
		Border(ui.Border{}.Radius(ui.L16)).
		Frame(ui.Frame{}.FullWidth())
}

// Toggle is a compact switch for the notifications of the browser, e.g. for a settings page. Only switching it on
// lets the browser ask for the permission. It shows nothing, if the push module is not enabled.
func Toggle(wnd core.Window) core.View {
	if !Enabled(wnd) {
		return nil
	}

	m := load(wnd, "nago.push.toggle")
	cur := m.Get()

	// the switch follows the browser, also if a subscription has not been allowed
	checked := core.StateOf[bool](wnd, "nago.push.toggle.checked").Init(func() bool { return cur.State.Registered })
	sync := func(s State) { checked.Set(s.Registered) }
	m.Observe(func(v model) { sync(v.State) })
	checked.Observe(func(on bool) {
		if on == m.Get().State.Registered {
			return
		}

		if on {
			subscribe(wnd, m, sync)
		} else {
			unsubscribe(wnd, m, sync)
		}
	})

	return ui.ToggleField(StrToggle.Get(wnd), checked.Get()).
		ID("nago-push-toggle").
		InputValue(checked).
		SupportingText(hint(wnd, cur)).
		Disabled(!cur.Loaded || !cur.State.CanSubscribe() || !wnd.Subject().Valid())
}
