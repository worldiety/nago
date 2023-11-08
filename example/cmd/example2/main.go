package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/persistence/kv"
	"go.wdy.de/nago/presentation/ui2"
)

type PID string

type Person struct {
	ID        PID
	Firstname string
}

func (p Person) Identity() PID {
	return p.ID
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example 2")

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

		cfg.Page2("hello-world", false, ui2.Scaffold{
			ApplicationName: "bla",
			Navigation: func(context ui2.Context) slice.Slice[ui2.NavItem] {
				return slice.Of(ui2.NavItem{
					Title: "hello world",
				})
			},
			Content: ui2.ListView[PID]{
				List: func() (slice.Slice[ui2.ListItem[PID]], ui2.Status) {
					s, err := persons.Filter(func(person Person) bool {
						return true
					})

					if err != nil {
						panic(err)
					}

					return slice.Map(s, func(idx int, v Person) ui2.ListItem[PID] {
						return ui2.ListItem[PID]{
							ID:    v.ID,
							Title: v.Firstname,
						}
					}), ui2.Ok
				},
			},
		})
	}).Run()
}
