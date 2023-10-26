package ui

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/internal/values"
	"go.wdy.de/nago/logging"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
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
		eventTypeDecoder:       make(map[string]func(ctx context.Context, in Model, encodedEvent []byte) (Model, error)),
		onPanic: func(p PanicContext[Model]) View {
			return Text("internal server error: " + p.IncidentTag)
		},
	}

	for _, option := range options {
		option(hnd)
	}

	frontendRequiresAuth := hnd.authenticationRequired
	if hnd.authenticationOptional {
		frontendRequiresAuth = false
	}

	for k := range hnd.eventTypeDecoder {
		// TODO where is our logger?
		slog.Default().Info("registered page event", slog.String("page", id), slog.String("eventType", k))
	}

	return PageHandler{
		id:                     id,
		authenticationRequired: frontendRequiresAuth,
		handler: func(w http.ResponseWriter, r *http.Request) {
			var model Model
			user := auth.FromContext(r.Context())

			defer func() {
				if p := recover(); p != nil {
					var err error
					if e, ok := p.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("recovered panic in page handler: %v", r)
					}

					var buf [16]byte
					_, _ = rand.Read(buf[:])
					incidentTag := IncidentTag(base64.StdEncoding.EncodeToString(buf[:]))
					fmt.Println(p)
					fmt.Println(string(debug.Stack()))
					logging.FromContext(r.Context()).Error("unexpected panic while rendering", slog.String("page", id), slog.String("incident", string(incidentTag)), slog.Any("err", err))
					if hnd.onPanic != nil {
						view := hnd.onPanic(PanicContext[Model]{
							User:        user,
							Model:       model,
							IncidentTag: incidentTag,
							Error:       err,
						})

						// try to write that out
						buf, err := json.Marshal(view)
						if err != nil {
							panic(fmt.Errorf("illegal state: %w", err)) // this would mean, that the UI model is broken
						}

						if _, err := w.Write(buf); err != nil {
							logging.FromContext(r.Context()).Error("cannot write panic handler response", slog.Any("err", err))
						}
					}
				}
			}()

			if frontendRequiresAuth && user == nil {
				// The HyperText Transfer Protocol (HTTP) 401 Unauthorized response status code indicates that the
				// client request has not been completed because it lacks valid authentication credentials for the
				// requested resource.
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if hnd.onRawRequest != nil {
				// handle any raw request
				model = hnd.onRawRequest(model, r)
			}

			if hnd.onRequest != nil {
				// handle just the simple request based "event" transformation
				model = hnd.onRequest(user, model)
			}

			if hnd.onRequestParams != nil {
				var err error
				model, err = hnd.onRequestParams(model, r)
				if err != nil {
					panic(fmt.Errorf("unexpected params error, invalid params model? cause: %w", err))
				}
			}

			if r.Method == http.MethodPost {
				// looks like a client-side triggered event
				var msg clientMessage
				dec := json.NewDecoder(r.Body)
				if err := dec.Decode(&msg); err != nil {
					logger := logging.FromContext(r.Context())
					logger.Error("invalid client message", slog.Any("err", err))
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if msg.Model != nil {
					if err := json.Unmarshal([]byte(msg.Model), &model); err != nil {
						logger := logging.FromContext(r.Context())
						logger.Error("invalid view model in client message", slog.Any("err", err))
						w.WriteHeader(http.StatusBadRequest)
						return
					}
				} else {
					logging.FromContext(r.Context()).Error("client has sent a post without the last view model, server cannot restore its state")
				}

				switch msg.EventType {
				case "!refresh":
				// no event processing, just render once again through the backend
				default:
					applyEvent, ok := hnd.eventTypeDecoder[msg.EventType]
					if !ok {
						logging.FromContext(r.Context()).Error(fmt.Sprintf("client requested event '%s' which has not been registered in page '%s'", msg.EventType, id))
					} else {
						var err error
						model, err = applyEvent(r.Context(), model, msg.EventData)
						if err != nil {
							logging.FromContext(r.Context()).Error("invalid event data in client message", slog.Any("err", err))
							w.WriteHeader(http.StatusBadRequest)
							return
						}
					}
				}
			}

			view := render(model)
			buf, err := json.Marshal(view)
			if err != nil {
				panic(fmt.Errorf("illegal state: %w", err)) // this would mean, that the UI model is broken
			}

			if _, err := w.Write(buf); err != nil {
				logging.FromContext(r.Context()).Error("cannot write response", slog.Any("err", err))
			}
		},
	}
}

type PanicContext[Model any] struct {
	User        auth.User
	Model       Model
	IncidentTag IncidentTag
	Error       error
}

type rHnd[Model any] struct {
	renderer func(Model) View
	//decoders  map[string]MsgHandler[Model]
	onRequest              func(auth.User, Model) Model
	onRawRequest           func(Model, *http.Request) Model
	onRequestParams        func(in Model, r *http.Request) (Model, error)
	onPanic                func(PanicContext[Model]) View
	maxMemory              int64
	authenticationRequired bool
	authenticationOptional bool
	eventTypeDecoder       map[string]func(ctx context.Context, in Model, encodedEvent []byte) (Model, error)
}

// RenderOption provides a package private Options pattern.
type RenderOption[Model any] func(hnd *rHnd[Model])

// UpdateFunc mutates the model by applying the Msg and returning an altered Model.

// OnEvent is invoked if the message alias is matched and tries to unmarshal the form value _eventData message into a new value
// of type Msg. It then calls the UpdateFunc to transform the given Model into a new state.
// To apply navigation, see also [Redirect].
func OnEvent[Model, Evt any](update func(model Model, evt Evt) Model) RenderOption[Model] {
	return func(hnd *rHnd[Model]) {
		var zeroEvt Evt
		t := reflect.TypeOf(zeroEvt)
		eventTypeName := t.PkgPath() + "." + t.Name()
		if _, ok := hnd.eventTypeDecoder[eventTypeName]; ok {
			panic(fmt.Errorf("the event type '%s' has already been registered", eventTypeName))
		}

		hnd.eventTypeDecoder[eventTypeName] = func(ctx context.Context, in Model, buf []byte) (Model, error) {
			var evt Evt
			if err := json.Unmarshal(buf, &evt); err != nil {
				return in, err
			}

			return update(in, evt), nil
		}
	}
}

func OnAuthEvent[Model, Evt any](update func(auth.User, Model, Evt) Model) RenderOption[Model] {
	return func(hnd *rHnd[Model]) {
		hnd.authenticationRequired = true

		var zeroEvt Evt
		eventTypeName := reflect.TypeOf(zeroEvt).String()
		if _, ok := hnd.eventTypeDecoder[eventTypeName]; ok {
			panic(fmt.Errorf("the event type '%s' has already been registered", eventTypeName))
		}

		hnd.eventTypeDecoder[eventTypeName] = func(ctx context.Context, in Model, buf []byte) (Model, error) {
			var evt Evt
			if err := json.Unmarshal(buf, &evt); err != nil {
				return in, err
			}

			user := auth.FromContext(ctx)
			return update(user, in, evt), nil
		}
	}
}

// AuthenticationOptional makes any required authentications optional.
func AuthenticationOptional[Model any]() RenderOption[Model] {
	return func(hnd *rHnd[Model]) {
		hnd.authenticationOptional = true
	}
}

// OnRawRequest is always the first transformation function. See also [OnRequest].
func OnRawRequest[Model any](f func(Model, *http.Request) Model) RenderOption[Model] {
	return func(hnd *rHnd[Model]) {
		if hnd.onRawRequest != nil {
			panic("already registered a raw request")
		}

		hnd.onRawRequest = f
	}
}

type Request[Params any] struct {
	User          auth.User // optional user or nil
	QueryOrHeader Params    //
}

// OnRequestParams tries to parse any header and url parameter into the given params type.
// It is called after [OnRequest].
func OnRequestParams[Model, Params any](f func(Model, Request[Params]) Model) RenderOption[Model] {
	return func(hnd *rHnd[Model]) {
		hnd.onRequestParams = func(in Model, r *http.Request) (Model, error) {
			var paramsModel Params

			if err := values.Unmarshal(&paramsModel, url.Values(r.Header)); err != nil {
				return in, fmt.Errorf("cannot parse header params: %w", err)
			}

			queryParams := r.URL.Query()
			if err := values.Unmarshal(&paramsModel, queryParams); err != nil {
				return in, fmt.Errorf("cannot parse query params: %w", err)
			}

			in = f(in, Request[Params]{
				User:          auth.FromContext(r.Context()),
				QueryOrHeader: paramsModel,
			})

			return in, nil
		}
	}
}

// OnRequest does not care if a user has been authenticated or not. See also [OnAuthRequest] to enforce
// authentication and [AuthenticationOptional] to make that optional.
// It is always called after [OnRawRequest].
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

type IncidentTag string

// OnPanic overwrites the default panic handler to catch assertions and unhandled (infrastructure) errors
// and can be used to display e.g. a help or contact form to the user.
func OnPanic[Model any](f func(PanicContext[Model]) View) RenderOption[Model] {
	return func(hnd *rHnd[Model]) {
		hnd.onPanic = f
	}
}

type clientMessage struct {
	// EventType is the absolute qualified type name as it was defined within the render tree.
	EventType string `json:"eventType"`
	// EventData is exactly the serialized payload of the Event which has been defined within the render tree.
	EventData json.RawMessage `json:"data"`
	// Model is whatever the server has used to build the render tree. This allows keeping the server stateless so far.
	Model json.RawMessage `json:"model"`
	// FormData is whatever the client wants to send, e.g. input text data, options or even file uploads.
	FormData json.RawMessage `json:"formData"`
}
