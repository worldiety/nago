package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/example/events/web"
	"go.wdy.de/nago/persistence/kv"
)

type Event struct {
	ID   string
	Name string
}

func (e Event) Identity() string {
	return e.ID
}

func main() {

	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example Event Planner")
		books := kv.NewCollection[Event](cfg.Store("planner-db"), "events")
		_ = books

		cfg.Page(web.Home())
	}).Run()
}
