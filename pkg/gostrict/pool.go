// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

// Pool distributes calls over several instances of a bundle, so that concurrent callers, e.g. the requests of a
// server, do not wait for each other. The package variables of a program never change, so the instances are
// interchangeable. A state, see [ValueState], belongs to the instance, which created it, so a pool serves only
// functions without states.
type Pool struct {
	free chan Instance
	all  []Instance
}

var _ Instance = (*Pool)(nil)

// NewPool loads size instances of the bundle.
func NewPool(ctx context.Context, e Engine, b *Bundle, host Funcs, size int) (*Pool, error) {
	if size < 1 {
		return nil, errors.New("a pool needs at least one instance")
	}
	p := &Pool{free: make(chan Instance, size)}
	for range size {
		in, err := e.Load(ctx, b, host)
		if err != nil {
			_ = p.Close()
			return nil, err
		}
		p.all = append(p.all, in)
		p.free <- in
	}
	return p, nil
}

// Call waits for a free instance and calls the function in it.
func (p *Pool) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	var in Instance
	select {
	case in = <-p.free:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { p.free <- in }()
	return in.Call(ctx, name, args...)
}

// Close closes all instances.
func (p *Pool) Close() error {
	var errs []error
	for _, in := range p.all {
		errs = append(errs, in.Close())
	}
	return errors.Join(errs...)
}

// Expect reports, whether the bundle exports the function with the given signature. An application requires it
// of a program, e.g. of a script, which a language model has written, before it calls the function.
func (b *Bundle) Expect(want Func) error {
	got, ok := b.Exports[want.Name]
	if !ok {
		return fmt.Errorf("the program does not export %s", want)
	}
	equal := func(a, b []Type) bool { return slices.EqualFunc(a, b, Type.Equal) }
	if !equal(got.Params, want.Params) || !equal(got.Results, want.Results) {
		return fmt.Errorf("the program exports %s instead of %s", got, want)
	}
	return nil
}
