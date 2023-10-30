package publicevents

import (
	"fmt"
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

type MyHeadersAndQueryParams struct {
	Test string
}

func Handler(f ShowAllPublicEventsFunc) PageHandler {
	return Page(
		"/events/public",
		Render,
		OnRequest(func(model PublicEventPageModel) PublicEventPageModel {
			events, err := f()
			serrors.OrPanic(err)
			model.Events = events
			return model
		}),

		OnEvent(func(model PublicEventPageModel, evt FormAbschicken) PublicEventPageModel {
			return model
		}),

		OnRequestParams(func(model PublicEventPageModel, r Request[MyHeadersAndQueryParams]) PublicEventPageModel {
			fmt.Println("Test Query Value ->", r.QueryOrHeader.Test)
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
					TableCell{Child: InputFile{Name: "CSVDatei"}},
					TableCell{Child: Button{OnClick: "hello world"}},
				),
			}
		}),
	}
}
