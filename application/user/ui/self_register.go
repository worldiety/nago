// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"errors"
	"log/slog"
	"time"

	"go.wdy.de/nago/application/consent"
	"go.wdy.de/nago/application/session"
	uisession "go.wdy.de/nago/application/session/ui"
	"go.wdy.de/nago/application/theme"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/footer"
	"go.wdy.de/nago/presentation/ui/form"
	"go.wdy.de/nago/presentation/ui/stepper"
)

// PageSelfRegister shows the self registration. Without a link, the visitor enters the mail address and gets a link
// by mail. The link opens the rest of the registration with the confirmed address. The account is created at the
// end and the window is logged in, if loginUser is not nil.
func PageSelfRegister(wnd core.Window, registration user.Registration, loginPath core.NavigationPath, loginUser session.LoginUser, subjectFromUser user.SubjectFromUser) core.View {
	themeSettings := core.GlobalSettings[theme.Settings](wnd)

	id := user.RegistrationID(wnd.Values()["id"])
	code := wnd.Values()["code"]

	var subcaption string
	var body core.View
	var actions []core.View

	// the steps are shown beside the content, which needs a wider card
	wide := true
	if id == "" {
		subcaption, body, actions = newRequestRegistration(wnd).view(registration.Request)
		body = besideProgress(wnd, registerProgress(), body)
	} else {
		subcaption, body, actions, wide = completeRegistration(wnd, registration, id, code, loginPath, loginUser, subjectFromUser)
	}

	cardWidth := ui.L480
	if wide && wnd.Info().SizeClass >= core.SizeClassMedium {
		cardWidth = ui.L880
	}

	hasFooter := themeSettings.ProviderName != "" || themeSettings.Impress != "" || themeSettings.GeneralTermsAndConditions != "" || themeSettings.PrivacyPolicy != ""

	return ui.VStack( //scaffold replacement
		ui.VStack(
			alert.BannerMessages(wnd),
			ui.WindowTitle("Konto erstellen"),
			uisession.LoginRegisterCardWithWidth(wnd, cardWidth, themeSettings.PageLogoLight, themeSettings.PageLogoDark, wnd.Application().Name()+"-Konto erstellen", subcaption, body, actions...),
			ui.TextLayout(
				ui.Text("Bereits registriert? Hier "),
				ui.LinkWithAction("anmelden!", func() {
					wnd.Navigation().ForwardTo(loginPath, nil)
				}),
			).Font(ui.BodySmall),

			ui.Spacer(),
			ui.IfFunc(hasFooter, func() core.View {
				return footer.Footer().
					ProviderName(themeSettings.ProviderName).
					Impress(themeSettings.Impress).
					PrivacyPolicy(themeSettings.PrivacyPolicy).
					TermsOfUse(themeSettings.TermsOfUse).
					Logo(ui.Image().Adaptive(themeSettings.PageLogoLight, themeSettings.PageLogoDark).ObjectFit(ui.FitContain).Frame(ui.Frame{Height: ui.L64})).
					GeneralTermsAndConditions(themeSettings.GeneralTermsAndConditions).
					Slogan(themeSettings.Slogan)
			}),
		).Gap(ui.L16).FullWidth(),
	).Frame(ui.Frame{}.MatchScreen())
}

type registerPage int

const (
	registerMail registerPage = iota
	registerPageNames
	registerPasswords
	registerCheck
)

func requiresAnyAdoption(s user.Settings) bool {
	return len(s.Consents) > 0
}

// localizedMessage returns the description of a localized error.
func localizedMessage(err error) (string, bool) {
	var lerr std.LocalizedError
	if errors.As(err, &lerr) {
		return lerr.Description(), true
	}

	return "", false
}

// invalidLink shows, that the link cannot be used, and offers to request a new one.
func invalidLink(wnd core.Window, registerPath core.NavigationPath) (string, core.View, []core.View) {
	body := ui.Text("Der Link ist ungültig, abgelaufen oder wurde bereits verwendet. Fordern Sie einfach einen neuen Link an.")
	actions := []core.View{
		ui.PrimaryButton(func() {
			wnd.Navigation().ResetTo(registerPath, nil)
		}).Title("Neuen Link anfordern"),
	}

	return "Link nicht gültig", body, actions
}

// registrationLink is the result of checking the link of the mail.
type registrationLink struct {
	email string
	valid bool
}

// completeRegistration is the rest of the registration, after the visitor opened the link from the mail.
// The last result tells, if the content needs a wide card.
func completeRegistration(wnd core.Window, registration user.Registration, id user.RegistrationID, code string, loginPath core.NavigationPath, loginUser session.LoginUser, subjectFromUser user.SubjectFromUser) (string, core.View, []core.View, bool) {
	userSettings := core.GlobalSettings[user.Settings](wnd)
	registerPath := wnd.Path()

	// the link is checked once per window, because every wrong check counts as attempt
	link := core.AutoState[registrationLink](wnd).Init(func() registrationLink {
		email, err := registration.Find(id, code)
		if err != nil && !errors.Is(err, user.ErrRegistrationInvalid) {
			slog.Error("cannot check registration link", "err", err)
		}

		return registrationLink{email: string(email), valid: err == nil}
	})

	if !link.Get().valid {
		subcaption, body, actions := invalidLink(wnd, registerPath)
		return subcaption, body, actions, false
	}

	linkEmail := core.AutoState[string](wnd).Init(func() string { return link.Get().email })
	stepIdx := core.AutoState[int](wnd).Init(func() int { return 1 })
	done := core.AutoState[bool](wnd)

	// contact
	firstname := core.AutoState[string](wnd)
	errFirstname := core.AutoState[string](wnd)

	lastname := core.AutoState[string](wnd)
	errLastname := core.AutoState[string](wnd)

	salutation := core.AutoState[string](wnd)
	errSalutation := core.AutoState[string](wnd)

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

	// mobile
	mobile := core.AutoState[string](wnd)
	errMobile := core.AutoState[string](wnd)

	// register
	regErr := core.AutoState[error](wnd)
	linkUsedUp := core.AutoState[bool](wnd)

	if linkUsedUp.Get() {
		subcaption, body, actions := invalidLink(wnd, registerPath)
		return subcaption, body, actions, false
	}

	if done.Get() {
		body := ui.Text("Ihr Konto wurde erstellt. Sie können sich jetzt anmelden.")
		return "Konto erstellt", body, []core.View{
			ui.PrimaryButton(func() {
				wnd.Navigation().ResetTo(loginPath, nil)
			}).Title("Jetzt anmelden"),
		}, false
	}

	submit := func() {
		regErr.Set(nil)
		if !validateConsents(userSettings, consentStates) {
			return
		}

		var myConsents []consent.Consent
		for _, option := range userSettings.Consents {
			status := consent.Revoked
			if consentStates[option.ID].Get() {
				status = consent.Approved
			}

			myConsents = append(myConsents, consent.Consent{
				ID:      option.ID,
				History: []consent.Action{{At: time.Now(), Status: status, Location: string(wnd.Path())}},
			})
		}

		usr, err := registration.Complete(id, code, user.ShortRegistrationUser{
			Firstname:         firstname.Get(),
			Lastname:          lastname.Get(),
			Password:          user.Password(password.Get()),
			PasswordRepeated:  user.Password(passwordRepeated.Get()),
			Consents:          myConsents,
			Salutation:        salutation.Get(),
			Title:             title.Get(),
			Position:          position.Get(),
			CompanyName:       companyName.Get(),
			City:              city.Get(),
			PostalCode:        postalCode.Get(),
			State:             state.Get(),
			Country:           country.Get(),
			ProfessionalGroup: professionalGroup.Get(),
			MobilePhone:       mobile.Get(),
			PreferredLanguage: wnd.Locale(),
		})

		switch {
		case errors.Is(err, user.ErrRegistrationInvalid):
			linkUsedUp.Set(true)
			return
		case err != nil && usr.ID == "":
			regErr.Set(err)
			return
		case err != nil:
			// the account exists, the visitor can still log in by password
			slog.Error("registration: account created with an error", "err", err)
		}

		if loginUser != nil && err == nil && login(wnd, loginUser, subjectFromUser, usr.ID) {
			wnd.Navigation().ResetTo(".", nil)
			return
		}

		done.Set(true)
	}

	// the mail address is the first step, which is already done by opening the link
	steps := []registerStep{
		{key: registerMail, headline: registerMail.headline(), subcaption: "E-Mail bestätigt", body: ui.Text("Ihre E-Mail-Adresse " + linkEmail.Get() + " ist bestätigt.")},
		{key: registerPageNames, headline: registerPageNames.headline(), subcaption: "Bitte den Kontakt eingeben", body: contact(
			userSettings,
			firstname, errFirstname,
			lastname, errLastname,
			salutation, errSalutation,
			title, errTitle,
			position, errPosition,
			companyName, errCompanyName,
			city, errCity,
			postalCode, errPostalCode,
			state, errState,
			country, errCountry,
			professionalGroup, errProfessionalGroup,
			mobile, errMobile,
		)},
		{key: registerPasswords, headline: registerPasswords.headline(), subcaption: "Bitte Passwort vergeben", body: passwords(wnd, password, passwordRepeated, errPasswordRepeated)},
	}

	// the last step shows the data and asks for the consents right below, before the account is created
	steps = append(steps, registerStep{key: registerCheck, headline: registerCheck.headline(), subcaption: "Fast geschafft...", body: ui.VStack(
		ui.If(regErr.Get() != nil, alert.BannerError(regErr.Get())),
		check(firstname, lastname, linkEmail),
		ui.If(requiresAnyAdoption(userSettings), consents(wnd, userSettings, consentStates)),
	).FullWidth().Gap(ui.L16)})

	// validate the step, which is left forwards
	validate := func(key registerPage) bool {
		switch key {
		case registerPageNames:
			return validateContact(
				userSettings,
				firstname, errFirstname,
				lastname, errLastname,
				salutation, errSalutation,
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
			return validatePasswords(password, passwordRepeated, errPasswordRepeated).Acceptable
		default:
			return true
		}
	}

	// the subcaption needs the current step before the MultiSteps clamps it the same way
	current := min(max(stepIdx.Get(), 1), len(steps)-1)

	body := form.MultiSteps(ui.ForEach(steps, func(s registerStep) form.TStep {
		return form.Step(s.body).Headline(s.headline)
	})...).
		InputValue(stepIdx).
		Layout(stepper.StepperLayoutSimpleList).
		StepperPosition(form.StepperPositionLeading).
		CanShow(func(currentIdx int, wantedIndex int) bool {
			// the mail step is done and cannot be visited again
			return wantedIndex > 0
		}).
		OnStepChange(func(from, to int) bool {
			regErr.Set(nil)
			return to < from || validate(steps[from].key)
		}).
		ButtonDone(ui.PrimaryButton(submit).Title("Konto erstellen")).
		Frame(ui.Frame{}.FullWidth())

	return steps[current].subcaption, body, nil, true
}

// registerPages are all steps of the registration, the mail address is the first one.
var registerPages = []registerPage{registerMail, registerPageNames, registerPasswords, registerCheck}

func (p registerPage) headline() string {
	switch p {
	case registerMail:
		return "E-Mail"
	case registerPageNames:
		return "Kontakt"
	case registerPasswords:
		return "Passwort"
	default:
		return "Prüfen"
	}
}

// registerProgress shows the steps of the registration with the mail address as the current one, in the same layout
// as the form after the link.
func registerProgress() core.View {
	return stepper.Stepper(ui.ForEach(registerPages, func(p registerPage) stepper.TStep {
		return stepper.Step().Headline(p.headline())
	})...).Layout(stepper.StepperLayoutSimpleList).Value(0)
}

// besideProgress places the steps left beside the content, like [form.StepperPositionLeading] does. Windows of the small
// size class show the steps above the content.
func besideProgress(wnd core.Window, progress, content core.View) core.View {
	if wnd.Info().SizeClass >= core.SizeClassMedium {
		return ui.HStack(
			// the steps keep the width of their labels, the content takes the rest
			ui.VStack(progress).Alignment(ui.TopLeading).NoClip(true).Frame(ui.Frame{}.PreventShrink()),
			content,
		).Gap(ui.L32).Alignment(ui.TopLeading).FullWidth()
	}

	return ui.VStack(
		ui.HStack(progress).FullWidth(),
		content,
	).Gap(ui.L16).FullWidth()
}

type registerStep struct {
	key        registerPage
	headline   string
	subcaption string
	body       core.View
}

// login logs the session and the window in.
func login(wnd core.Window, loginUser session.LoginUser, subjectFromUser user.SubjectFromUser, uid user.ID) bool {
	if err := loginUser(wnd.Session().ID(), uid); err != nil {
		slog.Error("registration: cannot log in the new account", "err", err)
		return false
	}

	if subjectFromUser != nil {
		optSubject, err := subjectFromUser(user.SU(), uid)
		if err != nil {
			slog.Error("registration: cannot resolve the subject of the new account", "err", err)
			return false
		}

		if optSubject.IsSome() {
			wnd.UpdateSubject(optSubject.Unwrap())
		}
	}

	return true
}
