// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"mime"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// File creates an in-memory file for [Window.Upload]. The mime type is derived from the file extension.
func File(name string, buf []byte) core.File {
	return core.MemFile{
		Filename:     name,
		MimeTypeHint: mime.TypeByExtension(filepath.Ext(name)),
		Bytes:        buf,
	}
}

// Upload completes the file import request with the given ID, like a user who picks the files in the
// browser dialog. See also [core.Window.ImportFiles] and [Window.Imports]. The window settles afterward.
func (w *Window) Upload(id string, files ...core.File) {
	w.t.Helper()

	opts, ok := w.scope.ImportFilesOptions(id)
	if !ok {
		w.t.Fatalf("nagotest: no file import requested with id %q", id)
	}

	if !opts.Multiple && len(files) > 1 {
		w.t.Fatalf("nagotest: file import %q accepts only a single file", id)
	}

	// like the HTTP upload handler, the completion is invoked outside the event loop
	opts.OnCompletion(files)
	w.Settle()
}

type listener struct {
	ptr  proto.Ptr
	call *proto.RegisterInputEventListener
}

// Input delivers the input event to all listeners registered for the element with [core.Window.AddInputListener]
// which accept its type. The test fails if there is none. The window settles afterward.
func (w *Window) Input(elemID string, evt core.InputEvent) {
	w.t.Helper()

	w.mutex.Lock()
	var targets []listener
	for _, l := range w.listeners {
		if string(l.call.Id) == elemID && slices.Contains(l.call.Types, proto.InputEventType(evt.Type)) {
			targets = append(targets, l)
		}
	}
	w.mutex.Unlock()

	if len(targets) == 0 {
		w.t.Fatalf("nagotest: no input listener for %q accepts event type %v", elemID, evt.Type)
	}

	slices.SortFunc(targets, func(a, b listener) int { return int(a.ptr) - int(b.ptr) })
	for _, l := range targets {
		w.dispatch(&proto.CallResolved{
			CallPtr: l.ptr,
			RID:     w.nextRID(),
			Ret: &proto.InputEvent{
				Handle: l.call.Handle,
				Type:   proto.InputEventType(evt.Type),
				X:      proto.Float(evt.X),
				Y:      proto.Float(evt.Y),
				Code:   proto.Str(evt.Code),
			},
		})
	}

	w.Settle()
}

// A Canvas gives access to the drawing commands of a canvas element, see the ui/canvas package.
type Canvas struct {
	w  *Window
	id string
}

// Canvas returns the canvas with the given ID.
func (w *Window) Canvas(id string) *Canvas {
	return &Canvas{w: w, id: id}
}

// Commands returns the drawing commands, e.g. *proto.CanvasFillRect, which have been issued for this canvas
// since the last call of Commands.
func (c *Canvas) Commands() []proto.CallArgs {
	c.w.mutex.Lock()
	defer c.w.mutex.Unlock()

	var res []proto.CallArgs
	for _, call := range c.w.calls[c.w.canvasSeen[c.id]:] {
		if strings.HasPrefix(reflect.Indirect(reflect.ValueOf(call.Call)).Type().Name(), "Canvas") && strField(call.Call, "Id") == c.id {
			res = append(res, call.Call)
		}
	}

	if c.w.canvasSeen == nil {
		c.w.canvasSeen = map[string]int{}
	}
	c.w.canvasSeen[c.id] = len(c.w.calls)

	return res
}

// Mount emulates that the frontend (re)mounts the canvas, which loses all display lists. Like the frontend,
// it notifies the listeners which requested [core.InputEventInvalidate], so that the application redraws.
// The window settles afterward.
func (c *Canvas) Mount() {
	c.w.t.Helper()

	c.w.Input(c.id, core.InputEvent{Type: core.InputEventInvalidate})
}
