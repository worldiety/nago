package core

import (
	"context"
	"fmt"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/ora"
	"golang.org/x/text/language"
	"io"
	"log/slog"
	"sync"
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
	// Note, that an invalidation is not triggered automatically. Either use [ViewRoot.Invalidate] manually or
	// even better use [Post], [PostDelayed] or [Schedule], because those have optimized lifecycle handling.
	Execute(task func())

	// Navigation allows access to the window navigation.
	Navigation() *NavigationController

	// Values contains those values, which have been passed from the callers, e.g. intent parameters or url query
	// parameters. This depends on the actual frontend.
	Values() Values

	// Subject is never nil. Use [auth.Subject.Audit] for permission handling.
	// You must not keep the identity instance over a long time, because
	// it will change over time, either due to refreshing tokens or because the user is logged out.
	Subject() auth.Subject

	// UpdateSubject sets the current subject.
	UpdateSubject(subject auth.Subject)

	// Context returns the wire-lifetime context. Contains additional injected types like User or Logger.
	Context() context.Context

	// SessionID is a unique identifier, which is assigned to a client using some sort of cookie mechanism. This is a
	// pure random string and belongs to a distinct client instance. It is shared across multiple windows on the client,
	// especially when using multiple tabs or activity windows. You may use this for authentication mechanics,
	// however be careful not to break external security concerns by never revisiting the actual user authentication
	// state.
	// It usually outlives a frontend process and e.g. is restored after a device restart.
	SessionID() SessionID

	// Authenticate triggers a round trip so that [Window.Subject] may contain a valid user afterward.
	// For sure, the user can always cancel that.
	Authenticate()

	// ViewRoot returns the access to the component which represents the current root.
	ViewRoot() ViewRoot

	// Locale returns the negotiated language tag or locale identifier between the frontend and the backend.
	Locale() language.Tag

	// Location returns negotiated time zone location
	Location() *time.Location
	// TODO add Locale, add Screen Metrics (density, pixel width and height, size classes etc)

	// SendFiles takes all contained files and tries to offer them to the user using whatever is native for the
	// actual frontend. For example, a browser may just download these files but an Android frontend may show
	// a _send multiple intent_. See also AsURI which does not trigger such intent.
	SendFiles(it iter.Seq2[File, error]) error

	// AsURI takes the open closure and provides a URI accessor for it. Whenever the URI is opened, the data
	// is returned from the open call. Note that open is usually not called from the event looper and the open call
	// must not modify your view tree. See also SendFiles to explicitly export binary into the user environment.
	AsURI(open func() (io.Reader, error)) (ora.URI, error)
}

type SessionID string

var _ Window = (*scopeWindow)(nil)

type scopeWindow struct {
	factory       ora.ComponentFactoryId
	scope         *Scope
	navController *NavigationController
	values        Values
	location      *time.Location
	viewRoot      *scopeViewRoot
	subject       auth.Subject
	mutex         sync.Mutex
	destroyed     bool
}

func (s *scopeWindow) UpdateSubject(subject auth.Subject) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.subject = subject
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

	s.subject = auth.InvalidSubject{}

	for _, observer := range s.scope.app.onWindowCreatedObservers {
		observer(s)
	}

	return s
}

func (s *scopeWindow) AsURI(open func() (io.Reader, error)) (ora.URI, error) {
	if s.destroyed {
		return "", nil
	}

	if callback := s.scope.app.onShareStream; callback != nil {
		return callback(s.scope, open)
	}

	return "", fmt.Errorf("no share stream platform adapter has been configured")
}

func (s *scopeWindow) SendFiles(it iter.Seq2[File, error]) error {
	if s.destroyed {
		return nil
	}

	if callback := s.scope.app.onSendFiles; callback != nil {
		return callback(s.scope, it)
	}

	return fmt.Errorf("no send files platform adapter has been configured")
}

func (s *scopeWindow) Execute(task func()) {
	if s.destroyed {
		return
	}

	s.scope.eventLoop.Post(task)
	s.scope.eventLoop.Tick()
}

func (s *scopeWindow) Navigation() *NavigationController {
	return s.navController
}

func (s *scopeWindow) Values() Values {
	return s.values
}

func (s *scopeWindow) Subject() auth.Subject {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.subject
}

func (s *scopeWindow) Context() context.Context {
	return s.scope.ctx
}

func (s *scopeWindow) SessionID() SessionID {
	return s.scope.sessionID
}

func (s *scopeWindow) Authenticate() {
	// TODO ????
}

func (s *scopeWindow) AuthenticatedObserver() {
	// TODO how to implement that? trigger navigation
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
