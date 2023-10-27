package main

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/container/serrors"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/example/domain/eventmanagement"
	"go.wdy.de/nago/example/domain/eventmanagement/web/publicevents"
	"go.wdy.de/nago/example/events/web"
	"go.wdy.de/nago/persistence/kv"
)

func main() {

	application.Configure(func(cfg *application.Configurator) {
		cfg.Name("Example Event Planner")

		cfg.KeycloakAuthentication()

		events := kv.NewCollection[eventmanagement.Event, eventmanagement.EventID](cfg.Store("planner-db"), "events")
		serrors.OrPanic(migrate(events))

		cfg.Page(web.Home(func(name string) {
			eventmanagement.ShowAllPublicEvents(events)
		}))

		cfg.Page(web.Profile())
		cfg.Page(web.Counter())

		// curry over to hide dependencies at the presentation side
		cfg.Page(publicevents.Handler(func() (slice.Slice[eventmanagement.Event], serrors.InfrastructureError) {
			return eventmanagement.ShowAllPublicEvents(events)
		}))
	}).Run()
}

func migrate(events kv.Collection[eventmanagement.Event, eventmanagement.EventID]) error {
	return events.Save(
		eventmanagement.Event{
			ID:     "1",
			Name:   "Winterzauber",
			Public: true,
		},
		eventmanagement.Event{
			ID:     "2",
			Name:   "Landpartie",
			Public: true,
		},
		eventmanagement.Event{
			ID:     "3",
			Name:   "Weihnachtsfeier 2023",
			Public: false,
		},
	)
}
