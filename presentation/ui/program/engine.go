// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package program

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// Engine executes bundles in the browser of a window, see [gostrict.Engine], e.g. a program of a language model,
// whose computation should not burden the server. A call waits for the answer of the browser, so it must not be
// made on the event loop of the window, e.g. in a callback of a button, but in a goroutine of its own, and its
// context should have a deadline, because the browser may never answer, e.g. after the window was closed. The
// functions of header packages run as their bindings in the browser, so a host must not implement them.
type Engine struct {
	Window core.Window
}

var _ gostrict.Engine = Engine{}

var instanceIDs atomic.Int64

// Load registers the bundle and returns an instance, which the browser creates on the first call.
func (e Engine) Load(ctx context.Context, b *gostrict.Bundle, host gostrict.Funcs) (gostrict.Instance, error) {
	if len(host) > 0 {
		return nil, errors.New("the browser runs the bindings of header packages and no functions of the host")
	}
	id := "engine-" + strconv.FormatInt(instanceIDs.Add(1), 10)
	return &browserInstance{wnd: e.Window, b: b, uri: Register(b), id: id}, nil
}

type browserInstance struct {
	wnd core.Window
	b   *gostrict.Bundle
	uri core.URI
	id  string
	// mu serializes the calls like the event loop of a browser
	mu sync.Mutex
}

// browserRef is the value of a token of a state, which the instance in the browser holds.
type browserRef struct {
	instance string
	ref      int64
}

func (in *browserInstance) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	f, ok := in.b.Exports[name]
	if !ok {
		return nil, fmt.Errorf("%s is not an exported function", name)
	}
	if err := gostrict.CheckArgs(name, f.Params, args); err != nil {
		return nil, err
	}
	cd := codec{
		b: in.b,
		state: func(v any) (any, error) {
			ref, ok := v.(gostrict.Token).Value().(browserRef)
			if !ok || ref.instance != in.id {
				return nil, errors.New("the state belongs to another instance")
			}
			return map[string]any{"$r": ref.ref}, nil
		},
		token: func(ref int64, t gostrict.Type) any {
			return gostrict.NewToken(t.Name, browserRef{instance: in.id, ref: ref})
		},
	}
	wire, err := cd.encode(args, f.Params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}

	in.mu.Lock()
	defer in.mu.Unlock()

	ret := make(chan proto.CallRet, 1)
	cancel := core.AsyncCall(in.wnd, &proto.GostrictCall{Instance: proto.Str(in.id), Bundle: proto.URI(in.uri), Name: proto.Str(name), Args: proto.Str(wire)}, func(r proto.CallRet) {
		ret <- r
	})
	var r proto.CallRet
	select {
	case r = <-ret:
	case <-ctx.Done():
		cancel()
		return nil, ctx.Err()
	}

	res, ok := r.(*proto.GostrictReturn)
	if !ok {
		return nil, fmt.Errorf("unexpected answer %T of the browser", r)
	}
	if res.Failure != "" {
		return nil, fmt.Errorf("the browser failed to call %s: %s", name, res.Failure)
	}
	values, err := cd.decode(string(res.Values), f.Results)
	if err != nil {
		return nil, fmt.Errorf("the results of %s: %w", name, err)
	}
	return values, nil
}

// Bundle returns the bundle of the instance.
func (in *browserInstance) Bundle() *gostrict.Bundle {
	return in.b
}

// Close drops the instance in the browser.
func (in *browserInstance) Close() error {
	core.AsyncCall(in.wnd, &proto.GostrictCall{Instance: proto.Str(in.id), Bundle: proto.URI(in.uri)}, nil)
	return nil
}
