// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uionboarding

import (
	"net/url"
	"strings"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/onboarding"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	flowbiteOutline "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"golang.org/x/text/language"
)

var (
	StrMethodSubtitle      = i18n.MustString("nago.onboarding.method_subtitle", i18n.Values{language.German: "Wie möchtest du dich künftig anmelden?", language.English: "How do you want to sign in from now on?"})
	StrMethodSSOTitle      = i18n.MustString("nago.onboarding.method_sso_title", i18n.Values{language.German: "Mit dem Nago-Login (SSO)", language.English: "With the Nago login (SSO)"})
	StrMethodSSOTextX      = i18n.MustVarString("nago.onboarding.method_sso_text_x", i18n.Values{language.German: "Du meldest dich mit deinem vorhandenen Konto bei {server} an. Ein eigenes Passwort für diese Anwendung brauchst du nicht.", language.English: "You sign in with your existing account at {server}. You need no password of its own for this application."})
	StrMethodPasswordTitle = i18n.MustString("nago.onboarding.method_password_title", i18n.Values{language.German: "Mit E-Mail und Passwort", language.English: "With mail and password"})
	StrMethodPasswordText  = i18n.MustString("nago.onboarding.method_password_text", i18n.Values{language.German: "Du legst hier ein Konto mit Namen und Passwort an.", language.English: "You create an account with your name and a password here."})
	StrRecommended         = i18n.MustString("nago.onboarding.recommended", i18n.Values{language.German: "Empfohlen", language.English: "Recommended"})
	StrSSOSubtitle         = i18n.MustString("nago.onboarding.sso_subtitle", i18n.Values{language.German: "Wer darf sich mit dem Nago-Login anmelden?", language.English: "Who may sign in with the Nago login?"})
	StrSSOIntro            = i18n.MustString("nago.onboarding.sso_intro", i18n.Values{language.German: "Wer sich zum ersten Mal anmeldet, bekommt automatisch ein Konto, aber noch keine Berechtigungen. Die vergibst du später in der Nutzerverwaltung.", language.English: "Whoever signs in for the first time gets an account automatically, but no permissions yet. You grant them later in the user management."})
	StrSSOOnlyMe           = i18n.MustString("nago.onboarding.sso_only_me", i18n.Values{language.German: "Nur ich", language.English: "Only me"})
	StrSSODomainX          = i18n.MustVarString("nago.onboarding.sso_domain_x", i18n.Values{language.German: "Alle von {domain}", language.English: "Everyone from {domain}"})
	StrSSODomainTextX      = i18n.MustVarString("nago.onboarding.sso_domain_text_x", i18n.Values{language.German: "Jede E-Mail-Adresse, die auf @{domain} endet.", language.English: "Every mail address ending with @{domain}."})
	StrSSOSelected         = i18n.MustString("nago.onboarding.sso_selected", i18n.Values{language.German: "Ich und ausgewählte Personen oder Firmen", language.English: "Me and selected people or companies"})
	StrSSOSelectedText     = i18n.MustString("nago.onboarding.sso_selected_text", i18n.Values{language.German: "Du trägst E-Mail-Adressen oder Firmen-Domains ein.", language.English: "You enter mail addresses or company domains."})
	StrSSOEveryone         = i18n.MustString("nago.onboarding.sso_everyone", i18n.Values{language.German: "Alle mit einem Konto beim Nago-Login", language.English: "Everyone with an account at the Nago login"})
	StrSSOEveryoneText     = i18n.MustString("nago.onboarding.sso_everyone_text", i18n.Values{language.German: "Jede Person, die sich dort anmelden kann, auch mit einem privaten Google-, Apple- oder GitHub-Konto.", language.English: "Every person, who can sign in there, also with a private Google, Apple or GitHub account."})
	StrSSOEntries          = i18n.MustString("nago.onboarding.sso_entries", i18n.Values{language.German: "E-Mail-Adressen oder Domains", language.English: "Mail addresses or domains"})
	StrSSOEntriesHint      = i18n.MustString("nago.onboarding.sso_entries_hint", i18n.Values{language.German: "Zum Beispiel anna@firma.de oder firma.de. Trennen kannst du sie, wie du möchtest, auch durch Kommas oder neue Zeilen.", language.English: "For example anna@company.com or company.com. Separate them as you like, also by commas or new lines."})
	StrSSOAllowedTitle     = i18n.MustString("nago.onboarding.sso_allowed_title", i18n.Values{language.German: "Anmelden dürfen dann", language.English: "Then, these may sign in"})
	StrSSOYouX             = i18n.MustVarString("nago.onboarding.sso_you_x", i18n.Values{language.German: "{email} (du)", language.English: "{email} (you)"})
	StrSSODomainEntryX     = i18n.MustVarString("nago.onboarding.sso_domain_entry_x", i18n.Values{language.German: "alle Adressen von {domain}", language.English: "all addresses of {domain}"})
	StrSSOInvalidX         = i18n.MustVarString("nago.onboarding.sso_invalid_x", i18n.Values{language.German: "Keine E-Mail-Adresse oder Domain: {entries}", language.English: "No mail address or domain: {entries}"})
	StrSSOPublicX          = i18n.MustVarString("nago.onboarding.sso_public_x", i18n.Values{language.German: "{domain} ist ein öffentlicher Mail-Anbieter, bei dem jede Person eine Adresse bekommt. Trage hier bitte einzelne Adressen ein.", language.English: "{domain} is a public mail provider, where anybody gets an address. Please enter single addresses instead."})
	StrSSONoEntries        = i18n.MustString("nago.onboarding.sso_no_entries", i18n.Values{language.German: "Trage mindestens eine Adresse oder Domain ein oder wähle „Nur ich“.", language.English: "Enter at least one address or domain or choose “Only me”."})
	StrSSOLaterX           = i18n.MustVarString("nago.onboarding.sso_later_x", i18n.Values{language.German: "Du kannst das später in den Einstellungen unter „{section}“ ändern.", language.English: "You can change this later in the settings under “{section}”."})
	StrSSOContinue         = i18n.MustString("nago.onboarding.sso_continue", i18n.Values{language.German: "Weiter zum Nago-Login", language.English: "Continue to the Nago login"})
	StrBack                = i18n.MustString("nago.onboarding.back", i18n.Values{language.German: "Zurück", language.English: "Back"})
)

// access options in the order of the radio group
const (
	accessOnlyMe = iota
	accessDomain
	accessSelected
	accessEveryone
	accessOptions
)

// methodStep lets the user choose between the single sign-on and a password.
func methodStep(wnd core.Window, flow *Flow) (subtitle string, body core.View) {
	server := serverName(flow.State().SSOServer)
	return StrMethodSubtitle.Get(wnd), ui.VStack(
		methodCard("method-sso", flowbiteOutline.ShieldCheck, StrMethodSSOTitle.Get(wnd), StrRecommended.Get(wnd),
			StrMethodSSOTextX.Get(wnd, i18n.String("server", server)), func() { flow.ChooseMethod(MethodSSO) }),
		methodCard("method-password", flowbiteOutline.Lock, StrMethodPasswordTitle.Get(wnd), "", StrMethodPasswordText.Get(wnd),
			func() { flow.ChooseMethod(MethodPassword) }),
	).Gap(ui.L12).FullWidth()
}

func methodCard(key string, icon core.SVG, title, badge, text string, action func()) core.View {
	// the badge stands above the title, which leaves the title the full width on a small screen
	var heading []core.View
	if badge != "" {
		heading = append(heading, badgeView(badge))
	}

	heading = append(heading, ui.Text(title).Font(ui.TitleMedium))

	return ui.HStack(
		ui.ImageIcon(icon).Frame(ui.Frame{}.Size(ui.L24, ui.L24)),
		ui.VStack(
			append(heading, ui.Text(text).Font(ui.BodySmall))...,
		).Alignment(ui.Leading).Gap(ui.L4).Frame(ui.Frame{}.FullWidth()),
		ui.ImageIcon(flowbiteOutline.ChevronRight).Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
	).Alignment(ui.Center).
		Gap(ui.L16).
		Action(action).
		Key("onboarding", key).
		HoveredBorder(ui.Border{}.Radius(ui.L12).Width(ui.L2).Color(ui.ColorInteractive)).
		FocusedBorder(ui.Border{}.Radius(ui.L12).Width(ui.L2).Color(ui.ColorInteractive)).
		WithPadding(ui.Padding{}.All(ui.L16)).
		FullWidth().
		Border(ui.Border{}.Radius(ui.L12).Width(ui.L2).Color(ui.ColorLine)).
		AccessibilityLabel(title)
}

// ssoStep lets the user decide who may sign in by single sign-on and opens the login service.
func ssoStep(wnd core.Window, flow *Flow) (subtitle string, body core.View, actions []core.View) {
	st := flow.State()
	domain, hasDomain := onboarding.CompanyDomain(st.Email)
	initial := accessOnlyMe
	if hasDomain {
		initial = accessDomain
	}

	group := ui.AutoRadioStateGroup(wnd, "nago.onboarding.sso.access", accessOptions).InitIndex(initial)
	text := core.AutoState[string](wnd)
	errText := core.AutoState[string](wnd)
	selected := group.SelectedIndex()

	entries, invalid := onboarding.ParseAccessEntries(text.Get())
	options := []core.View{
		accessOption(group, accessOnlyMe, StrSSOOnlyMe.Get(wnd), "", string(st.Email), nil),
	}

	if hasDomain {
		options = append(options, accessOption(group, accessDomain, StrSSODomainX.Get(wnd, i18n.String("domain", domain)),
			StrRecommended.Get(wnd), StrSSODomainTextX.Get(wnd, i18n.String("domain", domain)), nil))
	}

	var selectedDetails core.View
	if selected == accessSelected {
		selectedDetails = ui.VStack(
			ui.TextField(StrSSOEntries.Get(wnd), text.Get()).
				InputValue(text).
				Lines(3).
				SupportingText(StrSSOEntriesHint.Get(wnd)).
				ErrorText(errText.Get()).
				ID("onboarding-sso-entries").
				FullWidth(),
			allowedPreview(wnd, st.Email, entries, invalid),
		).Gap(ui.L8).FullWidth()
	}

	options = append(options,
		accessOption(group, accessSelected, StrSSOSelected.Get(wnd), "", StrSSOSelectedText.Get(wnd), selectedDetails),
		accessOption(group, accessEveryone, StrSSOEveryone.Get(wnd), "", StrSSOEveryoneText.Get(wnd), nil),
	)

	body = ui.VStack(
		ui.Text(StrSSOIntro.Get(wnd)).Font(ui.BodySmall),
		ui.VStack(options...).Gap(ui.L8).FullWidth(),
		ui.Text(StrSSOLaterX.Get(wnd, i18n.String("section", user.StrSettingsSSOTitle.Get(wnd)))).Font(ui.BodySmall),
	).Alignment(ui.Leading).Gap(ui.L16).FullWidth()

	actions = []core.View{
		ui.SecondaryButton(func() {
			flow.ChooseMethod(MethodUndecided)
		}).Title(StrBack.Get(wnd)).Key("onboarding", "back"),
		ui.PrimaryButton(func() {
			errText.Set("")
			access := onboarding.SSOAccess{Kind: onboarding.SSOOnlyAdmin}
			switch group.SelectedIndex() {
			case accessDomain:
				access.Kind = onboarding.SSODomain
			case accessEveryone:
				access.Kind = onboarding.SSOEveryone
			case accessSelected:
				entries, invalid := onboarding.ParseAccessEntries(text.Get())
				switch {
				case len(invalid) > 0:
					errText.Set(strings.Join(invalidTexts(wnd, invalid), " "))
					return
				case len(entries) == 0:
					errText.Set(StrSSONoEntries.Get(wnd))
					return
				}

				access = onboarding.SSOAccess{Kind: onboarding.SSOSelected, Entries: entries}
			}

			if err := flow.CompleteWithSSO(access); err != nil {
				alert.ShowBannerError(wnd, err)
			}
		}).Title(StrSSOContinue.Get(wnd)).Key("onboarding", "sso-continue"),
	}

	return StrSSOSubtitle.Get(wnd), body, actions
}

// accessOption is a radio button with a title and a text, which selects itself by a click anywhere. The details are
// shown below the text.
func accessOption(group ui.RadioStateGroup, index int, title, badge, text string, details core.View) core.View {
	state := stateAt(group, index)
	selected := group.SelectedIndex() == index
	// the badge stands above the title, which leaves the title the full width on a small screen
	var heading []core.View
	if badge != "" {
		heading = append(heading, badgeView(badge))
	}

	heading = append(heading, ui.Text(title).Font(ui.TitleMedium))

	border := ui.Border{}.Radius(ui.L12).Width(ui.L2).Color(ui.ColorLine)
	if selected {
		border = border.Color(ui.ColorInteractive)
	}

	content := append(heading, ui.Text(text).Font(ui.BodySmall))

	if details != nil {
		content = append(content, details)
	}

	row := ui.HStack(
		ui.RadioButton(selected).InputChecked(state).ID("onboarding-sso-access-"+string(rune('0'+index))),
		ui.VStack(content...).Alignment(ui.Leading).Gap(ui.L4).FullWidth(),
	).Alignment(ui.TopLeading).
		Gap(ui.L12).
		WithPadding(ui.Padding{}.All(ui.L12)).
		FullWidth()

	if selected {
		// a click into the details, e.g. the text field, must not select again
		return row.Border(border)
	}

	return row.Action(func() {
		group.SetSelectedIndex(index)
	}).HoveredBorder(border.Color(ui.ColorInteractive)).Border(border)
}

// badgeView is a small label like "recommended" next to a title.
func badgeView(text string) core.View {
	return ui.HStack(ui.Text(text).Font(ui.LabelSmall).Color(ui.ColorBlack)).
		BackgroundColor(ui.ColorSemanticGood).
		WithPadding(ui.Padding{}.Horizontal(ui.L8).Vertical(ui.L2)).
		Border(ui.Border{}.Radius(ui.L16))
}

func stateAt(group ui.RadioStateGroup, index int) *core.State[bool] {
	for i, state := range group.All() {
		if i == index {
			return state
		}
	}

	return nil
}

// allowedPreview lists, who may sign in by the entries, so that the user sees, how its text has been understood.
func allowedPreview(wnd core.Window, admin user.Email, entries []onboarding.AccessEntry, invalid []onboarding.InvalidEntry) core.View {
	rows := []core.View{
		ui.Text(StrSSOAllowedTitle.Get(wnd)).Font(ui.TitleSmall),
		allowedRow(flowbiteOutline.Check, ui.ColorSemanticGood, StrSSOYouX.Get(wnd, i18n.String("email", string(admin)))),
	}

	for _, entry := range entries {
		label := string(entry.Email)
		if entry.Domain != "" {
			label = StrSSODomainEntryX.Get(wnd, i18n.String("domain", entry.Domain))
		}

		rows = append(rows, allowedRow(flowbiteOutline.Check, ui.ColorSemanticGood, label))
	}

	for _, problem := range invalidTexts(wnd, invalid) {
		rows = append(rows, allowedRow(flowbiteOutline.ExclamationCircle, ui.ColorSemanticError, problem))
	}

	return ui.VStack(rows...).
		Alignment(ui.Leading).
		Gap(ui.L4).
		WithPadding(ui.Padding{}.All(ui.L12)).
		BackgroundColor(ui.ColorCardFooter).
		FullWidth().
		Border(ui.Border{}.Radius(ui.L8))
}

func allowedRow(icon core.SVG, color ui.Color, text string) core.View {
	return ui.HStack(
		ui.ImageIcon(icon).StrokeColor(color).Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
		ui.Text(text).Font(ui.BodySmall),
	).Alignment(ui.Leading).Gap(ui.L8)
}

// invalidTexts explain the parts of the text, which are no address or domain, and the domains of public providers.
func invalidTexts(wnd core.Window, invalid []onboarding.InvalidEntry) []string {
	var texts, unknown []string
	for _, entry := range invalid {
		if entry.PublicDomain {
			texts = append(texts, StrSSOPublicX.Get(wnd, i18n.String("domain", strings.TrimPrefix(entry.Text, "@"))))
			continue
		}

		unknown = append(unknown, entry.Text)
	}

	if len(unknown) > 0 {
		texts = append([]string{StrSSOInvalidX.Get(wnd, i18n.String("entries", strings.Join(unknown, ", ")))}, texts...)
	}

	return texts
}

// serverName returns the host of the login service, which people recognize.
func serverName(server string) string {
	if u, err := url.Parse(server); err == nil && u.Host != "" {
		return u.Host
	}

	return server
}
