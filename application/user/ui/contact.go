package uiuser

import (
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/image"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/crud"
)

type contactViewModel struct {
	Avatar image.ID `style:"avatar" section:"Profilbild"`
	// AcademicDegree is e.g. Diploma, Bachelor, Master or Doctor
	AcademicDegree string `label:"Akademischer Grad" section:"Daten"`
	// OfficialTitle is like Professor, Oberbürgermeister etc.
	OfficialTitle     string `label:"Amtsbezeichnung" section:"Daten"`
	Salutation        string `label:"Anrede" section:"Daten"`
	Firstname         string `label:"Vorname" section:"Daten"`
	Lastname          string `label:"Nachname" section:"Daten"`
	PreferredLanguage string `label:"Sprache" section:"Daten" supportingText:"Präferierte Sprache in BCP47 Kodierung, z.B. de_DE oder en_US oder en_GB."`
	EMail             string `label:"E-Mail Adresse" section:"Daten" disabled:"true" supportingText:"Die E-Mail Adresse kann hier nicht geändert werden, da sie Bestandteil der Identität ist."`

	// job
	Position    string `label:"Position" section:"Beruf"`
	CompanyName string `label:"Firma" section:"Beruf"`
	PostalCode  string `label:"Postleitzahl" section:"Beruf"`
	City        string `label:"Ort" section:"Beruf"`
	Country     string `label:"Land" section:"Beruf"`

	// contact
	Phone       string `label:"Telefon" section:"Kontakt"`
	MobilePhone string `label:"Mobil" section:"Kontakt"`
	LinkedIn    string `label:"LinkedIn" section:"Kontakt"`
	Website     string `label:"Website" section:"Kontakt"`
}

func (c contactViewModel) String() string {
	return c.Firstname + " " + c.Lastname
}

func (c contactViewModel) WithIdentity(id string) contactViewModel {
	return c
}

func (c contactViewModel) Identity() string {
	return "self"
}

func ContactPage(wnd core.Window, pages Pages, changeMyContact user.UpdateMyContact, readMyContact user.ReadMyContact) core.View {
	uc := rcrud.UseCasesFrom[contactViewModel, string](&rcrud.Funcs[contactViewModel, string]{})

	bnd := crud.AutoBinding[contactViewModel](crud.AutoBindingOptions{}, wnd, uc)
	state := core.AutoState[contactViewModel](wnd).Init(func() contactViewModel {
		c, err := readMyContact(wnd.Subject())
		if err != nil {
			alert.ShowBannerError(wnd, err)
			return contactViewModel{}
		}

		if c.PreferredLanguage == "" {
			c.PreferredLanguage = "de_DE"
		}

		if c.Country == "" {
			c.Country = "DE"
		}

		return contactViewModel{
			Avatar:            c.Avatar,
			AcademicDegree:    c.AcademicDegree,
			OfficialTitle:     c.OfficialTitle,
			Salutation:        c.Salutation,
			Firstname:         c.Firstname,
			Lastname:          c.Lastname,
			PreferredLanguage: c.PreferredLanguage,
			EMail:             wnd.Subject().Email(),
			Position:          c.Position,
			CompanyName:       c.CompanyName,
			PostalCode:        c.PostalCode,
			City:              c.City,
			Country:           c.Country,
			Phone:             c.Phone,
			MobilePhone:       c.MobilePhone,
			LinkedIn:          c.LinkedIn,
			Website:           c.Website,
		}
	})

	return ui.VStack(
		ui.VStack(
			ui.H1("Profil bearbeiten"),
			crud.Form[contactViewModel](bnd, state),
			ui.HLineWithColor(ui.ColorAccent),
			ui.HStack(
				ui.SecondaryButton(func() {
					wnd.Navigation().BackwardTo(pages.MyProfile, nil)
				}).Title("Abbrechen"),
				ui.PrimaryButton(func() {
					c := state.Get()
					err := changeMyContact(wnd.Subject(), user.Contact{
						Avatar:            c.Avatar,
						AcademicDegree:    c.AcademicDegree,
						OfficialTitle:     c.OfficialTitle,
						Salutation:        c.Salutation,
						Firstname:         c.Firstname,
						Lastname:          c.Lastname,
						Phone:             c.Phone,
						MobilePhone:       c.MobilePhone,
						Country:           c.Country,
						City:              c.City,
						PostalCode:        c.PostalCode,
						LinkedIn:          c.LinkedIn,
						Website:           c.Website,
						Position:          c.Position,
						CompanyName:       c.CompanyName,
						PreferredLanguage: c.PreferredLanguage,
					})

					if err != nil {
						alert.ShowBannerError(wnd, err)
						return
					}

					wnd.Navigation().BackwardTo(pages.MyProfile, nil)
				}).Title("Speichern"),
			).
				Alignment(ui.Trailing).
				Gap(ui.L16).
				FullWidth(),
		).
			Alignment(ui.Leading).
			Frame(ui.Frame{MaxWidth: ui.L480}.FullWidth()),
		ui.FixedSpacer("", ui.L16),
	).FullWidth()
}
