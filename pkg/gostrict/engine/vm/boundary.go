// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm

import (
	"context"
	"errors"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/internal/vmcode"
)

// boundary provides the host functions and the neutral values of package gostrict to the machine of an instance.
type boundary struct {
	in *Instance
}

var _ vmcode.Host = boundary{}

func (b boundary) Call(idx int, args []any) []any {
	h := b.in.host[idx]
	res := h.fn(args)
	if h.sig.Name == "" {
		// a variable, whose value the machine checks itself
		return res
	}
	if err := gostrict.CheckArgs(h.sig.Name, h.sig.Results, res); err != nil {
		panic(err)
	}
	return res
}

func (b boundary) Variant(tag string, fields []any) any {
	return gostrict.Variant{Tag: tag, Fields: fields}
}

func (b boundary) VariantOf(v any) (string, []any) {
	x := v.(gostrict.Variant)
	return x.Tag, x.Fields
}

func (b boundary) Token(name string, state any) any {
	return gostrict.NewToken(name, state)
}

func (b boundary) TokenValue(v any) (any, bool) {
	tok, ok := v.(gostrict.Token)
	if !ok {
		return nil, false
	}
	return tok.Value(), true
}

func (b boundary) Returned(v any) (error, bool) {
	if r, ok := v.(*gostrict.ResultError); ok {
		return r.Err, true
	}
	return nil, false
}

func (b boundary) Abort(err error) any {
	return &gostrict.Aborted{Err: err}
}

func (b boundary) ProgramError(msg string, pos int) error {
	return b.in.bundle.ProgramError(msg, pos)
}

func (b boundary) Callback(fn *vmcode.Closure) any {
	return &callback{in: b.in, fn: fn}
}

// callback is a function of the program, which the host has received.
type callback struct {
	in *Instance
	fn *vmcode.Closure
}

func (cb *callback) Call(args ...any) ([]any, error) {
	in := cb.in
	if in.depth.Load() == 0 {
		// between two calls, e.g. when a window renders, what a call has created
		in.lock()
		defer in.unlock()
	}
	if err := cb.fn.CheckArgs(args); err != nil {
		if in.err == nil {
			in.err = err
		}
		return nil, err
	}
	res, err := in.m.Invoke(in.m.Context(), cb.fn, args)
	if err != nil {
		err = in.failure(err)
		if in.err == nil {
			in.err = err
		}
	}
	return res, err
}

var _ gostrict.Detachable = (*callback)(nil)

// Detach returns the closure without its instance: any machine of the program can run it, because its captures never
// change, see [gostrict.Detached].
func (cb *callback) Detach() gostrict.Detached {
	return detached{prog: cb.in.prog, fn: cb.fn}
}

type detached struct {
	prog *vmcode.Program
	fn   *vmcode.Closure
}

// Home returns nil, because any instance of the bundle runs the closure.
func (d detached) Home() gostrict.Instance {
	return nil
}

// CallIn runs the closure in the instance.
func (d detached) CallIn(ctx context.Context, target gostrict.Instance, args ...any) ([]any, error) {
	in, ok := target.(*Instance)
	if !ok || in.prog != d.prog {
		return nil, errors.New("the function belongs to another bundle or engine")
	}
	if err := d.fn.CheckArgs(args); err != nil {
		return nil, err
	}
	in.lock()
	defer in.unlock()
	res, err := in.m.Invoke(ctx, d.fn, args)
	if err != nil {
		return nil, in.failure(err)
	}
	return res, nil
}

var _ gostrict.FastCallback = (*callback)(nil)

// Begin begins a fast call, see [gostrict.FastCallback].
func (cb *callback) Begin() gostrict.Writer {
	in := cb.in
	locked := in.depth.Load() == 0
	if locked {
		in.lock()
	}
	in.fastLocks = append(in.fastLocks, locked)
	return writer{in.m.BeginFast(in.m.Context(), cb.fn)}
}

// Invoke runs the fast call.
func (cb *callback) Invoke() (gostrict.Reader, error) {
	in := cb.in
	r, err := in.m.InvokeFast()
	if err != nil {
		err = in.failure(err)
		if in.err == nil {
			in.err = err
		}
		return nil, err
	}
	return reader{r}, nil
}

// End ends the fast call.
func (cb *callback) End() {
	in := cb.in
	in.m.EndFast()
	locked := in.fastLocks[len(in.fastLocks)-1]
	in.fastLocks = in.fastLocks[:len(in.fastLocks)-1]
	if locked {
		in.unlock()
	}
}

type reader struct {
	*vmcode.Reader
}

func (r reader) State() gostrict.Token {
	return r.StateValue().(gostrict.Token)
}

func (r reader) Func() gostrict.Callback {
	return r.FuncValue().(gostrict.Callback)
}

type writer struct {
	*vmcode.Writer
}

func (w writer) State(t gostrict.Token) {
	w.StateValue(t)
}
