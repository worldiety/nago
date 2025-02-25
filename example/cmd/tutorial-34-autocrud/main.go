// #[go.permission.generateTable]
package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/web/vuejs"
	"time"
)

type PersonID string

type Person struct {
	ID       PersonID   `visible:"false"`
	Vorname  string     `table-visible:"false"`
	Nachname string     `label:"Zuname"`
	Nr       string     `section:"Adressdaten"`
	Strasse  string     `section:"Adressdaten"`
	Anrede   string     `values:"[\"Herr\",\"Frau\"]"`
	Friends  []PersonID `source:"persons"`
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
		cfg.AddSystemService("persons", crud.AnyUseCaseList(useCases.FindAll))

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "8fb8724f-e604-444c-9671-58d07dd76164"))

		std.Must(std.Must(cfg.UserManagement()).UseCases.EnableBootstrapAdmin(time.Now().Add(time.Hour), "8fb8724f-e604-444c-9671-58d07dd76164"))

		std.Must(cfg.Authentication())
		cfg.SetDecorator(cfg.NewScaffold().
			MenuEntry().Icon(heroSolid.BellSnooze).Action(func(wnd core.Window) {
			alert.ShowBannerMessage(wnd, alert.Message{Title: "snack it", Message: "nom nom" + time.Now().String()})
		}).Private().
			MenuEntry().Icon(heroSolid.ArchiveBox).Title("Archiv").Action(func(wnd core.Window) {
			alert.ShowBannerError(wnd, fmt.Errorf("archiv not implemented, db password=1234"))
		}).Public().
			MenuEntry().Icon(heroSolid.Battery50).Title("Status").OneOf(group.PermFindByID).
			Decorator())

		cfg.RootView(".", cfg.DecorateRootView(crud.AutoRootView(crud.AutoRootViewOptions{
			Title: "Personen",
		}, useCases)))

		cfg.OnDestroy(func() {
			fmt.Println("regular shutdown")
		})
	}).Run()
}
