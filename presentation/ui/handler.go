package ui

import "net/http"

func Handler[Model any](render func(Model) Page, options ...RenderOption[Model]) http.HandlerFunc {
	return nil
}

type rHnd[Model any] struct {
	renderer func(Model) Page
	//decoders  map[string]MsgHandler[Model]
	//onRequest UpdReqFunc[Model]
	maxMemory int64
}

// RenderOption provides a package private Options pattern.
type RenderOption[Model any] func(hnd *rHnd[Model])

// UpdateFunc mutates the model by applying the Msg and returning an altered Model.
type UpdateFunc[Model, Evt any] func(model Model, evt Evt) Model

// OnEvent is invoked if the message alias is matched and tries to unmarshal the form value _eventData message into a new value
// of type Msg. It then calls the UpdateFunc to transform the given Model into a new state.
// To apply navigation, see also [Redirect].
func OnEvent[Model, Evt any](update UpdateFunc[Model, Evt]) RenderOption[Model] {
	return nil
}
