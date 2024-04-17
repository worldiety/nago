package core

import (
	"context"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/ora"
)

// A Realm owns the lifecycle of a component and is part of a Scope.
// Another association could be thinking of a "window", however a window-behavior is currently not specified and
// it depends completely on the clients rendering capabilities - especially if there is a relation between a
// component and a window.
type Realm interface {
	// Execute posts the task into the associated Executor. It will be executed in the associated event loop
	// to allow race free processing.
	// Use this to post updates from foreign goroutines into the ui components.
	Execute(task func())

	// Navigation allows access to the realms (respective window) navigation.
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
}

type SessionID string

type scopeRealm struct {
	factory       ora.ComponentFactoryId
	scope         *Scope
	navController *NavigationController
	values        Values
}

func newScopeRealm(scope *Scope, factory ora.ComponentFactoryId, values Values) *scopeRealm {
	s := &scopeRealm{factory: factory, scope: scope, values: values, navController: NewNavigationController(scope)}
	if values == nil {
		s.values = Values{}
	}

	return s
}

func (s *scopeRealm) Execute(task func()) {
	s.scope.eventLoop.Post(task)
}

func (s *scopeRealm) Navigation() *NavigationController {
	return s.navController
}

func (s *scopeRealm) Values() Values {
	return s.values
}

func (s *scopeRealm) User() auth.User {
	return invalidUser{} // TODO
}

func (s *scopeRealm) Context() context.Context {
	return s.scope.ctx
}

func (s *scopeRealm) SessionID() SessionID {
	return s.scope.sessionID
}
