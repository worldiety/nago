// #[go.permission.generateTable]
package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/image"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/web/vuejs"
)

type PersonID string

type Person struct {
	ID       PersonID `visible:"false"`
	Vorname  string   `table-visible:"false"`
	Nachname string   `label:"Zuname"`
	Nr       string   `section:"Adressdaten"`
	Strasse  string   `section:"Adressdaten"`
	Anrede   string   `values:"[\"Herr\",\"Frau\"]"`
	Profile  image.ID `style:"avatar"`
	Teaser   image.ID `json:"teaser2"`
}

func (p Person) Paraphe() string {
	return p.Vorname + " " + p.Nachname
}

func (p Person) Identity() PersonID {
	return p.ID
}

func (p Person) WithIdentity(id PersonID) Person {
	p.ID = id
	return p
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		std.Must(cfg.Authentication())

		persons := application.SloppyRepository[Person](cfg)
		useCases := crud.NewUseCases("de.tutorial.person", persons)
		cfg.SetDecorator(cfg.NewScaffold().Decorator())

		cfg.RootView(".", cfg.DecorateRootView(crud.AutoRootView(crud.AutoRootViewOptions{
			Title: "Personen",
		}, useCases)))

	}).Run()
}
