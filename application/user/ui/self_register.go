// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/footer"
	"time"
)

func PageSelfRegister(wnd core.Window, hasMail user.EMailUsed, createUser user.Create) core.View {
	isDesktop := wnd.Info().SizeClass > core.SizeClassSmall

	userSettings := core.GlobalSettings[user.Settings](wnd)
	_ = userSettings

	themeSettings := core.GlobalSettings[theme.Settings](wnd)

	registerPageCurrent := core.AutoState[registerPage](wnd)

	// contact
	firstname := core.AutoState[string](wnd)
	errFirstname := core.AutoState[string](wnd)

	lastname := core.AutoState[string](wnd)
	errLastname := core.AutoState[string](wnd)

	title := core.AutoState[string](wnd)
	errTitle := core.AutoState[string](wnd)

	position := core.AutoState[string](wnd)
	errPosition := core.AutoState[string](wnd)

	companyName := core.AutoState[string](wnd)
	errCompanyName := core.AutoState[string](wnd)

	city := core.AutoState[string](wnd)
	errCity := core.AutoState[string](wnd)

	postalCode := core.AutoState[string](wnd)
	errPostalCode := core.AutoState[string](wnd)

	state := core.AutoState[string](wnd)
	errState := core.AutoState[string](wnd)

	country := core.AutoState[string](wnd)
	errCountry := core.AutoState[string](wnd)

	professionalGroup := core.AutoState[string](wnd)
	errProfessionalGroup := core.AutoState[string](wnd)

	// password
	password := core.AutoState[string](wnd)
	passwordRepeated := core.AutoState[string](wnd)
	errPasswordRepeated := core.AutoState[string](wnd)

	// legal stuff
	consentStates := map[consent.ID]*core.State[bool]{}
	for _, consentOpt := range userSettings.Consents {
		consentStates[consentOpt.ID] = core.StateOf[bool](wnd, string(consentOpt.ID))
		consentStates[consentOpt.ID+"err"] = core.StateOf[bool](wnd, string(consentOpt.ID+"err"))
	}

	// email
	email := core.AutoState[string](wnd)
	emailRepeated := core.AutoState[string](wnd)
	errEmailRepeated := core.AutoState[string](wnd)

	// mobile
	mobile := core.AutoState[string](wnd)
	errMobile := core.AutoState[string](wnd)

	// register
	regErr := core.AutoState[error](wnd)

	var subcaption string
	var pageBody core.View
	nextCaption := "weiter"
	nextVisible := true
	switch registerPageCurrent.Get() {
	case registerPageNames:
		subcaption = "Bitte den Kontakt eingeben"
		pageBody = contact(
			userSettings,
			firstname, errFirstname,
			lastname, errLastname,
			title, errTitle,
			position, errPosition,
			companyName, errCompanyName,
			city, errCity,
			postalCode, errPostalCode,
			state, errState,
			country, errCountry,
			professionalGroup, errProfessionalGroup,
			mobile, errMobile,
		)
	case registerPasswords:
		subcaption = "Bitte Passwort vergeben"
		pageBody = passwords(password, passwordRepeated, errPasswordRepeated)
	case registerAdoptAny:
		subcaption = "Bitte zustimmen"
		pageBody = consents(wnd, userSettings, consentStates)
	case registerMails:
		subcaption = "Bitte die E-Mail eingeben"
		pageBody = emails(email, emailRepeated, errEmailRepeated)
	case registerCheck:
		subcaption = "Fast geschafft..."
		pageBody = check(firstname, lastname, email)
		nextCaption = "Registrieren"
	case registerRes:
		subcaption = "Konto verifizieren"
		pageBody = registerResult(regErr.Get())
		nextCaption = "Fertig"
		nextVisible = false
	}

	var content core.View
	var cardFrame ui.Frame
	if isDesktop {
		cardFrame = ui.Frame{}.MatchScreen()
		content = ui.Grid(
			ui.GridCell(ui.VStack(
				ui.If(themeSettings.AppIconLight != "" || themeSettings.AppIconDark != "",
					ui.Image().
						Adaptive(themeSettings.AppIconLight, themeSettings.AppIconDark).
						Frame(ui.Frame{}.Size(ui.L48, ui.L48)),
				),
				ui.Space(ui.L16),
				ui.Text(wnd.Application().Name()+"-Konto").Font(ui.Title),
				ui.Text("erstellen").Font(ui.Title),
				ui.Text(subcaption),
			).Alignment(ui.TopLeading)),

			ui.GridCell(pageBody),
		).Gap(ui.L16).Rows(1).FullWidth()
	} else {
		cardFrame = ui.Frame{}.MatchScreen()
		content = ui.VStack(
			ui.If(themeSettings.AppIconLight != "" || themeSettings.AppIconDark != "",
				ui.Image().
					Adaptive(themeSettings.AppIconLight, themeSettings.AppIconDark).
					Frame(ui.Frame{}.Size(ui.L48, ui.L48)),
			),

			ui.Space(ui.L16),
			ui.Text(wnd.Application().Name()+"-Konto").Font(ui.Title),
			ui.Text("erstellen").Font(ui.Title),
			ui.Text(subcaption),
			pageBody,
		).FullWidth().Alignment(ui.TopLeading)
	}

	cfgTheme := core.GlobalSettings[theme.Settings](wnd)
	hasFooter := cfgTheme.ProviderName != "" || cfgTheme.Impress != "" || cfgTheme.GeneralTermsAndConditions != "" || cfgTheme.PrivacyPolicy != ""

	return ui.VStack( //scaffold replacement
		alert.BannerMessages(wnd),
		ui.WindowTitle("Konto erstellen"),
		ui.Spacer(),
		cardlayout.Card("").Body(
			ui.VStack(content).Padding(ui.Padding{}.All(ui.L16)),
		).Padding(ui.Padding{}.All(ui.L40)).
			Frame(ui.Frame{MaxWidth: ui.L880}.FullWidth()).
			Footer(ui.HStack(
				ui.SecondaryButton(func() {
					registerPageCurrent.Set(registerPageCurrent.Get() - 1)
					if !requiresAnyAdoption(userSettings) && registerPageCurrent.Get() == registerAdoptAny {
						registerPageCurrent.Set(registerPageCurrent.Get() - 1)
					}
				}).Visible((registerPageCurrent.Get() > 0 && registerPageCurrent.Get() < registerRes) || regErr.Get() != nil).Title("Zurück"),
				ui.PrimaryButton(func() {
					switch registerPageCurrent.Get() {
					case registerPageNames:
						if validateContact(
							userSettings,
							firstname, errFirstname,
							lastname, errLastname,
							title, errTitle,
							position, errPosition,
							companyName, errCompanyName,
							city, errCity,
							postalCode, errPostalCode,
							state, errState,
							country, errCountry,
							professionalGroup, errProfessionalGroup,
							mobile, errMobile,
						) {
							registerPageCurrent.Set(registerPageCurrent.Get() + 1)
						}
					case registerPasswords:
						strength := validatePasswords(password, passwordRepeated, errPasswordRepeated)
						if strength.Acceptable {
							if requiresAnyAdoption(userSettings) {
								registerPageCurrent.Set(registerPageCurrent.Get() + 1)
							} else {
								registerPageCurrent.Set(registerPageCurrent.Get() + 2)
							}
						}
					case registerAdoptAny:
						if validateConsents(userSettings, consentStates) {
							registerPageCurrent.Set(registerPageCurrent.Get() + 1)
						}

					case registerMails:
						if validateEmails(hasMail, email, emailRepeated, errEmailRepeated) {
							registerPageCurrent.Set(registerPageCurrent.Get() + 1)
						}

					case registerCheck:
						var myConsents []consent.Consent
						for _, option := range userSettings.Consents {
							status := consent.Revoked
							if consentStates[option.ID].Get() {
								status = consent.Approved
							}

							myConsents = append(myConsents, consent.Consent{
								ID:      option.ID,
								History: []consent.Action{{At: time.Now(), Status: status}},
							})
						}

						_, err := createUser(user.SU(), user.ShortRegistrationUser{
							SelfRegistered:    true,
							Firstname:         firstname.Get(),
							Lastname:          lastname.Get(),
							Email:             user.Email(email.Get()),
							Password:          user.Password(password.Get()),
							PasswordRepeated:  user.Password(passwordRepeated.Get()),
							NotifyUser:        true,
							Verified:          false, // important, keep it always false
							Consents:          myConsents,
							Title:             title.Get(),
							Position:          position.Get(),
							CompanyName:       companyName.Get(),
							City:              city.Get(),
							PostalCode:        postalCode.Get(),
							State:             state.Get(),
							Country:           country.Get(),
							ProfessionalGroup: professionalGroup.Get(),
							MobilePhone:       mobile.Get(),
						})

						regErr.Set(err)
						registerPageCurrent.Set(registerPageCurrent.Get() + 1)
					}
				}).Title(nextCaption).Enabled(registerPageCurrent.Get() != registerRes).Visible(nextVisible),
			).Gap(ui.L8)),

		ui.Spacer(),
		ui.IfFunc(hasFooter, func() core.View {
			return footer.Footer().
				ProviderName(cfgTheme.ProviderName).
				Impress(cfgTheme.Impress).
				PrivacyPolicy(cfgTheme.PrivacyPolicy).
				TermsOfUse(cfgTheme.TermsOfUse).
				Logo(ui.Image().Adaptive(cfgTheme.PageLogoLight, cfgTheme.PageLogoDark)).
				GeneralTermsAndConditions(cfgTheme.GeneralTermsAndConditions).
				Slogan(cfgTheme.Slogan)
		}),
	).Frame(cardFrame)
}

func acceptedAt(b bool) time.Time {
	if b {
		return time.Now()
	}

	return time.Time{}
}

type registerPage int

const (
	registerPageNames = 0
	registerMails     = 1
	registerPasswords = 2
	registerAdoptAny  = 3
	registerCheck     = 4
	registerRes       = 5
)

func requiresAnyAdoption(s user.Settings) bool {
	return len(s.Consents) > 0
}
