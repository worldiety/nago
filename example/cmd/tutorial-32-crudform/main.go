package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/crud"
	"go.wdy.de/nago/web/vuejs"
)

type HelferID string

type Helfer struct {
	ID   HelferID
	Name string
}

func (h Helfer) Identity() HelferID {
	return h.ID
}

type EventID string

type Event struct {
	ID                  EventID
	Name                string
	Note1, Note2, Note3 string
	GeplanteHelfer      []HelferID
	Abfahrt             std.Option[Abfahrt]
}

type Abfahrt struct {
	Zeit       xtime.Date
	AbfahrtOrt AddressID
}

type AddressID string

func (e Event) Identity() EventID {
	return e.ID
}

type Events data.Repository[Event, EventID]
type Helfers data.Repository[Helfer, HelferID]

func main() {

	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial")
		cfg.Serve(vuejs.Dist())

		events := application.SloppyRepository[Event, EventID](cfg)
		helfers := application.SloppyRepository[Helfer, HelferID](cfg)

		/*	go func() {
				for range 1000 {
					events.Save(Event{ID: data.RandIdent[EventID]()})
					time.Sleep(100 * time.Millisecond)
					helfers.Save(Helfer{ID: data.RandIdent[HelferID]()})
				}
			}()

			go func() {
				for range 500 {
					events.Save(Event{ID: data.RandIdent[EventID]()})
					time.Sleep(10 * time.Millisecond)
					helfers.Save(Helfer{ID: data.RandIdent[HelferID]()})
				}
				os.Exit(-1)
			}()*/

		cfg.RootView(".", func(wnd core.Window) core.View {
			bnd := crud.NewBinding[Event](wnd)
			bnd.Add(
				crud.Text(crud.TextOptions{Label: "Name"}, crud.Ptr(func(entity *Event) *string {
					return &entity.Name
				})).WithValidation(func(evt Event) (errorText string, infrastructureError error) {
					if evt.Name == "" {
						return "Darf nicht leer sein", nil
					}

					return "", nil
				}).WithSupportingText("Gib Deinen Namen an"),
			)

			// assemble complex and nested section with rows
			var noteSectionFields []crud.Field[Event]
			noteSectionFields = append(noteSectionFields,
				crud.Text(crud.TextOptions{Label: "Note1"}, crud.Ptr(func(entity *Event) *string {
					return &entity.Note1
				})).WithValidation(func(event Event) (errorText string, infrastructureError error) {
					if event.Note1 == "" {
						return "Notiz 1 muss ausgefüllt sein", nil
					}

					return "", nil
				}),
				crud.Text(crud.TextOptions{Label: "Note2"}, crud.Ptr(func(entity *Event) *string {
					return &entity.Note2
				})).WithValidation(func(event Event) (errorText string, infrastructureError error) {
					if event.Note2 == "" {
						return "Notiz 2 muss ausgefüllt sein", nil
					}

					return "", nil
				}),
				crud.Text(crud.TextOptions{Label: "Note3"}, crud.Ptr(func(entity *Event) *string {
					return &entity.Note3
				})),
			)

			noteSectionFields = append(noteSectionFields, crud.Row(
				crud.FormColumn(crud.Text(crud.TextOptions{Label: "Note1"}, crud.Ptr(func(entity *Event) *string {
					return &entity.Note1
				})), 0.33),
				crud.FormColumn(crud.Text(crud.TextOptions{Label: "Note3"}, crud.Ptr(func(entity *Event) *string {
					return &entity.Note3
				})), 0.66),
			)...)

			noteSectionFields = append(noteSectionFields, crud.HLine[Event]())

			noteSectionFields = append(noteSectionFields, crud.Option(
				"Optionale Felder",
				false,
				crud.Row(
					crud.FormColumn(crud.Text(crud.TextOptions{Label: "Note1"}, crud.Ptr(func(entity *Event) *string {
						return &entity.Note1
					})), 0.33),
					crud.FormColumn(crud.Text(crud.TextOptions{Label: "Note3"}, crud.Ptr(func(entity *Event) *string {
						return &entity.Note3
					})), 0.66),
				)...,
			)...)

			bnd.Add(crud.Section("Notizen",
				noteSectionFields...,
			)...)

			// without section
			bnd.Add(crud.Row(
				crud.FormColumn(crud.Text(crud.TextOptions{Label: "Note1"}, crud.Ptr(func(entity *Event) *string {
					return &entity.Note1
				})), 0.33),
				crud.FormColumn(crud.Text(crud.TextOptions{Label: "Note3"}, crud.Ptr(func(entity *Event) *string {
					return &entity.Note3
				})), 0.66),
			)...)

			// foreign key helper
			helferBnd := crud.NewBinding[Helfer](wnd).Add(
				crud.Text(crud.TextOptions{Label: "Vorname"}, crud.Ptr(func(model *Helfer) *string {
					return &model.Name
				})),
			)

			bnd.Add(
				crud.Section[Event]("", crud.OneToManyTable[Event, Helfer, HelferID](
					crud.OneToManyTableOptions[Helfer, HelferID]{
						Label:           "Helfer",
						ForeignEntities: helfers.All(),
						ForeignBinding:  helferBnd,
						ForeignZero:     Helfer{},
						ForeignCreate: func(helfer Helfer) (errorText string, infrastructureError error) {
							helfer.ID = data.RandIdent[HelferID]()
							err := helfers.Save(helfer)
							return "", err
						},
						ForeignPickerRenderer: func(helfer Helfer) core.View {
							return ui.Text(helfer.Name)
						},
					},
					crud.Ptr(
						func(model *Event) *[]HelferID {
							return &model.GeplanteHelfer
						},
					),
				))...,
			)

			return ui.VStack(
				crud.View[Event, EventID](
					crud.Options[Event, EventID](bnd).
						Actions(
							crud.ButtonCreate[Event](bnd, Event{ID: "do not randomize here"}, func(evt Event) (errorText string, infrastructureError error) {
								if !bnd.Validates(evt) {
									return "irgendein validation fehler, gugg hin", nil
								}

								evt.ID = data.RandIdent[EventID]() // create a unique ID here

								return "", events.Save(evt)
							}),
						).
						FindAll(events.All()).
						Title("Events"),
				),
			).Padding(ui.Padding{}.All(ui.L16)).Frame(ui.Frame{}.FullWidth())
		})
	}).Run()
}
