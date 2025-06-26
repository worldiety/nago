// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiuser

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/xreflect"
	"go.wdy.de/nago/pkg/xstrings"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func contact(
	userSettings user.Settings,
	firstname, errLastname *core.State[string],
	lastname, errFirstname *core.State[string],
	salutation, errSalutation *core.State[string],
	title, errTitle *core.State[string],
	position, errPosition *core.State[string],
	companyName, errCompanyName *core.State[string],
	city, errCity *core.State[string],
	postalCode, errPostalCode *core.State[string],
	state, errState *core.State[string],
	country, errCountry *core.State[string],
	professionalGroup, errProfessionalGroup *core.State[string],
	mobilePhone, errMobilePhone *core.State[string],
) core.View {
	return ui.VStack(
		ui.Space(ui.L48),
		ui.Space(ui.L8), // -8 due to gap

		ui.IfFunc(!userSettings.Salutation.Hidden(), func() core.View {
			return ui.TextField("Anrede"+requiredChar(userSettings.Salutation), salutation.Get()).
				ErrorText(errSalutation.Get()).
				InputValue(salutation).
				SupportingText(supportingTextSalutation()).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.Title.Hidden(), func() core.View {
			return ui.TextField("Titel"+requiredChar(userSettings.Title), title.Get()).
				ErrorText(errTitle.Get()).
				InputValue(title).
				SupportingText(supportingTextTitle()).
				FullWidth()
		}),

		// firstname is always required by nago
		ui.TextField("Vorname*", firstname.Get()).
			ErrorText(errFirstname.Get()).
			InputValue(firstname).
			FullWidth(),

		// lastname is always required by nago
		ui.TextField("Nachname*", lastname.Get()).
			ErrorText(errLastname.Get()).
			InputValue(lastname).
			FullWidth(),

		ui.IfFunc(!userSettings.Position.Hidden(), func() core.View {
			return ui.TextField("Position"+requiredChar(userSettings.Position), position.Get()).
				ErrorText(errPosition.Get()).
				InputValue(position).
				SupportingText(supportingTextPosition()).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.CompanyName.Hidden(), func() core.View {
			return ui.TextField("Unternehmen"+requiredChar(userSettings.CompanyName), companyName.Get()).
				ErrorText(errCompanyName.Get()).
				InputValue(companyName).
				SupportingText(supportingTextCompanyName()).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.PostalCode.Hidden(), func() core.View {
			return ui.TextField("Postleitzahl"+requiredChar(userSettings.PostalCode), postalCode.Get()).
				ErrorText(errPostalCode.Get()).
				InputValue(postalCode).
				SupportingText(supportingTextPostalCode()).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.City.Hidden(), func() core.View {
			return ui.TextField("Ort"+requiredChar(userSettings.City), city.Get()).
				ErrorText(errCity.Get()).
				InputValue(city).
				SupportingText(supportingTextCity()).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.State.Hidden(), func() core.View {
			return ui.TextField("Bundesland"+requiredChar(userSettings.State), state.Get()).
				ErrorText(errState.Get()).
				InputValue(state).
				SupportingText(supportingTextState()).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.Country.Hidden(), func() core.View {
			return ui.TextField("Land"+requiredChar(userSettings.Country), country.Get()).
				ErrorText(errCountry.Get()).
				InputValue(country).
				SupportingText(supportingTextCountry()).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.ProfessionalGroup.Hidden(), func() core.View {
			return ui.TextField("Berufsgruppe"+requiredChar(userSettings.ProfessionalGroup), professionalGroup.Get()).
				ErrorText(errProfessionalGroup.Get()).
				InputValue(professionalGroup).
				SupportingText(supportingTextProfessionalGroup()).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.MobilePhone.Hidden(), func() core.View {
			return ui.TextField("Mobil"+requiredChar(userSettings.MobilePhone), mobilePhone.Get()).
				ErrorText(errMobilePhone.Get()).
				InputValue(mobilePhone).
				SupportingText(supportingTextMobilePhone()).
				FullWidth()
		}),
	).FullWidth().Gap(ui.L8)
}

func supportingTextSalutation() string {
	return xreflect.FieldTagFor[user.Settings]("Salutation", "supportingText")
}

func supportingTextTitle() string {
	return xreflect.FieldTagFor[user.Settings]("Title", "supportingText")
}

func supportingTextPosition() string {
	return xreflect.FieldTagFor[user.Settings]("Position", "supportingText")
}

func supportingTextCompanyName() string {
	return xreflect.FieldTagFor[user.Settings]("CompanyName", "supportingText")
}

func supportingTextCity() string {
	return xreflect.FieldTagFor[user.Settings]("City", "supportingText")
}

func supportingTextPostalCode() string {
	return xreflect.FieldTagFor[user.Settings]("PostalCode", "supportingText")
}

func supportingTextState() string {
	return xreflect.FieldTagFor[user.Settings]("State", "supportingText")
}

func supportingTextCountry() string {
	return xreflect.FieldTagFor[user.Settings]("Country", "supportingText")
}

func supportingTextProfessionalGroup() string {
	return xreflect.FieldTagFor[user.Settings]("ProfessionalGroup", "supportingText")
}

func supportingTextMobilePhone() string {
	return xreflect.FieldTagFor[user.Settings]("MobilePhone", "supportingText")
}

func validateContact(
	userSettings user.Settings,
	firstname, errLastname *core.State[string],
	lastname, errFirstname *core.State[string],
	salutation, errSalutation *core.State[string],
	title, errTitle *core.State[string],
	position, errPosition *core.State[string],
	companyName, errCompanyName *core.State[string],
	cityName, errCityName *core.State[string],
	postalCode, errPostalCode *core.State[string],
	state, errState *core.State[string],
	country, errCountry *core.State[string],
	professionalGroup, errProfessionalGroup *core.State[string],
	mobile, errMobile *core.State[string],
) bool {
	errFirstname.Set("")
	errLastname.Set("")
	errSalutation.Set("")
	errTitle.Set("")
	errPosition.Set("")
	errCompanyName.Set("")
	errCityName.Set("")
	errPostalCode.Set("")
	errCountry.Set("")
	errState.Set("")
	errProfessionalGroup.Set("")
	errMobile.Set("")
	anyError := false

	if firstname.Get() == "" {
		errFirstname.Set("Bitte einen Vornamen eingeben.")
		anyError = true
	}

	if lastname.Get() == "" {
		errLastname.Set("Bitte einen Nachnamen eingeben.")
		anyError = true
	}

	if errFirstname.Get() == "" && errLastname.Get() == "" && len(firstname.Get())+len(lastname.Get()) < 3 {
		errLastname.Set("Wurde der Name richtig eingegeben?")
		anyError = true
	}

	// other things
	if !userSettings.Title.Match(title.Get()) {
		errTitle.Set(xstrings.Space("Bitte den Titel eingeben.", supportingTextTitle()))
		anyError = true
	}

	if !userSettings.Salutation.Match(salutation.Get()) {
		errSalutation.Set(xstrings.Space("Bitte die Anrede eingeben.", supportingTextSalutation()))
		anyError = true
	}

	if !userSettings.Position.Match(position.Get()) {
		errPosition.Set(xstrings.Space("Bitte die Position eingeben.", supportingTextPosition()))
		anyError = true
	}

	if !userSettings.CompanyName.Match(companyName.Get()) {
		errCompanyName.Set(xstrings.Space("Bitte das Unternehmen eingeben.", supportingTextCompanyName()))
		anyError = true
	}

	if !userSettings.City.Match(cityName.Get()) {
		errCityName.Set(xstrings.Space("Bitte die Stadt eingeben.", supportingTextCity()))
		anyError = true
	}

	if !userSettings.PostalCode.Match(postalCode.Get()) {
		errPostalCode.Set(xstrings.Space("Bitte die Postleitzahl eingeben.", supportingTextPostalCode()))
		anyError = true
	}

	if !userSettings.State.Match(state.Get()) {
		errState.Set(xstrings.Space("Bitte das Bundesland eingeben.", supportingTextState()))
		anyError = true
	}

	if !userSettings.Country.Match(country.Get()) {
		errCountry.Set(xstrings.Space("Bitte das Land eingeben.", supportingTextCountry()))
		anyError = true
	}

	if !userSettings.ProfessionalGroup.Match(professionalGroup.Get()) {
		errProfessionalGroup.Set(xstrings.Space("Bitte die Berufsgruppe eingeben.", supportingTextProfessionalGroup()))
		anyError = true
	}

	if !userSettings.MobilePhone.Match(mobile.Get()) {
		errMobile.Set(xstrings.Space("Bitte die mobile Telefonnummer eingeben.", supportingTextMobilePhone()))
		anyError = true
	}

	return !anyError
}

func requiredChar(t user.FieldConstraint) string {
	if !t.Optional() {
		return "*"
	}

	return ""
}
