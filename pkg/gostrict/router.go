// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// Envelope is a message of a program, which a backend receives, independent of its transport: a [Message] of an
// engine in Go or a message of a program in a browser.
type Envelope interface {
	// MessageType returns the full name of the record, e.g. "example.com/arena/messages.Hit".
	MessageType() string
	// Decode stores the message in the value, to which ptr points: the record or a sealed interface, whose variant
	// it is.
	Decode(ptr any) error
}

// MessageType returns the full name of the record.
func (m Message) MessageType() string {
	return m.Type
}

// Decode stores the record in the value, to which ptr points, which may be a sealed interface with the variant.
func (m Message) Decode(ptr any) error {
	t := reflect.TypeOf(ptr)
	if t != nil && t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Interface {
		return FromNeutral(Variant{Tag: m.Type, Fields: m.Fields}, ptr)
	}
	return FromNeutral(m.Fields, ptr)
}

// MessageOf returns the message of a record or of the variant of a sealed interface, e.g. for
// [Session.Deliver].
func MessageOf(v any) (Message, error) {
	n, err := ToNeutral(v)
	if err != nil {
		return Message{}, err
	}
	switch n := n.(type) {
	case Variant:
		return Message{Type: n.Tag, Fields: n.Fields}, nil
	case []any:
		t := reflect.TypeOf(v)
		if t.Kind() != reflect.Struct {
			return Message{}, fmt.Errorf("%s is no record", t)
		}
		return Message{Type: TypeName(t), Fields: n}, nil
	}
	return Message{}, fmt.Errorf("%T is no record", v)
}

// Router passes the messages of a program to the typed functions of its backend, see [Handle].
type Router struct {
	mu     sync.RWMutex
	routes map[string]func(ctx context.Context, e Envelope) error
}

// Handle registers the function for the messages of type T, which is a record or a sealed interface, whose
// variants are registered with [Variants]. It returns the router.
func Handle[T any](r *Router, fn func(ctx context.Context, m T) error) *Router {
	route := func(ctx context.Context, e Envelope) error {
		var m T
		if err := e.Decode(&m); err != nil {
			return err
		}
		return fn(ctx, m)
	}

	t := reflect.TypeFor[T]()
	var names []string
	switch t.Kind() {
	case reflect.Struct:
		names = []string{TypeName(t)}
	case reflect.Interface:
		vs, ok := variants.Load(t)
		if !ok {
			panic(fmt.Errorf("register the variants of %s with gostrict.Variants", t))
		}
		for name := range vs.(map[string]reflect.Type) {
			names = append(names, name)
		}
	default:
		panic(fmt.Errorf("%s is no record and no sealed interface", t))
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.routes == nil {
		r.routes = map[string]func(ctx context.Context, e Envelope) error{}
	}
	for _, name := range names {
		r.routes[name] = route
	}
	return r
}

// Dispatch passes the message to its function.
func (r *Router) Dispatch(ctx context.Context, e Envelope) error {
	r.mu.RLock()
	route, ok := r.routes[e.MessageType()]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("the backend does not handle the message %s", e.MessageType())
	}
	return route(ctx, e)
}
