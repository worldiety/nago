// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"bytes"
	"fmt"
	"time"

	"go.wdy.de/nago/application/session"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// Open connects a new window as the given subject and allocates the root view of the given path, just like a
// browser which opens the according URL. If subject is nil, the window is anonymous.
// The returned window has already settled and is closed automatically at the end of the test.
func (a *App) Open(t TB, subject auth.Subject, path core.NavigationPath, opts ...OpenOption) *Window {
	t.Helper()

	w := newWindow(t, &localTransport{app: a}, path, opts)
	if subject != nil {
		a.subjects.Put(session.ID(w.opts.sessionID), subject)
	}

	w.open()
	return w
}

// localTransport connects a window to a scope within the same process.
type localTransport struct {
	app   *App
	scope *core.Scope
}

func (l *localTransport) connect(w *Window) error {
	l.scope = l.app.core.Connect(&channel{w: w, roundTrip: l.app.opts.roundTrip}, w.scopeID)
	l.app.windows.Put(w.scopeID, w)
	return l.send(w, &proto.SessionAssigned{SessionID: proto.Str(w.opts.sessionID)})
}

func (l *localTransport) send(w *Window, evt proto.NagoEvent) error {
	if l.app.opts.roundTrip {
		var err error
		if evt, err = roundTrip(evt); err != nil {
			return err
		}
	}

	return l.scope.Dispatch(evt)
}

func (l *localTransport) settle(w *Window) error {
	deadline := time.Now().Add(SettleTimeout)
	for {
		idle, ok := l.scope.FlushTimeout(max(time.Until(deadline), time.Millisecond))
		if idle {
			break
		}

		if !ok || time.Now().After(deadline) {
			return fmt.Errorf("window did not settle within %v: the event loop stays busy or blocks, a background task is still running or a view changes a state during each render", SettleTimeout)
		}

		// only background goroutines can keep us busy without posting to the event loop, thus don't spin
		time.Sleep(time.Millisecond)
	}

	return w.connError()
}

func (l *localTransport) upload(w *Window, id string, files []core.File) error {
	opts, ok := l.scope.ImportFilesOptions(id)
	if !ok {
		return fmt.Errorf("no file import requested with id %q", id)
	}

	// like the HTTP upload handler, the completion is invoked outside the event loop
	opts.OnCompletion(files)
	return nil
}

func (l *localTransport) close(w *Window) {
	if w.closed {
		// also remove the scope from the application, otherwise its update ticker keeps posting to it
		if !l.app.core.DestroyScope(l.scope.ID()) {
			l.scope.Destroy()
		}

		// wait for the destruction, but never hang the cleanup of a test on a blocked event loop
		if _, ok := l.scope.FlushTimeout(SettleTimeout); !ok {
			w.t.Logf("nagotest: the event loop of the closed window did not finish within %v", SettleTimeout)
		}
		l.app.windows.Delete(w.scopeID)
		return
	}

	// a reconnect just detaches the channel
	l.scope.Connect(nil)
}

// channel connects a scope with a window without encoding.
type channel struct {
	w         *Window
	roundTrip bool
}

func (c *channel) Subscribe(f func(msg []byte) error) (destroy func()) {
	return func() {}
}

func (c *channel) Publish(msg []byte) error {
	obj, err := proto.Unmarshal(proto.NewBinaryReader(bytes.NewBuffer(msg)))
	if err != nil {
		return err
	}

	evt, ok := obj.(proto.NagoEvent)
	if !ok {
		return unexpected(obj)
	}

	c.w.receive(evt)
	return nil
}

func (c *channel) PublishEvent(evt proto.NagoEvent) error {
	if c.roundTrip {
		var err error
		if evt, err = roundTrip(evt); err != nil {
			c.w.fail(err)
			return err
		}
	}

	c.w.receive(evt)
	return nil
}

// roundTrip encodes and decodes the event using the binary protocol.
func roundTrip(evt proto.NagoEvent) (proto.NagoEvent, error) {
	var buf bytes.Buffer
	if err := proto.Marshal(proto.NewBinaryWriter(&buf), evt); err != nil {
		return nil, fmt.Errorf("cannot encode %T: %w", evt, err)
	}

	obj, err := proto.Unmarshal(proto.NewBinaryReader(&buf))
	if err != nil {
		return nil, fmt.Errorf("cannot decode %T: %w", evt, err)
	}

	res, ok := obj.(proto.NagoEvent)
	if !ok {
		return nil, unexpected(obj)
	}

	return res, nil
}
