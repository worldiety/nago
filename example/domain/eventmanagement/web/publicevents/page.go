package publicevents

import (
	"go.wdy.de/nago/container/serrors"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/example/domain/eventmanagement"
	. "go.wdy.de/nago/presentation/ui"
	"strconv"
)

type ShowAllPublicEventsFunc func() (slice.Slice[eventmanagement.Event], serrors.InfrastructureError)

type PublicEventPageModel struct {
	Events slice.Slice[eventmanagement.Event]
}

type FormAbschicken struct {
	Firstname string
	CSVDatei  []byte
}

func Handler(f ShowAllPublicEventsFunc) PageHandler {
	return Page(
		"/events/public",
		Render,
		OnEvent(func(model PublicEventPageModel, evt FormAbschicken) PublicEventPageModel {
			return model
		}),
		OnRequest(func(model PublicEventPageModel) PublicEventPageModel {
			events, err := f()
			serrors.OrPanic(err)
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
					TableCell{Child: InputText{Name: "CSVDatei"}},
					TableCell{Child: Button{OnClick: "hello world"}},
				),
			}
		}),
	}
}
