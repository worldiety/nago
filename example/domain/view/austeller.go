package view

import (
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/example/domain/data"
	"go.wdy.de/nago/example/domain/usecase"
	"go.wdy.de/nago/presentation/ui"
)

type Action[T any] struct {
	Message  T
	OnAction func(T) error
}

type AusstellerViewID struct {
	ID data.Ausstellernummer
}

type DeleteAussteller data.Ausstellernummer

type MessageHandler struct {
	aussteller usecase.Ausstelleraggregat
}

func (h MessageHandler) OnDeleteAussteller(aussteller DeleteAussteller) any {
	h.aussteller.AusstellerLöschen(data.Ausstellernummer(aussteller))
	// return nothing? => just flicker?
	// return refresh? => no stack?
	// return a navigation redirect? => no stack?
	// return a new view, e.g. Dashboard?
	return h.AusstellerÜbersicht(ShowDashboard{})
}

type ShowDashboard struct{}
type ShowAusstellerDetail struct{ ID data.Ausstellernummer }

func (h MessageHandler) AusstellerÜbersicht(ShowDashboard) ui.View {
	return ui.Scaffold{
		Title: "Test Übersicht",
		Menu: slice.Of(
			ui.ListItem1L{
				Headline:    "Dashboard",
				ActionEvent: ShowDashboard{},
			},
		),
		Body: ui.MainDetail{
			Main:   h.ausstellerListView(),
			Detail: ui.Text("Bitte Aussteller wählen"),
		},
	}
}

func (h MessageHandler) ausstellerListView() ui.ListView {
	return ui.ListView{Items: slice.Map(h.aussteller.AusstellerAnzeigen(), func(idx int, v data.Aussteller) ui.ListItem1L {
		return ui.ListItem1L{
			Headline:    v.Vorname,
			ActionEvent: ShowAusstellerDetail{ID: v.ID},
		}
	})}
}

func (h MessageHandler) AusstellerDetail(detail ShowAusstellerDetail) ui.View {
	return ui.Scaffold{
		Title: "Test Übersicht",
		Menu: slice.Of(
			ui.ListItem1L{
				Headline:    "Dashboard",
				ActionEvent: ShowDashboard{},
			},
		),
		Body: ui.MainDetail{
			Main: func() ui.ListView {
				return h.ausstellerListView()
			},
			Detail: func(item ui.ListItem) ui.View {
				return ui.Text("Bitte Aussteller wählen")
			},
		},
	}
}
