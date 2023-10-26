package ui

import (
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/auth"
	"net/http"
)

type PageHandler struct {
	id                     string
	handler                http.HandlerFunc
	authenticationRequired bool
}

func (p *PageHandler) ID() string {
	return p.id
}

func (p *PageHandler) Authenticated() bool {
	return p.authenticationRequired
}

func (p *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.handler(w, r)
}

func Page[Model any](id string, render func(Model) View, options ...RenderOption[Model]) PageHandler {
	hnd := &rHnd[Model]{
		authenticationRequired: false,
		authenticationOptional: false,
	}

	for _, option := range options {
		option(hnd)
	}

	frontendRequiresAuth := hnd.authenticationRequired
	if hnd.authenticationOptional {
		frontendRequiresAuth = false
	}

	return PageHandler{
		id:                     id,
		authenticationRequired: frontendRequiresAuth,
		handler: func(w http.ResponseWriter, r *http.Request) {
			user := auth.FromContext(r.Context())
			if frontendRequiresAuth && user == nil {
				// The HyperText Transfer Protocol (HTTP) 401 Unauthorized response status code indicates that the
				// client request has not been completed because it lacks valid authentication credentials for the
				// requested resource.
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			var zero Model
			if hnd.onRequest != nil {
				zero = hnd.onRequest(user, zero)
			}
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
	onRequest              func(auth.User, Model) Model
	maxMemory              int64
	authenticationRequired bool
	authenticationOptional bool
}

// RenderOption provides a package private Options pattern.
type RenderOption[Model any] func(hnd *rHnd[Model])

// UpdateFunc mutates the model by applying the Msg and returning an altered Model.
type UpdateFunc[Model, Evt any] func(model Model, evt Evt) Model

// OnEvent is invoked if the message alias is matched and tries to unmarshal the form value _eventData message into a new value
// of type Msg. It then calls the UpdateFunc to transform the given Model into a new state.
// To apply navigation, see also [Redirect].
func OnEvent[Model, Evt any](update UpdateFunc[Model, Evt]) RenderOption[Model] {
	return func(hnd *rHnd[Model]) {

	}
}

// AuthenticationOptional makes any required authentications optional.
func AuthenticationOptional[Model any]() RenderOption[Model] {
	return func(hnd *rHnd[Model]) {
		hnd.authenticationOptional = true
	}
}

// OnRequest does not care if a user has been authenticated or not. See also [OnAuthRequest] to enforce
// authentication and [AuthenticationOptional] to make that optional.
func OnRequest[Model any](f func(Model) Model) RenderOption[Model] {
	return func(hnd *rHnd[Model]) {
		if hnd.onRequest != nil {
			panic("it is not allowed to configure an OnRequest handler twice")
		}

		hnd.onRequest = func(user auth.User, model Model) Model {
			return f(model)
		}

	}
}

// OnAuthRequest requires authentication, unless [AuthenticationOptional] overrides it. Then the provided user
// may be nil, depending on the actual authentication state.
func OnAuthRequest[Model any](f func(auth.User, Model) Model) RenderOption[Model] {
	return func(hnd *rHnd[Model]) {
		if hnd.onRequest != nil {
			panic("it is not allowed to configure an OnRequest handler twice")
		}

		hnd.onRequest = f
		hnd.authenticationRequired = true
	}
}
