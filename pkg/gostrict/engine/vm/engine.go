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
// one with a JavaScript binding.
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

// failure returns the error of a failed call: the error of an abort itself, see [gostrict.Abort].
func (in *Instance) failure(err error) error {
	var p *vmcode.PanicError
	if errors.As(err, &p) {
		if abortErr, ok := gostrict.AbortedErr(p.Value); ok {
			return abortErr
		}
		if e, ok := p.Value.(error); ok {
			return fmt.Errorf("vm: %w", e)
		}
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
