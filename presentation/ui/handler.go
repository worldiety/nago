package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PageHandler struct {
	id      string
	handler http.HandlerFunc
}

func (p *PageHandler) ID() string {
	return p.id
}

func (p *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.handler(w, r)
}

func Page[Model any](id string, render func(Model) View, options ...RenderOption[Model]) PageHandler {

	return PageHandler{
		id: id,
		handler: func(w http.ResponseWriter, r *http.Request) {
			var zero Model
			view := render(zero)
			buf, err := json.Marshal(view)
			if err != nil {
				panic(fmt.Errorf("illegal state: %w", err)) // this would mean, that the UI model is broken
			}

			if _, err := w.Write(buf); err != nil {
				// todo where is the app?
				fmt.Println(err)
			}
		},
	}
}

type rHnd[Model any] struct {
	renderer func(Model) View
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
