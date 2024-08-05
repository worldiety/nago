package core

import (
	"context"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/ora" // TODO remove this protocol leak from public contract
	"golang.org/x/text/language"
	"io"
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

	// Locale returns the negotiated language tag or locale identifier between the frontend and the backend.
	Locale() language.Tag

	// Location returns negotiated time zone location.
	Location() *time.Location

	// Info returns the screen metrics.
	Info() WindowInfo

	// SendFiles takes all contained files and tries to offer them to the user using whatever is native for the
	// actual frontend. For example, a browser may just download these files but an Android frontend may show
	// a _send multiple intent_. See also AsURI which does not trigger such intent.
	SendFiles(it iter.Seq2[File, error]) error

	// TODO
	// AsURI takes the open closure and provides a URI accessor for it. Whenever the URI is opened, the data
	// is returned from the open call. Note that open is usually not called from the event looper and the open call
	// must not modify your view tree. See also SendFiles to explicitly export binary into the user environment.
	AsURI(open func() (io.Reader, error)) (URI, error)

	// Application returns the parent application.
	Application() *Application

	// Factory returns the current active factory.
	Factory() ora.ComponentFactoryId // TODO remove this protocol leak from public contract

	// AddDestroyObserver registers an observer which is called, before the root component of the window is destroyed.
	AddDestroyObserver(fn func()) (removeObserver func())

	// Invalidate renders the tree and sends it to the actual frontend for displaying. Usually you should not use
	// this directly, because the request-response cycles triggers this automatically. However, if backend
	// data has changed due to other domain events, you have to notify the view tree to redraw and potentially
	// to load the data again from repositories. In those cases you likely want to use [core.Iterable.Iter] to
	// always rebuild the entire tree from the according property.
	Invalidate()

	// Once executes the given function only once for this window. The identity is asserted by the given id.
	Once(id string, fn func())
}

type SessionID string

// Colors returns a type safe value based ColorSet instance.
func Colors[CS ColorSet](wnd Window) CS {
	var zero CS
	scope := wnd.(*scopeWindow)

	scheme := scope.parent.windowInfo.ColorScheme
	colors, ok := scope.parent.app.colorSets[scheme]
	if !ok {
		slog.Error("could not find color set for scheme", "scheme", scheme)
		return zero.Default(scheme).(CS)
	}

	set, ok := colors[zero.Namespace()]
	if !ok {
		slog.Error("could not find color set for namespace", "scheme", scheme, "namespace", zero.Namespace())
		return zero.Default(scheme).(CS)
	}

	return set.(CS)
}
