// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// Channel connects a program, which a component of a window runs in the frontend, with its backend in Go. The
// backend sends messages to the handlers of the program with [Channel.Send] at any time, e.g. from a use case or
// another window, and receives the messages, which the exported functions of the program return, on the event
// loop of the window, see [Channel.OnReceive]. The messages of each direction keep their order. A message is
// delivered at most once: if the connection breaks, the messages in transit are lost, and a program, which needs
// them, asks its backend for the state again, e.g. in its Init function.
type Channel struct {
	wnd    core.Window
	id     string
	mu     sync.Mutex
	bundle *gostrict.Bundle
	recv   func(ctx context.Context, e gostrict.Envelope) error
	cancel func()
}

const channelState = "nago.program.channel."

// ChannelOf returns the channel of the component with the id in the window. Call it in each render, e.g. through
// the component, because the channel ends with the window or, when the component is no longer rendered, with its
// state.
func ChannelOf(wnd core.Window, id string) *Channel {
	st := core.StateOf[*Channel](wnd, channelState+id)
	st.Init(func() *Channel {
		c := &Channel{wnd: wnd, id: id}
		c.cancel = core.AsyncListen(wnd, &proto.GostrictListen{Id: proto.Str(id)}, c.received)
		st.AddDestroyObserver(c.cancel)
		return c
	})
	return st.Get()
}

// ID returns the id of the component.
func (c *Channel) ID() string {
	return c.id
}

// OnReceive sets the function, which handles the messages of the program, usually a router with the typed
// functions of the backend:
//
//	router := new(gostrict.Router)
//	gostrict.Handle(router, func(ctx context.Context, m messages.Clicked) error { ... })
//	ch.OnReceive(router.Dispatch)
func (c *Channel) OnReceive(fn func(ctx context.Context, e gostrict.Envelope) error) *Channel {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recv = fn
	return c
}

// attach sets the bundle, whose schema decodes the messages of the program.
func (c *Channel) attach(b *gostrict.Bundle) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bundle = b
}

// Send passes records or variants of sealed interfaces to the handlers of the program. A program, whose types the
// backend knows only at runtime, e.g. one of a language model, gets a [gostrict.Message] with neutral values. It may
// be called from any goroutine.
func (c *Channel) Send(values ...any) error {
	c.mu.Lock()
	b := c.bundle
	c.mu.Unlock()
	if b == nil {
		return fmt.Errorf("the component %s has no program yet", c.id)
	}

	msgs := make(proto.GostrictMessages, 0, len(values))
	for _, v := range values {
		// a program, which the backend knows only at runtime, gets its messages as neutral values
		m, ok := v.(gostrict.Message)
		if !ok {
			var err error
			if m, err = gostrict.MessageOf(v); err != nil {
				return err
			}
		}
		if _, ok := b.Handlers[m.Type]; !ok {
			return fmt.Errorf("the program does not handle the message %s", m.Type)
		}
		src, err := codec{b: b}.encodeRecord(m.Type, m.Fields)
		if err != nil {
			return fmt.Errorf("message %s: %w", m.Type, err)
		}
		msgs = append(msgs, proto.GostrictMessage{Type: proto.Str(m.Type), Json: proto.Str(src)})
	}
	core.AsyncCall(c.wnd, &proto.GostrictDeliver{Id: proto.Str(c.id), Messages: msgs}, nil)
	return nil
}

// received decodes the messages of the program and passes them to the backend.
func (c *Channel) received(ret proto.CallRet) {
	r, ok := ret.(*proto.GostrictReceived)
	if !ok {
		return
	}

	c.mu.Lock()
	b, recv := c.bundle, c.recv
	c.mu.Unlock()
	if b == nil || recv == nil {
		slog.Warn("the program sent messages, which no backend receives", "id", c.id, "count", len(r.Messages))
		return
	}

	cd := codec{b: b}
	for _, pm := range r.Messages {
		name := string(pm.Type)
		if !cd.isOutgoing(name) {
			slog.Error("the program sent a message, which it does not declare", "id", c.id, "type", name)
			continue
		}
		fields, err := cd.decodeRecord(name, string(pm.Json))
		if err != nil {
			slog.Error("cannot decode a message of the program", "id", c.id, "type", name, "err", err)
			continue
		}
		m := gostrict.Message{Type: name, Fields: fields}
		if err := recv(c.wnd.Context(), m); err != nil {
			slog.Error("the backend failed to handle a message of the program", "id", c.id, "type", m.Type, "err", err)
		}
	}
}
