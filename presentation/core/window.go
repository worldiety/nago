package core

import (
	"context"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/ora"
	"golang.org/x/text/language"
	"log/slog"
	"time"
)

// A Window owns the lifecycle of a component and is part of a Scope.
// Ora does not define (yet) what a window is.
// However, obviously every component lives inside a window of the frontend and navigation is related to that.
// Today, a frontend client must treat a scope as a window.
type Window interface {
	// Execute posts the task into the associated Executor. It will be executed in the associated event loop
	// to allow race free processing.
	// Use this to post updates from foreign goroutines into the ui components.
	Execute(task func())

	// Navigation allows access to the window navigation.
	Navigation() *NavigationController

	// Values contains those values, which have been passed from the callers, e.g. intent parameters or url query
	// parameters. This depends on the actual frontend.
	Values() Values

	// User is never nil. Check [auth.User.Valid]. You must not keep the User instance over a long time, because
	// it will change over time, either due to refreshing tokens or because the user is logged out.
	User() auth.User

	// Context returns the wire-lifetime context. Contains additional injected types like User or Logger.
	Context() context.Context

	// SessionID is a unique identifier, which is assigned to a client using some sort of cookie mechanism. This is a
	// pure random string and belongs to a distinct client instance. It is shared across multiple windows on the client,
	// especially when using multiple tabs or activity windows. You may use this for authentication mechanics,
	// however be careful not to break external security concerns by never revisiting the actual user authentication
	// state.
	// It usually outlives a frontend process and e.g. is restored after a device restart.
	SessionID() SessionID

	// Authenticate triggers a round trip so that [Window.User] may contain a valid user afterward.
	// For sure, the user can always cancel that.
	Authenticate()

	// ViewRoot returns the access to the component which represents the current root.
	ViewRoot() ViewRoot

	// Locale returns the negotiated language tag or locale identifier between the frontend and the backend.
	Locale() language.Tag

	// Location returns negotiated time zone location
	Location() *time.Location
	// TODO add Locale, add Screen Metrics (density, pixel width and height, size classes etc)

}

type SessionID string

type scopeWindow struct {
	factory       ora.ComponentFactoryId
	scope         *Scope
	navController *NavigationController
	values        Values
	location      *time.Location
	viewRoot      *scopeViewRoot
}

func newScopeWindow(scope *Scope, factory ora.ComponentFactoryId, values Values) *scopeWindow {
	s := &scopeWindow{factory: factory, scope: scope, values: values, navController: NewNavigationController(scope), viewRoot: newScopeViewRoot(scope)}
	if values == nil {
		s.values = Values{}
	}

	loc, err := time.LoadLocation("Europe/Berlin") // TODO implement me
	if err != nil {
		slog.Error("cannot load location", slog.Any("err", err))
		loc = time.UTC
	}
	s.location = loc

	return s
}

func (s *scopeWindow) Execute(task func()) {
	s.scope.eventLoop.Post(task)
}

func (s *scopeWindow) Navigation() *NavigationController {
	return s.navController
}

func (s *scopeWindow) Values() Values {
	return s.values
}

func (s *scopeWindow) User() auth.User {
	return invalidUser{} // TODO
}

func (s *scopeWindow) Context() context.Context {
	return s.scope.ctx
}

func (s *scopeWindow) SessionID() SessionID {
	return s.scope.sessionID
}

func (s *scopeWindow) Authenticate() {
	// TODO implement me
}

func (s *scopeWindow) ViewRoot() ViewRoot {
	return s.viewRoot
}

func (s *scopeWindow) Locale() language.Tag {
	return language.German // TODO implement me
}

func (s *scopeWindow) Location() *time.Location {
	return s.location
}
