// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vm

import (
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

func (b boundary) Callback(fn *vmcode.Func) any {
	return &callback{in: b.in, fn: fn}
}

// callback is a function of the program, which the host has received.
type callback struct {
	in *Instance
	fn *vmcode.Func
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
