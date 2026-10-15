// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"context"
	"fmt"
	gotoken "go/token"
	"go/types"
	"slices"
	"sync"
)

// Message is a record, which a program and its backend exchange, see package std/msg.
type Message struct {
	// Type is the full name of the record, e.g. "example.com/arena.Hit".
	Type string
	// Fields are the fields of the record in the order of its [Record].
	Fields []any
}

// isHandler tells, whether the exported function handles messages: it takes a state and a record and returns
// nothing or an outbox.
func isHandler(f Func) bool {
	if len(f.Params) != 2 || f.Params[0].Kind != ValueState || (f.Params[1].Kind != ValueRecord && f.Params[1].Kind != ValueUnion) {
		return false
	}
	return len(f.Results) == 0 || (len(f.Results) == 1 && f.Results[0].Kind == ValueOut)
}

// handlers maps each message to its handler. A handler of a union handles each of its variants.
func handlers(exports map[string]Func, unions map[string]Union) map[string]string {
	res := map[string]string{}
	for _, name := range sortedNames(exports) {
		f := exports[name]
		if !isHandler(f) {
			continue
		}
		if u, ok := unions[f.Params[1].Name]; ok {
			for _, v := range u.Variants {
				res[v] = name
			}
			continue
		}
		res[f.Params[1].Name] = name
	}
	return res
}

// checkHandlers reports two handlers of the same message.
func checkHandlers(p *Package, d *diagnoser) {
	bnd := newBoundary()
	seen := map[string]string{}
	// in the order of the source, so that the later declaration is reported
	var funcs []*types.Func
	for _, name := range p.Types.Scope().Names() {
		if fn, ok := p.Types.Scope().Lookup(name).(*types.Func); ok && fn.Exported() {
			funcs = append(funcs, fn)
		}
	}
	slices.SortFunc(funcs, func(a, b *types.Func) int { return int(a.Pos() - b.Pos()) })
	for _, fn := range funcs {
		name := fn.Name()
		f, ok := bnd.exportSignature(name, fn.Signature())
		if !ok || !isHandler(f) {
			continue
		}
		msgs := []string{f.Params[1].Name}
		if u, ok := bnd.unions[f.Params[1].Name]; ok {
			msgs = u.Variants
		}
		for _, msg := range msgs {
			if other, ok := seen[msg]; ok {
				pos := fn.Pos()
				d.report(pos, pos+gotoken.Pos(len(name)), CodeHandler, "both "+other+" and "+name+" handle the message "+msg, "the host passes each message to exactly one function", "handle the message in one function")
				break
			}
			seen[msg] = name
		}
	}
}

// Session couples an instance of a program and its state with the backend. It passes each message of the backend
// to its handler and sends the messages, which the program returns, see package std/msg. A session serializes its
// calls, so the program handles one message after the other, like on the event loop of a browser.
type Session struct {
	mu    sync.Mutex
	in    Instance
	b     *Bundle
	state Token
	send  func(ctx context.Context, m Message) error
}

// NewSession returns a session of the instance with the state, which an exported function has returned. The
// session calls send with each message of the program.
func NewSession(in Instance, b *Bundle, state Token, send func(ctx context.Context, m Message) error) *Session {
	return &Session{in: in, b: b, state: state, send: send}
}

// State returns the token of the state of the program.
func (s *Session) State() Token {
	return s.state
}

// Deliver passes a message of the backend to its handler, which the program may receive at any time, and sends
// the messages, which the handler returns.
func (s *Session) Deliver(ctx context.Context, m Message) error {
	name, ok := s.b.Handlers[m.Type]
	if !ok {
		return fmt.Errorf("the program does not handle the message %s", m.Type)
	}
	var arg any = m.Fields
	if s.b.Exports[name].Params[1].Kind == ValueUnion {
		arg = Variant{Tag: m.Type, Fields: m.Fields}
	}
	_, err := s.Call(ctx, name, arg)
	return err
}

// Call calls an exported function with the state as its first argument, if it takes one, e.g. the frame of an
// animation, and sends the messages, which it returns. It returns the other results.
func (s *Session) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	f, ok := s.b.Exports[name]
	if !ok {
		return nil, fmt.Errorf("%s is not an exported function", name)
	}
	if len(f.Params) > 0 && f.Params[0].Kind == ValueState && f.Params[0].Name == s.state.Type() {
		args = append([]any{s.state}, args...)
	}

	s.mu.Lock()
	res, err := s.in.Call(ctx, name, args...)
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}

	// the messages are sent after the call, so that the backend may answer with a message right away
	var rest []any
	var out []Message
	for i, t := range f.Results {
		if t.Kind != ValueOut {
			rest = append(rest, res[i])
			continue
		}
		for _, e := range res[i].([]any) {
			if v, ok := e.(Variant); ok {
				out = append(out, Message{Type: v.Tag, Fields: v.Fields})
				continue
			}
			out = append(out, Message{Type: t.Elem.Name, Fields: e.([]any)})
		}
	}
	for _, m := range out {
		if err := s.send(ctx, m); err != nil {
			return rest, err
		}
	}
	return rest, nil
}
