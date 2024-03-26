package main

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
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

func PersonName(p Person) string {
	return p.Firstname + " " + p.Lastname
}

func (p Person) Identity() PersonID {
	return p.ID
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
