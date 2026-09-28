// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"fmt"
	"io"
	"sync"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// An App is a nago application running within the test process. It has no HTTP server and no websocket,
// instead each [Window] is connected directly to the core application.
type App struct {
	t         testing.TB
	cfg       *application.Configurator
	core      *core.Application
	opts      options
	subjects  concurrent.RWMap[session.ID, auth.Subject]
	windows   concurrent.RWMap[proto.ScopeID, *Window]
	mutex     sync.Mutex
	shared    map[core.URI]func() (io.Reader, error)
	sharedSeq int
}

type options struct {
	roundTrip bool
}

// Option configures an [App].
type Option func(*options)

// WithRoundTrip encodes and decodes every event in both directions using the binary protocol, as a real
// frontend would do. This is slower, but also verifies that the rendered trees survive the wire format.
func WithRoundTrip() Option {
	return func(o *options) {
		o.roundTrip = true
	}
}

// New configures a new nago application within the test process. The data directory is a temporary directory
// of the test, thus each App starts with empty stores. Environment variables and .env files are not loaded.
// The application is destroyed automatically when the test and all its subtests complete.
func New(t testing.TB, configure func(cfg *application.Configurator), opts ...Option) *App {
	t.Helper()

	a := &App{t: t, shared: map[core.URI]func() (io.Reader, error){}}
	for _, opt := range opts {
		opt(&a.opts)
	}

	a.cfg = application.NewConfigurator()
	a.cfg.SetDataDir(t.TempDir())
	configure(a.cfg)

	app, err := a.cfg.NewInProcessApplication()
	if err != nil {
		t.Fatalf("nagotest: cannot create application: %v", err)
	}

	a.core = app
	t.Cleanup(app.Destroy)

	// runs after all configured observers, e.g. the session management, which would otherwise reset the subject
	app.AddOnWindowCreatedObserver(func(wnd core.Window) {
		if subject, ok := a.subjects.Get(wnd.Session().ID()); ok {
			wnd.UpdateSubject(subject)
		}
	})

	app.SetOnSendFiles(func(scope *core.Scope, options core.ExportFilesOptions) error {
		w, ok := a.windows.Get(scope.ID())
		if !ok {
			return fmt.Errorf("nagotest: no window for scope %s", scope.ID())
		}

		w.mutex.Lock()
		defer w.mutex.Unlock()
		w.downloads = append(w.downloads, options)
		return nil
	})

	app.SetOnShareStream(func(scope *core.Scope, open func() (io.Reader, error)) (core.URI, error) {
		a.mutex.Lock()
		defer a.mutex.Unlock()

		a.sharedSeq++
		uri := core.URI(fmt.Sprintf("nagotest://share/%d", a.sharedSeq))
		a.shared[uri] = open
		return uri, nil
	})

	return a
}

// Configurator returns the configurator, e.g. to access use cases for preparing or verifying test data.
func (a *App) Configurator() *application.Configurator {
	return a.cfg
}

// Core returns the core application.
func (a *App) Core() *core.Application {
	return a.core
}

// Shared opens the stream behind a URI, which has been created by [core.Window.AsURI].
func (a *App) Shared(uri core.URI) (io.Reader, error) {
	a.mutex.Lock()
	open, ok := a.shared[uri]
	a.mutex.Unlock()

	if !ok {
		return nil, fmt.Errorf("nagotest: unknown shared uri %s", uri)
	}

	return open()
}
