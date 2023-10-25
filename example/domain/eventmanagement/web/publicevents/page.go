package publicevents

import (
	"go.wdy.de/nago/container/errors"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/example/domain/eventmanagement"
	"go.wdy.de/nago/persistence/kv"
	. "go.wdy.de/nago/presentation/ui"
	"strconv"
)

type PublicEventPageModel struct {
	Events slice.Slice[eventmanagement.Event]
}

func Handler(c kv.Collection[eventmanagement.Event, eventmanagement.EventID]) PageHandler {
	return Page(
		"/events/public",
		Render,
		OnRequest(func(model PublicEventPageModel) PublicEventPageModel {
			events, err := eventmanagement.ShowAllPublicEvents(c)
			errors.OrPanic(err)
			model.Events = events
			return model
		}),
	)
}

func Render(model PublicEventPageModel) View {
	return Table{
		Rows: Map(model.Events, func(idx int, in eventmanagement.Event) TableRow {
			return TableRow{
				Columns: slice.Of(
					TableCell{Child: Text(strconv.Itoa(idx))},
					TableCell{Child: Text(in.Name)},
				),
			}
		}),
	}
}
