package eventmanagement

import (
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/persistence"
	"go.wdy.de/nago/persistence/kv"
)

type EventID string

type Event struct {
	ID     EventID
	Name   string
	Public bool
}

func (e Event) Identity() EventID {
	return e.ID
}

func ShowAllPublicEvents(c kv.Collection[Event, EventID]) (slice.Slice[Event], persistence.InfrastructureError) {
	return c.Filter(func(event Event) bool {
		return event.Public
	})
}
