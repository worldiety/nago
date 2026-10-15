// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package native executes the Go code of a package, which a bundle was translated from. A program of the subset
// is valid Go, so the native code is the reference, against which the other engines are tested, and it runs on a
// server without any JavaScript.
package native

import (
	"context"
	"fmt"
	"reflect"

	"go.wdy.de/nago/pkg/gostrict"
)

// Engine calls the exported functions of the main package natively. The functions of header packages are native
// anyway, so the host is ignored.
type Engine struct {
	// Exports are the exported functions of the main package, usually its generated Exports function.
	Exports gostrict.Funcs
}

var _ gostrict.Engine = Engine{}

func (e Engine) Load(ctx context.Context, b *gostrict.Bundle, host gostrict.Funcs) (gostrict.Instance, error) {
	for name := range b.Exports {
		if _, ok := e.Exports[name]; !ok {
			return nil, fmt.Errorf("the native code lacks the exported function %s", name)
		}
	}
	return instance{engine: e, bundle: b}, nil
}

type instance struct {
	engine Engine
	bundle *gostrict.Bundle
}

func (in instance) Call(ctx context.Context, name string, args ...any) ([]any, error) {
	sig, ok := in.bundle.Exports[name]
	if !ok {
		return nil, fmt.Errorf("%s is not an exported function", name)
	}
	if err := gostrict.CheckArgs(name, sig.Params, args); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return in.engine.Exports[name](args), nil
}

func (instance) Close() error {
	return nil
}

// Reflect adapts Go functions by reflection, e.g. in a test, where no generated functions exist.
func Reflect(funcs map[string]any) gostrict.Funcs {
	res := gostrict.Funcs{}
	for name, f := range funcs {
		fn := reflect.ValueOf(f)
		res[name] = func(args []any) []any {
			in := make([]reflect.Value, len(args))
			for i, a := range args {
				in[i] = reflect.ValueOf(a)
			}
			var out []any
			for _, v := range fn.Call(in) {
				out = append(out, v.Interface())
			}
			return out
		}
	}
	return res
}
