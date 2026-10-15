// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package gojaengine executes bundles in goja, a JavaScript engine written in Go, e.g. on a server.
package gojaengine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/dop251/goja"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/internal/lru"
	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// Engine runs bundles in goja. It compiles each bundle once and shares the compiled program between runtimes.
// The zero value is ready to use.
type Engine struct {
	// MaxPrograms limits the cached programs, 256 if zero.
	MaxPrograms int
	// MaxCallStackSize limits the depth of calls, which ends an endless recursion with an error, 10000 if zero.
	MaxCallStackSize int

	once     sync.Once
	mu       sync.Mutex
	programs *lru.Cache[string, *goja.Program]
}

var _ gostrict.Engine = (*Engine)(nil)

// Error is an error of JavaScript, e.g. an exception of a binding, which a checked program itself never throws.
type Error struct {
	Msg string
}

func (e *Error) Error() string {
	return "javascript: " + e.Msg
}

func (e *Engine) program(b *gostrict.Bundle) (*goja.Program, error) {
	e.once.Do(func() {
		size := e.MaxPrograms
		if size == 0 {
			size = 256
		}
		e.programs = lru.New[string, *goja.Program](size)
	})

	e.mu.Lock()
	defer e.mu.Unlock()
	if p, ok := e.programs.Get(b.Hash); ok {
		return p, nil
	}
	p, err := goja.Compile("gostrict-"+b.Hash+".js", b.JS, true)
	if err != nil {
		return nil, fmt.Errorf("cannot compile bundle: %w", err)
	}
	e.programs.Put(b.Hash, p)
	return p, nil
}

// Load evaluates the bundle in a runtime of its own.
func (e *Engine) Load(ctx context.Context, b *gostrict.Bundle, host gostrict.Funcs) (gostrict.Instance, error) {
	return e.NewRuntime().Load(ctx, b, host)
}

// NewRuntime returns a runtime, which may hold several bundles, e.g. the views of a window. A program of the
// subset can not reach the globals of JavaScript, so the bundles can not interfere, while they share the memory of
// one runtime. The calls of the instances of a runtime are serialized.
func (e *Engine) NewRuntime() *Runtime {
	rt := goja.New()
	size := e.MaxCallStackSize
	if size == 0 {
		size = 10000
	}
	rt.SetMaxCallStackSize(size)

	r := &Runtime{engine: e, rt: rt}
	r.newMap = r.helper(`(entries) => new Map(entries)`)
	r.mapEntries = r.helper(`(m) => Array.from(m.entries())`)
	errFn, err := rt.RunString(`(s) => ({ v: { s: s }, m: { Error: (e) => e.s } })`)
	if err != nil {
		panic(err) // the expression is constant
	}
	r.newErr, _ = goja.AssertFunction(errFn)
	return r
}

// Runtime is a goja runtime, which holds instances of bundles.
type Runtime struct {
	engine     *Engine
	mu         sync.Mutex
	rt         *goja.Runtime
	newErr     goja.Callable
	newMap     goja.Callable
	mapEntries goja.Callable
}

// helper compiles a constant JavaScript function.
func (r *Runtime) helper(src string) goja.Callable {
	v, err := r.rt.RunString(src)
	if err != nil {
		panic(err)
	}
	fn, _ := goja.AssertFunction(v)
	return fn
}

// Load evaluates the bundle in the runtime.
func (r *Runtime) Load(ctx context.Context, b *gostrict.Bundle, host gostrict.Funcs) (gostrict.Instance, error) {
	prog, err := r.engine.program(b)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	in := &instance{rt: r, bundle: b, exports: map[string]goja.Callable{}, externs: map[string]goja.Callable{}, records: map[string]goja.Callable{}}
	v, err := r.run(ctx, func() (goja.Value, error) { return r.rt.RunProgram(prog) })
	if err != nil {
		return nil, err
	}
	fn, ok := goja.AssertFunction(v)
	if !ok {
		return nil, errors.New("bundle does not evaluate to a function")
	}

	h := r.rt.NewObject()
	for _, x := range b.Externs {
		if f, ok := host[x.Name]; ok {
			if err := h.Set(x.Name, r.hostFunc(x.Func, f)); err != nil {
				return nil, err
			}
		}
	}

	exports, err := r.run(ctx, func() (goja.Value, error) { return fn(goja.Undefined(), h) })
	if err != nil {
		return nil, err
	}
	obj := exports.ToObject(r.rt)
	for name := range b.Exports {
		in.exports[name], _ = goja.AssertFunction(obj.Get(name))
	}
	externs := obj.Get("$externs").ToObject(r.rt)
	for _, x := range b.Externs {
		in.externs[x.Name], _ = goja.AssertFunction(externs.Get(x.Name))
	}
	news := obj.Get("$new").ToObject(r.rt)
	for name := range b.Records {
		in.records[name], _ = goja.AssertFunction(news.Get(name))
	}
	in.variants = map[string]map[string]variant{}
	unions := obj.Get("$unions").ToObject(r.rt)
	for name, u := range b.Unions {
		vs := unions.Get(name).ToObject(r.rt)
		in.variants[name] = map[string]variant{}
		for _, tag := range u.Variants {
			pair := vs.Get(tag).ToObject(r.rt)
			mk, _ := goja.AssertFunction(pair.Get("1"))
			in.variants[name][tag] = variant{methods: pair.Get("0"), make: mk}
		}
	}
	return in, nil
}

// nativeRef wraps a handle, which a native host function created, while it is in JavaScript.
type nativeRef struct {
	ref js.Ref
}

type instance struct {
	rt      *Runtime
	bundle  *gostrict.Bundle
	exports map[string]goja.Callable
	externs map[string]goja.Callable
	// records are the constructors of the records
	records map[string]goja.Callable
	// variants are the method tables and constructors of the variants of the unions
	variants map[string]map[string]variant
}

type variant struct {
	methods goja.Value
	make    goja.Callable
}

// run interrupts the runtime, when the context ends.
func (r *Runtime) run(ctx context.Context, f func() (goja.Value, error)) (goja.Value, error) {
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			r.rt.Interrupt(ctx.Err())
		case <-done:
		}
	}()

	v, err := f()
	r.rt.ClearInterrupt()
	if err != nil {
		var ex *goja.Exception
		if errors.As(err, &ex) {
			return nil, &Error{Msg: ex.Value().String()}
		}
		var ie *goja.InterruptedError
		if errors.As(err, &ie) {
			return nil, ctx.Err()
		}
		return nil, err
	}
	return v, nil
}

func (in *instance) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	sig, ok := in.bundle.Exports[name]
	if !ok {
		return nil, fmt.Errorf("%s is not an exported function", name)
	}
	if err := gostrict.CheckArgs(name, sig.Params, args); err != nil {
		return nil, err
	}

	in.rt.mu.Lock()
	defer in.rt.mu.Unlock()
	return in.callJS(ctx, in.exports[name], sig, args)
}

// CallExtern calls the function of a header package, as the program calls it: the host function or the binding.
func (in *instance) CallExtern(ctx context.Context, x gostrict.Func, args ...any) ([]any, error) {
	fn, ok := in.externs[x.Name]
	if !ok {
		return nil, fmt.Errorf("%s is not a function of a header package of the bundle", x.Name)
	}

	in.rt.mu.Lock()
	defer in.rt.mu.Unlock()
	return in.callJS(ctx, fn, x, args)
}

func (in *instance) callJS(ctx context.Context, fn goja.Callable, sig gostrict.Func, args []any) ([]any, error) {
	c := converter{r: in.rt, in: in}
	jsArgs := make([]goja.Value, len(args))
	for i, a := range args {
		v, err := c.toJS(a, sig.Params[i])
		if err != nil {
			return nil, fmt.Errorf("value %d of %s: %w", i, sig.Name, err)
		}
		jsArgs[i] = v
	}

	v, err := in.rt.run(ctx, func() (goja.Value, error) { return fn(goja.Undefined(), jsArgs...) })
	if err != nil {
		return nil, err
	}

	values := []goja.Value{v}
	if len(sig.Results) > 1 {
		arr := v.ToObject(in.rt.rt)
		values = values[:0]
		for i := range sig.Results {
			values = append(values, arr.Get(strconv.Itoa(i)))
		}
	}

	res := make([]any, len(sig.Results))
	for i, t := range sig.Results {
		r, err := c.toGo(values[i], t)
		if err != nil {
			return nil, err
		}
		res[i] = r
	}
	return res, nil
}

// converter converts between the neutral values of Go and JavaScript. The records of an instance need its
// constructors, a host function has none.
type converter struct {
	r  *Runtime
	in *instance
}

func (c converter) toGo(v goja.Value, t gostrict.Type) (any, error) {
	rt := c.r.rt
	switch t.Kind {
	case gostrict.ValueBool:
		return v.ToBoolean(), nil
	case gostrict.ValueInt32:
		return int32(v.ToInteger()), nil
	case gostrict.ValueUint32:
		return uint32(v.ToInteger()), nil
	case gostrict.ValueFloat64:
		return v.ToFloat(), nil
	case gostrict.ValueString:
		return v.String(), nil
	case gostrict.ValueHandle:
		if obj, ok := v.(*goja.Object); ok {
			if n, ok := obj.Export().(nativeRef); ok {
				return n.ref, nil
			}
		}
		return js.NewRef(v), nil
	case gostrict.ValueState:
		return gostrict.NewToken(t.Name, v), nil
	case gostrict.ValueMap:
		entries, err := c.r.mapEntries(goja.Undefined(), v)
		if err != nil {
			return nil, err
		}
		obj := entries.ToObject(rt)
		n := int(obj.Get("length").ToInteger())
		res := make([]any, n)
		for i := range n {
			kv := obj.Get(strconv.Itoa(i)).ToObject(rt)
			k, err := c.toGo(kv.Get("0"), *t.Key)
			if err != nil {
				return nil, err
			}
			val, err := c.toGo(kv.Get("1"), *t.Elem)
			if err != nil {
				return nil, err
			}
			res[i] = []any{k, val}
		}
		return res, nil
	case gostrict.ValueList, gostrict.ValueOut:
		obj := v.ToObject(rt)
		n := int(obj.Get("length").ToInteger())
		res := make([]any, n)
		for i := range n {
			e, err := c.toGo(obj.Get(strconv.Itoa(i)), *t.Elem)
			if err != nil {
				return nil, err
			}
			res[i] = e
		}
		return res, nil
	case gostrict.ValueUnion:
		if c.in == nil {
			return nil, fmt.Errorf("unions do not cross the boundary of a host function: %s", t)
		}
		box := v.ToObject(rt)
		methods := box.Get("m")
		for _, tag := range c.in.bundle.Unions[t.Name].Variants {
			if c.in.variants[t.Name][tag].methods.SameAs(methods) {
				fields, err := c.toGo(box.Get("v"), gostrict.Type{Kind: gostrict.ValueRecord, Name: tag})
				if err != nil {
					return nil, err
				}
				return gostrict.Variant{Tag: tag, Fields: fields.([]any)}, nil
			}
		}
		return nil, fmt.Errorf("the value is no variant of %s", t.Name)
	case gostrict.ValueRecord:
		rec, err := c.record(t)
		if err != nil {
			return nil, err
		}
		obj := v.ToObject(rt)
		res := make([]any, len(rec.Fields))
		for i, f := range rec.Fields {
			fv, err := c.toGo(obj.Get(f.JS), f.Type)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", f.Name, err)
			}
			res[i] = fv
		}
		return res, nil
	case gostrict.ValueError:
		if goja.IsNull(v) || goja.IsUndefined(v) {
			return nil, nil
		}
		box := v.ToObject(rt)
		methods := box.Get("m").ToObject(rt)
		errorFn, ok := goja.AssertFunction(methods.Get("Error"))
		if !ok {
			return nil, errors.New("value is not an error")
		}
		msg, err := errorFn(goja.Undefined(), box.Get("v"))
		if err != nil {
			return nil, err
		}
		return errors.New(msg.String()), nil
	}
	return nil, fmt.Errorf("unsupported type %s", t)
}

func (c converter) record(t gostrict.Type) (gostrict.Record, error) {
	if c.in == nil {
		return gostrict.Record{}, fmt.Errorf("records do not cross the boundary of a host function: %s", t)
	}
	rec, ok := c.in.bundle.Records[t.Name]
	if !ok {
		return rec, fmt.Errorf("unknown record %s", t.Name)
	}
	return rec, nil
}

func (c converter) toJS(v any, t gostrict.Type) (goja.Value, error) {
	rt := c.r.rt
	switch t.Kind {
	case gostrict.ValueHandle:
		ref := v.(js.Ref)
		if gv, ok := ref.Value().(goja.Value); ok {
			return gv, nil
		}
		// a handle, which the host created natively, returns to the host as the same reference
		return rt.ToValue(nativeRef{ref: ref}), nil
	case gostrict.ValueError:
		if v == nil {
			return goja.Null(), nil
		}
		return c.r.newErr(goja.Undefined(), rt.ToValue(v.(error).Error()))
	case gostrict.ValueState:
		tok := v.(gostrict.Token)
		gv, ok := tok.Value().(goja.Value)
		if !ok {
			return nil, fmt.Errorf("the state of %s belongs to another engine", tok.Type())
		}
		return gv, nil
	case gostrict.ValueMap:
		entries := v.([]any)
		pairs := make([]any, len(entries))
		for i, e := range entries {
			kv := e.([]any)
			k, err := c.toJS(kv[0], *t.Key)
			if err != nil {
				return nil, err
			}
			val, err := c.toJS(kv[1], *t.Elem)
			if err != nil {
				return nil, err
			}
			pairs[i] = rt.NewArray(k, val)
		}
		return c.r.newMap(goja.Undefined(), rt.NewArray(pairs...))
	case gostrict.ValueList, gostrict.ValueOut:
		list := v.([]any)
		values := make([]any, len(list))
		for i, e := range list {
			ev, err := c.toJS(e, *t.Elem)
			if err != nil {
				return nil, err
			}
			values[i] = ev
		}
		return rt.NewArray(values...), nil
	case gostrict.ValueUnion:
		if c.in == nil {
			return nil, fmt.Errorf("unions do not cross the boundary of a host function: %s", t)
		}
		x := v.(gostrict.Variant)
		vr, ok := c.in.variants[t.Name][x.Tag]
		if !ok {
			return nil, fmt.Errorf("%s is no variant of %s", x.Tag, t.Name)
		}
		rec, err := c.toJS(x.Fields, gostrict.Type{Kind: gostrict.ValueRecord, Name: x.Tag})
		if err != nil {
			return nil, err
		}
		// the constructor takes the fields, which the record has just been built of
		obj := rec.ToObject(rt)
		var args []goja.Value
		for _, f := range c.in.bundle.Records[x.Tag].Fields {
			args = append(args, obj.Get(f.JS))
		}
		return vr.make(goja.Undefined(), args...)
	case gostrict.ValueRecord:
		rec, err := c.record(t)
		if err != nil {
			return nil, err
		}
		fields := v.([]any)
		if len(fields) != len(rec.Fields) {
			return nil, fmt.Errorf("%s has %d fields, got %d values", rec.Name, len(rec.Fields), len(fields))
		}
		args := make([]goja.Value, len(fields))
		for i, f := range rec.Fields {
			if err := gostrict.CheckArgs(rec.Name+"."+f.Name, []gostrict.Type{f.Type}, fields[i:i+1]); err != nil {
				return nil, err
			}
			fv, err := c.toJS(fields[i], f.Type)
			if err != nil {
				return nil, err
			}
			args[i] = fv
		}
		return c.in.records[t.Name](goja.Undefined(), args...)
	}
	return rt.ToValue(v), nil
}

// hostFunc adapts a Go function to the calling convention of the bundle: several results are an array.
func (r *Runtime) hostFunc(sig gostrict.Func, f gostrict.Fn) func(goja.FunctionCall) goja.Value {
	c := converter{r: r}
	return func(call goja.FunctionCall) goja.Value {
		args := make([]any, len(sig.Params))
		for i, k := range sig.Params {
			v, err := c.toGo(call.Argument(i), k)
			if err != nil {
				panic(r.rt.NewGoError(err))
			}
			args[i] = v
		}

		res := f(args)
		if err := gostrict.CheckArgs(sig.Name, sig.Results, res); err != nil {
			panic(r.rt.NewGoError(err))
		}

		values := make([]goja.Value, len(res))
		for i, k := range sig.Results {
			v, err := c.toJS(res[i], k)
			if err != nil {
				panic(r.rt.NewGoError(err))
			}
			values[i] = v
		}
		switch len(values) {
		case 0:
			return goja.Undefined()
		case 1:
			return values[0]
		}
		return r.rt.NewArray(toAny(values)...)
	}
}

func toAny(values []goja.Value) []any {
	res := make([]any, len(values))
	for i, v := range values {
		res[i] = v
	}
	return res
}

// Close releases the instance. The runtime is released, when no instance refers to it anymore.
func (in *instance) Close() error {
	return nil
}
