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
	"reflect"
	"slices"
	"sync"
)

// Pool distributes calls over several instances of a bundle, so that concurrent callers, e.g. the requests of a
// server, do not wait for each other. The package variables of a program never change, so the instances are
// interchangeable for functions without a state. A state, see [ValueState], belongs to the instance, which created
// it, so the pool calls a function, which takes a state, in that instance: it is sticky. The calls of a state are
// serialized, the calls of different instances run concurrently. [Pool.Release] forgets a state, which is no longer
// used.
type Pool struct {
	free  chan int
	all   []Instance
	locks []sync.Mutex

	mu sync.Mutex
	// owners maps the value of a token to the index of the instance, which created it
	owners map[any]int
}

var _ Instance = (*Pool)(nil)

// NewPool loads size instances of the bundle.
func NewPool(ctx context.Context, e Engine, b *Bundle, host Funcs, size int) (*Pool, error) {
	if size < 1 {
		return nil, errors.New("a pool needs at least one instance")
	}
	p := &Pool{free: make(chan int, size), locks: make([]sync.Mutex, size), owners: map[any]int{}}
	for i := range size {
		in, err := e.Load(ctx, b, host)
		if err != nil {
			_ = p.Close()
			return nil, err
		}
		p.all = append(p.all, in)
		p.free <- i
	}
	return p, nil
}

// Call calls the function in the instance, which owns the state among the arguments, or in a free instance.
func (p *Pool) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	if i, ok := p.owner(args); ok {
		return p.callIn(ctx, i, name, args)
	}

	var i int
	select {
	case i = <-p.free:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { p.free <- i }()
	return p.callIn(ctx, i, name, args)
}

func (p *Pool) callIn(ctx context.Context, i int, name string, args []any) ([]any, error) {
	p.locks[i].Lock()
	defer p.locks[i].Unlock()
	res, err := p.all[i].Call(ctx, name, args...)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, r := range res {
		if tok, ok := r.(Token); ok && comparable(tok.v) {
			p.owners[tok.v] = i
		}
	}
	return res, nil
}

func (p *Pool) owner(args []any) (int, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, a := range args {
		if tok, ok := a.(Token); ok && comparable(tok.v) {
			if i, ok := p.owners[tok.v]; ok {
				return i, true
			}
		}
	}
	return 0, false
}

// Release forgets the state, after the caller has dropped it, so that the pool does not keep it.
func (p *Pool) Release(tok Token) {
	if !comparable(tok.v) {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.owners, tok.v)
}

func comparable(v any) bool {
	return v != nil && reflect.TypeOf(v).Comparable()
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
