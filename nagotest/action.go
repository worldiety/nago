// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package nagotest

import (
	"time"

	"go.wdy.de/nago/presentation/proto"
)

// clickFields are the callback fields a frontend invokes on click, in order of precedence.
var clickFields = []string{"Action", "OnClick"}

// Click clicks the single selected node, like a user. As in a browser, the click bubbles up to the nearest
// ancestor with an action, thus selecting the text of a button is fine. The test fails, if nothing is
// clickable or if the clicked component or any ancestor is disabled or invisible. The window settles afterward.
func (w *Window) Click(s Selection) {
	w.t.Helper()

	start := time.Now()
	w.act(s, func(s Selection) proto.NagoEvent {
		n := s.Node()
		ptr, target := callbackOf(n, clickFields...)
		if ptr == 0 {
			w.t.Fatalf("nagotest: %s: %T is not clickable", s.matcher, n.Component)
		}

		w.assertInteractive(s, target)
		return &proto.FunctionCallRequested{Ptr: ptr, RID: w.nextRID()}
	})
	w.observe("click", s.matcher.desc, start)
}

// PressEnter presses the enter key within the single selected text or password field.
func (w *Window) PressEnter(s Selection) {
	w.t.Helper()

	start := time.Now()
	w.act(s, func(s Selection) proto.NagoEvent {
		n := s.Node()
		ptr := ptrField(n.Component, "KeydownEnter")
		if ptr == 0 {
			w.t.Fatalf("nagotest: %s: %T has no enter key handler", s.matcher, n.Component)
		}

		w.assertInteractive(s, n)
		return &proto.FunctionCallRequested{Ptr: ptr, RID: w.nextRID()}
	})
	w.observe("enter", s.matcher.desc, start)
}

// Type sets the value of the single selected input component, e.g. a text field. Other inputs use the
// string representation of their state, e.g. "true" for a toggle. The window settles afterward.
func (w *Window) Type(s Selection, value string) {
	w.t.Helper()

	start := time.Now()
	w.act(s, func(s Selection) proto.NagoEvent {
		n := s.Node()
		ptr := ptrField(n.Component, "InputValue")
		if ptr == 0 {
			w.t.Fatalf("nagotest: %s: %T has no input binding", s.matcher, n.Component)
		}

		w.assertInteractive(s, n)
		return &proto.UpdateStateValueRequested{StatePointer: ptr, Value: proto.Str(value), RID: w.nextRID()}
	})
	w.observe("type", s.matcher.desc, start)
}

// act sends the event for the single selected node and settles. If the backend discarded the event, because
// the window rendered a newer tree in the meantime, the selection is resolved again within the latest tree and
// the event is sent again, see [StaleRetries]. A stale event is never executed, so repeating it is safe.
func (w *Window) act(s Selection, event func(s Selection) proto.NagoEvent) {
	w.t.Helper()

	for attempt := 0; ; attempt++ {
		evt := event(s)
		before := w.staleCalls()
		w.dispatch(evt)
		w.Settle()

		if w.staleCalls() == before {
			return
		}

		if attempt >= w.opts.staleRetries {
			w.t.Fatalf("nagotest: %s: the backend discarded the action %d times as stale, because the window rendered again in between. Does a view change a state during each render?", s.matcher, attempt+1)
		}

		s = s.refresh()
	}
}

// staleCalls returns the amount of stale requests discarded by the backend. Through a websocket, they cannot be
// detected and 0 is returned.
func (w *Window) staleCalls() int64 {
	if tr, ok := w.tr.(*localTransport); ok && tr.scope != nil {
		return tr.scope.StaleCalls()
	}

	return 0
}

func (w *Window) assertInteractive(s Selection, n Node) {
	w.t.Helper()

	if c := firstOf(n, "Disabled"); c != nil {
		w.t.Fatalf("nagotest: %s: cannot interact, %T is disabled", s.matcher, c)
	}

	if c := firstOf(n, "Invisible"); c != nil {
		w.t.Fatalf("nagotest: %s: cannot interact, %T is invisible", s.matcher, c)
	}
}

// callbackOf returns the first callback pointer of the node or its nearest ancestor together with the node
// which owns it.
func callbackOf(n Node, fields ...string) (proto.Ptr, Node) {
	for i := len(n.Parents); i >= 0; i-- {
		c := n.Component
		if i < len(n.Parents) {
			c = n.Parents[i]
		}

		for _, f := range fields {
			if ptr := ptrField(c, f); ptr != 0 {
				return ptr, Node{Component: c, Parents: n.Parents[:i]}
			}
		}
	}

	return 0, Node{}
}
