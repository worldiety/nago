// #[go.permission.generateTable]
package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/image"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/web/vuejs"
)

type PersonID string

type Person struct {
	ID        PersonID `visible:"false"`
	Vorname   string   `table-visible:"false"`
	Nachname  string   `label:"Zuname"`
	Nr        string   `section:"Adressdaten"`
	Strasse   string   `section:"Adressdaten"`
	Anrede    string   `values:"[\"Herr\",\"Frau\"]"`
	Profile   image.ID `style:"avatar"`
	Teaser    image.ID `json:"teaser2"`
	Favorites []image.Image
	Gallery   image.Image `style:"gallery"`
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

		// this must happen before IAM init, otherwise the permissions are missing
		persons := application.SloppyRepository[Person](cfg)
		useCases := crud.NewUseCases("de.tutorial.person", persons)

		iamCfg := application.IAMSettings{}
		iamCfg.Decorator = cfg.NewScaffold().Decorator()
		iamCfg = cfg.IAM(iamCfg)

		cfg.RootView(".", iamCfg.DecorateRootView(crud.AutoRootView(crud.AutoRootViewOptions{
			Title: "Personen",
		}, useCases)))

		cfg.OnDestroy(func() {
			fmt.Println("regular shutdown")
		})
	}).Run()
}
