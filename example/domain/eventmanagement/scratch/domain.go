package scratch

import (
	"cmp"
	"go.wdy.de/nago/container/data"
	"go.wdy.de/nago/container/slice"
	dm "go.wdy.de/nago/domain"
	"go.wdy.de/nago/presentation/ui"
	"time"
)

type Ɂ[LOL any] struct{}
type PersonID string

type PersonRepo dm.Repository[Person, PersonID]

type Person struct {
	ID        PersonID
	Firstname data.Validateable[string]
	Lastname  data.Validateable[string]
	Age       data.Validateable[int]
	Address   data.Validateable[Address]
}

func (p Person) Identity() PersonID {
	return p.ID
}

type Address struct {
	Street string
}

func RenamePerson(r PersonRepo, p Person) Person {
	op, err := r.FindOne(p.ID)
	dm.OrTechnicalSupport(err)
	if op.IsNone() {
		panic("race zwischen 2 bearbeitern???")
	}

	if p.Lastname.Value == "" {
		p.Lastname.ErrorText = "Der Nachname darf nicht leer sein"
		return p
	}

	if p.Age.Value < 18 {
		p.Age.ErrorText = "Volljährig du musst sein"
		return p
	}

	p.Address.LabelText = "Bitte dies als Formular in eigener Sektion darstellen"

	dm.OrTechnicalSupport(r.Save(p))

	return p
}

func WithPersonFormHints(p Person) Person {
	p.Age.SupportingText = "Dein Alter zwischen 18 und 99"
	p.Firstname.SupportingText = "Dein Vorname, darf nicht leer sein"
	return p
}

type AppLayer struct {
	personsRepo PersonRepo
}

func (a AppLayer) RenamePerson(p Person) Person {
	return RenamePerson(a.personsRepo, p)
}

func PresentationOverview(r PersonRepo) any {
	return ui.ListPersona{
		ID: nil,
		Items: slice.Map(must(r.Filter(func(person Person) bool {
			return true
		})), func(idx int, v Person) ui.ListItem {
			return ui.ListItem{
				ID: v.ID,
			}
		}),
		Delete:        ui.Opt[ui.DeleteView]{},
		Edit:          ui.Opt[ui.EditView]{},
		DeleteBatch:   ui.Opt[ui.Confirmation]{},
		BatchBeliebig: nil,
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

type SavePersonCmd struct {
	Command
	Firstname string
	Lastname  string
	Age       int
}

type Command struct {
}

func (Command) isCmd() {}

type Event struct {
	IssuedAt time.Time
}

func (Event) isEvent() {

}

type PersonCreated struct {
	Event
	ID PersonID
}

type PersonCreationFailed struct {
	Event
	Person
	Error string
}

// well this app events can be as sync as everything
func SavePerson(app Application, cmd SavePersonCmd) {
	repo := Persistence[Person, PersonID](app, "persons")
	if err := repo.Save(); err != nil {
		app.OutgoingEvents().Send(PersonCreationFailed{
			Error: err.Error(),
		})
		return
	}

	app.OutgoingEvents().Send(PersonCreated{
		Event: Event{},
		ID:    "1234",
	})
}

// weired but discoverable using reflection and actually also a native sum type declaration
func SavePerson2(app Application, cmd SavePersonCmd) (createdEvent data.Option[PersonCreated], failedEvent data.Option[PersonCreationFailed]) {
	repo := Persistence[Person, PersonID](app, "persons")
	if err := repo.Save(); err != nil {
		failedEvent = data.Some(PersonCreationFailed{
			Error: err.Error(),
		})
		return
	}

	createdEvent = data.Some(PersonCreated{
		Event: Event{},
		ID:    "1234",
	})

	return
}

func RegisterUseCase[Cmd interface{ isCmd() }](app Application, desc string, f func(Application, Cmd)) {

}

func OnDomainEvent[T interface{ isEvent() }](application Application, f func(T)) {

}

type OutgoingEvents interface {
	Send(msg interface{ isEvent() }) // no error, callee can never do anything about that
}

type Application interface {
	OutgoingEvents() OutgoingEvents
	OnReceive(f func(msg []byte))
	Store(name string) any
}

func Persistence[T dm.Entity[ID], ID cmp.Ordered](application Application, name string) dm.Repository[T, ID] {
	panic("")
}

func main() {
	var app Application
	RegisterUseCase(app, "Save person use case persists a new person", SavePerson)
	OnDomainEvent(app, func(t PersonCreationFailed) {
		// e.g. send mail
		// e.g. perform the presentation rendering???
	})

}
