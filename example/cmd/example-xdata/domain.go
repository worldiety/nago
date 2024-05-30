package main

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/std"
	"math/rand"
	"time"
)

type PersonID string
type Rank int64

type Person struct {
	ID         PersonID
	Firstname  string
	Lastname   string
	Age        int
	Rank       Rank
	Friends    []PersonID
	BestFriend PersonID
	CoolGuy    bool
	Birthday   time.Time
}

func (p Person) Identity() PersonID {
	return p.ID
}

func PersonName(p Person) string {
	return p.Firstname + " " + p.Lastname
}

type PersonView struct {
	ID        PersonID `ignore:"true"`
	Firstname string   `caption:"Vorname" sortable:"true"`
	Lastname  string   `caption:"Nachname" sortable:"true"`
	Age       int      `caption:"Alter" sortable:"true"`
	Rank      Rank     `sortable:"true"`
	Friends   int      `caption:"Anzahl Freunde"`
}

type PersonService struct {
	repo Persons
}

func NewPersonService(repo Persons) *PersonService {
	return &PersonService{repo: repo}
}

func (p *PersonService) ViewPersons() iter.Seq2[PersonView, error] {
	return iter.Map2[Person, error, PersonView, error](func(person Person, err error) (PersonView, error) {
		return PersonView{
			ID:        person.ID,
			Firstname: person.Firstname,
			Lastname:  person.Lastname,
			Age:       person.Age,
			Rank:      person.Rank,
			Friends:   len(person.Friends),
		}, err
	}, p.repo.Each)
}

func (p *PersonService) Persons() iter.Seq2[Person, error] {
	return p.repo.Each
}

func (p *PersonService) Update(person Person) error {
	return p.repo.Save(person)
}

func (p *PersonService) RemoveByPersonView(person PersonView) error {
	return p.repo.DeleteByID(person.ID)
}

func (p *PersonService) FindPerson(id PersonID) (std.Option[Person], error) {
	return p.repo.FindByID(id)
}

type Persons data.Repository[Person, PersonID]

func initUsers(repo Persons) error {
	names := []string{
		"Paco",
		"Benni",
		"Shiva",
		"Bo",
		"Noah",
		"Finn",
		"Amar",
		"Robin",
		"Mika",
		"Jona",
		"Yuki",
		"Luca",
		"Kim",
	}

	for i := range 20 {
		err := repo.Save(Person{
			ID:        PersonID(fmt.Sprintf("p%d", i)),
			Firstname: names[rand.Intn(len(names))],
			Lastname:  names[rand.Intn(len(names))],
			Age:       rand.Intn(45) + 1,
			Rank:      Rank(rand.Intn(1000)),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
