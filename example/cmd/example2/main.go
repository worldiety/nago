package main

import (
	"fmt"
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/web/vuejs"
)

type PID string

type Person struct {
	ID        PID
	Firstname string
}

func (p Person) Identity() PID {
	return p.ID
}

type PetID int

type Pet struct {
	ID   PetID
	Name string
}

func (p Pet) Identity() PetID {
	return p.ID
}

type IntlIke int

type EntByInt struct {
	ID IntlIke
}

func (e EntByInt) Identity() IntlIke {
	return e.ID
}

type MyForm struct {
	Name    ui.TextField
	Check   ui.SwitchField
	DueDate ui.DateField
	Comment ui.TextAreaField
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example 2")

		cfg.KeycloakAuthentication()
		persons := kv.NewCollection[Person, PID](cfg.Store("test2-db"), "persons")
		err := persons.Save(
			Person{
				ID:        "1",
				Firstname: "Frodo",
			},
			Person{
				ID:        "2",
				Firstname: "Sam",
			},
			Person{
				ID:        "3",
				Firstname: "Pippin",
			},
		)
		if err != nil {
			panic(err)
		}

		pets := kv.NewCollection[Pet, PetID](cfg.Store("test2-db"), "pets")
		pets.Save(
			Pet{
				ID:   1,
				Name: "Katze",
			},
			Pet{
				ID:   2,
				Name: "Hund",
			},
			Pet{
				ID:   3,
				Name: "Esel",
			},
			Pet{
				ID:   4,
				Name: "Stadtmusikant",
			},
		)

		testPet, err2 := pets.Find(3)
		if err2 != nil {
			panic(err2)
		}
		fmt.Println(testPet)

		cfg.Serve(vuejs.Dist())

		counter := 0
		cfg.LivePage("1234", func(w ui.Wire) *ui.LivePage {
			page := ui.NewLivePage(w)
			page.SetBody(
				ui.NewVBox().With(func(vbox *ui.VBox) {
					vbox.Append(
						ui.NewTextField().With(func(t *ui.TextField) {
							t.Label().Set("Vorname")
							t.Hint().Set("dieses Feld ist ohne Fehler")
						}),

						ui.NewTextField().With(func(t *ui.TextField) {
							t.Label().Set("Nachname")
							t.OnTextChanged().Set(func() {
								fmt.Printf("ontext changed to '%v'\n", t.Value().Value())
								t.Error().Set("Malte sagt doch gut: " + t.Value().Get())
								if t.Value().Get() == "magic" {
									vbox.Append(ui.NewTextField().With(func(t *ui.TextField) {
										t.Label().Set("magic field")
										t.Disabled().Set(true)
									}))
								}
							})
						}),

						ui.NewButton().With(func(btn *ui.Button) {
							btn.Caption().Set("hello world")
							btn.Action().Set(func() {
								counter++
								btn.Caption().Set(fmt.Sprintf("clicked %d", counter))
							})
						}),
						ui.NewButton().With(func(btn *ui.Button) {
							btn.Caption().Set("primary")
							btn.Style().Set(ui.PrimaryIntent)
						}),
						ui.NewButton().With(func(btn *ui.Button) {
							btn.Caption().Set("secondary")
							btn.Style().Set(ui.SecondaryIntent)
						}),
						ui.NewButton().With(func(btn *ui.Button) {
							btn.Caption().Set("destructive")
							btn.Style().Set(ui.Destructive)
						}),

						ui.NewButton().With(func(btn *ui.Button) {
							btn.Caption().Set("subtile")
							btn.Style().Set(ui.SubtileIntent)
						}),
						ui.NewButton().With(func(btn *ui.Button) {
							btn.Caption().Set("preicon")
							btn.PreIcon().Set(`<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 18 21">
<path d="M15 12a1 1 0 0 0 .962-.726l2-7A1 1 0 0 0 17 3H3.77L3.175.745A1 1 0 0 0 2.208 0H1a1 1 0 0 0 0 2h.438l.6 2.255v.019l2 7 .746 2.986A3 3 0 1 0 9 17a2.966 2.966 0 0 0-.184-1h2.368c-.118.32-.18.659-.184 1a3 3 0 1 0 3-3H6.78l-.5-2H15Z"/>
</svg>`)
							btn.Style().Set(ui.PrimaryIntent)
						}),

						ui.NewButton().With(func(btn *ui.Button) {
							btn.Caption().Set("post")
							btn.PostIcon().Set(`<svg class="rtl:rotate-180" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 10">
<path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M1 5h12m0 0L9 1m4 4L9 9"/>
</svg>`)
							btn.Style().Set(ui.PrimaryIntent)
						}),

						ui.NewButton().With(func(btn *ui.Button) {
							btn.PreIcon().Set(`<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 18 18">
<path d="M3 7H1a1 1 0 0 0-1 1v8a2 2 0 0 0 4 0V8a1 1 0 0 0-1-1Zm12.954 0H12l1.558-4.5a1.778 1.778 0 0 0-3.331-1.06A24.859 24.859 0 0 1 6 6.8v9.586h.114C8.223 16.969 11.015 18 13.6 18c1.4 0 1.592-.526 1.88-1.317l2.354-7A2 2 0 0 0 15.954 7Z"/>
</svg>`)
							btn.Style().Set(ui.PrimaryIntent)
						}),
					)
				}),
			)

			return page
		})

	}).Run()
}

type ExampleForm struct {
	Vorname  ui.TextField `class:"col-start-2 col-span-4"`
	Nachname ui.TextField
	Avatar   ui.FileUploadField
	Chooser  ui.SelectField[PID]
}
