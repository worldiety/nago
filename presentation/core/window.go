package core

import (
	"context"
	"go.wdy.de/nago/application/session"
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

func ExportFileBytes(name string, buf []byte) ExportFilesOptions {
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

// A Window owns the lifecycle of a component and is part of a Scope.
// Ora does not define (yet) what a window is.
// However, obviously every component lives inside a window of the frontend and navigation is related to that.
// Today, a frontend client must treat a scope as a window.
type Window interface {

	// Navigation allows access to the window navigation.
	Navigation() Navigation

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

	// Session returns access to the technical client identity. This identity is usually assigned by the server
	// using a cookie mechanics. It is the same for all browser windows and tabs of the same browser instance.
	//
	// Note, that this is different from the allocated scope which is connected over the wire (usually a websocket).
	// A scope is uniquely assigned either to none or exact one client instance (usually a browser tab).
	// If you hit the refresh button in a browser, the scope id is lost and the client will (re-)connect using
	// a new unique random scope id.
	Session() session.UserSession

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

	// SetColorScheme requests that the frontend changes the theme or color scheme, e.g. to dark or light mode.
	// The frontend may ignore or just not support a specific theme.
	SetColorScheme(ColorScheme)

	// Application returns the parent application.
	Application() *Application

	// Path returns the current active navigation path.
	Path() NavigationPath

	// AddDestroyObserver registers an observer which is called, before the root component of the window is destroyed.
	AddDestroyObserver(fn func()) (removeObserver func())

	// Clipboard gives access to the frontend clipboard.
	Clipboard() Clipboard
}

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
