package uiuser

import (
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func contact(
	userSettings user.Settings,
	firstname, errLastname *core.State[string],
	lastname, errFirstname *core.State[string],
	title, errTitle *core.State[string],
	position, errPosition *core.State[string],
	companyName, errCompanyName *core.State[string],
	city, errCity *core.State[string],
	postalCode, errPostalCode *core.State[string],
	state, errState *core.State[string],
	country, errCountry *core.State[string],
	professionalGroup, errProfessionalGroup *core.State[string],
) core.View {
	return ui.VStack(
		ui.Space(ui.L48),
		ui.Space(ui.L8), // -8 due to gap
		ui.IfFunc(!userSettings.Title.Hidden(), func() core.View {
			return ui.TextField("Titel"+requiredChar(userSettings.Title), title.Get()).
				ErrorText(errTitle.Get()).
				InputValue(title).
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

		ui.IfFunc(!userSettings.CompanyName.Hidden(), func() core.View {
			return ui.TextField("Position"+requiredChar(userSettings.Position), position.Get()).
				ErrorText(errPosition.Get()).
				InputValue(position).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.CompanyName.Hidden(), func() core.View {
			return ui.TextField("Unternehmen"+requiredChar(userSettings.CompanyName), companyName.Get()).
				ErrorText(errCompanyName.Get()).
				InputValue(companyName).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.State.Hidden(), func() core.View {
			return ui.TextField("Postleitzahl"+requiredChar(userSettings.PostalCode), postalCode.Get()).
				ErrorText(errPostalCode.Get()).
				InputValue(postalCode).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.State.Hidden(), func() core.View {
			return ui.TextField("Ort"+requiredChar(userSettings.City), city.Get()).
				ErrorText(errCity.Get()).
				InputValue(city).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.State.Hidden(), func() core.View {
			return ui.TextField("Bundesland"+requiredChar(userSettings.State), state.Get()).
				ErrorText(errState.Get()).
				InputValue(state).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.State.Hidden(), func() core.View {
			return ui.TextField("Land"+requiredChar(userSettings.Country), country.Get()).
				ErrorText(errCountry.Get()).
				InputValue(country).
				FullWidth()
		}),

		ui.IfFunc(!userSettings.State.Hidden(), func() core.View {
			return ui.TextField("Berufsgruppe"+requiredChar(userSettings.ProfessionalGroup), professionalGroup.Get()).
				ErrorText(errProfessionalGroup.Get()).
				InputValue(professionalGroup).
				FullWidth()
		}),
	).FullWidth().Gap(ui.L8)
}

func validateContact(
	userSettings user.Settings,
	firstname, errLastname *core.State[string],
	lastname, errFirstname *core.State[string],
	title, errTitle *core.State[string],
	position, errPosition *core.State[string],
	companyName, errCompanyName *core.State[string],
	cityName, errCityName *core.State[string],
	postalCode, errPostalCode *core.State[string],
	state, errState *core.State[string],
	country, errCountry *core.State[string],
	professionalGroup, errProfessionalGroup *core.State[string],
) bool {
	errFirstname.Set("")
	errLastname.Set("")
	errTitle.Set("")
	errPosition.Set("")
	errCompanyName.Set("")
	errCityName.Set("")
	errPostalCode.Set("")
	errCountry.Set("")
	errState.Set("")
	errProfessionalGroup.Set("")
	anyError := false

	if firstname.Get() == "" {
		errFirstname.Set("Bitte geben Sie einen Vornamen ein.")
		anyError = true
	}

	if lastname.Get() == "" {
		errLastname.Set("Bitte einen Nachnamen ein.")
		anyError = true
	}

	if errFirstname.Get() == "" && errLastname.Get() == "" && len(firstname.Get())+len(lastname.Get()) < 3 {
		errLastname.Set("Wurde der Name richtig eingegeben?")
		anyError = true
	}

	// other things
	if !userSettings.Title.Match(title.Get()) {
		errTitle.Set("Bitte Titel eingeben.")
		anyError = true
	}

	if !userSettings.Position.Match(position.Get()) {
		errPosition.Set("Bitte Position eingeben.")
		anyError = true
	}

	if !userSettings.CompanyName.Match(companyName.Get()) {
		errCompanyName.Set("Bitte Unternehmen eingeben.")
		anyError = true
	}

	if !userSettings.City.Match(cityName.Get()) {
		errCityName.Set("Bitte Stadt eingeben.")
		anyError = true
	}

	if !userSettings.PostalCode.Match(postalCode.Get()) {
		errPostalCode.Set("Bitte Postleitzahl eingeben.")
		anyError = true
	}

	if !userSettings.State.Match(state.Get()) {
		errState.Set("Bitte Bundesland eingeben.")
		anyError = true
	}

	if !userSettings.Country.Match(country.Get()) {
		errCountry.Set("Bitte Land eingeben.")
		anyError = true
	}

	if !userSettings.ProfessionalGroup.Match(professionalGroup.Get()) {
		errProfessionalGroup.Set("Bitte Berufsgruppe eingeben.")
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
