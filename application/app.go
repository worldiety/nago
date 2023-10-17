package application

import (
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/presentation/rest"
	"go.wdy.de/nago/presentation/ui"
)

type Application struct {
}

func NewApplication(presentation PresentationLayer) *Application {
	return nil
}

func (a *Application) Run() error {
	return nil
}

type PresentationLayer struct {
	Pages  slice.Slice[ui.Route]
	Events slice.Slice[ui.Event]
	Http   slice.Slice[rest.Route]
}
