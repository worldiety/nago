package eventmanagement

import (
	"go.wdy.de/nago/container/serrors"
	"go.wdy.de/nago/container/slice"
)

type EventFilterRepository interface {
	Filter(p func(event Event) bool) (slice.Slice[Event], serrors.InfrastructureError)
}

type EventID string

type Event struct {
	ID     EventID
	Name   string
	Public bool
}

func (e Event) Identity() EventID {
	return e.ID
}

func ShowAllPublicEvents(r EventFilterRepository) (slice.Slice[Event], serrors.InfrastructureError) {
	return r.Filter(func(event Event) bool {
		return event.Public
	})
}
