// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"context"
	"go.wdy.de/nago/pkg/std/concurrent"
	"reflect"
)

type syskey string

type myCtx struct {
	byName *concurrent.RWMap[string, any]
	byType *concurrent.RWMap[reflect.Type, any]
}

const (
	keyMyCtx syskey = "m"
)

// FromContext is a simple kind of dynamic inversion of control container,
// but has a simplified implementation strategy to
// improve lookup performance to O(1) instead of O(n).
// Why is this useful at all? It helps to create especially system components which react
// dynamically to whatever has been provided. Usually, the success of the component can degrade partially
// and may still work without any problems but just omits some features.
//
// # Should I use it?
//
// In general, you should avoid using this method and instead pass your funcs or use cases explicitly.
// When in doubt, choose the explicit way.
//
// # Implementation details
//
// If name is not empty, an exact lookup by name is performed.
// If that is not found, the zero value of T and false is returned.
//
// If name is empty, the type lookup table is used and the exact type is matched. There is no polymorphism involved.
// If no exact match is found, false and the zero value of T is returned.
//
// In any case, whatever has been registered the last time for any name or type has the highest precedence.
func FromContext[T any](ctx context.Context, name string) (T, bool) {
	mCtx, ok := ctx.Value(keyMyCtx).(myCtx)
	if !ok {
		var zero T
		return zero, false
	}

	if name != "" {
		v, ok := mCtx.byName.Get(name)
		if !ok {
			var zero T
			return zero, false
		}

		if v, ok := v.(T); ok {
			return v, true
		}
	}

	v, ok := mCtx.byType.Get(reflect.TypeFor[T]())
	if !ok {
		var zero T
		return zero, false
	}

	if v, ok := v.(T); ok {
		return v, true
	}

	var zero T
	return zero, false
}

func ContextValue[T any](name string, v T) CtxOption {
	return ctxValue[T]{
		Name:  name,
		Value: v,
	}
}

type ctxValue[T any] struct {
	Name  string
	Value T
}

func (c ctxValue[T]) configureByName(m *concurrent.RWMap[string, any]) {
	if c.Name == "" {
		return
	}

	m.Put(c.Name, c.Value)
}

func (c ctxValue[T]) configureByType(m *concurrent.RWMap[reflect.Type, any]) {
	m.Put(reflect.TypeFor[T](), c.Value)
}

type CtxOption interface {
	configureByName(*concurrent.RWMap[string, any])
	configureByType(*concurrent.RWMap[reflect.Type, any])
}

// WithContext constructs a new context with all values and names from the given context and applies
// the options on it. Using this approach, you can inject at any time additional services or replace
// existing services to pass them into dynamic ui components, e.g. for generating automatic forms.
//
// See also [Application.Context] and [FromContext].
func WithContext(ctx context.Context, cfg ...CtxOption) context.Context {
	mCtx, ok := ctx.Value(keyMyCtx).(myCtx)
	if !ok {
		mCtx = myCtx{
			byName: &concurrent.RWMap[string, any]{},
			byType: &concurrent.RWMap[reflect.Type, any]{},
		}
	} else {
		tmp := myCtx{
			byName: mCtx.byName.Clone(),
			byType: mCtx.byType.Clone(),
		}

		mCtx = tmp
	}

	for _, option := range cfg {
		option.configureByName(mCtx.byName)
		option.configureByType(mCtx.byType)
	}

	return context.WithValue(ctx, keyMyCtx, mCtx)
}
