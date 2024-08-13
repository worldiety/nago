package core

import (
	"context"
	"go.wdy.de/nago/auth"
	"golang.org/x/text/language"
	"log/slog"
	"time"
)

type ImportFilesOptions struct {
	// ID of your import request. If empty, an automatic ID based on your structure
	// is created, which may work. If your structure changes between renderings,
	// consider using a unique static ID.
	ID       string
	Multiple bool
	// MaxBytes is a recommendation and an attacker may ignore that, so don't
	// count on that for security reasons.
	MaxBytes         int64
	AllowedMimeTypes []string
	// OnCompletion may be performance optimized and thus may be called from
	// a non-ui goroutine. Thus, be careful of data races, when modifying your state.
	// Note, that keeping the file reference is illegal and may cause anything from resource
	// leaks to mal functions. Process data (e.g. by copying) within the completion handler entirely.
	OnCompletion func(files []File)
}

type ExportFilesOptions struct {
	// ID of your export request. If empty, an automatic ID based on your structure
	// is created, which may work. If your structure changes between renderings,
	// consider using a unique static ID.
	ID string
	// Files is the pull way of sending a file.
	Files []File
}

func ExportFile(name string, buf []byte) ExportFilesOptions {
	return ExportFilesOptions{
		ID: name,
		Files: []File{
			MemFile{
				Filename: name,
				Bytes:    buf,
			},
		},
	}
}

type RootViewFactory string

// A Window owns the lifecycle of a component and is part of a Scope.
// Ora does not define (yet) what a window is.
// However, obviously every component lives inside a window of the frontend and navigation is related to that.
// Today, a frontend client must treat a scope as a window.
type Window interface {

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

	// ExportFiles takes all contained files and tries to offer them to the user using whatever is native for the
	// actual frontend. For example, a browser may just download these files but an Android frontend may show
	// a _send multiple intent_. See also AsURI which does not trigger such intent and is used
	// to stream data into the frontend.
	ExportFiles(options ExportFilesOptions)

	// ImportFiles is the opposite of SendFiles. The identity of the request is
	// derived by the given identifier. If the ID is empty, a structural identifier is
	// automatically created.
	ImportFiles(options ImportFilesOptions)

	// Application returns the parent application.
	Application() *Application

	// Factory returns the current active factory.
	Factory() RootViewFactory

	// AddDestroyObserver registers an observer which is called, before the root component of the window is destroyed.
	AddDestroyObserver(fn func()) (removeObserver func())
}

// Execute posts the task into the associated Executor. It will be executed in the associated event loop
// to allow race free processing.
// Use this to post updates from foreign goroutines into the ui components.
// Note, that an invalidation is not triggered automatically. Either use [ViewRoot.Invalidate] manually or
// even better use [Post], [PostDelayed] or [Schedule], because those have optimized lifecycle handling.
//Execute(task func())

// Invalidate enforces a render cycle sometimes in the future.
// Usually you should not use
// this directly, because the request-response cycles triggers this automatically. However, if backend
// data has changed due to other domain events, you have to notify the view tree to redraw and potentially
// to load the data again from repositories.
//Invalidate()

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
