package ui

import "go.wdy.de/nago/container/slice"

// TODO problem between view and persona => persona creates endpoint behavior (eventually???)
type VBox struct {
	ID       ComponentID
	Children slice.Slice[Persona]
}

func (v VBox) Id() ComponentID {
	return v.ID
}

func (v VBox) Endpoints(page PageID, authenticated bool) []Endpoint {
	//TODO implement me
	panic("implement me")
}
