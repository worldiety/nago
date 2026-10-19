// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package vm executes bundles in an interpreter of register bytecode, which the compiler creates from the checked
// program next to the JavaScript, see [gostrict.Bundle.VM]. It needs no JavaScript engine: values keep their types,
// structs live in registers and the program and its package variables are shared by all instances.
package vm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/internal/vmcode"
)

// Engine runs bundles in the interpreter. The zero value is ready to use.
type Engine struct {
	// MaxCallDepth limits the depth of calls, which ends an endless recursion with an error, 10000 if zero.
	MaxCallDepth int
}

var _ gostrict.Engine = (*Engine)(nil)

// Load creates an instance of the bundle. Each function of a header package needs a function of the host, also
// one with a JavaScript binding, and each variable of [gostrict.Bundle.HostVars] a function without parameters,
// which returns its value.
func (e *Engine) Load(ctx context.Context, b *gostrict.Bundle, host gostrict.Funcs) (gostrict.Instance, error) {
	p, err := b.VM()
	if err != nil {
		return nil, err
	}
	in := &Instance{bundle: b, prog: p}
	sigs := map[string]gostrict.Func{}
	for _, x := range b.Externs {
		sigs[x.Name] = x.Func
	}
	for _, name := range p.Externs() {
		fn, ok := host[name]
		if !ok {
			return nil, fmt.Errorf("the host does not implement %s, which the vm needs", name)
		}
		in.host = append(in.host, hostFunc{sig: sigs[name], fn: fn})
	}
	in.lock()
	defer in.unlock()
	in.m, err = vmcode.NewMachine(ctx, p, boundary{in}, e.MaxCallDepth)
	if err != nil {
		return nil, in.failure(err)
	}
	return in, nil
}

type hostFunc struct {
	sig gostrict.Func
	fn  gostrict.Fn
}

// Instance is a loaded bundle. It owns only its stacks, the program and its package variables are shared.
type Instance struct {
	bundle *gostrict.Bundle
	prog   *vmcode.Program
	host   []hostFunc
	m      *vmcode.Machine
	mu     sync.Mutex
	// depth counts the calls, which hold mu, so that a callback within a call runs without locking again
	depth atomic.Int32
	// err is the first error of a callback since the last call, see [gostrict.Callback]
	err error
	// fastLocks tells for each fast call in progress, whether it locked the instance
	fastLocks []bool
}

// Call calls an exported function.
func (in *Instance) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	sig, ok := in.bundle.Exports[name]
	fn := in.prog.Export(name)
	if !ok || fn == nil {
		return nil, fmt.Errorf("%s is not an exported function", name)
	}
	if err := gostrict.CheckArgs(name, sig.Params, args); err != nil {
		return nil, err
	}
	in.lock()
	defer in.unlock()
	if err := in.callbackErr(); err != nil {
		return nil, err
	}
	res, err := in.m.Invoke(ctx, fn, args)
	if err != nil {
		return nil, in.failure(err)
	}
	return res, in.callbackErr()
}

// Run runs f, which calls the callbacks of the program, which the host has received, e.g. to render what a call has
// created, like a call with the context ctx: it ends the callbacks, which f calls, and f holds the instance. Run
// returns the first error of these callbacks. Without Run, a callback between two calls runs without a context.
func (in *Instance) Run(ctx context.Context, f func()) error {
	outer := in.depth.Load() == 0
	if outer {
		in.lock()
		defer in.unlock()
	}
	prev := in.m.SetContext(ctx)
	defer in.m.SetContext(prev)
	f()
	if !outer {
		// the error belongs to the call, which runs f, too
		return in.err
	}
	return in.callbackErr()
}

// Error is a failure of the program run, e.g. a panic of the program or of a function of the host, unlike an error
// of the API, e.g. an unknown export or wrong arguments, an ended context, an exhausted stack, see
// [gostrict.ErrStackExhausted], or an abort, see [gostrict.Abort].
type Error struct {
	Msg string
	// Err is the error of the failure, if any.
	Err error
}

func (e *Error) Error() string {
	return "vm: " + e.Msg
}

func (e *Error) Unwrap() error {
	return e.Err
}

// failure returns the error of a failed call: the error of an abort itself, see [gostrict.Abort], or an [Error] for a
// panic.
func (in *Instance) failure(err error) error {
	if errors.Is(err, vmcode.ErrDepth) {
		return fmt.Errorf("%w: %w", gostrict.ErrStackExhausted, err)
	}
	var p *vmcode.PanicError
	if errors.As(err, &p) {
		if abortErr, ok := gostrict.AbortedErr(p.Value); ok {
			return abortErr
		}
		e, _ := p.Value.(error)
		return &Error{Msg: fmt.Sprint(p.Value), Err: e}
	}
	return err
}

func (in *Instance) lock() {
	in.mu.Lock()
	in.depth.Add(1)
}

func (in *Instance) unlock() {
	in.depth.Add(-1)
	in.mu.Unlock()
}

// callbackErr returns and clears the error of a callback.
func (in *Instance) callbackErr() error {
	err := in.err
	in.err = nil
	if err != nil {
		return fmt.Errorf("a callback of the program failed: %w", err)
	}
	return nil
}

// Close releases the instance.
func (in *Instance) Close() error {
	return nil
}

// Bundle returns the bundle of the instance.
func (in *Instance) Bundle() *gostrict.Bundle {
	return in.bundle
}
